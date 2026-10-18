package wranglerasynq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"
)

const emailBodyPdfFilenamePrefix = "email-body"
const emailBodyPdfSubjectSlugMax = 60

type EmailProcessTaskPayload struct {
	GroupSettingsId uint
	ImageForOcrPath string
	TempFilePath    string
	Metadata        structs.EmailMetadata
	Attachment      structs.Attachment
	RenderBodyPdf   bool
}

func HandleEmailProcessTask(context context.Context, task *asynq.Task) error {
	var payload EmailProcessTaskPayload

	taskId, err := GetTaskIdFromContext(context)
	if err != nil {
		return HandleError(err)
	}

	err = json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return HandleError(err)
	}

	return handleEmailProcessPayload(taskId, payload)
}

// handleEmailProcessPayload contains the core email ingest logic. It is a
// separate function so that tests in this package can call it directly without
// needing a real asynq context (whose task-id key is unexported).
func handleEmailProcessPayload(taskId string, payload EmailProcessTaskPayload) error {
	db := repositories.GetDB()
	systemTaskService := services.NewSystemTaskService(nil)
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	systemTaskRepository := repositories.NewSystemTaskRepository(nil)

	hasAttachmentImage := len(payload.ImageForOcrPath) > 0

	var fileBytes []byte
	if len(payload.TempFilePath) > 0 {
		// os.ReadFile, not utils.ReadFile: the latter returns (nil, nil) on any
		// read error. A missing attachment would then sail past this check as
		// empty bytes, and because hasAttachmentImage is derived from the payload
		// string rather than the file, processing would continue and persist a
		// zero-byte receipt image carrying the real name and size.
		var readErr error
		fileBytes, readErr = os.ReadFile(payload.TempFilePath)
		if readErr != nil {
			return HandleError(readErr)
		}
	}

	groupSettingsIdString := utils.UintToString(payload.GroupSettingsId)
	groupSettingsToUse, err := groupSettingsRepository.GetGroupSettingsById(groupSettingsIdString)
	if err != nil {
		return HandleError(err)
	}

	if groupSettingsToUse.ID == 0 {
		return HandleError(fmt.Errorf("could not find group settings with id %d", payload.GroupSettingsId))
	}

	groupIdString := utils.UintToString(groupSettingsToUse.GroupId)

	bodyPdfBytes, bodyImagePath, cleanupBodyImage, htmlPdfTaskCmd, renderErr := renderBodyPdfIfRequested(payload)
	if cleanupBodyImage != nil {
		defer cleanupBodyImage()
	}

	// Persist HTML_TO_PDF orphaned (without parent) when render fails, so
	// the failure is still visible in the system-task UI before we
	// short-circuit. asynq retries handle re-renders.
	if renderErr != nil {
		persistHtmlToPdfSystemTask(systemTaskRepository, htmlPdfTaskCmd, groupSettingsToUse, taskId, nil)
		return HandleError(renderErr)
	}

	hasBodyImage := len(bodyImagePath) > 0
	hasAnyImage := hasAttachmentImage || hasBodyImage

	var baseCommand commands.UpsertReceiptCommand
	var processingMetadata commands.ReceiptProcessingMetadata
	var processingErr error

	start := time.Now()
	if hasAnyImage {
		imagePaths := []string{}
		if hasAttachmentImage {
			imagePaths = append(imagePaths, payload.ImageForOcrPath)
		}
		if hasBodyImage {
			imagePaths = append(imagePaths, bodyImagePath)
		}
		baseCommand, processingMetadata, processingErr = services.ReadReceiptImagesWithEmailBody(
			imagePaths,
			payload.Metadata.Body,
			hasBodyImage,
			groupIdString,
		)
	} else {
		baseCommand, processingMetadata, processingErr = services.ReadReceiptFromTextOnly(payload.Metadata.Body, groupIdString)
	}
	end := time.Now()

	metadataBytes, err := json.Marshal(payload.Metadata)
	if err != nil {
		return HandleError(err)
	}

	status := models.SYSTEM_TASK_SUCCEEDED
	if processingErr != nil {
		status = models.SYSTEM_TASK_FAILED
	}

	resultDescription := string(metadataBytes)
	if processingErr != nil {
		resultDescription = processingErr.Error()
	}

	emailReadSystemTask, err := systemTaskRepository.CreateSystemTask(
		commands.UpsertSystemTaskCommand{
			Type:                 models.EMAIL_READ,
			Status:               status,
			AssociatedEntityType: models.SYSTEM_EMAIL,
			AssociatedEntityId:   groupSettingsToUse.SystemEmail.ID,
			StartedAt:            start,
			EndedAt:              &end,
			RanByUserId:          nil,
			ResultDescription:    resultDescription,
			AsynqTaskId:          taskId,
		},
	)
	if err != nil {
		return HandleError(err)
	}

	// Persist HTML_TO_PDF chained under EMAIL_READ so the render appears
	// nested alongside the OCR / prompt / completion tasks for this email
	// in the system-task UI.
	persistHtmlToPdfSystemTask(systemTaskRepository, htmlPdfTaskCmd, groupSettingsToUse, taskId, &emailReadSystemTask.ID)

	processingSystemTasks, err := systemTaskService.CreateSystemTasksFromMetadata(
		processingMetadata,
		start,
		end,
		models.EMAIL_UPLOAD,
		nil,
		&groupSettingsToUse.GroupId,
		taskId,
		func(command commands.UpsertSystemTaskCommand) *uint {
			return &emailReadSystemTask.ID
		},
	)
	if err != nil {
		return HandleError(err)
	}

	if processingErr != nil {
		return HandleError(processingErr)
	}

	command := baseCommand
	command.GroupId = groupSettingsToUse.GroupId

	if len(command.Status) == 0 {
		command.Status = groupSettingsToUse.EmailDefaultReceiptStatus
	}

	if command.PaidByUserID == 0 {
		command.PaidByUserID = *groupSettingsToUse.EmailDefaultReceiptPaidById
	}

	command.CreatedByString = "Email Integration"

	// Attach the group's default custom fields (as empty values) when the group opted into applying
	// them to server-created receipts. Same helper as quick scan so the two ingest paths cannot drift.
	err = services.ApplyGroupDefaultCustomFields(nil, command.GroupId, &command)
	if err != nil {
		return HandleError(err)
	}

	// Resolve id-only categories/tags from the AI response. The default prompt instructs the
	// model to return { Id: N } with no name, which Validate() would reject. This mirrors the
	// equivalent step in QuickScan (resolveQuickScanCategories/Tags). Ids that don't resolve
	// (hallucinated or deleted) are dropped rather than failing the whole ingest.
	receiptService := services.NewReceiptService(nil)
	command.Categories, err = receiptService.ResolveIngestCategories(command.Categories, command.GroupId)
	if err != nil {
		return HandleError(err)
	}
	command.Tags, err = receiptService.ResolveIngestTags(command.Tags, command.GroupId)
	if err != nil {
		return HandleError(err)
	}

	vErr := command.Validate(0, true)
	if len(vErr.Errors) > 0 {
		errBytes, _ := json.Marshal(vErr.Errors)
		return HandleError(fmt.Errorf("receipt validation failed: %s", string(errBytes)))
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository := repositories.NewReceiptRepository(tx)
		receiptImageRepository := repositories.NewReceiptImageRepository(tx)
		systemTaskService.SetTransaction(tx)

		createdReceipt, err := receiptRepository.CreateReceipt(command, 0, false)
		_, taskErr := systemTaskService.CreateReceiptUploadedSystemTask(
			err,
			createdReceipt,
			processingSystemTasks,
			time.Now(),
		)
		if taskErr != nil {
			return HandleError(taskErr)
		}
		if err != nil {
			tx.Commit()
			return HandleError(taskErr)
		}

		err = systemTaskService.AssociateProcessingSystemTasksToReceipt(processingSystemTasks, createdReceipt.ID)
		if err != nil {
			return HandleError(err)
		}

		if hasAttachmentImage {
			fileData := models.FileData{
				ReceiptId: createdReceipt.ID,
				Name:      payload.Attachment.Filename,
				FileType:  payload.Attachment.FileType,
				Size:      payload.Attachment.Size,
			}

			_, err = receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
			if err != nil {
				return HandleError(err)
			}
		}

		if len(bodyPdfBytes) > 0 {
			bodyFileData := models.FileData{
				ReceiptId: createdReceipt.ID,
				Name:      buildEmailBodyPdfFilename(payload.Metadata),
				FileType:  constants.ApplicationPdf,
				Size:      uint(len(bodyPdfBytes)),
			}

			_, err = receiptImageRepository.CreateReceiptImage(bodyFileData, bodyPdfBytes)
			if err != nil {
				return HandleError(err)
			}
		}

		return nil
	})

	return err
}

