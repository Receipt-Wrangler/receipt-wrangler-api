package wranglerasynq

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/emailclient"
	config "receipt-wrangler/api/internal/env"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

const (
	// maxMatchedMessagesPerMailboxPoll bounds the work of one poll. Messages that
	// match no group do not count, so a backlog of unrelated unread mail can never
	// starve new receipts. Whatever is left stays unread for the next poll.
	maxMatchedMessagesPerMailboxPoll = 50
	// maxEmailMessageBytes matches the 50 MB request cap nginx puts on uploads.
	maxEmailMessageBytes = 50 << 20
)

type emailMessageOutcome int

const (
	// leaveUnread keeps the message for the next poll.
	leaveUnread emailMessageOutcome = iota
	// markRead means the message is finished with, processed or failed.
	markRead
)

// emailPoller reads the group mailboxes and queues their receipts. Its fields are
// the seams that keep it testable without an IMAP server, DNS or Redis.
type emailPoller struct {
	connect   func(credentials emailclient.Credentials) (emailclient.Mailbox, error)
	lookupTXT emailclient.TXTLookup
	enqueue   func(payload EmailProcessTaskPayload) error
}

// newEmailPoller builds the production poller. A variable so tests can drive
// PollEmail end to end with fakes in place of IMAP, DNS and Redis.
var newEmailPoller = func() emailPoller {
	return emailPoller{
		connect:   emailclient.Connector{}.Connect,
		lookupTXT: net.LookupTXT,
		enqueue:   enqueueEmailProcessTask,
	}
}

func enqueueEmailProcessTask(payload EmailProcessTaskPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = EnqueueTask(asynq.NewTask(EmailProcess, payloadBytes), models.EmailReceiptProcessingQueue)
	return err
}

// mailboxPollLocks holds one mutex per system email. Messages stay unread while
// they are processed, so the scheduled poll and a manual one reading the same
// mailbox at once would both see them and queue each twice.
var mailboxPollLocks sync.Map

func mailboxPollLock(systemEmailId uint) *sync.Mutex {
	lock, _ := mailboxPollLocks.LoadOrStore(systemEmailId, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// poll reads every mailbox used by the requested groups. Messages are routed
// against every enabled group on that mailbox, not just the requested ones, so a
// manual poll for one group never swallows another group's mail.
//
// Each mailbox is independent: one that fails is logged and the rest are still
// read. The poll only fails when every mailbox did.
func (poller emailPoller) poll(requested []models.GroupSettings, enabled []models.GroupSettings) error {
	systemEmails := mailboxesOf(requested)
	failures := 0

	for _, systemEmail := range systemEmails {
		err := poller.pollMailbox(systemEmail, groupsReading(enabled, systemEmail.ID))
		if err != nil {
			failures++
			logging.LogStd(logging.LOG_LEVEL_ERROR, fmt.Sprintf(
				"Polling system email %d (%s on %s) failed: %s", systemEmail.ID, systemEmail.Username, systemEmail.Host, err.Error(),
			))
		}
	}

	if len(systemEmails) > 0 && failures == len(systemEmails) {
		return fmt.Errorf("polling failed for all %d mailboxes", failures)
	}

	return nil
}

func (poller emailPoller) pollMailbox(systemEmail models.SystemEmail, groups []models.GroupSettings) error {
	lock := mailboxPollLock(systemEmail.ID)
	if !lock.TryLock() {
		logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf("System email %d is already being polled; skipping it this time", systemEmail.ID))
		return nil
	}
	defer lock.Unlock()

	password, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
	if err != nil {
		return err
	}

	mailbox, err := poller.connect(emailclient.Credentials{
		Host:        systemEmail.Host,
		Port:        systemEmail.Port,
		Username:    systemEmail.Username,
		Password:    password,
		UseStartTLS: systemEmail.UseStartTLS,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := mailbox.Close(); err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not close the IMAP session: ", err.Error())
		}
	}()

	uids, err := mailbox.UnseenUIDs()
	if err != nil || len(uids) == 0 {
		return err
	}

	headers, err := mailbox.FetchHeaders(uids)
	if err != nil {
		return err
	}

	handled := 0
	for _, header := range headers {
		fields := emailclient.ParseHeaderFields(header.Header)
		matched := groupsWithIds(groups, emailclient.Route(fields.From, fields.Subject, systemEmail.ID, groups))
		if len(matched) == 0 {
			// Not for any group: left unread for whoever else reads this inbox.
			continue
		}

		if handled == maxMatchedMessagesPerMailboxPoll {
			logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf(
				"System email %d has more than %d matching emails; the rest wait for the next poll", systemEmail.ID, maxMatchedMessagesPerMailboxPoll,
			))
			break
		}
		handled++

		outcome, err := poller.processMessage(mailbox, systemEmail, header, fields, matched)
		if err != nil {
			return err
		}

		if outcome == markRead {
			if err := mailbox.MarkSeen(header.UID); err != nil {
				return err
			}
		}
	}

	return nil
}

