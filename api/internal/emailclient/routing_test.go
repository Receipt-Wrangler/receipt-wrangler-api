package emailclient

import (
	"slices"
	"testing"

	"receipt-wrangler/api/internal/models"
)

func whitelist(emails ...string) []models.GroupSettingsWhiteListEmail {
	result := make([]models.GroupSettingsWhiteListEmail, len(emails))
	for i, email := range emails {
		result[i] = models.GroupSettingsWhiteListEmail{Email: email}
	}
	return result
}

func subjectRegexes(regexes ...string) []models.SubjectLineRegex {
	result := make([]models.SubjectLineRegex, len(regexes))
	for i, regex := range regexes {
		result[i] = models.SubjectLineRegex{Regex: regex}
	}
	return result
}

func groupOnMailbox(id uint, systemEmailId uint, emails []models.GroupSettingsWhiteListEmail, regexes []models.SubjectLineRegex) models.GroupSettings {
	settings := models.GroupSettings{
		SystemEmailId:      &systemEmailId,
		EmailWhiteList:     emails,
		SubjectLineRegexes: regexes,
	}
	settings.ID = id
	return settings
}

// --- Parity with the Python client's utils tests --------------------------------

func TestSenderAllowed(t *testing.T) {
	cases := []struct {
		name      string
		from      []string
		whitelist []models.GroupSettingsWhiteListEmail
		want      bool
	}{
		{"empty whitelist allows anyone", []string{"test@example.com"}, nil, true},
		{"in whitelist", []string{"test@example.com"}, whitelist("test@example.com", "admin@example.com"), true},
		{"not in whitelist", []string{"test@example.com"}, whitelist("admin@example.com"), false},
		// New: the Python client compared exact strings.
		{"case-insensitive sender", []string{"Test@Example.COM"}, whitelist("test@example.com"), true},
		{"case-insensitive entry", []string{"test@example.com"}, whitelist("TEST@example.com"), true},
		{"entry stored with a display name", []string{"bob@example.com"}, whitelist("Bob <bob@example.com>"), true},
		{"no sender address", nil, whitelist("bob@example.com"), false},
		{"any of several addresses", []string{"x@example.com", "bob@example.com"}, whitelist("bob@example.com"), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SenderAllowed(tc.from, tc.whitelist); got != tc.want {
				t.Errorf("SenderAllowed(%v) = %v, want %v", tc.from, got, tc.want)
			}
		})
	}
}

func TestSubjectAllowed(t *testing.T) {
	cases := []struct {
		name    string
		subject string
		regexes []models.SubjectLineRegex
		want    bool
	}{
		{"no regexes allows any subject", "Subject", nil, true},
		{"match found", "Test subject", subjectRegexes(`^Test`), true},
		{"no match", "Test subject", subjectRegexes(`^Hello`), false},
		{"unanchored search like re.search", "Your Receipt #42", subjectRegexes(`Receipt #\d+`), true},
		{"any regex matching is enough", "Invoice", subjectRegexes(`^Receipt`, `Invoice`), true},
		// New: an invalid regex used to raise and fail the whole poll.
		{"python-only lookahead is ignored, not fatal", "Receipt", subjectRegexes(`^(?=Receipt)`), false},
		{"invalid regex skipped, valid one still matches", "Receipt", subjectRegexes(`(unclosed`, `^Receipt$`), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SubjectAllowed(tc.subject, tc.regexes); got != tc.want {
				t.Errorf("SubjectAllowed(%q) = %v, want %v", tc.subject, got, tc.want)
			}
		})
	}
}

// --- Routing ----------------------------------------------------------------------

func TestRoute_OnlyGroupsOnTheReceivingMailbox(t *testing.T) {
	// Group 2 reads a different mailbox and accepts anyone. The Python client
	// routed mailbox 1's message to it anyway.
	groups := []models.GroupSettings{
		groupOnMailbox(1, 10, whitelist("bob@example.com"), nil),
		groupOnMailbox(2, 20, nil, nil),
	}

	got := Route([]string{"bob@example.com"}, "Receipt", 10, groups)

	if !slices.Equal(got, []uint{1}) {
		t.Errorf("Route() = %v, want [1]", got)
	}
}

func TestRoute_SharedSenderReachesOnlyTheReceivingMailboxsGroup(t *testing.T) {
	groups := []models.GroupSettings{
		groupOnMailbox(1, 10, whitelist("bob@example.com"), nil),
		groupOnMailbox(2, 20, whitelist("bob@example.com"), nil),
	}

	if got := Route([]string{"bob@example.com"}, "Receipt", 20, groups); !slices.Equal(got, []uint{2}) {
		t.Errorf("Route() = %v, want [2]", got)
	}
}

func TestRoute_GroupsKeepTheirOwnFilters(t *testing.T) {
	// The Python client unioned every group's whitelist and regexes before
	// fetching, so group 1's whitelist narrowed group 2, which accepts anyone.
	groups := []models.GroupSettings{
		groupOnMailbox(1, 10, whitelist("bob@example.com"), subjectRegexes(`^Receipt`)),
		groupOnMailbox(2, 10, nil, nil),
	}

	cases := []struct {
		name    string
		from    string
		subject string
		want    []uint
	}{
		{"matches both", "bob@example.com", "Receipt 1", []uint{1, 2}},
		{"stranger reaches the open group only", "stranger@example.com", "Receipt 1", []uint{2}},
		{"other subject reaches the open group only", "bob@example.com", "Hello", []uint{2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Route([]string{tc.from}, tc.subject, 10, groups); !slices.Equal(got, tc.want) {
				t.Errorf("Route() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRoute_GroupWithoutAMailboxIsIgnored(t *testing.T) {
	settings := models.GroupSettings{}
	settings.ID = 1

	if got := Route([]string{"bob@example.com"}, "Receipt", 10, []models.GroupSettings{settings}); len(got) != 0 {
		t.Errorf("Route() = %v, want none", got)
	}
}

func TestRoute_NoMatchIsEmptyNotNil(t *testing.T) {
	got := Route([]string{"bob@example.com"}, "Receipt", 10, nil)
	if got == nil || len(got) != 0 {
		t.Errorf("Route() = %#v, want an empty slice", got)
	}
}
