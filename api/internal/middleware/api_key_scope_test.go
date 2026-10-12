package middleware

import (
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/utils"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// scopeRequest drives one request through the middleware with an API key.
func scopeRequest(t *testing.T, key string, method string, path string) int {
	t.Helper()

	r := httptest.NewRequest(method, path, nil)
	r.Header.Set("Authorization", key)
	w := httptest.NewRecorder()

	UnifiedAuthMiddleware(createFakeHandler()).ServeHTTP(w, r)

	return w.Result().StatusCode
}

// routeCase is one row of the inventory: a real route, and what it does.
type routeCase struct {
	method string
	path   string
	class  requestClass
}

// scopePermits states the whole rule as a matrix rather than deriving it, so a
// change to apiKeyScopePermits cannot quietly change what the table expects of
// it. Read across: "w" is literal, and neither single scope reaches an endpoint
// that both discloses and changes.
var scopePermits = map[string]map[requestClass]bool{
	"r":  {classRead: true, classWrite: false, classReadWrite: false},
	"w":  {classRead: false, classWrite: true, classReadWrite: false},
	"rw": {classRead: true, classWrite: true, classReadWrite: true},
}

func className(class requestClass) string {
	switch class {
	case classRead:
		return "read"
	case classWrite:
		return "write"
	case classReadWrite:
		return "read-write"
	default:
		return "unknown"
	}
}

// everyRoute is the whole authenticated surface, classified.
//
// This table is what makes the allowlists maintainable. A route added without
// thought about its scope shows up here as a missing row, and the two sections
// at the bottom -- writes that disclose, and writes with read-ish names -- are
// the ones a reviewer would otherwise wave through.
var everyRoute = []routeCase{
	// Reads over GET -- the ordinary case.
	{http.MethodGet, "/api/receipt/1", classRead},
	{http.MethodGet, "/api/group/1", classRead},
	{http.MethodGet, "/api/category/", classRead},
	{http.MethodGet, "/api/search/", classRead},
	{http.MethodGet, "/api/report/template/1", classRead},
	{http.MethodGet, "/api/systemTask/1/sourceFile", classRead},

	// Reads over POST -- the allowlist.
	{http.MethodPost, "/api/apiKey/paged", classRead},
	{http.MethodPost, "/api/category/getPagedCategories", classRead},
	{http.MethodPost, "/api/customField/getPagedCustomFields", classRead},
	{http.MethodPost, "/api/group/getPagedGroups", classRead},
	{http.MethodPost, "/api/prompt/getPagedPrompts", classRead},
	{http.MethodPost, "/api/tag/getPagedTags", classRead},
	{http.MethodPost, "/api/user/getPagedUsers", classRead},
	{http.MethodPost, "/api/receiptProcessingSettings/getPagedProcessingSettings", classRead},
	{http.MethodPost, "/api/systemEmail/getSystemEmails", classRead},
	{http.MethodPost, "/api/systemTask/getPagedSystemTasks", classRead},
	{http.MethodPost, "/api/systemTask/getPagedActivities", classRead},
	{http.MethodPost, "/api/report/template/list", classRead},
	{http.MethodPost, "/api/report/generate", classRead},
	{http.MethodPost, "/api/report/preview", classRead},
	{http.MethodPost, "/api/report/receipts", classRead},
	{http.MethodPost, "/api/receiptImage/convertToJpg", classRead},
	{http.MethodPost, "/api/export/", classRead},
	{http.MethodPost, "/api/export/7", classRead},
	{http.MethodPost, "/api/receipt/group/7", classRead},
	{http.MethodPost, "/api/receipt/group/7/summary", classRead},
	{http.MethodPost, "/api/widget/pieChart/7", classRead},
	{http.MethodPost, "/api/report/template/7/generate", classRead},
	{http.MethodPost, "/api/report/template/7/render", classRead},

	// Writes.
	{http.MethodPost, "/api/receipt/", classWrite},
	{http.MethodPut, "/api/receipt/1", classWrite},
	{http.MethodDelete, "/api/receipt/1", classWrite},
	{http.MethodPost, "/api/receipt/quickScan", classWrite},
	// Returns the receipt it was just handed, so it discloses nothing the caller
	// did not supply -- unlike the duplicate endpoints below.
	{http.MethodPost, "/api/receipt/withFiles", classWrite},
	{http.MethodPost, "/api/receipt/bulkStatusUpdate", classWrite},
	{http.MethodPost, "/api/apiKey/", classWrite},
	{http.MethodPut, "/api/apiKey/1", classWrite},
	{http.MethodDelete, "/api/apiKey/1", classWrite},
	{http.MethodPost, "/api/category/", classWrite},
	{http.MethodPost, "/api/group/", classWrite},
	{http.MethodPut, "/api/group/1", classWrite},
	{http.MethodPost, "/api/report/template", classWrite},
	{http.MethodPut, "/api/report/template/1", classWrite},
	{http.MethodDelete, "/api/report/template/1", classWrite},
	{http.MethodPost, "/api/user/1/resetPassword", classWrite},
	{http.MethodDelete, "/api/user/bulk", classWrite},
	{http.MethodPost, "/api/import/importConfigJson", classWrite},
	{http.MethodPost, "/api/systemSettings/restartTaskServer", classWrite},

	// Writes that DISCLOSE. Both take nothing but an id and answer with the
	// source's contents, so a write-only key could read through them.
	{http.MethodPost, "/api/receipt/1/duplicate", classReadWrite},
	{http.MethodPost, "/api/report/template/1/duplicate", classReadWrite},

	// Writes that LOOK like reads. Each persists a row or enqueues a task while
	// carrying a read-ish name, and two of them carry a ".read" permission.
	{http.MethodPost, "/api/systemEmail/checkConnectivity", classWrite},
	{http.MethodPost, "/api/receiptProcessingSettings/checkConnectivity", classWrite},
	{http.MethodPost, "/api/prompt/createDefaultPrompt", classWrite},
	{http.MethodPost, "/api/group/1/pollGroupEmail", classWrite},
	{http.MethodPost, "/api/systemTask/rerunActivity/1", classWrite},
}

// TestApiKeyScopeIsEnforcedAcrossEveryRoute is the regression test for the
// finding: the scope was stored, validated and then never consulted, so a key
// created as "Read" could delete anything its owner could.
func TestApiKeyScopeIsEnforcedAcrossEveryRoute(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()

	keys := map[string]string{}
	for scope := range scopePermits {
		_, key, err := createTestApiKey(user.ID, scope)
		if err != nil {
			t.Fatalf("failed to create a %q key: %v", scope, err)
		}

		keys[scope] = key
	}

	for _, route := range everyRoute {
		for scope, key := range keys {
			want := http.StatusForbidden
			if scopePermits[scope][route.class] {
				want = http.StatusOK
			}

			got := scopeRequest(t, key, route.method, route.path)
			if got != want {
				t.Errorf(
					"scope %q on %s %s (%s): got %d, want %d",
					scope, route.method, route.path, className(route.class), got, want,
				)
			}
		}
	}
}

// TestReadOnlyPostPathMatching pins the matcher itself, where a near-miss would
// either open a write to a read-only key or break a legitimate read.
func TestReadOnlyPostPathMatching(t *testing.T) {
	cases := []struct {
		path string
		read bool
	}{
		{"/api/receipt/group/7", true},
		{"/api/receipt/group/7/summary", true},
		{"/api/receipt/group/7/", true}, // trailing slash
		{"/api/receipt/group", false},   // one segment short
		{"/api/receipt/group/", false},  // empty parameter
		{"/api/receipt/group//", false}, // ditto
		{"/api/receipt/group/7/other", false},
		// "group" is literal, not a parameter. Both duplicate routes are
		// read-AND-write, which classifyRequest settles before this matcher runs.
		{"/api/receipt/1/duplicate", false},
		{"/api/export", true},
		{"/api/export/", true},
		{"/api/export/7", true},
		{"/api/export/7/extra", false},
		{"/api/report/template/list", true},
		{"/api/report/template", false},   // create
		{"/api/report/template/7", false}, // update is a PUT; a POST here is not a read
		{"/api/report/template/7/duplicate", false},
		{"/api/report/generate", true},
		{"/api/report/generatex", false}, // a prefix of an allowlisted path is not one
		{"/api/apiKey/paged", true},
		{"/api/apiKey/pagedx", false},
		{"", false},
		{"/", false},
	}

	for _, c := range cases {
		if got := isReadOnlyPostPath(c.path); got != c.read {
			t.Errorf("isReadOnlyPostPath(%q) = %v, want %v", c.path, got, c.read)
		}
	}
}

// TestApiKeyScopePermitsIsLiteral covers the rule directly, including the
// fail-closed behaviour for a scope that should be unreachable.
func TestApiKeyScopePermitsIsLiteral(t *testing.T) {
	for scope, expected := range scopePermits {
		for _, class := range []requestClass{classRead, classWrite, classReadWrite} {
			if got := apiKeyScopePermits(scope, class); got != expected[class] {
				t.Errorf(
					"apiKeyScopePermits(%q, %s) = %v, want %v",
					scope, className(class), got, expected[class],
				)
			}
		}
	}

	// Neither the command validator nor Claims.Validate lets these through, so
	// reaching here with one is a bug -- and a bug must not widen access.
	for _, scope := range []string{"", "x", "R", "W"} {
		for _, class := range []requestClass{classRead, classWrite, classReadWrite} {
			if apiKeyScopePermits(scope, class) {
				t.Errorf("scope %q permitted a %s request", scope, className(class))
			}
		}
	}
}

// TestApiKeyScopeStillStampsLastUsedOnARead: the last-used write is
// authentication bookkeeping, not the caller's operation, so a read-only key must
// still record that it was used. It is the one write a scope check must not gate,
// and it lives in the same function as the gate.
func TestApiKeyScopeStillStampsLastUsedOnARead(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-key")
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()

	dbApiKey, key, err := createTestApiKey(user.ID, "r")
	if err != nil {
		t.Fatalf("failed to create the key: %v", err)
	}

	if status := scopeRequest(t, key, http.MethodGet, "/api/receipt/1"); status != http.StatusOK {
		t.Fatalf("expected the read to be allowed, got %d", status)
	}

	// The stamp is written from a goroutine, so give it a moment to land.
	repository := repositories.NewApiKeyRepository(nil)
	for i := 0; i < 50; i++ {
		refreshed, err := repository.GetApiKeyById(dbApiKey.ID)
		if err == nil && refreshed.LastUsedAt != nil {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	utils.PrintTestError(t, "last_used_at never set", "a read-only key must still stamp its own last use")
}

// TestApiKeyScopeDoesNotAffectJwtCallers: the scope belongs to an API key. A
// session token carries none, and gating on an empty one would log every browser
// user out.
func TestApiKeyScopeDoesNotAffectJwtCallers(t *testing.T) {
	defer teardownAuthTest()
	setupAuthTest()

	user := createTestUser()

	jwt, _, _, err := services.GenerateJWT(user.ID)
	if err != nil {
		t.Fatalf("failed to mint a jwt: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/api/receipt/", nil)
	r.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	UnifiedAuthMiddleware(createFakeHandler()).ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("a jwt write must not be gated by API key scope, got %d", w.Result().StatusCode)
	}
}

// TestMountedRouterPreservesUrlPathForScopeMatching pins the assumption the whole
// allowlist rests on.
//
// The matcher reads r.URL.Path, and every router is reached through
// chi's Mount at an "/api/..." prefix. If Mount rewrote URL.Path to be relative
// to the mount point — as some routers do — no allowlist entry would ever match
// and every read-over-POST would 403 for a read-only key.
//
// The tests above drive the middleware directly with a path they choose
// themselves, so they cannot catch that. This one goes through a real chi mount.
//
// (chi tracks routing position in RouteContext.RoutePath and leaves URL.Path
// alone, which is also why the matcher does NOT use RoutePattern(): a mux runs
// its middleware chain before it routes, so the pattern is not populated yet.)
func TestMountedRouterPreservesUrlPathForScopeMatching(t *testing.T) {
	var seen string

	capture := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = r.URL.Path
			next.ServeHTTP(w, r)
		})
	}

	receiptRouter := chi.NewRouter()
	receiptRouter.Use(capture)
	receiptRouter.Post("/group/{groupId}/summary", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	root := chi.NewRouter()
	root.Mount("/api/receipt", receiptRouter)

	request := httptest.NewRequest(http.MethodPost, "/api/receipt/group/7/summary", nil)
	root.ServeHTTP(httptest.NewRecorder(), request)

	if seen != "/api/receipt/group/7/summary" {
		t.Fatalf("middleware saw URL.Path %q, want the full mounted path", seen)
	}

	if !isReadOnlyPostPath(seen) {
		t.Errorf("the path a mounted middleware actually sees (%q) does not match the allowlist", seen)
	}
}

// TestClassifyRequest pins the classifier, and with it the one assumption the
// method-based model makes: that a write does not disclose.
//
// The duplicate endpoints break that assumption -- they take nothing but an id
// and answer with the source's contents -- so they must NOT fall through to the
// plain-write branch, where a write-only key would reach them and read through
// them.
func TestClassifyRequest(t *testing.T) {
	cases := []struct {
		method string
		path   string
		class  requestClass
	}{
		{http.MethodGet, "/api/receipt/1", classRead},
		{http.MethodHead, "/api/receipt/1", classRead},
		{http.MethodOptions, "/api/receipt/1", classRead},
		{http.MethodPost, "/api/receipt/group/7", classRead},
		{http.MethodPost, "/api/receipt/", classWrite},
		{http.MethodPut, "/api/receipt/1", classWrite},
		{http.MethodPatch, "/api/receipt/1", classWrite},
		{http.MethodDelete, "/api/receipt/1", classWrite},

		{http.MethodPost, "/api/receipt/1/duplicate", classReadWrite},
		{http.MethodPost, "/api/receipt/1/duplicate/", classReadWrite},
		{http.MethodPost, "/api/report/template/7/duplicate", classReadWrite},

		// Near misses, which must stay plain writes rather than widen the rule.
		{http.MethodPost, "/api/receipt/duplicate", classWrite},
		{http.MethodPost, "/api/receipt//duplicate", classWrite},
		{http.MethodPost, "/api/receipt/1/duplicatex", classWrite},
		{http.MethodPost, "/api/receipt/1/duplicate/extra", classWrite},
		{http.MethodPost, "/api/report/template/duplicate", classWrite},

		// The method still decides: only POST consults the pattern lists.
		{http.MethodPut, "/api/receipt/1/duplicate", classWrite},
		{http.MethodGet, "/api/receipt/1/duplicate", classRead},
	}

	for _, c := range cases {
		got := classifyRequest(httptest.NewRequest(c.method, c.path, nil))
		if got != c.class {
			t.Errorf(
				"classifyRequest(%s %s) = %s, want %s",
				c.method, c.path, className(got), className(c.class),
			)
		}
	}
}