// persistHtmlToPdfSystemTask writes an HTML_TO_PDF system task (if one was
// produced — i.e. render was attempted, indicated by a non-empty Type).
// parentSystemTaskId, when non-nil, chains the task under that parent
// (typically the EMAIL_READ task) so it appears nested in the UI.
// Persistence failures are logged, not propagated, so they don't mask the
// underlying render result.
func persistHtmlToPdfSystemTask(
	systemTaskRepository repositories.SystemTaskRepository,
	cmd commands.UpsertSystemTaskCommand,
	groupSettingsToUse models.GroupSettings,
	taskId string,
	parentSystemTaskId *uint,
) {
	if cmd.Type == "" {
		return
	}
	cmd.AssociatedEntityType = models.SYSTEM_EMAIL
	cmd.AssociatedEntityId = groupSettingsToUse.SystemEmail.ID
	cmd.AsynqTaskId = taskId
	cmd.AssociatedSystemTaskId = parentSystemTaskId
	if _, err := systemTaskRepository.CreateSystemTask(cmd); err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "failed to persist HTML_TO_PDF system task: ", err.Error())
	}
}

// buildEmailBodyPdfFilename produces a per-email display name for the body
// PDF FileData saved on a receipt. Result shape is
// "email-body-{subject-slug}-{YYYY-MM-DD}.pdf"; pieces are omitted when the
// corresponding metadata is missing. Subject is slugified to ASCII so the
// resulting string is safe to use as a filename.
func buildEmailBodyPdfFilename(metadata structs.EmailMetadata) string {
	parts := []string{emailBodyPdfFilenamePrefix}
	if slug := slugifySubject(metadata.Subject); slug != "" {
		parts = append(parts, slug)
	}
	if !metadata.Date.IsZero() {
		parts = append(parts, metadata.Date.UTC().Format("2006-01-02"))
	}
	return strings.Join(parts, "-") + ".pdf"
}