// processMessage turns one routed message into queued tasks. Anything wrong with
// the message itself is recorded as a failed activity and the message is marked
// read, so it is never retried forever. An error is returned only for problems
// with the connection or the queue, which abort the mailbox and leave the
// message unread for the next poll.
func (poller emailPoller) processMessage(
	mailbox emailclient.Mailbox,
	systemEmail models.SystemEmail,
	header emailclient.MessageHeader,
	fields emailclient.HeaderFields,
	groups []models.GroupSettings,
) (emailMessageOutcome, error) {
	failure := emailFailure{systemEmail: systemEmail, sender: firstOrEmpty(fields.From), subject: fields.Subject}

	if header.Size > maxEmailMessageBytes {
		failure.record(groups, fmt.Sprintf("the email is %d MB, over the %d MB limit", header.Size>>20, maxEmailMessageBytes>>20))
		return markRead, nil
	}

	raw, err := mailbox.FetchRaw(header.UID)
	if err != nil {
		return leaveUnread, err
	}

	message, err := emailclient.ParseMessage(raw)
	if err != nil {
		failure.record(groups, fmt.Sprintf("the email could not be read: %s", err.Error()))
		return markRead, nil
	}
	failure.sender = message.From.Email
	failure.subject = message.Subject

	requiring, notRequiring := splitBySignedSenderRequirement(groups)
	if len(requiring) > 0 {
		verdict := emailclient.VerifySender(raw, message.From.Email, poller.lookupTXT)
		switch verdict.Kind {
		case emailclient.RetryLater:
			logging.LogStd(logging.LOG_LEVEL_INFO, "Leaving an email unread to retry its sender check: ", verdict.Reason)
			return leaveUnread, nil
		case emailclient.Unauthenticated:
			failure.record(requiring, "the sender could not be verified: "+verdict.Reason)
			groups = notRequiring
		}
	}

	if len(groups) == 0 {
		return markRead, nil
	}

	if !message.HasContent() {
		logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf("Email %q from %s has no attachment or body to process", message.Subject, message.From.Email))
		return markRead, nil
	}

	attachments := make([]savedEmailAttachment, 0, len(message.Attachments))
	for _, attachment := range message.Attachments {
		saved, err := saveEmailAttachment(attachment)
		if err != nil {
			failure.record(groups, fmt.Sprintf("the attachment %q could not be processed: %s", attachment.Filename, err.Error()))
			continue
		}
		attachments = append(attachments, saved)
	}

	if len(message.Attachments) > 0 && len(attachments) == 0 {
		// Every attachment failed. Falling back to the body alone would create a
		// receipt without the image it was sent for.
		return markRead, nil
	}

	for _, payload := range buildEmailProcessPayloads(emailMetadataOf(message, attachments, groups), attachments, groups) {
		if err := poller.enqueue(payload); err != nil {
			return leaveUnread, fmt.Errorf("could not queue the email for processing: %w", err)
		}
	}

	logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf(
		"Email captured: subject=%q from=%q attachments=%d bodyLen=%d bodyHtmlLen=%d groups=%v",
		message.Subject, message.From.Email, len(attachments), len(message.Body), len(message.BodyHtml), groupSettingsIdsOf(groups),
	))

	return markRead, nil
}

