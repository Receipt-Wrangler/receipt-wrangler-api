package wranglerasynq

import (
	"errors"
	"maps"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"receipt-wrangler/api/internal/emailclient"
	"receipt-wrangler/api/internal/models"
)

func TestPollEmail_HeadersOfNonMatchingMailAreFetchedOnce(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})
	first := mailbox.deliver(rawEmail("stranger@example.com", "Hello", "not a receipt"))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	second := mailbox.deliver(rawEmail("stranger@example.com", "Hello again", "not a receipt"))
	for range 2 {
		if err := PollEmail(true, nil); err != nil {
			t.Fatal(err)
		}
	}

	if got := mailbox.fetchedHeadersOf(); !slices.Equal(got, []uint32{first, second}) {
		t.Errorf("fetched headers of %v over three polls, want each non-matching email once: %v", got, []uint32{first, second})
	}
	if mailbox.isSeen(first) || mailbox.isSeen(second) {
		t.Error("a remembered non-matching email was marked read")
	}
}

func TestPollEmail_RescansRememberedMail(t *testing.T) {
	cases := []struct {
		name   string
		change func(fixture *pollFixture, mailbox *fakeMailbox, settings models.GroupSettings, uid uint32)
	}{
		{
			name: "the whitelist is edited",
			change: func(fixture *pollFixture, _ *fakeMailbox, settings models.GroupSettings, _ uint32) {
				entry := models.GroupSettingsWhiteListEmail{GroupSettingsId: settings.ID, Email: "alice@example.com"}
				if err := fixture.db.Create(&entry).Error; err != nil {
					fixture.t.Fatal(err)
				}
			},
		},
		{
			// The mailbox was rebuilt, and the UID now names a different message.
			name: "UIDVALIDITY changes",
			change: func(_ *pollFixture, mailbox *fakeMailbox, _ models.GroupSettings, uid uint32) {
				mailbox.uidValidity++
				mailbox.messages[uid].raw = rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
			settings := fixture.addGroup(systemEmail, groupOptions{whitelist: []string{"bob@example.com"}})
			uid := mailbox.deliver(rawEmail("alice@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

			if err := PollEmail(true, nil); err != nil {
				t.Fatal(err)
			}
			tc.change(fixture, mailbox, settings, uid)
			if err := PollEmail(true, nil); err != nil {
				t.Fatal(err)
			}

			if len(fixture.enqueued) != 1 || !mailbox.isSeen(uid) {
				t.Errorf("enqueued %d; a remembered non-match was not looked at again after the change", len(fixture.enqueued))
			}
		})
	}
}

func TestPollEmail_RetriesAreBackedOff(t *testing.T) {
	cases := []struct {
		name string
		fail func(fixture *pollFixture, mailbox *fakeMailbox, uid uint32)
	}{
		{
			name: "the sender check could not be done",
			fail: func(fixture *pollFixture, _ *fakeMailbox, _ uint32) { fixture.dnsTemporarilyDown() },
		},
		{
			name: "the server could not return the email",
			fail: func(_ *pollFixture, mailbox *fakeMailbox, uid uint32) {
				mailbox.fetchRawErrs = map[uint32]error{uid: &emailclient.MessageFetchError{UID: uid, Temporary: true, Err: errors.New("NO [UNAVAILABLE]")}}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
			fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
			uid := mailbox.deliver(signEmail(t, rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)), "example.com"))
			tc.fail(fixture, mailbox, uid)

			downloadsAfter := func(wait time.Duration) int {
				t.Helper()
				fixture.advance(wait)
				if err := PollEmail(true, nil); err != nil {
					t.Fatal(err)
				}
				return len(mailbox.rawFetches)
			}

			if got := downloadsAfter(0); got != 1 {
				t.Fatalf("downloaded %d times on the first poll, want 1", got)
			}
			if got := downloadsAfter(time.Minute); got != 1 {
				t.Errorf("downloaded again a minute later (%d times); it should wait out %s", got, retryBackoff)
			}
			if got := downloadsAfter(retryBackoff); got != 2 {
				t.Errorf("downloaded %d times once %s had passed, want it tried again", got, retryBackoff)
			}
			if len(fixture.enqueued) != 0 || mailbox.isSeen(uid) {
				t.Error("an email that could not be checked was processed or marked read")
			}
		})
	}
}

