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

	"receipt-wrangler/api/internal/emailclient"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
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
		if !bytes.Equal(readFile(t, payload.TempFilePath), pollJpeg) {
			t.Error("the temp file does not hold the attachment")
		}
		// The task makes the OCR copy; the poll only names it.
		if filepath.Base(payload.ImageForOcrPath) != "image-"+filepath.Base(payload.TempFilePath) {
			t.Errorf("ImageForOcrPath = %s, want the image- copy of %s", payload.ImageForOcrPath, payload.TempFilePath)
		}
		if _, err := os.Stat(payload.ImageForOcrPath); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("the poll wrote the OCR copy (stat: %v); converting is the task's job", err)
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

func TestPollEmail_AttachmentsAreQueuedUnconverted(t *testing.T) {
	// A file ImageMagick cannot convert used to fail here: the original was
	// deleted and the email marked read, so the image was gone for good. Now it is
	// queued like any other, and the task records the failure with the image kept.
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})

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
		if len(readFile(t, payload.TempFilePath)) == 0 {
			t.Errorf("%s/%s was queued without its original", payload.Metadata.Subject, payload.Attachment.Filename)
		}
	}
	if !slices.Equal(subjects, []string{"Broken/receipt.pdf", "Partial/bad.pdf", "Partial/good.jpg", "Later/receipt.jpg"}) {
		t.Errorf("enqueued %v, want every attachment", subjects)
	}
	if failures := fixture.systemTasks(models.EMAIL_READ); len(failures) != 0 {
		t.Errorf("recorded %d failures at poll time, want none: conversion happens in the task", len(failures))
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
			// The likeliest cause on Exchange and Office 365 is the mailbox, so the
			// failure says which setting to turn off.
			if tc.wantFailedRead > 0 && !strings.Contains(failures[0].ResultDescription, `turn off "Require Signed Sender (DKIM)"`) {
				t.Errorf("ResultDescription = %q, want it to name the setting", failures[0].ResultDescription)
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

	// Once the queue is back, the email goes through exactly once.
	fixture.enqueueErr = nil
	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 1 || !mailbox.isSeen(uid) {
		t.Errorf("after the outage: enqueued %d, seen %v; want 1 and true", len(fixture.enqueued), mailbox.isSeen(uid))
	}
}

func TestPollEmail_AnEmailIsClaimedBeforeAnythingIsQueuedOrRecorded(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
	notRequiring := fixture.addGroup(systemEmail, groupOptions{})
	// Unsigned: queued for one group, a failed activity for the other.
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))
	mailbox.markSeenErr = errors.New("connection reset")

	if err := PollEmail(true, nil); err == nil {
		t.Error("PollEmail() = nil, want the mark error")
	}
	if len(fixture.enqueued) != 0 || len(fixture.systemTasks(models.EMAIL_READ)) != 0 {
		t.Fatalf("enqueued %d and recorded %d failures for an email that could not be claimed",
			len(fixture.enqueued), len(fixture.systemTasks(models.EMAIL_READ)))
	}

	mailbox.markSeenErr = nil
	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 1 || len(fixture.payloadsFor(notRequiring.ID)) != 1 {
		t.Errorf("enqueued %+v, want exactly one task for the group not requiring a signature", fixture.enqueued)
	}
	if failures := fixture.systemTasks(models.EMAIL_READ); len(failures) != 1 {
		t.Errorf("recorded %d failed EMAIL_READ across the two polls, want exactly 1", len(failures))
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
}

func TestPollEmail_PartlyQueuedEmailIsRecordedAndNeverQueuedTwice(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	first := fixture.addGroup(systemEmail, groupOptions{})
	second := fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))
	fixture.failEnqueueCall = 2

	if err := PollEmail(true, nil); err == nil {
		t.Error("PollEmail() = nil, want the queue error")
	}
	if len(fixture.payloadsFor(first.ID)) != 1 || len(fixture.payloadsFor(second.ID)) != 0 {
		t.Fatalf("enqueued %+v, want the first group's task only", fixture.enqueued)
	}
	if !mailbox.isSeen(uid) {
		t.Error("a partly queued email went back to unread; the next poll would queue the first task again")
	}
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(uploads) != 1 || *uploads[0].GroupId != second.GroupId || !strings.Contains(uploads[0].ResultDescription, "could not queue") {
		t.Errorf("failed EMAIL_UPLOAD = %+v, want one for the group whose task was not queued", uploads)
	}

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 1 {
		t.Errorf("enqueued %d after a second poll, want the first task only once", len(fixture.enqueued))
	}
}

func TestPollEmail_UnqueuedEmailThatCannotGoBackToUnreadIsRecorded(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	settings := fixture.addGroup(systemEmail, groupOptions{})
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))
	fixture.failEnqueueCall = 1
	mailbox.markUnseenErr = errors.New("connection reset")

	if err := PollEmail(true, nil); err == nil {
		t.Error("PollEmail() = nil, want the queue error")
	}
	if !slices.Equal(mailbox.unmarked, []uint32{uid}) {
		t.Errorf("MarkUnseen calls = %v, want one attempt for %d", mailbox.unmarked, uid)
	}
	if !mailbox.isSeen(uid) {
		t.Fatal("precondition: the email should still be read")
	}
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(uploads) != 1 || *uploads[0].GroupId != settings.GroupId {
		t.Errorf("failed EMAIL_UPLOAD = %+v; an email left read and unqueued must not vanish silently", uploads)
	}
}