// savedEmailAttachment is an attachment written to temp/, ready for a task.
type savedEmailAttachment struct {
	tempFilePath    string
	imageForOcrPath string
	attachment      structs.Attachment
}

// saveEmailAttachment writes the attachment under a random name, never the
// sender's: two emails sending "receipt.pdf", or Outlook naming every inline
// image "image001.png", used to overwrite each other in temp/, putting one
// email's image on another email's receipt. The sender's name is kept on the
// Attachment for display and download.
func saveEmailAttachment(attachment emailclient.Attachment) (savedEmailAttachment, error) {
	fileRepository := repositories.NewFileRepository(nil)

	tempFilePath, err := fileRepository.WriteTempFile(attachment.Data)
	if err != nil {
		return savedEmailAttachment{}, err
	}

	ocrBytes, err := fileRepository.GetBytesFromImageBytes(attachment.Data)
	if err != nil {
		os.Remove(tempFilePath)
		return savedEmailAttachment{}, err
	}

	imageForOcrPath := filepath.Join(filepath.Dir(tempFilePath), "image-"+filepath.Base(tempFilePath))
	if err := utils.WriteFile(imageForOcrPath, ocrBytes); err != nil {
		os.Remove(tempFilePath)
		os.Remove(imageForOcrPath)
		return savedEmailAttachment{}, err
	}

	return savedEmailAttachment{
		tempFilePath:    tempFilePath,
		imageForOcrPath: imageForOcrPath,
		attachment: structs.Attachment{
			Filename: attachment.Filename,
			FileType: attachment.FileType,
			Size:     uint(len(attachment.Data)),
		},
	}, nil
}

func emailMetadataOf(message emailclient.Message, attachments []savedEmailAttachment, groups []models.GroupSettings) structs.EmailMetadata {
	metadataAttachments := make([]structs.Attachment, len(attachments))
	for i, saved := range attachments {
		metadataAttachments[i] = saved.attachment
	}

	return structs.EmailMetadata{
		Date:             message.Date,
		Subject:          message.Subject,
		To:               message.To,
		FromName:         message.From.Name,
		FromEmail:        message.From.Email,
		Body:             message.Body,
		BodyHtml:         message.BodyHtml,
		Attachments:      metadataAttachments,
		GroupSettingsIds: groupSettingsIdsOf(groups),
	}
}

// buildEmailProcessPayloads applies the per-group rules email ingestion has
// always used: one task per attachment per group, with the body blanked for
// groups that do not process bodies, and for an email without attachments, one
// task per group that does.
func buildEmailProcessPayloads(
	metadata structs.EmailMetadata,
	attachments []savedEmailAttachment,
	groups []models.GroupSettings,
) []EmailProcessTaskPayload {
	lookup := make(map[uint]models.GroupSettings, len(groups))
	for _, settings := range groups {
		lookup[settings.ID] = settings
	}

	payloads := make([]EmailProcessTaskPayload, 0)

	for _, saved := range attachments {
		for _, settings := range groups {
			taskMetadata := metadata
			if !settings.EmailBodyProcessingEnabled {
				taskMetadata.Body = ""
				taskMetadata.BodyHtml = ""
			}

			payloads = append(payloads, EmailProcessTaskPayload{
				GroupSettingsId: settings.ID,
				ImageForOcrPath: saved.imageForOcrPath,
				TempFilePath:    saved.tempFilePath,
				Metadata:        taskMetadata,
				Attachment:      saved.attachment,
				RenderBodyPdf:   shouldRenderEmailBodyPdfForGroup(metadata, settings.ID, lookup),
			})
		}
	}

	if len(metadata.Attachments) == 0 && (len(metadata.Body) > 0 || len(metadata.BodyHtml) > 0) {
		for _, settings := range groups {
			if !settings.EmailBodyProcessingEnabled {
				continue
			}

			payloads = append(payloads, EmailProcessTaskPayload{
				GroupSettingsId: settings.ID,
				Metadata:        metadata,
				RenderBodyPdf:   shouldRenderEmailBodyPdfForGroup(metadata, settings.ID, lookup),
			})
		}
	}

	return payloads
}

