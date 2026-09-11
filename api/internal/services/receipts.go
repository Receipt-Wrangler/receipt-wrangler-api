package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strconv"
	"strings"
	"time"
)

// Sentinel errors for the shared, enforced read operations. They let each ingress
// point (REST handlers, MCP tools) map a single enforcement outcome to its own
// transport without re-implementing the checks.
var (
	// ErrReceiptAccessDenied is returned when a receipt is missing, the caller
	// lacks group.receipts.read, or the receipt is hidden by paid-by visibility.
	// It is intentionally indistinct so callers don't leak a receipt's existence.
	ErrReceiptAccessDenied = errors.New("receipt access denied")
	// ErrSearchForbidden is returned when the caller lacks app.receipts.search.
	ErrSearchForbidden = errors.New("not authorized to search receipts")
)

type ReceiptService struct {
	BaseService
}

func NewReceiptService(tx *gorm.DB) ReceiptService {
	service := ReceiptService{BaseService: BaseService{
		DB: repositories.GetDB(),
		TX: tx,
	}}
	return service
}

// GetReceiptForUser is the single, shared "read one receipt" operation used by
// both the REST handler and the MCP tool. It fetches the receipt and applies the
// full read enforcement chain — group.receipts.read, paid-by visibility, and
// category/tag grant stripping — so the two ingress points cannot drift. Missing,
// forbidden, and paid-by-hidden receipts all collapse to ErrReceiptAccessDenied so
// the caller cannot infer a receipt's existence.
func (service ReceiptService) GetReceiptForUser(userId uint, receiptId string) (models.Receipt, error) {
	receiptRepository := repositories.NewReceiptRepository(service.TX)
	permissionService := NewPermissionService(service.TX)

	// Authorize on the lightweight auth fields first (this fetch uses First, so a
	// missing row surfaces as ErrRecordNotFound). Only load the full receipt with
	// its associations once the read is allowed.
	authReceipt, err := receiptRepository.GetReceiptForAuthorization(receiptId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Receipt{}, ErrReceiptAccessDenied
		}
		return models.Receipt{}, err
	}

	hasAccess, err := permissionService.HasGroupPermissions(userId, authReceipt.GroupId, permissions.GroupReceiptsRead)
	if err != nil {
		return models.Receipt{}, err
	}
	if !hasAccess {
		return models.Receipt{}, ErrReceiptAccessDenied
	}

	visible, err := permissionService.ReceiptPaidByVisible(userId, authReceipt.GroupId, authReceipt.PaidByUserID)
	if err != nil {
		return models.Receipt{}, err
	}
	if !visible {
		return models.Receipt{}, ErrReceiptAccessDenied
	}

	receipt, err := receiptRepository.GetFullyLoadedReceiptById(receiptId)
	if err != nil {
		return models.Receipt{}, err
	}

	// Guard the window between the authorization read and this full load: if the
	// receipt's identity, group, or payer changed (or it was deleted — Find leaves
	// a zero-value row), the prior authorization no longer applies, so deny rather
	// than return data that was never authorized.
	if receipt.ID != authReceipt.ID ||
		receipt.GroupId != authReceipt.GroupId ||
		receipt.PaidByUserID != authReceipt.PaidByUserID {
		return models.Receipt{}, ErrReceiptAccessDenied
	}

	if err := permissionService.FilterReceiptCategoriesTagsForReceipt(userId, &receipt); err != nil {
		return models.Receipt{}, err
	}

	// Mask user references (created-by, charged-to) and drop non-visible comment
	// authors outside the caller's member-visible set.
	if err := permissionService.MaskReceiptForMemberVisibility(userId, &receipt); err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

