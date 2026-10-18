package wranglerasynq

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"receipt-wrangler/api/internal/models"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func TestPollEmail_QueuesEachAttachmentForEachGroup(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	withoutBody := fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})
	withBody := fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	uid := mailbox.deliver(rawEmail("Bob <bob@example.com>", "Receipt 42", "Thanks for your order", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.connects) != 1 || fixture.connects[0].Password != pollTestPassword || fixture.connects[0].Host != "imap.example.org" {
		t.Errorf("connects = %+v, want one with the decrypted password", fixture.connects)
	}
	if !mailbox.isSeen(uid) {
		t.Error("a processed email was left unread")
	}
	if len(fixture.enqueued) != 2 {
		t.Fatalf("enqueued %d payloads, want 2", len(fixture.enqueued))
	}

	for _, payload := range fixture.enqueued {
		if !bytes.Equal(readFile(t, payload.TempFilePath), pollJpeg) || !bytes.Equal(readFile(t, payload.ImageForOcrPath), pollJpeg) {
			t.Error("the temp files do not hold the attachment")
		}
		if filepath.Base(payload.ImageForOcrPath) != "image-"+filepath.Base(payload.TempFilePath) {
			t.Errorf("ImageForOcrPath = %s, want the image- copy of %s", payload.ImageForOcrPath, payload.TempFilePath)
		}
		if payload.Attachment.Filename != "receipt.jpg" || payload.Attachment.FileType != "image/jpeg" || payload.Attachment.Size != uint(len(pollJpeg)) {
			t.Errorf("Attachment = %+v", payload.Attachment)
		}

		metadata := payload.Metadata
		if metadata.Subject != "Receipt 42" || metadata.FromEmail != "bob@example.com" || metadata.FromName != "Bob" || metadata.To != "receipts@example.org" {
			t.Errorf("Metadata = %+v", metadata)
		}
		if !metadata.Date.Equal(time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)) {
			t.Errorf("Date = %v, want the instant in UTC", metadata.Date)
		}
		if !slices.Equal(metadata.GroupSettingsIds, []uint{withoutBody.ID, withBody.ID}) {
			t.Errorf("GroupSettingsIds = %v", metadata.GroupSettingsIds)
		}
		if len(metadata.Attachments) != 1 || metadata.Attachments[0] != payload.Attachment {
			t.Errorf("Metadata.Attachments = %+v", metadata.Attachments)
		}
		if payload.RenderBodyPdf {
			t.Error("RenderBodyPdf = true for an email without an HTML body")
		}
	}

	if body := fixture.payloadsFor(withoutBody.ID)[0].Metadata.Body; body != "" {
		t.Errorf("a group without body processing got the body %q", body)
	}
	if body := fixture.payloadsFor(withBody.ID)[0].Metadata.Body; body != "Thanks for your order" {
		t.Errorf("a group with body processing got the body %q", body)
	}
}

func TestPollEmail_SameAttachmentNamesNeverShareATempFile(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})

	first := append(append([]byte{}, pollJpeg...), []byte("first email")...)
	second := append(append([]byte{}, pollJpeg...), []byte("second email")...)
	logoA := append(append([]byte{}, pollJpeg...), []byte("outlook inline A")...)
	logoB := append(append([]byte{}, pollJpeg...), []byte("outlook inline B")...)

	mailbox.deliver(rawEmail("a@example.com", "One", "", jpegNamed("receipt.jpg", first)))
	mailbox.deliver(rawEmail("b@example.com", "Two", "", jpegNamed("receipt.jpg", second)))
	mailbox.deliver(rawEmail("c@example.com", "Three", "", jpegNamed("image001.jpg", logoA), jpegNamed("image001.jpg", logoB)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 4 {
		t.Fatalf("enqueued %d payloads, want 4", len(fixture.enqueued))
	}

	want := map[string][]byte{"One": first, "Two": second}
	paths := make(map[string]bool)
	logos := make([][]byte, 0)
	for _, payload := range fixture.enqueued {
		if paths[payload.TempFilePath] {
			t.Errorf("temp path %s is shared by two attachments", payload.TempFilePath)
		}
		paths[payload.TempFilePath] = true

		data := readFile(t, payload.TempFilePath)
		if expected, ok := want[payload.Metadata.Subject]; ok {
			if !bytes.Equal(data, expected) {
				t.Errorf("email %q got another email's attachment", payload.Metadata.Subject)
			}
		} else {
			logos = append(logos, data)
		}
	}

	if len(logos) != 2 || bytes.Equal(logos[0], logos[1]) {
		t.Error("two same-named attachments in one email collapsed into one file")
	}
}

func TestPollEmail_BodyOnlyEmailGoesToBodyProcessingGroups(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	withoutBody := fixture.addGroup(systemEmail, groupOptions{})
	withBody := fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Order", "<p>Your order total is <b>$45.00</b></p>"))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.payloadsFor(withoutBody.ID)) != 0 {
		t.Error("a group without body processing got a body-only email")
	}
	payloads := fixture.payloadsFor(withBody.ID)
	if len(payloads) != 1 {
		t.Fatalf("payloads = %d, want 1", len(payloads))
	}
	payload := payloads[0]
	if payload.TempFilePath != "" || payload.ImageForOcrPath != "" {
		t.Errorf("a body-only payload has files: %+v", payload)
	}
	if payload.Metadata.Body != "Your order total is $45.00" || payload.Metadata.BodyHtml != "<p>Your order total is <b>$45.00</b></p>" {
		t.Errorf("Metadata = %+v", payload.Metadata)
	}
	if !payload.RenderBodyPdf {
		t.Error("RenderBodyPdf = false for an HTML body with text")
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
}

func TestPollEmail_NonMatchingEmailStaysUnreadAndIsNotDownloaded(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}, regexes: []string{"^Receipt"}})
	stranger := mailbox.deliver(rawEmail("stranger@example.com", "Receipt", "body", jpegNamed("receipt.jpg", pollJpeg)))
	otherSubject := mailbox.deliver(rawEmail("bob@example.com", "Lunch on Friday?", "body"))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if mailbox.isSeen(stranger) || mailbox.isSeen(otherSubject) {
		t.Error("an email matching no group was marked read")
	}
	if len(mailbox.rawFetches) != 0 {
		t.Errorf("downloaded %v; only headers should be fetched for non-matching mail", mailbox.rawFetches)
	}
	if len(fixture.enqueued) != 0 {
		t.Errorf("enqueued %d payloads", len(fixture.enqueued))
	}
}

