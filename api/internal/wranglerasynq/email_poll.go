package wranglerasynq

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"path/filepath"
	"slices"
	"strings"
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
	// retryBackoff is how long a message whose check could not finish waits before
	// it is tried again. Without it, a sender whose DNS keeps failing took a slot
	// under the cap on every poll, and 50 of them stopped the mailbox.
	retryBackoff = 30 * time.Minute
	// senderCheckGiveUpAfter ends the wait on a sender check that keeps failing
	// temporarily, counted from delivery. The wait also holds back the matched
	// groups that do not require the check, so it cannot be open-ended.
	senderCheckGiveUpAfter = 24 * time.Hour
)

// signedSenderHint follows a failed sender check, because on some providers the
// likeliest cause is the mailbox rather than the sender.
const signedSenderHint = `If this mailbox cannot verify signatures (Exchange and Office 365 rebuild messages served over IMAP), turn off "Require Signed Sender (DKIM)" in the group's email settings.`

// emailPoller reads the group mailboxes and queues their receipts. Its fields are
// the seams that keep it testable without an IMAP server, DNS or Redis.
type emailPoller struct {
	connect   func(credentials emailclient.Credentials) (emailclient.Mailbox, error)
	lookupTXT emailclient.TXTLookup
	enqueue   func(payload EmailProcessTaskPayload) error
	now       func() time.Time
}

