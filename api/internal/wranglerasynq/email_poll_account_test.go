package wranglerasynq

import (
	"bytes"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/utils"
)

// captureLog sends the file logger's output to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	logger := logging.GetLogger()
	original := logger.Writer()
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)
	t.Cleanup(func() { logger.SetOutput(original) })
	return &buffer
}

func TestPollEmail_TwoRowsForOneAccountAreReadOnce(t *testing.T) {
	cases := []struct {
		name   string
		manual bool
	}{
		{name: "every group"},
		{name: "a manual poll of the first row's group", manual: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newPollFixture(t)
			first, mailbox := fixture.addSystemEmail("receipts@example.org")
			second, _ := fixture.addSystemEmail("Receipts@Example.org")
			onFirst := fixture.addGroup(first, groupOptions{})
			onSecond := fixture.addGroup(second, groupOptions{})
			uid := mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

			var groupIds []string
			if tc.manual {
				groupIds = []string{utils.UintToString(onFirst.GroupId)}
			}
			if err := PollEmail(!tc.manual, groupIds); err != nil {
				t.Fatal(err)
			}

			// Polled per row, the first row consumed the mail and the second row's
			// group never saw it.
			if len(fixture.connects) != 1 || fixture.connects[0].Username != first.Username {
				t.Errorf("connected as %+v, want once, with the lowest-id row", fixture.connects)
			}
			if len(fixture.payloadsFor(onFirst.ID)) != 1 || len(fixture.payloadsFor(onSecond.ID)) != 1 {
				t.Errorf("enqueued %+v, want one payload for each row's group", fixture.enqueued)
			}
			if !mailbox.isSeen(uid) {
				t.Error("the email was left unread")
			}
		})
	}
}

func TestPollEmail_UsesTheNextRowWhenOneCannotLogIn(t *testing.T) {
	fixture := newPollFixture(t)
	stale, mailbox := fixture.addSystemEmailWithPassword("receipts@example.org", "an old password")
	current, _ := fixture.addSystemEmail("receipts@example.org")
	onStale := fixture.addGroup(stale, groupOptions{})
	onCurrent := fixture.addGroup(current, groupOptions{})
	mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.connects) != 2 {
		t.Errorf("tried %d logins, want 2: the stale row, then the current one", len(fixture.connects))
	}
	if len(fixture.payloadsFor(onStale.ID)) != 1 || len(fixture.payloadsFor(onCurrent.ID)) != 1 {
		t.Errorf("enqueued %+v; a stale password on one row locked out the account", fixture.enqueued)
	}
}

func TestPollEmail_AccountsAreIsolated(t *testing.T) {
	// The second group accepts anyone. The Python client routed the first
	// account's mail to it anyway.
	fixture := newPollFixture(t)
	firstEmail, firstMailbox := fixture.addSystemEmail("first@example.org")
	secondEmail, _ := fixture.addSystemEmail("second@example.org")
	onFirst := fixture.addGroup(firstEmail, groupOptions{whitelist: []string{"bob@example.com"}})
	onSecond := fixture.addGroup(secondEmail, groupOptions{})
	firstMailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if len(fixture.payloadsFor(onFirst.ID)) != 1 || len(fixture.payloadsFor(onSecond.ID)) != 0 {
		t.Errorf("enqueued %+v, want the first account's mail for its own group only", fixture.enqueued)
	}
}

func TestPollEmail_InvalidRegexIsLoggedOncePerPoll(t *testing.T) {
	fixture := newPollFixture(t)
	log := captureLog(t)
	systemEmail, mailbox := fixture.addSystemEmail("receipts@example.org")
	broken := fixture.addGroup(systemEmail, groupOptions{regexes: []string{`^(?=Receipt)`}})
	open := fixture.addGroup(systemEmail, groupOptions{})
	for range 3 {
		mailbox.deliver(rawEmail("bob@example.com", "Receipt", "", jpegNamed("receipt.jpg", pollJpeg)))
	}

	if err := PollEmail(true, nil); err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(log.String(), "Ignoring subject line regex"); got != 1 {
		t.Errorf("logged the invalid regex %d times for 3 emails, want once per poll", got)
	}
	if len(fixture.payloadsFor(broken.ID)) != 0 || len(fixture.payloadsFor(open.ID)) != 3 {
		t.Errorf("enqueued %+v, want all 3 for the open group only", fixture.enqueued)
	}
}