// slugifySubject converts a free-form email subject into a lowercase
// ASCII slug suitable for use inside a filename. Non-alphanumerics collapse
// to single hyphens; leading/trailing hyphens are trimmed; the result is
// truncated to emailBodyPdfSubjectSlugMax characters.
func slugifySubject(subject string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range strings.ToLower(subject) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	slug := strings.TrimRight(b.String(), "-")
	if len(slug) > emailBodyPdfSubjectSlugMax {
		slug = strings.TrimRight(slug[:emailBodyPdfSubjectSlugMax], "-")
	}
	return slug
}

// renderBodyPdfIfRequested converts the email's HTML body to a PDF (when
// requested by the payload) and writes a JPEG version to a per-task temp
// file for OCR/vision processing. Returns the PDF bytes, the JPEG path, a
// cleanup function that callers must invoke (typically via defer), and the
// HTML_TO_PDF system task command (un-persisted) so the caller can chain
// it under the EMAIL_READ system task once that exists. When no render is
// requested, the returned command is the zero value (its Type is empty).
func renderBodyPdfIfRequested(
	payload EmailProcessTaskPayload,
) ([]byte, string, func(), commands.UpsertSystemTaskCommand, error) {
	if !payload.RenderBodyPdf || len(payload.Metadata.BodyHtml) == 0 {
		return nil, "", nil, commands.UpsertSystemTaskCommand{}, nil
	}

	htmlToPdfService := services.NewHtmlToPdfService(nil)
	pdfBytes, htmlPdfTaskCmd, renderErr := htmlToPdfService.Render(payload.Metadata.BodyHtml)
	if renderErr != nil {
		return nil, "", nil, htmlPdfTaskCmd, renderErr
	}

	fileRepository := repositories.NewFileRepository(nil)
	imageBytes, err := fileRepository.GetBytesFromImageBytes(pdfBytes)
	if err != nil {
		return nil, "", nil, htmlPdfTaskCmd, err
	}

	randId, err := utils.GetRandomString(8)
	if err != nil {
		return nil, "", nil, htmlPdfTaskCmd, err
	}
	bodyImagePath := filepath.Join(fileRepository.GetTempDirectoryPath(), "image-body-"+randId+".jpg")
	cleanup := func() {
		if !utils.FileExists(bodyImagePath) {
			return
		}
		if err := os.Remove(bodyImagePath); err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "failed to remove body image temp file: ", err.Error())
		}
	}
	// Set up cleanup before the write so a partial WriteFile failure also
	// removes the orphaned bytes.
	if err := utils.WriteFile(bodyImagePath, imageBytes); err != nil {
		cleanup()
		return nil, "", nil, htmlPdfTaskCmd, err
	}

	return pdfBytes, bodyImagePath, cleanup, htmlPdfTaskCmd, nil
}