func TestPollEmail_UncheckableSendersDoNotStarveTheMailbox(t *testing.T) {
	fixture := newPollFixture(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
	fixture.publishKey("example.com")
	published := fixture.lookupTXT
	fixture.lookupTXT = func(name string) ([]string, error) {
		if strings.HasSuffix(name, ".slow.example") {
			return nil, &net.DNSError{Err: "server misbehaving", Name: name, IsTemporary: true}
		}
		return published(name)
	}

	slow := make([]uint32, 0, maxMatchedMessagesPerMailboxPoll)
	for range maxMatchedMessagesPerMailboxPoll {
		raw := rawEmail("shop@slow.example", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg))
		slow = append(slow, mailbox.deliver(signEmail(t, raw, "slow.example")))
	}
	good := mailbox.deliver(signEmail(t, rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)), "example.com"))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}
	if len(fixture.enqueued) != 0 {
		t.Fatalf("precondition: the uncheckable emails should fill the first poll, enqueued %d", len(fixture.enqueued))
	}
	downloadedBefore := len(mailbox.rawFetches)

	fixture.advance(time.Minute)
	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.enqueued) != 1 || !mailbox.isSeen(good) {
		t.Errorf("enqueued %d; the good email was starved by %d that could not be checked", len(fixture.enqueued), len(slow))
	}
	if got := mailbox.rawFetches[downloadedBefore:]; !slices.Equal(got, []uint32{good}) {
		t.Errorf("the second poll downloaded %v, want only the good email", got)
	}
}

func TestPollEmail_SenderCheckGivesUpAfter24Hours(t *testing.T) {
	cases := []struct {
		name     string
		age      time.Duration
		wantGive bool
	}{
		{name: "23 hours old: still waiting", age: 23 * time.Hour},
		{name: "25 hours old: gives up", age: 25 * time.Hour, wantGive: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
			requiring := fixture.addGroup(systemEmail, groupOptions{requireSigned: true})
			notRequiring := fixture.addGroup(systemEmail, groupOptions{})
			fixture.dnsTemporarilyDown()
			raw := signEmail(t, rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)), "example.com")
			uid := mailbox.deliverAt(raw, fixture.clock.Add(-tc.age))

			if err := PollEmail(true, nil); err != nil {
				t.Fatal(err)
			}

			uploads := fixture.systemTasks(models.EMAIL_UPLOAD)
			if !tc.wantGive {
				// Waiting holds back the group that does not require the check too.
				if mailbox.isSeen(uid) || len(fixture.enqueued) != 0 || len(uploads) != 0 {
					t.Errorf("seen=%v enqueued=%d failures=%d; the email should still be waiting", mailbox.isSeen(uid), len(fixture.enqueued), len(uploads))
				}
				return
			}

			if !mailbox.isSeen(uid) {
				t.Error("the email was left unread after the sender check gave up")
			}
			if len(fixture.payloadsFor(requiring.ID)) != 0 || len(fixture.payloadsFor(notRequiring.ID)) != 1 {
				t.Errorf("enqueued %+v, want only the group not requiring the check", fixture.enqueued)
			}
			if len(uploads) != 1 || *uploads[0].GroupId != requiring.GroupId || !strings.Contains(uploads[0].ResultDescription, "24 hours") {
				t.Errorf("failed EMAIL_UPLOAD = %+v, want one for the requiring group naming 24 hours", uploads)
			}
		})
	}
}

func TestMailboxMemory_ForgetsUIDsNoLongerUnread(t *testing.T) {
	memory := &mailboxMemory{}
	memory.refresh(1, "routing")
	memory.unmatched[1] = struct{}{}
	memory.unmatched[2] = struct{}{}
	memory.deferredUntil[3] = time.Now()
	memory.deferredUntil[4] = time.Now()

	memory.forgetAllBut([]uint32{2, 4, 5})

	if got := slices.Sorted(maps.Keys(memory.unmatched)); !slices.Equal(got, []uint32{2}) {
		t.Errorf("unmatched = %v, want [2]", got)
	}
	if got := slices.Sorted(maps.Keys(memory.deferredUntil)); !slices.Equal(got, []uint32{4}) {
		t.Errorf("deferredUntil = %v, want [4]", got)
	}
}
