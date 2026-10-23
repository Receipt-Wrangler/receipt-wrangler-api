package emailclient

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
)

// Router matches messages to the groups reading one IMAP account. Each group is
// judged by its own whitelist and subject regexes. The Python client pooled every
// mailbox's messages and matched them against every group, so a group with an
// empty whitelist received mail sent to any other group's inbox, and one group's
// filters narrowed another's.
type Router struct {
	groups []routedGroup
}

type routedGroup struct {
	settings models.GroupSettings
	// The filters flags record that the group configured a whitelist or regexes,
	// even if no entry survived: a group whose every regex is invalid matches
	// nothing rather than everything.
	filtersSender  bool
	senders        []string
	filtersSubject bool
	subjects       []*regexp.Regexp
}

// NewRouter prepares the groups for routing, once per poll. Each subject regex is
// compiled here, so a regex Go cannot compile (for example a lookahead saved
// before regexes were validated) is logged once rather than once per message. It
// is treated as not matching, so it can never fail a poll.
func NewRouter(groups []models.GroupSettings) Router {
	router := Router{groups: make([]routedGroup, 0, len(groups))}

	for _, settings := range groups {
		group := routedGroup{
			settings:       settings,
			filtersSender:  len(settings.EmailWhiteList) > 0,
			filtersSubject: len(settings.SubjectLineRegexes) > 0,
		}

		for _, entry := range settings.EmailWhiteList {
			if sender := normalizeAddress(entry.Email); len(sender) > 0 {
				group.senders = append(group.senders, sender)
			}
		}

		for _, subjectRegex := range settings.SubjectLineRegexes {
			compiled, err := regexp.Compile(subjectRegex.Regex)
			if err != nil {
				logging.LogStd(logging.LOG_LEVEL_ERROR, fmt.Sprintf(
					"Ignoring subject line regex %q on group settings %d: %s",
					subjectRegex.Regex, subjectRegex.GroupSettingsId, err.Error(),
				))
				continue
			}
			group.subjects = append(group.subjects, compiled)
		}

		router.groups = append(router.groups, group)
	}

	return router
}

// Route returns the groups a message belongs to, in the order they were given.
func (router Router) Route(fromAddresses []string, subject string) []models.GroupSettings {
	matched := make([]models.GroupSettings, 0)
	for _, group := range router.groups {
		if group.allowsSender(fromAddresses) && group.allowsSubject(subject) {
			matched = append(matched, group.settings)
		}
	}
	return matched
}

// allowsSender reports whether any of the sender addresses is on the whitelist.
// No whitelist allows every sender. Addresses compare case-insensitively, and a
// whitelist entry in "Name <address>" form is reduced to its address.
func (group routedGroup) allowsSender(fromAddresses []string) bool {
	if !group.filtersSender {
		return true
	}

	for _, allowed := range group.senders {
		for _, from := range fromAddresses {
			if normalizeAddress(from) == allowed {
				return true
			}
		}
	}

	return false
}

// allowsSubject reports whether the subject matches any of the regexes, searched
// unanchored like Python's re.search. No regexes allows every subject.
func (group routedGroup) allowsSubject(subject string) bool {
	if !group.filtersSubject {
		return true
	}

	for _, compiled := range group.subjects {
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
