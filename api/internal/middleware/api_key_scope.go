package middleware

import (
	"net/http"
	"strings"
)

// API key scopes are finally enforced here.
//
// models.ApiKeyScope ("r" / "w" / "rw") has always been stored on the key, copied
// into the JWT claims and validated for well-formedness -- and then consulted by
// nothing. Both IsValid() calls check the VALUE is one of the three, never that
// the request is allowed, and enforcePermissions resolves everything from the
// caller's user id without ever reading the scope. So a key created as "Read"
// could create, update and delete anything its owner could.
//
// That is not a dormant field: the desktop form presents a required
// Read / Write / Read-Write picker and DEFAULTS to Read, and swagger calls it
// "Scope/permissions for API keys". The product promised a restriction it did not
// apply, and the most restrictive label was the default.
//
// The scopes are LITERAL: "r" denies writes, "w" denies reads, "rw" allows both.

// requestClass is what a request does to the caller, which is not always what it
// does to the database.
//
// "Read" here means DISCLOSURE, not "touched a row". The gate operates on HTTP
// requests, never on queries: UpdateReceipt internally reads the receipt,
// re-resolves permissions and merges hidden associations, and none of that is a
// read in this sense because none of it reaches the caller. A write-scoped key
// may PUT all day and the server reads whatever it needs.
//
// The corollary is the one easy to miss: a WRITE that returns data is a read
// too. See classReadWrite.
type requestClass int

const (
	classRead requestClass = iota
	classWrite
	classReadWrite
)

// classifyRequest decides what a request discloses and what it changes.
//
// The HTTP method decides, which fails closed: a write endpoint added later is
// denied to an "r" key with nobody having to remember. The opposite mistake -- a
// new read-over-POST wrongly denied -- surfaces immediately as a broken call
// rather than sitting there as a silent hole.
//
// The allowlists below are the exceptions, and the read one is not small: this
// API serves 23 reads over POST, because PagedRequestCommand and friends read
// their filter from a JSON body.
func classifyRequest(r *http.Request) requestClass {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return classRead
	case http.MethodPost:
		// Read-and-write is checked first: both lists are POST paths, and an
		// endpoint that discloses must never be classified by the weaker rule.
		if matchesAny(r.URL.Path, readWritePostPatterns) {
			return classReadWrite
		}

		if isReadOnlyPostPath(r.URL.Path) {
			return classRead
		}

		return classWrite
	default:
		return classWrite
	}
}

// apiKeyScopePermits applies the scope to a request already classified.
//
// An unrecognized or empty scope permits nothing -- UpsertApiKeyCommand and
// Claims.Validate both reject anything outside the three values, so reaching here
// with something else is a bug, and a bug must not widen access.
func apiKeyScopePermits(scope string, class requestClass) bool {
	reads := strings.Contains(scope, "r")
	writes := strings.Contains(scope, "w")

	switch class {
	case classRead:
		return reads
	case classWrite:
		return writes
	case classReadWrite:
		return reads && writes
	default:
		return false
	}
}

// readWritePostPatterns are endpoints that BOTH disclose and change, so only a
// key holding read and write may call them.
//
// Both duplicate endpoints take nothing but an id, supply no content of their
// own, and return the SOURCE's contents in the response. Classified as plain
// writes -- which is what the method says -- a write-only key could read any
// receipt or report template it may duplicate simply by duplicating it and
// reading the answer. That defeats the one thing write-without-read exists for:
// a key that feeds data in and cannot pull data out.
//
// This is the assumption the method-based model hides: it takes for granted that
// a write does not disclose. That is an assumption, not a guarantee, so an
// endpoint that breaks it is named here rather than forced into a binary that
// does not fit.
var readWritePostPatterns = [][]string{
	{"api", "receipt", "*", "duplicate"},
	{"api", "report", "template", "*", "duplicate"},
}

// readOnlyPostPaths are the endpoints that READ but answer over POST, because
// their filter arrives in the request body.
//
// Kept beside the matcher rather than near the routes so the whole set reads as
// one unit -- the list is the security boundary, and a reviewer needs to see it
// whole to judge it.
var readOnlyPostPaths = map[string]bool{
	// Paged / filtered lists. All of these carry a PagedRequestCommand.
	"/api/apiKey/paged":                                         true,
	"/api/category/getPagedCategories":                          true,
	"/api/customField/getPagedCustomFields":                     true,
	"/api/group/getPagedGroups":                                 true,
	"/api/prompt/getPagedPrompts":                               true,
	"/api/tag/getPagedTags":                                     true,
	"/api/user/getPagedUsers":                                   true,
	"/api/receiptProcessingSettings/getPagedProcessingSettings": true,
	"/api/systemEmail/getSystemEmails":                          true,
	"/api/systemTask/getPagedSystemTasks":                       true,
	"/api/systemTask/getPagedActivities":                        true,
	"/api/report/template/list":                                 true,

	// Reports. Verified non-persisting: report_service.go makes no write call and
	// HtmlToPdfService.Render touches no file.
	"/api/report/generate": true,
	"/api/report/preview":  true,
	"/api/report/receipts": true,

	// Stateless image utility -- no database, no disk.
	"/api/receiptImage/convertToJpg": true,
}

// readOnlyPostPatterns are the same, for routes carrying a path parameter.
// "*" matches exactly one segment.
var readOnlyPostPatterns = [][]string{
	{"api", "export"},                              // POST / on the export router
	{"api", "export", "*"},                         // by group
	{"api", "receipt", "group", "*"},               // the paged receipts list
	{"api", "receipt", "group", "*", "summary"},    // the totals under it
	{"api", "widget", "pieChart", "*"},             // dashboard pie data
	{"api", "report", "template", "*", "generate"}, // runs a stored config
	{"api", "report", "template", "*", "render"},   // ditto, for the dashboard widget
}

// Deliberately NOT allowlisted, because they look like reads and are not:
//
//   - /api/systemEmail/checkConnectivity and
//     /api/receiptProcessingSettings/checkConnectivity both PERSIST a SystemTask
//     row (services/system_email.go, services/ai.go) while carrying a ".read"
//     permission. They are the clearest writes-dressed-as-reads in the codebase.
//   - /api/prompt/createDefaultPrompt inserts a row.
//   - /api/group/{groupId}/pollGroupEmail and /api/systemTask/rerunActivity/{id}
//     enqueue asynq tasks.
//
// This is also why the declared permission strings are not a usable signal for
// read-vs-write: ".read" verbs sit on persisting endpoints, and
// app.reports.generate sits on a pure read. Deriving from them would fail OPEN.

func isReadOnlyPostPath(path string) bool {
	trimmed := strings.Trim(path, "/")
	if len(trimmed) == 0 {
		return false
	}

	if readOnlyPostPaths["/"+trimmed] {
		return true
	}

	return matchesAny(path, readOnlyPostPatterns)
}

// matchesAny reports whether a path matches any of the segment patterns.
func matchesAny(path string, patterns [][]string) bool {
	trimmed := strings.Trim(path, "/")
	if len(trimmed) == 0 {
		return false
	}

	segments := strings.Split(trimmed, "/")
	for _, pattern := range patterns {
		if segmentsMatch(segments, pattern) {
			return true
		}
	}

	return false
}

func segmentsMatch(segments []string, pattern []string) bool {
	if len(segments) != len(pattern) {
		return false
	}

	for i, want := range pattern {
		if want == "*" {
			// A parameter, but never an empty one -- "/api/export//" must not read
			// as "/api/export/{groupId}".
			if len(segments[i]) == 0 {
				return false
			}

			continue
		}

		if segments[i] != want {
			return false
		}
	}

	return true
}