func TestPollEmail_WhitelistIsCaseInsensitive(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})
	mailbox.deliver(rawEmail("Bob <Bob@Example.COM>", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 1 {
		t.Errorf("enqueued %d payloads, want 1", len(fixture.enqueued))
	}
}

func TestPollEmail_CorruptAttachmentFailsAloneAndLaterEmailsAreProcessed(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	settings := fixture.addGroup(systemEmail, groupOptions{})

	broken := mailbox.deliver(rawEmail("bob@example.com", "Broken", "", fixtureAttachment{filename: "receipt.pdf", contentType: "application/pdf", data: corruptPdf}))
	partial := mailbox.deliver(rawEmail("bob@example.com", "Partial", "", fixtureAttachment{filename: "bad.pdf", contentType: "application/pdf", data: corruptPdf}, jpegNamed("good.jpg", pollJpeg)))
	later := mailbox.deliver(rawEmail("bob@example.com", "Later", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	for _, uid := range []uint32{broken, partial, later} {
		if !mailbox.isSeen(uid) {
			t.Errorf("email %d was left unread", uid)
		}
	}

	subjects := make([]string, 0)
	for _, payload := range fixture.enqueued {
		subjects = append(subjects, payload.Metadata.Subject+"/"+payload.Attachment.Filename)
	}
	if !slices.Equal(subjects, []string{"Partial/good.jpg", "Later/receipt.jpg"}) {
		t.Errorf("enqueued %v; one bad attachment must not stop the rest", subjects)
	}

	emailReads := fixture.systemTasks(models.EMAIL_READ)
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(emailReads) != 2 || len(uploads) != 2 {
		t.Fatalf("EMAIL_READ = %d, EMAIL_UPLOAD = %d, want one failure each per broken attachment", len(emailReads), len(uploads))
	}
	for i, task := range emailReads {
		if task.Status != models.SYSTEM_TASK_FAILED || task.AssociatedEntityType != models.SYSTEM_EMAIL || task.AssociatedEntityId != systemEmail.ID {
			t.Errorf("EMAIL_READ = %+v", task)
		}
		upload := uploads[i]
		if upload.Status != models.SYSTEM_TASK_FAILED || upload.GroupId == nil || *upload.GroupId != settings.GroupId ||
			upload.AssociatedSystemTaskId == nil || *upload.AssociatedSystemTaskId != task.ID || upload.RanByUserId != nil {
			t.Errorf("EMAIL_UPLOAD = %+v", upload)
		}
	}
	if !strings.Contains(emailReads[0].ResultDescription, "receipt.pdf") || !strings.Contains(emailReads[0].ResultDescription, "Broken") {
		t.Errorf("ResultDescription = %q, want the attachment and subject named", emailReads[0].ResultDescription)
	}
}

func TestPollEmail_UnreadableEmailIsRecordedAndMarkedRead(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})

	// A second From header: the classic way to ride a forged sender alongside a
	// genuine DKIM signature.
	forged := append([]byte("From: ceo@victim.example\r\n"), rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg))...)
	uid := mailbox.deliver(forged)

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
	if len(fixture.enqueued) != 0 {
		t.Errorf("enqueued %d payloads", len(fixture.enqueued))
	}
	if tasks := fixture.systemTasks(models.EMAIL_READ); len(tasks) != 1 || !strings.Contains(tasks[0].ResultDescription, "more than one sender") {
		t.Errorf("EMAIL_READ = %+v", tasks)
	}
}