// newEmailPoller builds the production poller. A variable so tests can drive
// PollEmail end to end with fakes in place of IMAP, DNS and Redis.
var newEmailPoller = func() emailPoller {
	return emailPoller{
		connect:   emailclient.Connector{}.Connect,
		lookupTXT: net.LookupTXT,
		enqueue:   enqueueEmailProcessTask,
		now:       time.Now,
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

// accountKey identifies an IMAP account. System Email rows with the same key read
// the same INBOX, so they are polled as one. The username compares
// case-insensitively, as email-address logins do on mainstream servers.
type accountKey struct {
	host     string
	port     string
	username string
}

func accountKeyOf(systemEmail models.SystemEmail) accountKey {
	return accountKey{
		host:     strings.ToLower(strings.TrimSpace(systemEmail.Host)),
		port:     strings.TrimSpace(systemEmail.Port),
		username: strings.ToLower(strings.TrimSpace(systemEmail.Username)),
	}
}

// mailAccount is one IMAP account and everything that reads it.
type mailAccount struct {
	key accountKey
	// systemEmails are the account's rows, by id. Any of them can log in.
	systemEmails []models.SystemEmail
	// groups are the enabled groups on any of those rows.
	groups []models.GroupSettings
}

func (account mailAccount) String() string {
	ids := make([]string, len(account.systemEmails))
	for i, systemEmail := range account.systemEmails {
		ids[i] = utils.UintToString(systemEmail.ID)
	}
	return fmt.Sprintf("%s on %s (system email %s)", account.key.username, account.key.host, strings.Join(ids, ", "))
}

// mailboxPollLocks holds one mutex per IMAP account. Messages stay unread while
// they are processed, so two polls of one account at once (the scheduled poll and
// a manual one) would both see them and queue each twice.
var mailboxPollLocks sync.Map

func mailboxPollLock(key accountKey) *sync.Mutex {
	lock, _ := mailboxPollLocks.LoadOrStore(key, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// mailboxMemory is what one poll learned about an account's unread mail that the
// next can rely on. A message's From and Subject never change, so one that matched
// no group matches none until the routing changes. Without it, the headers of
// every unrelated unread message were fetched again on every poll, forever.
//
// It is kept in memory only, so a restart costs one full header scan. It is read
// and written only under the account's poll lock.
type mailboxMemory struct {
	uidValidity uint32
	fingerprint string
	unmatched   map[uint32]struct{}
	// deferredUntil holds the messages waiting to be tried again, and until when.
	deferredUntil map[uint32]time.Time
}

var mailboxMemories sync.Map

func mailboxMemoryOf(key accountKey) *mailboxMemory {
	memory, _ := mailboxMemories.LoadOrStore(key, &mailboxMemory{})
	return memory.(*mailboxMemory)
}

// refresh forgets everything once a UID may name a different message (a new
// UIDVALIDITY) or the routing has changed (a new fingerprint).
func (memory *mailboxMemory) refresh(uidValidity uint32, fingerprint string) {
	if memory.unmatched != nil && memory.uidValidity == uidValidity && memory.fingerprint == fingerprint {
		return
	}
	*memory = mailboxMemory{
		uidValidity:   uidValidity,
		fingerprint:   fingerprint,
		unmatched:     make(map[uint32]struct{}),
		deferredUntil: make(map[uint32]time.Time),
	}
}

// forgetAllBut drops the messages that are no longer unread, which bounds the
// memory to the mailbox's unread mail. unseen is in UID order.
func (memory *mailboxMemory) forgetAllBut(unseen []uint32) {
	gone := func(uid uint32) bool {
		_, found := slices.BinarySearch(unseen, uid)
		return !found
	}
	maps.DeleteFunc(memory.unmatched, func(uid uint32, _ struct{}) bool { return gone(uid) })
	maps.DeleteFunc(memory.deferredUntil, func(uid uint32, _ time.Time) bool { return gone(uid) })
}

// worthFetching returns the unread messages whose headers are worth fetching: not
// known to match no group, and not waiting out a retry.
func (memory *mailboxMemory) worthFetching(unseen []uint32, now time.Time) []uint32 {
	result := make([]uint32, 0, len(unseen))
	for _, uid := range unseen {
		if _, unmatched := memory.unmatched[uid]; unmatched {
			continue
		}
		if until, deferred := memory.deferredUntil[uid]; deferred && now.Before(until) {
			continue
		}
		result = append(result, uid)
	}
	return result
}

// fingerprintOf summarizes everything that decides where an account's mail goes
// and whether its sender is checked. A change to any of it starts a fresh memory:
// a whitelist or regex edit, a group enabled, disabled or moved, or the DKIM
// requirement turned off to release mail waiting on its sender check.
func fingerprintOf(groups []models.GroupSettings) string {
	type groupRouting struct {
		Id                  uint
		RequireSignedSender bool
		Whitelist           []string
		Regexes             []string
	}

	routing := make([]groupRouting, 0, len(groups))
	for _, settings := range groups {
		entry := groupRouting{Id: settings.ID, RequireSignedSender: settings.EmailRequireSignedSender}
		for _, allowed := range settings.EmailWhiteList {
			entry.Whitelist = append(entry.Whitelist, allowed.Email)
		}
		for _, subjectRegex := range settings.SubjectLineRegexes {
			entry.Regexes = append(entry.Regexes, subjectRegex.Regex)
		}
		slices.Sort(entry.Whitelist)
		slices.Sort(entry.Regexes)
		routing = append(routing, entry)
	}
	slices.SortFunc(routing, func(a, b groupRouting) int { return cmp.Compare(a.Id, b.Id) })

	// Marshalling strings, bools and numbers cannot fail.
	encoded, _ := json.Marshal(routing)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// poll reads every IMAP account used by the requested groups. Messages are routed
// against every enabled group on that account, not just the requested ones, so a
// manual poll for one group never swallows another group's mail.
//
// Each account is independent: one that fails is logged and the rest are still
// read. The poll only fails when every account did.
func (poller emailPoller) poll(requested []models.GroupSettings, enabled []models.GroupSettings) error {
	accounts := accountsOf(requested, enabled)
	failures := 0

	for _, account := range accounts {
		if err := poller.pollAccount(account); err != nil {
			failures++
			logging.LogStd(logging.LOG_LEVEL_ERROR, fmt.Sprintf("Polling %s failed: %s", account, err.Error()))
		}
	}

	if len(accounts) > 0 && failures == len(accounts) {
		return fmt.Errorf("polling failed for all %d mailboxes", failures)
	}

	return nil
}

func (poller emailPoller) pollAccount(account mailAccount) error {
	lock := mailboxPollLock(account.key)
	if !lock.TryLock() {
		logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf("%s is already being polled; skipping it this time", account))
		return nil
	}
	defer lock.Unlock()

	mailbox, systemEmail, err := poller.connectAccount(account)
	if err != nil {
		return err
	}
	defer func() {
		if err := mailbox.Close(); err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not close the IMAP session: ", err.Error())
		}
	}()

	now := poller.now()
	memory := mailboxMemoryOf(account.key)
	memory.refresh(mailbox.UIDValidity(), fingerprintOf(account.groups))

	uids, err := mailbox.UnseenUIDs()
	if err != nil {
		return err
	}
	memory.forgetAllBut(uids)

	uids = memory.worthFetching(uids, now)
	if len(uids) == 0 {
		return nil
	}

	headers, err := mailbox.FetchHeaders(uids)
	if err != nil {
		return err
	}

	router := emailclient.NewRouter(account.groups)
	handled := 0
	for _, header := range headers {
		fields := emailclient.ParseHeaderFields(header.Header)
		matched := router.Route(fields.From, fields.Subject)
		if len(matched) == 0 {
			// Not for any group: left unread for whoever else reads this inbox.
			memory.unmatched[header.UID] = struct{}{}
			continue
		}

		if handled == maxMatchedMessagesPerMailboxPoll {
			logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf(
				"%s has more than %d matching emails; the rest wait for the next poll", account, maxMatchedMessagesPerMailboxPoll,
			))
			break
		}
		handled++

		plan, err := poller.planMessage(mailbox, systemEmail, header, fields, matched)
		if err != nil {
			return err
		}
		if plan.retryLater {
			memory.deferredUntil[header.UID] = now.Add(retryBackoff)
			continue
		}

		if err := poller.finishMessage(mailbox, header.UID, plan); err != nil {
			return err
		}
	}

	return nil
}

// connectAccount logs in with the account's rows in id order, so a stale password
// on one row does not lock out the groups on the others. It returns the row that
// logged in; the poll's failed activities are recorded against it.
func (poller emailPoller) connectAccount(account mailAccount) (emailclient.Mailbox, models.SystemEmail, error) {
	errs := make([]error, 0, len(account.systemEmails))

	for i, systemEmail := range account.systemEmails {
		mailbox, err := poller.connectAs(systemEmail)
		if err == nil {
			return mailbox, systemEmail, nil
		}

		err = fmt.Errorf("system email %d: %w", systemEmail.ID, err)
		errs = append(errs, err)
		if i < len(account.systemEmails)-1 {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not log in; trying the account's next system email: ", err.Error())
		}
	}

	return nil, models.SystemEmail{}, errors.Join(errs...)
}

func (poller emailPoller) connectAs(systemEmail models.SystemEmail) (emailclient.Mailbox, error) {
	password, err := utils.DecryptB64EncodedData(config.GetEncryptionKey(), systemEmail.Password)
	if err != nil {
		return nil, err
	}

	return poller.connect(emailclient.Credentials{
		Host:        systemEmail.Host,
		Port:        systemEmail.Port,
		Username:    systemEmail.Username,
		Password:    password,
		UseStartTLS: systemEmail.UseStartTLS,
	})
}

// pendingFailure is a failed activity waiting to be recorded.
type pendingFailure struct {
	groups []models.GroupSettings
	reason string
}

// messagePlan is everything one message will cause, worked out before any of it
// happens: the tasks to queue and the failures to record.
type messagePlan struct {
	// retryLater leaves the message unread and untouched, to be tried again once
	// retryBackoff has passed.
	retryLater bool
	context    emailFailure
	groups     []models.GroupSettings
	payloads   []EmailProcessTaskPayload
	failures   []pendingFailure
	// summary is logged once the message's tasks are queued.
	summary string
}

func (plan *messagePlan) fail(groups []models.GroupSettings, reason string) {
	plan.failures = append(plan.failures, pendingFailure{groups: groups, reason: reason})
}

// planMessage works out what one routed message should become. Its only side
// effect is writing attachments to temp/; queuing and recording wait for
// finishMessage. Anything wrong with the message itself becomes a failure in the
// plan. An error is returned only when the session failed, which abandons the
// mailbox for this poll.
func (poller emailPoller) planMessage(
	mailbox emailclient.Mailbox,
	systemEmail models.SystemEmail,
	header emailclient.MessageHeader,
	fields emailclient.HeaderFields,
	groups []models.GroupSettings,
) (messagePlan, error) {
	plan := messagePlan{
		context: emailFailure{systemEmail: systemEmail, sender: firstOrEmpty(fields.From), subject: fields.Subject},
		groups:  groups,
	}

	if header.Size > maxEmailMessageBytes {
		plan.fail(groups, fmt.Sprintf("the email is %d MB, over the %d MB limit", header.Size>>20, maxEmailMessageBytes>>20))
		return plan, nil
	}

	raw, err := mailbox.FetchRaw(header.UID)
	if err != nil {
		var fetchErr *emailclient.MessageFetchError
		if !errors.As(err, &fetchErr) {
			return messagePlan{}, err
		}

		switch {
		case fetchErr.Gone:
			logging.LogStd(logging.LOG_LEVEL_INFO, "Skipping an email that is no longer in the mailbox: ", fetchErr.Error())
			return messagePlan{retryLater: true}, nil
		case fetchErr.Temporary:
			logging.LogStd(logging.LOG_LEVEL_INFO, "Leaving an email unread that the server could not return right now: ", fetchErr.Error())
			return messagePlan{retryLater: true}, nil
		}

		// Retrying a message the server refuses would block every later one, which
		// always comes after it in UID order.
		plan.fail(groups, "the email server would not return it: "+fetchErr.Err.Error())
		return plan, nil
	}

	message, err := emailclient.ParseMessage(raw)
	if err != nil {
		plan.fail(groups, fmt.Sprintf("the email could not be read: %s", err.Error()))
		return plan, nil
	}
	plan.context.sender = message.From.Email
	plan.context.subject = message.Subject

	requiring, notRequiring := splitBySignedSenderRequirement(groups)
	if len(requiring) > 0 {
		verdict := emailclient.VerifySender(raw, message.From.Email, poller.lookupTXT)
		switch verdict.Kind {
		case emailclient.RetryLater:
			if !poller.senderCheckExpired(header) {
				logging.LogStd(logging.LOG_LEVEL_INFO, "Leaving an email unread to retry its sender check: ", verdict.Reason)
				return messagePlan{retryLater: true}, nil
			}
			plan.fail(requiring, fmt.Sprintf(
				"the sender could not be verified: its DKIM signature still could not be checked %d hours after the email arrived", int(senderCheckGiveUpAfter.Hours()),
			))
			groups = notRequiring
		case emailclient.Unauthenticated:
			plan.fail(requiring, fmt.Sprintf("the sender could not be verified: %s. %s", verdict.Reason, signedSenderHint))
			groups = notRequiring
		}
	}
	plan.groups = groups

	if len(groups) == 0 {
		return plan, nil
	}

	// Recorded before the content check, so an email whose only attachment is
	// unusable is reported rather than marked read as having nothing in it.
	for _, rejected := range message.RejectedAttachments {
		plan.fail(groups, fmt.Sprintf("the attachment %q could not be processed: %s", rejected.Filename, rejected.Reason))
	}

	if !message.HasContent() {
		logging.LogStd(logging.LOG_LEVEL_INFO, fmt.Sprintf("Email %q from %s has no attachment or body to process", message.Subject, message.From.Email))
		return plan, nil
	}

	attachments := make([]savedEmailAttachment, 0, len(message.Attachments))
	for _, attachment := range message.Attachments {
		saved, err := saveEmailAttachment(attachment)
		if err != nil {
			plan.fail(groups, fmt.Sprintf("the attachment %q could not be processed: %s", attachment.Filename, err.Error()))
			continue
		}
		attachments = append(attachments, saved)
	}

	attempted := len(message.Attachments) + len(message.RejectedAttachments)
	if attempted > 0 && len(attachments) == 0 {
		// Every attachment failed. Falling back to the body alone would create a
		// receipt without the image it was sent for.
		return plan, nil
	}

	plan.payloads = buildEmailProcessPayloads(emailMetadataOf(message, attachments, groups), attachments, groups)
	plan.summary = fmt.Sprintf(
		"Email captured: subject=%q from=%q attachments=%d bodyLen=%d bodyHtmlLen=%d groups=%v",
		message.Subject, message.From.Email, len(attachments), len(message.Body), len(message.BodyHtml), groupSettingsIdsOf(groups),
	)

	return plan, nil
}

// senderCheckExpired reports whether a message has waited long enough on a sender
// check that keeps failing temporarily. A message without a delivery time waits
// indefinitely, at retryBackoff intervals; every IMAP server sends one.
func (poller emailPoller) senderCheckExpired(header emailclient.MessageHeader) bool {
	return !header.ReceivedAt.IsZero() && poller.now().Sub(header.ReceivedAt) >= senderCheckGiveUpAfter
}

// finishMessage carries out a plan, claiming the message first: it is marked
// read before anything is queued or recorded. Marking it last, as before, meant
// a queue error part-way through, or a failed mark, sent the whole message
// through again on the next poll and duplicated the receipts already queued.
// On a mailbox that never accepts the mark, that repeated on every poll.
//
// The cost is the opposite failure: if the process dies between claiming and
// queuing, that one message is lost without a record.
func (poller emailPoller) finishMessage(mailbox emailclient.Mailbox, uid uint32, plan messagePlan) error {
	if err := mailbox.MarkSeen(uid); err != nil {
		return err
	}

	queued := 0
	var queueErr error
	for _, payload := range plan.payloads {
		if queueErr = poller.enqueue(payload); queueErr != nil {
			break
		}
		queued++
	}

	if queueErr != nil {
		queueErr = fmt.Errorf("could not queue the email for processing: %w", queueErr)

		if queued == 0 {
			// Nothing has happened yet, so the message can go back for a clean retry.
			unmarkErr := mailbox.MarkUnseen(uid)
			if unmarkErr == nil {
				return queueErr
			}
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not return an unqueued email to unread: ", unmarkErr.Error())
		}

		// Retrying now would queue the earlier tasks twice, so the message stays read
		// and what was not queued is recorded instead.
		plan.fail(groupsOfPayloads(plan.groups, plan.payloads[queued:]), queueErr.Error())
	}

	for _, failure := range plan.failures {
		plan.context.record(failure.groups, failure.reason)
	}

	if queueErr == nil && len(plan.summary) > 0 {
		logging.LogStd(logging.LOG_LEVEL_INFO, plan.summary)
	}

	return queueErr
}

// groupsOfPayloads returns the groups the payloads are for, in the order given.
func groupsOfPayloads(groups []models.GroupSettings, payloads []EmailProcessTaskPayload) []models.GroupSettings {
	result := make([]models.GroupSettings, 0)
	for _, settings := range groups {
		for _, payload := range payloads {
			if payload.GroupSettingsId == settings.ID {
				result = append(result, settings)
				break
			}
		}
	}
	return result
}

// savedEmailAttachment is an attachment written to temp/, ready for a task.
type savedEmailAttachment struct {
	tempFilePath string
	// imageForOcrPath is where the task writes the converted copy.
	imageForOcrPath string
	attachment      structs.Attachment
}

// saveEmailAttachment writes the attachment under a random name, never the
// sender's: two emails sending "receipt.pdf", or Outlook naming every inline
// image "image001.png", used to overwrite each other in temp/, putting one
// email's image on another email's receipt. The sender's name is kept on the
// Attachment for display and download.
//
// Only the original is written. The OCR copy at imageForOcrPath is made by the
// task (ensureOcrCopy), so a file ImageMagick cannot convert fails there, where
// the original is kept and the activity can be rerun. Converting here deleted
// the original and marked the email read, losing the image for good. It also
// kept ImageMagick running between IMAP commands.
func saveEmailAttachment(attachment emailclient.Attachment) (savedEmailAttachment, error) {
	fileRepository := repositories.NewFileRepository(nil)

	tempFilePath, err := fileRepository.WriteTempFile(attachment.Data)
	if err != nil {
		return savedEmailAttachment{}, err
	}

	imageForOcrPath := filepath.Join(filepath.Dir(tempFilePath), "image-"+filepath.Base(tempFilePath))

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
	// asynqTaskId is set when the failure happened inside a queued task. It is
	// what lets the activity offer its source file and, once archived, a rerun.
	asynqTaskId string
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
		AsynqTaskId:          failure.asynqTaskId,
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
			AsynqTaskId:            failure.asynqTaskId,
		})
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, "Could not record the email failure for group ", groupId, ": ", err.Error())
		}
	}
}

