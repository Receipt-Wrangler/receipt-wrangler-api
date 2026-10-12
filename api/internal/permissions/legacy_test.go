package permissions

import (
	"slices"
	"sort"
	"testing"
)

// sortedCopy returns a sorted copy of keys, leaving the input untouched.
func sortedCopy(keys []string) []string {
	out := append([]string(nil), keys...)
	sort.Strings(out)
	return out
}

// equalSet reports whether two key slices contain the same keys (order-independent).
func equalSet(a []string, b []string) bool {
	return slices.Equal(sortedCopy(a), sortedCopy(b))
}

// countScope returns how many registry descriptors belong to the given scope.
func countScope(scope Scope) int {
	count := 0
	for _, descriptor := range registry {
		if descriptor.Scope == scope {
			count++
		}
	}
	return count
}

func TestLegacyAppUserKeys(t *testing.T) {
	expected := []string{
		AppCategoriesCreate,
		AppTagsCreate,
		AppCustomFieldsCreate,
		AppCustomFieldsRead,
		AppGroupsCreate,
		AppApiKeysCreate,
		AppApiKeysRead,
		AppApiKeysUpdate,
		AppApiKeysDelete,
		AppNotificationsRead,
		AppNotificationsDelete,
		AppUserPreferencesRead,
		AppUserPreferencesUpdate,
		AppAccountRead,
		AppAccountUpdate,
		AppAccountDelete,
		AppReceiptsSearch,
	}

	got := LegacyAppUserKeys()
	if !equalSet(got, expected) {
		t.Helper()
		utilPrint(t, got, expected)
	}

	// Every key must be app-scoped.
	for _, key := range got {
		descriptor, ok := Get(key)
		if !ok || descriptor.Scope != ScopeApp {
			utilPrint(t, key, "an app-scope permission")
		}
	}
}

func TestLegacyAppAdminIncludesReadAnyApiKeys(t *testing.T) {
	// Legacy Admin is every app permission, so a newly added app permission such
	// as app.api-keys.read-any must flow into it automatically (upgrade keeps the
	// admin's "view all API keys" ability).
	if !slices.Contains(LegacyAppAdminKeys(), AppApiKeysReadAny) {
		utilPrint(t, "Legacy Admin missing "+AppApiKeysReadAny, "present")
	}
}

func TestLegacyAppUserExcludesReadAnyApiKeys(t *testing.T) {
	// Legacy User is a fixed subset and must NOT grant the privileged
	// "view all API keys" permission.
	if slices.Contains(LegacyAppUserKeys(), AppApiKeysReadAny) {
		utilPrint(t, "Legacy User contains "+AppApiKeysReadAny, "absent")
	}
}

func TestLegacyAppAdminIncludesCustomFieldsUpdate(t *testing.T) {
	// Legacy Admin is every app permission, so app.custom-fields.update flows into
	// it automatically — an upgrading install's admin can edit custom fields on the
	// next boot via the seeder's add-only reconciliation, with no data migration.
	if !slices.Contains(LegacyAppAdminKeys(), AppCustomFieldsUpdate) {
		utilPrint(t, "Legacy Admin missing "+AppCustomFieldsUpdate, "present")
	}
}

func TestLegacyAppUserExcludesCustomFieldsUpdate(t *testing.T) {
	// Legacy User is a fixed subset. It holds custom-fields create + read so a
	// normal user can add a field for their own receipts, but editing an existing
	// definition changes it for every group's receipts, so it stays admin-only —
	// the same shape as categories and tags, where create is granted and
	// update/delete are not.
	if slices.Contains(LegacyAppUserKeys(), AppCustomFieldsUpdate) {
		utilPrint(t, "Legacy User contains "+AppCustomFieldsUpdate, "absent")
	}
}

func TestLegacyAppAdminIncludesReportsRead(t *testing.T) {
	// Legacy Admin is every app permission, so the newly added app.reports.read
	// must flow into it automatically — an upgrading install's admin keeps access
	// to the report builder with zero manual reconciliation.
	if !slices.Contains(LegacyAppAdminKeys(), AppReportsRead) {
		utilPrint(t, "Legacy Admin missing "+AppReportsRead, "present")
	}
}

func TestLegacyAppUserExcludesReportsRead(t *testing.T) {
	// Legacy User is a fixed subset: reporting is admin-by-default, granted to
	// non-admins only via a custom role that includes app.reports.read.
	if slices.Contains(LegacyAppUserKeys(), AppReportsRead) {
		utilPrint(t, "Legacy User contains "+AppReportsRead, "absent")
	}
}