func TestPollEmail_OneFailingMailboxDoesNotStopTheOthers(t *testing.T) {
	fixture := newPollFixture(t)
	brokenEmail, brokenMailbox := fixture.addSystemEmail("broken@example.org")
	workingEmail, workingMailbox := fixture.addSystemEmail("working@example.org")
	fixture.addGroup(brokenEmail, groupOptions{})
	fixture.addGroup(workingEmail, groupOptions{})
	brokenMailbox.connectErr = errors.New("authentication failed")
	uid := workingMailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Errorf("PollEmail() = %v, want nil when only one mailbox failed", err)
	}
	if len(fixture.enqueued) != 1 || !workingMailbox.isSeen(uid) {
		t.Error("the working mailbox was not processed")
	}
}

func TestPollEmail_FailsWhenEveryMailboxFails(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("broken@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	mailbox.connectErr = errors.New("authentication failed")

	if err := PollEmail(true, nil); err == nil {
		t.Error("PollEmail() = nil, want an error when no mailbox could be read")
	}
}

func TestPollEmail_ManualPollRoutesOtherGroupsMailOnTheSameMailbox(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("shared@example.org")
	requested := fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"alice@example.com"}})
	other := fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	// A manual poll for the first group only.
	if err := PollEmail(false, []string{"1"}); err != nil {
		t.Fatal(err)
	}

	if requested.GroupId != 1 {
		t.Fatalf("precondition: requested group id = %d", requested.GroupId)
	}
	if len(fixture.payloadsFor(other.ID)) != 1 || len(fixture.payloadsFor(requested.ID)) != 0 {
		t.Errorf("enqueued %+v; the other group's email must reach the other group", fixture.enqueued)
	}
	if !mailbox.isSeen(uid) {
		t.Error("the routed email was left unread")
	}
}

func TestPollEmail_ManualPollOnlyReadsTheRequestedGroupsMailboxes(t *testing.T) {
	fixture := newPollFixture(t)
	requestedEmail, _ := fixture.addSystemEmail("requested@example.org")
	otherEmail, otherMailbox := fixture.addSystemEmail("other@example.org")
	fixture.addGroup(requestedEmail, groupOptions{})
	fixture.addGroup(otherEmail, groupOptions{})
	otherMailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(false, []string{"1"}); err != nil {
		t.Fatal(err)
	}

	if len(fixture.connects) != 1 || fixture.connects[0].Username != "requested@example.org" {
		t.Errorf("connects = %+v, want only the requested group's mailbox", fixture.connects)
	}
}

func TestPollEmail_DisabledGroupsAreNeitherPolledNorRoutedTo(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{disabled: true})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.connects) != 0 || mailbox.isSeen(uid) {
		t.Error("a mailbox only used by a disabled group was read")
	}
}

func TestPollEmail_SignedSender(t *testing.T) {
	cases := []struct {
		name           string
		sign           bool
		dnsDown        bool
		wantSeen       bool
		wantEnqueued   int
		wantFailedRead int
	}{
		{name: "signed by the sender's domain", sign: true, wantSeen: true, wantEnqueued: 1},
		{name: "unsigned", wantSeen: true, wantFailedRead: 1},
		{name: "dns temporarily down", sign: true, dnsDown: true, wantSeen: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
			fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
			fixture.publishKey("example.com")
			if tc.dnsDown {
				fixture.dnsTemporarilyDown()
			}

			raw := rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg))
			if tc.sign {
				raw = signEmail(t, raw, "example.com")
			}
			uid := mailbox.deliver(raw)

			if err := PollEmail(true, nil); err != nil {
				t.Fatal(err)
			}

			if mailbox.isSeen(uid) != tc.wantSeen {
				t.Errorf("seen = %v, want %v", mailbox.isSeen(uid), tc.wantSeen)
			}
			if len(fixture.enqueued) != tc.wantEnqueued {
				t.Errorf("enqueued %d, want %d", len(fixture.enqueued), tc.wantEnqueued)
			}
			failures := fixture.systemTasks(models.EMAIL_READ)
			if len(failures) != tc.wantFailedRead {
				t.Errorf("failed EMAIL_READ = %d, want %d", len(failures), tc.wantFailedRead)
			}
			if tc.wantFailedRead > 0 && !strings.Contains(failures[0].ResultDescription, "sender could not be verified") {
				t.Errorf("ResultDescription = %q", failures[0].ResultDescription)
			}
		})
	}
}