// emailFailure records why an email could not become a receipt: a failed
// EMAIL_READ for the admin System Tasks page, with a failed EMAIL_UPLOAD per
// affected group so the group's activity feed shows it too. That is the same
// parent/child shape a processed email produces.
type emailFailure struct {
	systemEmail models.SystemEmail
	sender      string
	subject     string
}

func (failure emailFailure) record(groups []models.GroupSettings, reason string) {
	description := fmt.Sprintf("Email from %s with subject %q was not processed: %s", failure.sender, failure.subject, reason)
	logging.LogStd(logging.LOG_LEVEL_ERROR, description)

	systemTaskRepository := repositories.NewSystemTaskRepository(nil)
	now := time.Now()

	emailRead, err := systemTaskRepository.CreateSystemTask(commands.UpsertSystemTaskCommand{
		Type:                 models.EMAIL_READ,
		Status:               models.SYSTEM_TASK_FAILED,
		AssociatedEntityType: models.SYSTEM_EMAIL,
		AssociatedEntityId:   failure.systemEmail.ID,
		StartedAt:            now,
		EndedAt:              &now,
		ResultDescription:    description,
	})
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not record the email failure: ", err.Error())
		return
	}

	for _, settings := range groups {
		groupId := settings.GroupId
		_, err := systemTaskRepository.CreateSystemTask(commands.UpsertSystemTaskCommand{
			Type:                   models.EMAIL_UPLOAD,
			Status:                 models.SYSTEM_TASK_FAILED,
			AssociatedEntityType:   models.SYSTEM_EMAIL,
			AssociatedEntityId:     failure.systemEmail.ID,
			StartedAt:              now,
			EndedAt:                &now,
			ResultDescription:      description,
			GroupId:                &groupId,
			AssociatedSystemTaskId: &emailRead.ID,
		})
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not record the email failure for group ", groupId, ": ", err.Error())
		}
	}
}

// mailboxesOf returns the distinct system emails the groups read, by id.
func mailboxesOf(groupSettings []models.GroupSettings) []models.SystemEmail {
	byId := make(map[uint]models.SystemEmail)
	for _, settings := range groupSettings {
		if settings.SystemEmailId == nil || settings.SystemEmail.ID != *settings.SystemEmailId {
			continue
		}
		byId[settings.SystemEmail.ID] = settings.SystemEmail
	}

	systemEmails := make([]models.SystemEmail, 0, len(byId))
	for _, systemEmail := range byId {
		systemEmails = append(systemEmails, systemEmail)
	}
	slices.SortFunc(systemEmails, func(a, b models.SystemEmail) int { return int(a.ID) - int(b.ID) })

	return systemEmails
}

func groupsReading(groupSettings []models.GroupSettings, systemEmailId uint) []models.GroupSettings {
	result := make([]models.GroupSettings, 0)
	for _, settings := range groupSettings {
		if settings.SystemEmailId != nil && *settings.SystemEmailId == systemEmailId {
			result = append(result, settings)
		}
	}
	return result
}

func groupsWithIds(groupSettings []models.GroupSettings, ids []uint) []models.GroupSettings {
	result := make([]models.GroupSettings, 0, len(ids))
	for _, settings := range groupSettings {
		if slices.Contains(ids, settings.ID) {
			result = append(result, settings)
		}
	}
	return result
}

func splitBySignedSenderRequirement(groupSettings []models.GroupSettings) ([]models.GroupSettings, []models.GroupSettings) {
	requiring := make([]models.GroupSettings, 0)
	notRequiring := make([]models.GroupSettings, 0)
	for _, settings := range groupSettings {
		if settings.EmailRequireSignedSender {
			requiring = append(requiring, settings)
		} else {
			notRequiring = append(notRequiring, settings)
		}
	}
	return requiring, notRequiring
}

func groupSettingsIdsOf(groupSettings []models.GroupSettings) []uint {
	ids := make([]uint, len(groupSettings))
	for i, settings := range groupSettings {
		ids[i] = settings.ID
	}
	return ids
}

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