func TestLegacyAppAdminIncludesReportAllPerms(t *testing.T) {
	// The per-action report bypass permissions are app-scoped, so Legacy Admin auto-
	// grants them — an upgrading install's admin keeps system-wide report reach even
	// after the group-access ceiling tightens the six template handlers.
	for _, key := range []string{
		AppReportsReadAll, AppReportsCreateAll, AppReportsUpdateAll,
		AppReportsDeleteAll, AppReportsDuplicateAll, AppReportsGenerateAll,
	} {
		if !slices.Contains(LegacyAppAdminKeys(), key) {
			utilPrint(t, "Legacy Admin missing "+key, "present")
		}
	}
}

func TestLegacyAppUserExcludesReportAllPerms(t *testing.T) {
	// Legacy User is a fixed subset: the "all templates" bypass permissions are
	// admin-by-default and must never flow into it.
	for _, key := range []string{
		AppReportsReadAll, AppReportsCreateAll, AppReportsUpdateAll,
		AppReportsDeleteAll, AppReportsDuplicateAll, AppReportsGenerateAll,
	} {
		if slices.Contains(LegacyAppUserKeys(), key) {
			utilPrint(t, "Legacy User contains "+key, "absent")
		}
	}
}

func TestLegacyAppAdminIncludesDeleteAnyGroup(t *testing.T) {
	// Legacy Admin is every app permission, so app.groups.delete flows into it
	// automatically — an upgrading install's admin can clean up abandoned groups
	// with no manual reconciliation.
	if !slices.Contains(LegacyAppAdminKeys(), AppGroupsDelete) {
		utilPrint(t, "Legacy Admin missing "+AppGroupsDelete, "present")
	}
}

func TestLegacyAppUserExcludesDeleteAnyGroup(t *testing.T) {
	// Legacy User is a fixed subset: deleting a group the caller is not a member
	// of is an administrative capability and must never flow into it. A normal
	// user still deletes their own groups through the group-scoped group.delete.
	if slices.Contains(LegacyAppUserKeys(), AppGroupsDelete) {
		utilPrint(t, "Legacy User contains "+AppGroupsDelete, "absent")
	}
}

func TestLegacyAppAdminIncludesOidcProviderPermissions(t *testing.T) {
	// Legacy Admin is every app permission, so the four app.oidc-providers.* keys
	// flow into it automatically on the next boot via SeedSystemRoles' add-only
	// reconciliation — an upgrading install's admin can configure OIDC with no
	// manual role edit and no data migration.
	admin := LegacyAppAdminKeys()

	for _, key := range oidcProviderKeys() {
		if !slices.Contains(admin, key) {
			utilPrint(t, "Legacy Admin missing "+key, "present")
		}
	}
}

func TestLegacyAppUserDoesNotGetOidcProviderPermissions(t *testing.T) {
	// Legacy User is a fixed list. Registering an identity provider decides who
	// can sign in to the whole install and whether unknown users are provisioned
	// accounts, so it is administrative by definition and must never flow into
	// the ordinary-user set.
	user := LegacyAppUserKeys()

	for _, key := range oidcProviderKeys() {
		if slices.Contains(user, key) {
			utilPrint(t, "Legacy User contains "+key, "absent")
		}
	}
}

func oidcProviderKeys() []string {
	return []string{
		AppOidcProvidersCreate,
		AppOidcProvidersRead,
		AppOidcProvidersUpdate,
		AppOidcProvidersDelete,
	}
}

func TestLegacyAppUserExcludesUsersRead(t *testing.T) {
	// app.users.read gates only the admin "Manage Users" listing (GET /user/),
	// which no client calls. Legacy User must NOT hold it, so normal users don't
	// reach the admin Users page.
	if slices.Contains(LegacyAppUserKeys(), AppUsersRead) {
		utilPrint(t, "Legacy User contains "+AppUsersRead, "absent")
	}
}

func TestLegacyAppUserExcludesGlobalCategoryTagRead(t *testing.T) {
	// app.categories.read / app.tags.read gate the GLOBAL category/tag lists
	// (GET /category, GET /tag, and the flat AppData arrays). Legacy User must NOT
	// hold them — normal users receive only the per-group filtered catalogs, so
	// granting global read would leak categories/tags outside their grants.
	if slices.Contains(LegacyAppUserKeys(), AppCategoriesRead) {
		utilPrint(t, "Legacy User contains "+AppCategoriesRead, "absent")
	}
	if slices.Contains(LegacyAppUserKeys(), AppTagsRead) {
		utilPrint(t, "Legacy User contains "+AppTagsRead, "absent")
	}
}