func TestPollEmail_PerMessageFetchFailures(t *testing.T) {
	cases := []struct {
		name        string
		err         error
		wantSeen    bool
		wantFailure bool
		wantPollErr bool
	}{
		{name: "vanished", err: &emailclient.MessageFetchError{UID: 1, Gone: true, Err: errors.New("the server did not return it")}},
		{name: "temporarily unavailable", err: &emailclient.MessageFetchError{UID: 1, Temporary: true, Err: errors.New("NO [UNAVAILABLE]")}},
		{name: "refused", err: &emailclient.MessageFetchError{UID: 1, Err: errors.New("NO message is damaged")}, wantSeen: true, wantFailure: true},
		{name: "connection lost", err: errors.New("connection reset"), wantPollErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
			fixture.addGroup(systemEmail, groupOptions{})
			bad := mailbox.deliver(rawEmail("bob@example.com", "Bad", "", jpegNamed("bad.jpg", pollJpeg)))
			good := mailbox.deliver(rawEmail("bob@example.com", "Good", "", jpegNamed("good.jpg", pollJpeg)))
			mailbox.fetchRawErrs = map[uint32]error{bad: tc.err}

			err := PollEmail(true, nil)
			if (err != nil) != tc.wantPollErr {
				t.Fatalf("PollEmail() = %v, want an error: %v", err, tc.wantPollErr)
			}

			if mailbox.isSeen(bad) != tc.wantSeen {
				t.Errorf("failing email seen = %v, want %v", mailbox.isSeen(bad), tc.wantSeen)
			}
			failures := fixture.systemTasks(models.EMAIL_READ)
			if (len(failures) == 1) != tc.wantFailure || (tc.wantFailure && !strings.Contains(failures[0].ResultDescription, "would not return it")) {
				t.Errorf("failed EMAIL_READ = %+v, want one: %v", failures, tc.wantFailure)
			}

			// One bad message must not hold back the next, except when the session
			// itself is gone.
			if tc.wantPollErr {
				if len(fixture.enqueued) != 0 {
					t.Errorf("enqueued %d after the session failed", len(fixture.enqueued))
				}
				return
			}
			if len(fixture.enqueued) != 1 || fixture.enqueued[0].Metadata.Subject != "Good" || !mailbox.isSeen(good) {
				t.Errorf("enqueued %+v; the email after the failing one was not processed", fixture.enqueued)
			}
		})
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
	// The lock belongs to the account, so a poll through another row for the same
	// account holds it too.
	otherRow, _ := fixture.addSystemEmail("Receipts@Example.org")

	lock := mailboxPollLock(accountKeyOf(otherRow))
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

func TestPollEmail_UnconvertibleAttachmentIsQueuedNotTheBody(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	mailbox.deliver(rawEmail("bob@example.com", "Receipt", "See the attached receipt", fixtureAttachment{filename: "receipt.pdf", contentType: "application/pdf", data: corruptPdf}))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	// A body-only payload would become a receipt without the image it was sent for.
	if len(fixture.enqueued) != 1 || fixture.enqueued[0].Attachment.Filename != "receipt.pdf" || len(fixture.enqueued[0].TempFilePath) == 0 {
		t.Errorf("enqueued %+v, want the attachment and no body-only payload", fixture.enqueued)
	}
}

func TestPollEmail_AttachmentThatCannotBeSavedIsRecordedNotReplacedByTheBody(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	settings := fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	// A file where temp/ belongs makes every attachment write fail.
	if err := os.WriteFile(repositories.NewFileRepository(nil).GetTempDirectoryPath(), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "See the attached receipt", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatalf("PollEmail() = %v; one email's attachment must not abandon the mailbox", err)
	}

	if len(fixture.enqueued) != 0 {
		t.Errorf("enqueued %+v; a receipt without its image must not be created from the body", fixture.enqueued)
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
	uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
	if len(uploads) != 1 || *uploads[0].GroupId != settings.GroupId || !strings.Contains(uploads[0].ResultDescription, `"receipt.jpg" could not be processed`) {
		t.Errorf("failed EMAIL_UPLOAD = %+v, want one naming the attachment", uploads)
	}
}

func TestPollEmail_RejectedAttachmentDoesNotFallBackToTheBody(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{bodyProcessing: true})
	mislabelled := fixtureAttachment{filename: "receipt.pdf", contentType: "application/pdf", data: []byte("<html>not a pdf</html>")}
	uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "Total $5.00", mislabelled))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.enqueued) != 0 {
		t.Errorf("enqueued %+v; a receipt made from the body alone would be missing its image", fixture.enqueued)
	}
	failures := fixture.systemTasks(models.EMAIL_READ)
	if len(failures) != 1 || !strings.Contains(failures[0].ResultDescription, "receipt.pdf") ||
		!strings.Contains(failures[0].ResultDescription, "not a supported image or PDF") {
		t.Errorf("failed EMAIL_READ = %+v, want one naming receipt.pdf", failures)
	}
	if !mailbox.isSeen(uid) {
		t.Error("the email was left unread")
	}
}

func TestPollEmail_RejectedAttachmentBesideAGoodOne(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{})
	mislabelled := fixtureAttachment{filename: "receipt.pdf", contentType: "application/pdf", data: []byte("<html>not a pdf</html>")}
	mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", mislabelled, jpegNamed("photo.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.enqueued) != 1 || fixture.enqueued[0].Attachment.Filename != "photo.jpg" {
		t.Errorf("enqueued %+v, want photo.jpg only", fixture.enqueued)
	}
	if failures := fixture.systemTasks(models.EMAIL_READ); len(failures) != 1 || !strings.Contains(failures[0].ResultDescription, "receipt.pdf") {
		t.Errorf("failed EMAIL_READ = %+v, want one naming receipt.pdf", failures)
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
