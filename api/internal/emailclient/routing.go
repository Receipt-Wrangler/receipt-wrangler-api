package emailclient

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
)

// Route returns the ids of the group settings a message belongs to.
//
// Only groups reading the mailbox the message arrived in are considered, and
// each is judged by its own whitelist and subject regexes. The Python client
// pooled every mailbox's messages and matched them against every group, so a
// group with an empty whitelist received mail sent to any other group's inbox,
// and one group's filters narrowed another's.
func Route(fromAddresses []string, subject string, systemEmailId uint, groupSettings []models.GroupSettings) []uint {
	groupSettingsIds := make([]uint, 0)

	for _, settings := range groupSettings {
		if settings.SystemEmailId == nil || *settings.SystemEmailId != systemEmailId {
			continue
		}

		if SenderAllowed(fromAddresses, settings.EmailWhiteList) && SubjectAllowed(subject, settings.SubjectLineRegexes) {
			groupSettingsIds = append(groupSettingsIds, settings.ID)
		}
	}

	return groupSettingsIds
}

// SenderAllowed reports whether any of the sender addresses is on the whitelist.
// An empty whitelist allows every sender. Addresses compare case-insensitively,
// and a whitelist entry in "Name <address>" form is reduced to its address.
func SenderAllowed(fromAddresses []string, whitelist []models.GroupSettingsWhiteListEmail) bool {
	if len(whitelist) == 0 {
		return true
	}

	for _, entry := range whitelist {
		allowed := normalizeAddress(entry.Email)
		if len(allowed) == 0 {
			continue
		}

		for _, from := range fromAddresses {
			if normalizeAddress(from) == allowed {
				return true
			}
		}
	}

	return false
}

// SubjectAllowed reports whether the subject matches any of the regexes, searched
// unanchored like Python's re.search. No regexes allows every subject. A regex Go
// cannot compile (for example a lookahead saved before regexes were validated) is
// logged and treated as not matching, so it can never fail a poll.
func SubjectAllowed(subject string, regexes []models.SubjectLineRegex) bool {
	if len(regexes) == 0 {
		return true
	}

	for _, subjectRegex := range regexes {
		compiled, err := regexp.Compile(subjectRegex.Regex)
		if err != nil {
			logging.LogStd(logging.LOG_LEVEL_ERROR, fmt.Sprintf(
				"Ignoring subject line regex %q on group settings %d: %s",
				subjectRegex.Regex, subjectRegex.GroupSettingsId, err.Error(),
			))
			continue
		}

		if compiled.MatchString(subject) {
			return true
		}
	}

	return false
}

func normalizeAddress(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := mail.ParseAddress(value); err == nil {
		value = parsed.Address
	}
	return strings.ToLower(value)
}