// SearchReceiptsForUser is the single, shared receipt-search operation used by both
// the REST handler and the MCP tool. It enforces app.receipts.search, narrows the
// caller's groups to those granting group.receipts.read, applies paid-by visibility
// in SQL before the limit, and maps to SearchResult. A blank query returns no
// results (matching the REST search bar).
func (service ReceiptService) SearchReceiptsForUser(userId uint, query string, limit int) ([]structs.SearchResult, error) {
	permissionService := NewPermissionService(service.TX)

	hasAccess, err := permissionService.HasAppPermissions(userId, permissions.AppReceiptsSearch)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrSearchForbidden
	}

	results := make([]structs.SearchResult, 0)
	if len(strings.TrimSpace(query)) == 0 {
		return results, nil
	}

	groupMemberRepository := repositories.NewGroupMemberRepository(service.TX)
	groupIds, err := groupMemberRepository.GetGroupIdsByUserId(utils.UintToString(userId))
	if err != nil {
		return nil, err
	}

	// Membership alone is not read access: narrow to the groups actually granting
	// group.receipts.read, so search matches every other receipt read surface. This
	// runs before the query so the paid-by resolution below only walks those groups.
	// Having none is "no results", not a denial — the caller does hold the app-level
	// search permission.
	readableGroupIds, err := permissionService.GroupIdsWithPermission(userId, groupIds, permissions.GroupReceiptsRead)
	if err != nil {
		return nil, err
	}
	if len(readableGroupIds) == 0 {
		return results, nil
	}

	receiptRepository := repositories.NewReceiptRepository(service.TX)
	receipts, err := receiptRepository.SearchReceiptsByGroupIds(readableGroupIds, query, limit, permissionService.PaidByListResolver(userId))
	if err != nil {
		return nil, err
	}

	for _, receipt := range receipts {
		results = append(results, structs.SearchResult{
			ID:            receipt.ID,
			GroupID:       receipt.GroupId,
			Name:          receipt.Name,
			Date:          receipt.Date,
			Type:          "Receipt",
			Amount:        receipt.Amount,
			ReceiptStatus: receipt.Status,
			PaidByUserId:  receipt.PaidByUserID,
			CreatedAt:     receipt.CreatedAt,
		})
	}

	return results, nil
}

func (service ReceiptService) GetReceiptByReceiptImageId(receiptImageId string) (models.Receipt, error) {
	db := service.GetDB()
	var fileData models.FileData

	err := db.Model(models.FileData{}).Where("id = ?", receiptImageId).Select("receipt_id").First(&fileData).Error
	if err != nil {
		return models.Receipt{}, err
	}

	receiptRepository := repositories.NewReceiptRepository(service.TX)
	receipt, err := receiptRepository.GetReceiptById(strconv.FormatUint(uint64(fileData.ReceiptId), 10))
	if err != nil {
		return models.Receipt{}, err
	}

	return receipt, nil
}