// accountsOf returns the IMAP accounts the requested groups read, each with every
// enabled group on any of its rows. Rows for one account used to be polled
// separately, so whichever ran first consumed the mail and the other row's groups
// never saw it.
func accountsOf(requested []models.GroupSettings, enabled []models.GroupSettings) []mailAccount {
	wanted := make(map[accountKey]bool)
	for _, settings := range requested {
		if systemEmail, ok := systemEmailOf(settings); ok {
			wanted[accountKeyOf(systemEmail)] = true
		}
	}

	byKey := make(map[accountKey]*mailAccount)
	for _, settings := range enabled {
		systemEmail, ok := systemEmailOf(settings)
		if !ok {
			continue
		}
		key := accountKeyOf(systemEmail)
		if !wanted[key] {
			continue
		}

		account, exists := byKey[key]
		if !exists {
			account = &mailAccount{key: key}
			byKey[key] = account
		}
		if !slices.ContainsFunc(account.systemEmails, func(row models.SystemEmail) bool { return row.ID == systemEmail.ID }) {
			account.systemEmails = append(account.systemEmails, systemEmail)
		}
		account.groups = append(account.groups, settings)
	}

	accounts := make([]mailAccount, 0, len(byKey))
	for _, account := range byKey {
		slices.SortFunc(account.systemEmails, func(a, b models.SystemEmail) int { return cmp.Compare(a.ID, b.ID) })
		accounts = append(accounts, *account)
	}
	slices.SortFunc(accounts, func(a, b mailAccount) int { return cmp.Compare(a.systemEmails[0].ID, b.systemEmails[0].ID) })

	return accounts
}

func systemEmailOf(settings models.GroupSettings) (models.SystemEmail, bool) {
	if settings.SystemEmailId == nil || settings.SystemEmail.ID != *settings.SystemEmailId {
		return models.SystemEmail{}, false
	}
	return settings.SystemEmail, true
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