func TestPollEmail_SignedSenderCheckOnlyAppliesToGroupsRequiringIt(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	requiring := fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
	notRequiring := fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.payloadsFor(requiring.ID)) != 0 || len(fixture.payloadsFor(notRequiring.ID)) != 1 {
		t.Errorf("enqueued %+v; only the group not requiring a signature should get it", fixture.enqueued)
	}
	if fixture.payloadsFor(notRequiring.ID)[0].Metadata.GroupSettingsIds[0] != notRequiring.ID {
		t.Error("Metadata.GroupSettingsIds still lists the rejecting group")
	}
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(uploads) != 1 || *uploads[0].GroupId != requiring.GroupId {
		t.Errorf("failed EMAIL_UPLOAD = %+v, want one for the requiring group only", uploads)
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
}

func TestPollEmail_OversizedEmailIsRecordedWithoutDownloading(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Huge", "", jpegNamed("receipt.jpg", pollJpeg)))
	mailbox.messages[uid].size = maxEmailMessageBytes + 1

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(mailbox.rawFetches) != 0 {
		t.Error("an oversized email was downloaded")
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
	if tasks := fixture.systemTasks(models.EMAIL_READ); len(tasks) != 1 || !strings.Contains(tasks[0].ResultDescription, "limit") {
		t.Errorf("EMAIL_READ = %+v", tasks)
	}
}

func TestPollEmail_QueueFailureLeavesTheEmailUnread(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))
	fixture.enqueueErr = errors.New("redis unavailable")

	if err := PollEmail(true, nil); err == nil {
		t.Error("PollEmail() = nil, want the queue error")
	}
	if mailbox.isSeen(uid) {
		t.Error("an email that was never queued was marked read")
	}
	if len(fixture.systemTasks(models.EMAIL_READ)) != 0 {
		t.Error("a queue outage was recorded as a problem with the email")
	}
}

func TestPollEmail_ProcessesAtMostFiftyMatchingEmailsPerPoll(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}, bodyProcessing: true})

	// Unrelated unread mail first: it must not use up the allowance.
	for range 5 {
		mailbox.deliver(rawEmail("stranger@example.com", "Hello", "not a receipt"))
	}
	matching := make([]uint32, 0)
	for range maxMatchedMessagesPerMailboxPoll + 3 {
		matching = append(matching, mailbox.deliver(rawEmail("bob@example.com", "Receipt", "Total $1")))
	}

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.enqueued) != maxMatchedMessagesPerMailboxPoll {
		t.Errorf("enqueued %d, want %d", len(fixture.enqueued), maxMatchedMessagesPerMailboxPoll)
	}
	unread, _ := mailbox.UnseenUIDs()
	if len(unread) != 5+3 || !slices.Contains(unread, matching[len(matching)-1]) {
		t.Errorf("unread = %v, want the 5 unrelated and the last 3 matching", unread)
	}
}

func TestPollEmail_SkipsAMailboxAlreadyBeingPolled(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	lock := mailboxPollLock(systemEmail.ID)
	lock.Lock()
	err := PollEmail(true, nil)
	lock.Unlock()

	if err != nil {
		t.Errorf("PollEmail() = %v, want nil: a busy mailbox is not a failure", err)
	}
	if len(fixture.connects) != 0 || len(fixture.enqueued) != 0 {
		t.Error("a mailbox already being polled was polled again")
	}
}

func TestPollEmail_EmailWithNothingToProcessIsMarkedRead(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Notes", "", fixtureAttachment{filename: "notes.txt", contentType: "text/plain", data: []byte("not a receipt")}))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if !mailbox.isSeen(uid) || len(fixture.enqueued) != 0 || len(fixture.systemTasks(models.EMAIL_READ)) != 0 {
		t.Error("an email with nothing to process should be marked read silently, as before")
	}
}

func TestPollEmail_FailedAttachmentsDoNotFallBackToTheBody(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	mailbox.deliver(rawEmail("bob@example.com", "Receipt", "See the attached receipt", fixtureAttachment{filename: "receipt.pdf", contentType: "application/pdf", data: corruptPdf}))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.enqueued) != 0 {
		t.Errorf("enqueued %+v; a receipt without its image must not be created from the body", fixture.enqueued)
	}
}

func TestPollEmail_NoEnabledGroupsDoesNothing(t *testing.T) {
	fixture := newPollFixture(t)

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.connects) != 0 {
		t.Error("connected without any enabled group")
	}
}