func (service ReceiptService) DeleteReceipt(id string) error {
	db := service.GetDB()
	var receipt models.Receipt
	receiptRepository := repositories.NewReceiptRepository(service.TX)

	receipt, err := receiptRepository.GetFullyLoadedReceiptById(id)
	if err != nil {
		return err
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var imagesToDelete []string
		fileRepository := repositories.NewFileRepository(tx)
		fileRepository.SetTransaction(tx)

		for _, f := range receipt.ImageFiles {
			path, _ := fileRepository.BuildFilePath(utils.UintToString(f.ReceiptId), utils.UintToString(f.ID), f.Name)
			imagesToDelete = append(imagesToDelete, path)
		}

		for _, r := range receipt.ReceiptItems {
			err = tx.Model(&r).Association("Categories").Clear()
			if err != nil {
				return err
			}

			err = tx.Model(&r).Association("Tags").Clear()
			if err != nil {
				return err
			}
		}

		err = tx.Model(&receipt).Association("ReceiptItems").Clear()
		if err != nil {
			return err
		}

		err = tx.Select(clause.Associations).Delete(&receipt).Error
		if err != nil {
			return err
		}

		for _, path := range imagesToDelete {
			utils.RemoveDataPath(path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// QuickScanParams carries the inputs for a single quick-scanned file: the caller's claims, the
// per-file values the handler already resolved against the group's quick-scan configuration, and the
// uploaded file's location. Grouped into a struct rather than passed positionally because the
// positional form ends in several interchangeable strings, which is easy to transpose silently.
type QuickScanParams struct {
	Token            *structs.Claims
	PaidByUserId     uint
	GroupId          uint
	Status           models.ReceiptStatus
	CategoryIds      []uint
	TagIds           []uint
	Comment          string
	TempPath         string
	OriginalFileName string
	AsynqTaskId      string
}

func (service ReceiptService) QuickScan(params QuickScanParams) (models.Receipt, error) {
	token := params.Token
	groupId := params.GroupId
	db := repositories.GetDB()
	systemTaskService := NewSystemTaskService(service.TX)
	var createdReceipt models.Receipt

	fileRepository := repositories.NewFileRepository(service.TX)
	fileBytes, err := utils.ReadFile(params.TempPath)
	if err != nil {
		return models.Receipt{}, err
	}

	fileInfo, err := os.Stat(params.TempPath)
	if err != nil {
		return models.Receipt{}, err
	}

	validatedFileType, err := fileRepository.ValidateFileType(fileBytes)
	if err != nil {
		return models.Receipt{}, err
	}

	magicFillCommand := commands.MagicFillCommand{
		ImageData: fileBytes,
		Filename:  params.OriginalFileName,
	}

	receiptRepository := repositories.NewReceiptRepository(service.TX)
	receiptImageRepository := repositories.NewReceiptImageRepository(service.TX)

	groupIdString := utils.UintToString(groupId)

	now := time.Now()
	receiptCommand, receiptProcessingMetadata, magicFillErr := MagicFillFromImage(magicFillCommand, groupIdString, token.UserId)
	finishedAt := time.Now()

	quickScanSystemTasks, taskErr := systemTaskService.CreateSystemTasksFromMetadata(
		receiptProcessingMetadata,
		now,
		finishedAt,
		models.QUICK_SCAN,
		&token.UserId,
		&groupId,
		params.AsynqTaskId, nil)
	if taskErr != nil {
		return models.Receipt{}, taskErr
	}

	if magicFillErr != nil {
		return models.Receipt{}, magicFillErr
	}

	if receiptCommand.PaidByUserID == 0 {
		receiptCommand.PaidByUserID = params.PaidByUserId
	}

	if len(receiptCommand.Status) == 0 {
		receiptCommand.Status = models.ReceiptStatus(params.Status)
	}

	receiptCommand.GroupId = groupId

	// Record a FAILED system task for any error that aborts the quick scan *before* the
	// create-receipt transaction (category/tag resolution or receipt validation). The
	// in-transaction CreateReceipt failure is already covered by CreateReceiptUploadedSystemTask;
	// this closes the gap where a pre-transaction failure (notably a validation error) left the AI
	// processing tasks marked SUCCEEDED and no record of why no receipt was created. Chaining a
	// FAILED child to the quick-scan parent flips that parent to FAILED (see
	// SystemTaskRepository.CreateSystemTask) so the failure surfaces in the activity feed.
	// RanByUserId is deliberately left nil so the child stays hidden and only the flipped parent
	// shows, mirroring the successful RECEIPT_UPLOADED child.
	recordEarlyQuickScanFailure := func(failureErr error) error {
		parentTask := quickScanSystemTasks.SystemTask
		if quickScanSystemTasks.FallbackSystemTask.Status == models.SYSTEM_TASK_SUCCEEDED {
			parentTask = quickScanSystemTasks.FallbackSystemTask
		}
		if parentTask.ID == 0 {
			return failureErr
		}

		parentId := parentTask.ID
		_, taskErr := systemTaskService.CreateSystemTaskFromError(commands.UpsertSystemTaskCommand{
			Type:                   models.RECEIPT_UPLOADED,
			AssociatedEntityType:   models.RECEIPT_PROCESSING_SETTINGS,
			AssociatedEntityId:     parentTask.AssociatedEntityId,
			StartedAt:              finishedAt,
			AsynqTaskId:            params.AsynqTaskId,
			GroupId:                &groupId,
			AssociatedSystemTaskId: &parentId,
		}, failureErr)
		return combineEarlyFailureErrors(failureErr, taskErr)
	}

	// Resolve the AI-assigned and user-picked category/tag ids to real records: names are filled
	// from the database (the AI returns ids only, which would otherwise fail receipt validation),
	// ids that don't resolve are dropped (hallucinated/non-existent), and ids the triggering user
	// isn't allowed to see are dropped too (defense-in-depth alongside the prompt's grant filter).
	receiptCommand.Categories, err = service.resolveQuickScanCategories(receiptCommand.Categories, params.CategoryIds, token.UserId, groupId)
	if err != nil {
		return models.Receipt{}, recordEarlyQuickScanFailure(err)
	}

	receiptCommand.Tags, err = service.resolveQuickScanTags(receiptCommand.Tags, params.TagIds, token.UserId, groupId)
	if err != nil {
		return models.Receipt{}, recordEarlyQuickScanFailure(err)
	}

	// Append the user's quick-scan comment rather than replacing whatever the AI produced: the
	// default prompt doesn't ask for comments, but the response is unmarshalled straight into an
	// UpsertReceiptCommand, so a group running a custom prompt can produce them and they must not be
	// dropped. UserId is required — UpsertCommentCommand.Validate rejects a nil one, which would fail
	// the whole receipt. ReceiptId stays unset; it is only required when updating, and CreateReceipt
	// fills the foreign key through the association.
	if len(params.Comment) > 0 {
		commentUserId := token.UserId
		receiptCommand.Comments = append(receiptCommand.Comments, commands.UpsertCommentCommand{
			Comment: params.Comment,
			UserId:  &commentUserId,
		})
	}

	// Attach the group's default custom fields (as empty values) when the group opted into applying
	// them to server-created receipts. Runs after the AI's own custom fields are in the command so
	// a field the prompt already produced is not duplicated.
	err = ApplyGroupDefaultCustomFields(service.TX, groupId, &receiptCommand)
	if err != nil {
		return models.Receipt{}, recordEarlyQuickScanFailure(err)
	}

	vErr := receiptCommand.Validate(token.UserId, true)
	if len(vErr.Errors) > 0 {
		errBytes, _ := json.Marshal(vErr.Errors)
		return models.Receipt{}, recordEarlyQuickScanFailure(fmt.Errorf("receipt validation failed: %s", string(errBytes)))
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		receiptRepository.SetTransaction(tx)
		receiptImageRepository.SetTransaction(tx)
		systemTaskService.SetTransaction(tx)
		uploadStart := time.Now()

		createdReceipt, err = receiptRepository.CreateReceipt(receiptCommand, token.UserId, false)
		_, taskErr := systemTaskService.CreateReceiptUploadedSystemTask(
			err,
			createdReceipt,
			quickScanSystemTasks,
			uploadStart,
		)
		if taskErr != nil {
			return taskErr
		}
		if err != nil {
			tx.Commit()
			return err
		}

		taskErr = systemTaskService.AssociateProcessingSystemTasksToReceipt(quickScanSystemTasks, createdReceipt.ID)
		if taskErr != nil {
			return taskErr
		}

		fileData := models.FileData{
			Name:      params.OriginalFileName,
			Size:      uint(fileInfo.Size()),
			ReceiptId: createdReceipt.ID,
			FileType:  validatedFileType,
		}
		_, err := receiptImageRepository.CreateReceiptImage(fileData, fileBytes)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Receipt{}, err
	}

	os.Remove(params.TempPath)
	return createdReceipt, nil
}

// combineEarlyFailureErrors keeps failureErr as the primary, inspectable cause (the reason no receipt
// was created) while also preserving taskErr in the unwrap chain when system-task recording itself
// failed, so errors.Is/As can reach either.
func combineEarlyFailureErrors(failureErr, taskErr error) error {
	if taskErr == nil {
		return failureErr
	}
	return fmt.Errorf("%w (recording the failure system task also failed: %w)", failureErr, taskErr)
}

// resolveQuickScanCategories turns the AI-assigned and user-picked category ids into a validated
// selection. The union of ids (AI-assigned first, then the user's picks, deduped) is resolved against
// the database so each carries its real name — the AI returns ids only, which would otherwise fail
// receipt validation. Ids that don't resolve are dropped (hallucinated / deleted), and ids the
// triggering user isn't allowed to see are dropped too (defense-in-depth alongside the prompt's grant
// filter). Returns an empty slice when there is nothing to resolve.
func (service ReceiptService) resolveQuickScanCategories(
	aiCategories []commands.UpsertCategoryCommand,
	userCategoryIds []uint,
	userId uint,
	groupId uint,
) ([]commands.UpsertCategoryCommand, error) {
	orderedIds := make([]uint, 0, len(aiCategories)+len(userCategoryIds))
	seen := make(map[uint]bool)
	appendId := func(id uint) {
		if !seen[id] {
			seen[id] = true
			orderedIds = append(orderedIds, id)
		}
	}
	for _, category := range aiCategories {
		if category.Id != nil {
			appendId(*category.Id)
		}
	}
	for _, id := range userCategoryIds {
		appendId(id)
	}
	if len(orderedIds) == 0 {
		return []commands.UpsertCategoryCommand{}, nil
	}

	categoryRepository := repositories.NewCategoryRepository(service.TX)
	records, err := categoryRepository.GetByIds(orderedIds)
	if err != nil {
		return nil, err
	}
	recordsById := make(map[uint]models.Category, len(records))
	for _, record := range records {
		recordsById[record.ID] = record
	}

	allowed, unrestricted, err := service.resolveAllowedCategoryIds(userId, groupId)
	if err != nil {
		return nil, err
	}

	resolved := make([]commands.UpsertCategoryCommand, 0, len(orderedIds))
	for _, id := range orderedIds {
		record, ok := recordsById[id]
		if !ok {
			continue // id did not resolve to a real category
		}
		if !unrestricted {
			if _, visible := allowed[id]; !visible {
				continue // category the user is not allowed to see
			}
		}

		recordId := record.ID
		resolved = append(resolved, commands.UpsertCategoryCommand{
			Id:          &recordId,
			Name:        record.Name,
			Description: record.Description,
		})
	}

	return resolved, nil
}

// resolveQuickScanTags is the tag counterpart of resolveQuickScanCategories.
func (service ReceiptService) resolveQuickScanTags(
	aiTags []commands.UpsertTagCommand,
	userTagIds []uint,
	userId uint,
	groupId uint,
) ([]commands.UpsertTagCommand, error) {
	orderedIds := make([]uint, 0, len(aiTags)+len(userTagIds))
	seen := make(map[uint]bool)
	appendId := func(id uint) {
		if !seen[id] {
			seen[id] = true
			orderedIds = append(orderedIds, id)
		}
	}
	for _, tag := range aiTags {
		if tag.Id != nil {
			appendId(*tag.Id)
		}
	}
	for _, id := range userTagIds {
		appendId(id)
	}
	if len(orderedIds) == 0 {
		return []commands.UpsertTagCommand{}, nil
	}

	tagsRepository := repositories.NewTagsRepository(service.TX)
	records, err := tagsRepository.GetByIds(orderedIds)
	if err != nil {
		return nil, err
	}
	recordsById := make(map[uint]models.Tag, len(records))
	for _, record := range records {
		recordsById[record.ID] = record
	}

	allowed, unrestricted, err := service.resolveAllowedTagIds(userId, groupId)
	if err != nil {
		return nil, err
	}

	resolved := make([]commands.UpsertTagCommand, 0, len(orderedIds))
	for _, id := range orderedIds {
		record, ok := recordsById[id]
		if !ok {
			continue // id did not resolve to a real tag
		}
		if !unrestricted {
			if _, visible := allowed[id]; !visible {
				continue // tag the user is not allowed to see
			}
		}

		recordId := record.ID
		resolved = append(resolved, commands.UpsertTagCommand{
			Id:          &recordId,
			Name:        record.Name,
			Description: record.Description,
		})
	}

	return resolved, nil
}

// resolveAllowedCategoryIds returns the set of category ids the triggering user may see in the group,
// or unrestricted=true (see-all) when the user bypasses grants (holds app.categories.read) or their
// group role grants nothing for categories. The returned set is shared grant-cache state and must
// only be read.
func (service ReceiptService) resolveAllowedCategoryIds(userId uint, groupId uint) (map[uint]struct{}, bool, error) {
	permissionService := NewPermissionService(service.TX)
	bypass, err := permissionService.userBypassesGrants(userId, permissions.AppCategoriesRead)
	if err != nil {
		return nil, false, err
	}
	if bypass {
		return nil, true, nil
	}
	return permissionService.GetGroupCategoryIdsForUser(userId, groupId)
}

// resolveAllowedTagIds is the tag counterpart of resolveAllowedCategoryIds.
func (service ReceiptService) resolveAllowedTagIds(userId uint, groupId uint) (map[uint]struct{}, bool, error) {
	permissionService := NewPermissionService(service.TX)
	bypass, err := permissionService.userBypassesGrants(userId, permissions.AppTagsRead)
	if err != nil {
		return nil, false, err
	}
	if bypass {
		return nil, true, nil
	}
	return permissionService.GetGroupTagIdsForUser(userId, groupId)
}

func (service ReceiptService) DuplicateReceipt(
	userId uint,
	receiptId string,
) (models.Receipt, error) {
	db := repositories.GetDB()
	newReceipt := models.Receipt{}

	systemTaskCommand := commands.UpsertSystemTaskCommand{
		Type:                 models.RECEIPT_UPLOADED,
		Status:               models.SYSTEM_TASK_SUCCEEDED,
		AssociatedEntityType: models.RECEIPT,
		AssociatedEntityId:   0,
		StartedAt:            time.Now(),
		EndedAt:              nil,
		ResultDescription:    "",
		RanByUserId:          &userId,
		ReceiptId:            nil,
		GroupId:              nil,
	}

	receiptRepository := repositories.NewReceiptRepository(nil)
	receipt, err := receiptRepository.GetFullyLoadedReceiptById(receiptId)
	defer func() {
		systemTaskService := NewSystemTaskService(nil)
		systemTaskService.CreateSystemTaskFromError(systemTaskCommand, err)
	}()
	if err != nil {
		return models.Receipt{}, err
	}

	systemTaskCommand.GroupId = &receipt.GroupId

	// Strip categories/tags the duplicating user cannot see so they are not
	// copied onto the new receipt.
	permissionService := NewPermissionService(nil)
	err = permissionService.FilterReceiptCategoriesTagsForReceipt(userId, &receipt)
	if err != nil {
		return models.Receipt{}, err
	}

	// Mask user references outside the caller's member-visible set (e.g. an item
	// charged to a non-visible user) so they are not carried onto the copy.
	err = permissionService.MaskReceiptForMemberVisibility(userId, &receipt)
	if err != nil {
		return models.Receipt{}, err
	}

	copier.Copy(&newReceipt, receipt)

	newReceipt.ID = 0
	newReceipt.Name = newReceipt.Name + " duplicate"
	newReceipt.ImageFiles = make([]models.FileData, 0)
	newReceipt.ReceiptItems = make([]models.Item, 0)
	newReceipt.Comments = make([]models.Comment, 0)
	newReceipt.CreatedAt = time.Now()
	newReceipt.UpdatedAt = time.Now()
	newReceipt.CreatedBy = &userId

	// Remove fks from any related data
	for _, fileData := range receipt.ImageFiles {
		var newFileData models.FileData
		copier.Copy(&newFileData, fileData)

		newFileData.ID = 0
		newFileData.ReceiptId = 0
		newFileData.Receipt = models.Receipt{}
		newReceipt.ImageFiles = append(newReceipt.ImageFiles, newFileData)
	}

	// Copy items
	for _, item := range receipt.ReceiptItems {
		var newItem models.Item
		copier.Copy(&newItem, item)

		newItem.ID = 0
		newItem.ReceiptId = 0
		newItem.Receipt = models.Receipt{}
		newReceipt.ReceiptItems = append(newReceipt.ReceiptItems, newItem)
	}

	// Copy comments
	for _, comment := range receipt.Comments {
		var newComment models.Comment
		copier.Copy(&newComment, comment)

		newComment.ID = 0
		newComment.ReceiptId = 0
		newComment.Receipt = models.Receipt{}
		newReceipt.Comments = append(newReceipt.Comments, newComment)
	}

	err = db.Create(&newReceipt).Error
	if err != nil {
		return models.Receipt{}, err
	}
	systemTaskCommand.AssociatedEntityId = newReceipt.ID
	systemTaskCommand.ReceiptId = &newReceipt.ID

	resultString, err := newReceipt.ToString()
	if err != nil {
		return models.Receipt{}, err
	}

	systemTaskCommand.ResultDescription = resultString

	// Copy receipt images
	fileRepository := repositories.NewFileRepository(nil)
	for i, fileData := range newReceipt.ImageFiles {
		srcFileData := receipt.ImageFiles[i]
		srcImageBytes, err := fileRepository.GetBytesForFileData(srcFileData)
		if err != nil {
			return models.Receipt{}, err
		}

		dstPath, err := fileRepository.BuildFilePath(
			utils.UintToString(newReceipt.ID),
			utils.UintToString(fileData.ID),
			fileData.Name,
		)
		if err != nil {
			return models.Receipt{}, err
		}

		err = utils.WriteDataFile(dstPath, srcImageBytes)
		if err != nil {
			return models.Receipt{}, err
		}
	}

	return newReceipt, nil
}
