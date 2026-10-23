package emailclient

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"receipt-wrangler/api/internal/logging"
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

func routedGroupSettings(id uint, emails []models.GroupSettingsWhiteListEmail, regexes []models.SubjectLineRegex) models.GroupSettings {
	settings := models.GroupSettings{EmailWhiteList: emails, SubjectLineRegexes: regexes}
	settings.ID = id
	return settings
}

func routedIds(groups []models.GroupSettings) []uint {
	ids := make([]uint, len(groups))
	for i, settings := range groups {
		ids[i] = settings.ID
	}
	return ids
}

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

// --- Parity with the Python client's utils tests --------------------------------

func TestRouter_SenderWhitelist(t *testing.T) {
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
		{"a whitelist of blank entries allows nobody", []string{"bob@example.com"}, whitelist("  "), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter([]models.GroupSettings{routedGroupSettings(1, tc.whitelist, nil)})
			if got := len(router.Route(tc.from, "Receipt")) == 1; got != tc.want {
				t.Errorf("Route(%v) matched = %v, want %v", tc.from, got, tc.want)
			}
		})
	}
}

func TestRouter_SubjectRegexes(t *testing.T) {
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
		{"every regex invalid matches nothing", "anything", subjectRegexes(`(unclosed`, `^(?=x)`), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter([]models.GroupSettings{routedGroupSettings(1, nil, tc.regexes)})
			if got := len(router.Route([]string{"bob@example.com"}, tc.subject)) == 1; got != tc.want {
				t.Errorf("Route(%q) matched = %v, want %v", tc.subject, got, tc.want)
			}
		})
	}
}

// --- Routing ----------------------------------------------------------------------

func TestRouter_GroupsKeepTheirOwnFilters(t *testing.T) {
	// The Python client unioned every group's whitelist and regexes before
	// fetching, so group 1's whitelist narrowed group 2, which accepts anyone.
	router := NewRouter([]models.GroupSettings{
		routedGroupSettings(1, whitelist("bob@example.com"), subjectRegexes(`^Receipt`)),
		routedGroupSettings(2, nil, nil),
	})

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
			if got := routedIds(router.Route([]string{tc.from}, tc.subject)); !slices.Equal(got, tc.want) {
				t.Errorf("Route() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRouter_InvalidRegexIsLoggedOnce(t *testing.T) {
	log := captureLog(t)
	router := NewRouter([]models.GroupSettings{
		routedGroupSettings(1, nil, subjectRegexes(`(unclosed`, `^Receipt`)),
		routedGroupSettings(2, nil, subjectRegexes(`^(?=x)`)),
	})

	for range 10 {
		router.Route([]string{"bob@example.com"}, "Receipt")
	}

	if got := strings.Count(log.String(), "Ignoring subject line regex"); got != 2 {
		t.Errorf("logged %d invalid regexes, want 2 (once each, however many messages):\n%s", got, log.String())
	}
}

func TestRouter_NoMatchIsEmptyNotNil(t *testing.T) {
	got := NewRouter(nil).Route([]string{"bob@example.com"}, "Receipt")
	if got == nil || len(got) != 0 {
		t.Errorf("Route() = %#v, want an empty slice", got)
	}
}