func TestLegacyAppUserRetainsCategoryTagCreate(t *testing.T) {
	// The create permissions are retained so inline category/tag creation in the
	// receipt form still works (gated on create, per the lock-down rules).
	if !slices.Contains(LegacyAppUserKeys(), AppCategoriesCreate) {
		utilPrint(t, "Legacy User missing "+AppCategoriesCreate, "present")
	}
	if !slices.Contains(LegacyAppUserKeys(), AppTagsCreate) {
		utilPrint(t, "Legacy User missing "+AppTagsCreate, "present")
	}
}

func TestLegacyGroupViewerKeys(t *testing.T) {
	expected := []string{
		GroupView,
		GroupEmailPoll,
		GroupReceiptsRead,
		GroupReceiptsMagicFill,
		GroupCommentsCreate,
		GroupCommentsDelete,
		GroupDashboardsCreate,
		GroupDashboardsRead,
		GroupDashboardsUpdate,
		GroupDashboardsDelete,
		GroupWidgetsRead,
		GroupActivitiesRead,
	}

	got := LegacyGroupViewerKeys()
	if !equalSet(got, expected) {
		utilPrint(t, got, expected)
	}

	for _, key := range got {
		descriptor, ok := Get(key)
		if !ok || descriptor.Scope != ScopeGroup {
			utilPrint(t, key, "a group-scope permission")
		}
	}
}

func TestLegacyGroupEditorKeysSupersetOfViewer(t *testing.T) {
	viewer := LegacyGroupViewerKeys()
	editor := LegacyGroupEditorKeys()

	// Editor is a strict superset of viewer.
	for _, key := range viewer {
		if !slices.Contains(editor, key) {
			utilPrint(t, "editor missing viewer key "+key, "present")
		}
	}

	editorOnly := []string{
		GroupReceiptsCreate,
		GroupReceiptsUpdate,
		GroupReceiptsDelete,
		GroupReceiptsDuplicate,
		GroupReceiptsQuickScan,
		GroupActivitiesRerun,
	}
	for _, key := range editorOnly {
		if !slices.Contains(editor, key) {
			utilPrint(t, "editor missing key "+key, "present")
		}
	}

	if len(editor) != len(viewer)+len(editorOnly) {
		utilPrint(t, len(editor), len(viewer)+len(editorOnly))
	}
}

func TestLegacyAppAdminKeysAreAllAppScope(t *testing.T) {
	got := LegacyAppAdminKeys()

	if len(got) != countScope(ScopeApp) {
		utilPrint(t, len(got), countScope(ScopeApp))
	}

	for _, key := range got {
		descriptor, ok := Get(key)
		if !ok || descriptor.Scope != ScopeApp {
			utilPrint(t, key, "an app-scope permission")
		}
	}
}

func TestLegacyGroupOwnerKeysAreAllGroupScope(t *testing.T) {
	got := LegacyGroupOwnerKeys()

	if len(got) != countScope(ScopeGroup) {
		utilPrint(t, len(got), countScope(ScopeGroup))
	}

	for _, key := range got {
		descriptor, ok := Get(key)
		if !ok || descriptor.Scope != ScopeGroup {
			utilPrint(t, key, "a group-scope permission")
		}
	}

	// Owner is a superset of editor.
	for _, key := range LegacyGroupEditorKeys() {
		if !slices.Contains(got, key) {
			utilPrint(t, "owner missing editor key "+key, "present")
		}
	}
}

func TestLegacyKeysAllExistInRegistry(t *testing.T) {
	groups := [][]string{
		LegacyAppAdminKeys(),
		LegacyAppUserKeys(),
		LegacyGroupViewerKeys(),
		LegacyGroupEditorKeys(),
		LegacyGroupOwnerKeys(),
	}

	for _, keys := range groups {
		for _, key := range keys {
			if !Exists(key) {
				utilPrint(t, key, "a key that exists in the registry")
			}
		}
	}
}

func TestLegacyGroupEditorDoesNotMutateViewer(t *testing.T) {
	// Build the editor set (which appends onto a viewer slice internally), then
	// confirm the viewer helper is unaffected.
	_ = LegacyGroupEditorKeys()

	viewer := LegacyGroupViewerKeys()
	if len(viewer) != 12 {
		utilPrint(t, len(viewer), 12)
	}
}

// utilPrint mirrors utils.PrintTestError without importing utils (which would
// create an import cycle for the permissions package's test).
func utilPrint(t *testing.T, actual any, expected any) {
	t.Helper()
	t.Errorf("Test failed!\nExpected: %v\nActual: %v\n", expected, actual)
}
