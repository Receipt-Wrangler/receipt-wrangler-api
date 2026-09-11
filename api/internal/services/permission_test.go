package services

import (
	"receipt-wrangler/api/internal/commands"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"slices"
	"sort"
	"testing"
)

// seedUserWithAppRole creates an app role granting perms and a user assigned to
// it, returning the user id and role id.
func seedUserWithAppRole(t *testing.T, username string, perms []string) (uint, uint) {
	t.Helper()
	db := repositories.GetDB()

	roleRepository := repositories.NewRoleRepository(nil)
	role, err := roleRepository.CreateAppRole("App Role", "", perms, false)
	if err != nil {
		t.Fatalf("seed app role: %v", err)
	}

	user := models.User{Username: username, Password: "password", AppRoleID: &role.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return user.ID, role.ID
}

// seedMemberWithGroupRole creates a group, a group role granting perms, and a
// group member assigned to it, returning the user id, group id, and role id.
func seedMemberWithGroupRole(t *testing.T, username string, perms []string) (uint, uint, uint) {
	t.Helper()
	db := repositories.GetDB()

	group := models.Group{Name: "perm-group"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}

	roleRepository := repositories.NewRoleRepository(nil)
	role, err := roleRepository.CreateGroupRole("Group Role", "", perms, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("seed group role: %v", err)
	}

	user := models.User{Username: username, Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	member := models.GroupMember{GroupID: group.ID, UserID: user.ID, GroupRoleID: &role.ID}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed group member: %v", err)
	}

	return user.ID, group.ID, role.ID
}

func TestHasAppPermissions(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, _ := seedUserWithAppRole(t, "app-user", []string{permissions.AppUsersRead, permissions.AppUsersCreate})
	service := NewPermissionService(nil)

	tests := []struct {
		name     string
		required []string
		want     bool
	}{
		{"single granted", []string{permissions.AppUsersRead}, true},
		{"single not granted", []string{permissions.AppUsersDelete}, false},
		{"AND all granted", []string{permissions.AppUsersRead, permissions.AppUsersCreate}, true},
		{"AND one missing", []string{permissions.AppUsersRead, permissions.AppUsersDelete}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := service.HasAppPermissions(userId, test.required...)
			if err != nil {
				t.Fatalf("HasAppPermissions: %v", err)
			}
			if got != test.want {
				t.Errorf("HasAppPermissions(%v) = %v, want %v", test.required, got, test.want)
			}
		})
	}
}

func TestHasAnyAppPermission(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, _ := seedUserWithAppRole(t, "app-user", []string{permissions.AppUsersRead})
	service := NewPermissionService(nil)

	got, err := service.HasAnyAppPermission(userId, permissions.AppUsersDelete, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("HasAnyAppPermission: %v", err)
	}
	if !got {
		t.Error("expected OR check to pass when one permission is granted")
	}

	got, err = service.HasAnyAppPermission(userId, permissions.AppUsersDelete, permissions.AppUsersCreate)
	if err != nil {
		t.Fatalf("HasAnyAppPermission: %v", err)
	}
	if got {
		t.Error("expected OR check to fail when no permission is granted")
	}
}

func TestHasAppPermissionsWildcardGrant(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, _ := seedUserWithAppRole(t, "wildcard-user", []string{"app.*"})
	service := NewPermissionService(nil)

	got, err := service.HasAppPermissions(userId, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("HasAppPermissions: %v", err)
	}
	if !got {
		t.Error("expected app.* wildcard grant to satisfy a concrete app permission")
	}
}

func TestHasAppPermissionsNoRoleDenies(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()
	db := repositories.GetDB()

	user := models.User{Username: "no-role-user", Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := NewPermissionService(nil)
	got, err := service.HasAppPermissions(user.ID, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("HasAppPermissions: %v", err)
	}
	if got {
		t.Error("expected user with no app role to be denied")
	}
}

func TestHasGroupPermissions(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, groupId, _ := seedMemberWithGroupRole(t, "group-user", []string{permissions.GroupReceiptsRead})
	service := NewPermissionService(nil)

	got, err := service.HasGroupPermissions(userId, groupId, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("HasGroupPermissions: %v", err)
	}
	if !got {
		t.Error("expected granted group permission to pass")
	}

	got, err = service.HasGroupPermissions(userId, groupId, permissions.GroupReceiptsDelete)
	if err != nil {
		t.Fatalf("HasGroupPermissions: %v", err)
	}
	if got {
		t.Error("expected ungranted group permission to fail")
	}
}

func TestHasAnyGroupPermission(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, groupId, _ := seedMemberWithGroupRole(t, "group-user", []string{permissions.GroupReceiptsRead})
	service := NewPermissionService(nil)

	got, err := service.HasAnyGroupPermission(userId, groupId, permissions.GroupReceiptsDelete, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("HasAnyGroupPermission: %v", err)
	}
	if !got {
		t.Error("expected OR check to pass when one group permission is granted")
	}

	got, err = service.HasAnyGroupPermission(userId, groupId, permissions.GroupReceiptsDelete, permissions.GroupReceiptsCreate)
	if err != nil {
		t.Fatalf("HasAnyGroupPermission: %v", err)
	}
	if got {
		t.Error("expected OR check to fail when no group permission is granted")
	}
}

func TestHasGroupPermissionsNonMemberDenies(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	// Seed a member of one group, then check a different group the user is not in.
	userId, _, _ := seedMemberWithGroupRole(t, "group-user", []string{permissions.GroupReceiptsRead})
	service := NewPermissionService(nil)

	got, err := service.HasGroupPermissions(userId, 999, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("HasGroupPermissions: %v", err)
	}
	if got {
		t.Error("expected non-member to be denied")
	}
}

func TestHasGroupPermissionsNoGroupRoleDenies(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()
	db := repositories.GetDB()

	group := models.Group{Name: "legacy-group"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	user := models.User{Username: "legacy-member", Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	// Member with no GroupRoleID assigned (legacy membership mid-transition).
	member := models.GroupMember{GroupID: group.ID, UserID: user.ID}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("seed group member: %v", err)
	}

	service := NewPermissionService(nil)
	got, err := service.HasGroupPermissions(user.ID, group.ID, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("HasGroupPermissions: %v", err)
	}
	if got {
		t.Error("expected member with no group role to be denied")
	}
}

func TestPermissionScopeGuards(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, _ := seedUserWithAppRole(t, "guard-user", []string{permissions.AppUsersRead})
	service := NewPermissionService(nil)

	// Group-scoped key passed to an app check.
	if _, err := service.HasAppPermissions(userId, permissions.GroupReceiptsRead); err == nil {
		t.Error("expected error passing a group permission to an app check")
	}
	// App-scoped key passed to a group check.
	if _, err := service.HasGroupPermissions(userId, 1, permissions.AppUsersRead); err == nil {
		t.Error("expected error passing an app permission to a group check")
	}
	// Unknown permission key.
	if _, err := service.HasAppPermissions(userId, "app.not.a.real.permission"); err == nil {
		t.Error("expected error for unknown permission key")
	}
	// No required permissions.
	if _, err := service.HasAppPermissions(userId); err != ErrNoRequiredPermissions {
		t.Errorf("expected ErrNoRequiredPermissions, got %v", err)
	}
}

func TestPermissionCacheInvalidatedOnRoleUpdate(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	userId, roleId := seedUserWithAppRole(t, "cache-user", []string{permissions.AppUsersRead})
	service := NewPermissionService(nil)

	// Prime the cache with the original permission set.
	got, err := service.HasAppPermissions(userId, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("HasAppPermissions: %v", err)
	}
	if !got {
		t.Fatal("expected initial permission to be granted")
	}

	// Update the role to swap the permission; this must invalidate the cache.
	roleService := NewRoleService(nil)
	_, err = roleService.UpdateRole(roleId, commands.UpsertRoleCommand{
		Name:        "App Role",
		Scope:       permissions.ScopeApp,
		Permissions: []string{permissions.AppUsersCreate},
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	got, err = service.HasAppPermissions(userId, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("HasAppPermissions: %v", err)
	}
	if got {
		t.Error("expected revoked permission to be denied after role update")
	}

	got, err = service.HasAppPermissions(userId, permissions.AppUsersCreate)
	if err != nil {
		t.Fatalf("HasAppPermissions: %v", err)
	}
	if !got {
		t.Error("expected newly granted permission to be allowed after role update")
	}
}

func TestPermissionCacheInvalidatedOnRoleDelete(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	// An assigned role can't be deleted (delete guard), and the user->role
	// lookup is always fresh, so prime the cache directly for an unassigned role
	// and verify DeleteRole evicts that role's cached permission list.
	roleRepository := repositories.NewRoleRepository(nil)
	role, err := roleRepository.CreateAppRole("Deletable Role", "", []string{permissions.AppUsersRead}, false)
	if err != nil {
		t.Fatalf("seed role: %v", err)
	}

	setCachedRolePermissions(permissions.ScopeApp, role.ID, []string{permissions.AppUsersRead}, rolePermissionCacheGen())
	if _, ok := getCachedRolePermissions(permissions.ScopeApp, role.ID); !ok {
		t.Fatal("precondition: role permissions should be cached")
	}

	if err := NewRoleService(nil).DeleteRole(role.ID, permissions.ScopeApp); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	if _, ok := getCachedRolePermissions(permissions.ScopeApp, role.ID); ok {
		t.Error("expected role's cached permissions to be evicted after delete")
	}
}

func TestRolePermissionCacheRejectsStaleWrite(t *testing.T) {
	clearRolePermissionCacheAll()

	// Generation observed before a (simulated) concurrent eviction.
	staleGen := rolePermissionCacheGen()
	clearRolePermissionCache(permissions.ScopeApp, 1)

	// A write carrying the pre-eviction generation must be dropped, otherwise a
	// concurrent miss could resurrect revoked permissions.
	setCachedRolePermissions(permissions.ScopeApp, 1, []string{permissions.AppUsersRead}, staleGen)
	if _, ok := getCachedRolePermissions(permissions.ScopeApp, 1); ok {
		t.Error("expected stale-generation write to be rejected")
	}

	// A write with the current generation is stored normally.
	setCachedRolePermissions(permissions.ScopeApp, 1, []string{permissions.AppUsersRead}, rolePermissionCacheGen())
	if _, ok := getCachedRolePermissions(permissions.ScopeApp, 1); !ok {
		t.Error("expected current-generation write to be cached")
	}
}

func TestHasAppPermissionsMissingUserDenies(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	service := NewPermissionService(nil)

	// A missing/deleted user must deny cleanly, not surface a record-not-found error.
	got, err := service.HasAppPermissions(99999, permissions.AppUsersRead)
	if err != nil {
		t.Fatalf("expected no error for missing user, got %v", err)
	}
	if got {
		t.Error("expected missing user to be denied")
	}
}

func TestGetAppPermissionsForUser(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	granted := []string{permissions.AppUsersRead, permissions.AppUsersCreate}
	userId, _ := seedUserWithAppRole(t, "appperms-user", granted)
	service := NewPermissionService(nil)

	got, err := service.GetAppPermissionsForUser(userId)
	if err != nil {
		t.Fatalf("GetAppPermissionsForUser: %v", err)
	}
	if !slices.Equal(sortedCopy(got), sortedCopy(granted)) {
		t.Errorf("GetAppPermissionsForUser = %v, want %v", got, granted)
	}
}

func TestGetAppPermissionsForUserNoRole(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()
	db := repositories.GetDB()

	user := models.User{Username: "appperms-no-role", Password: "password"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	service := NewPermissionService(nil)
	got, err := service.GetAppPermissionsForUser(user.ID)
	if err != nil {
		t.Fatalf("GetAppPermissionsForUser: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty permissions for user with no app role, got %v", got)
	}
}

func TestGetAppPermissionsForUserMissingUser(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	service := NewPermissionService(nil)
	got, err := service.GetAppPermissionsForUser(99999)
	if err != nil {
		t.Fatalf("expected no error for missing user, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty permissions for missing user, got %v", got)
	}
}

func TestGetGroupPermissionsForUser(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	granted := []string{permissions.GroupReceiptsRead, permissions.GroupReceiptsUpdate}
	userId, groupId, _ := seedMemberWithGroupRole(t, "groupperms-user", granted)
	service := NewPermissionService(nil)

	got, err := service.GetGroupPermissionsForUser(userId, groupId)
	if err != nil {
		t.Fatalf("GetGroupPermissionsForUser: %v", err)
	}
	if !slices.Equal(sortedCopy(got), sortedCopy(granted)) {
		t.Errorf("GetGroupPermissionsForUser = %v, want %v", got, granted)
	}
}

func TestGetGroupPermissionsForUserNonMember(t *testing.T) {
	defer repositories.TruncateTestDb()
	clearRolePermissionCacheAll()

	// Seed a member of one group, then resolve a different group the user is not in.
	userId, _, _ := seedMemberWithGroupRole(t, "groupperms-nonmember", []string{permissions.GroupReceiptsRead})
	service := NewPermissionService(nil)

	got, err := service.GetGroupPermissionsForUser(userId, 999)
	if err != nil {
		t.Fatalf("GetGroupPermissionsForUser: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty permissions for non-member, got %v", got)
	}
}

func TestGroupIdsWithPermissionFiltersToPermittedGroups(t *testing.T) {
	defer repositories.TruncateTestDb()

	// Three groups: one the user can read, one they are a member of without a role
	// (so no permissions), and one they do not belong to at all.
	userId, readableGroupId, _ := seedMemberWithGroupRole(t, "gidfilter", []string{permissions.GroupReceiptsRead})
	roleLessGroupId := addMemberToNewGroup(t, userId, "gidfilter-noread")

	foreignGroup := models.Group{Name: "gidfilter-foreign"}
	if err := repositories.GetDB().Create(&foreignGroup).Error; err != nil {
		t.Fatalf("seed foreign group: %v", err)
	}

	service := NewPermissionService(nil)
	got, err := service.GroupIdsWithPermission(
		userId,
		[]uint{roleLessGroupId, readableGroupId, foreignGroup.ID},
		permissions.GroupReceiptsRead,
	)
	if err != nil {
		t.Fatalf("GroupIdsWithPermission: %v", err)
	}
	if len(got) != 1 || got[0] != readableGroupId {
		t.Errorf("expected only the readable group %d, got %v", readableGroupId, got)
	}
}

func TestGroupIdsWithPermissionPreservesOrder(t *testing.T) {
	defer repositories.TruncateTestDb()

	// The caller pairs the result with other per-group state, so the input order
	// must survive the filter.
	userId, firstGroupId, roleId := seedMemberWithGroupRole(t, "gidorder", []string{permissions.GroupReceiptsRead})

	db := repositories.GetDB()
	var permittedGroupIds []uint
	permittedGroupIds = append(permittedGroupIds, firstGroupId)
	for _, name := range []string{"gidorder-b", "gidorder-c"} {
		group := models.Group{Name: name}
		if err := db.Create(&group).Error; err != nil {
			t.Fatalf("seed group: %v", err)
		}
		member := models.GroupMember{GroupID: group.ID, UserID: userId, GroupRoleID: &roleId}
		if err := db.Create(&member).Error; err != nil {
			t.Fatalf("seed group member: %v", err)
		}
		permittedGroupIds = append(permittedGroupIds, group.ID)
	}

	reversed := []uint{permittedGroupIds[2], permittedGroupIds[1], permittedGroupIds[0]}
	got, err := NewPermissionService(nil).GroupIdsWithPermission(userId, reversed, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("GroupIdsWithPermission: %v", err)
	}
	if !slices.Equal(got, reversed) {
		t.Errorf("expected input order %v preserved, got %v", reversed, got)
	}
}

func TestGroupIdsWithPermissionEmptyInputReturnsEmptyNonNil(t *testing.T) {
	defer repositories.TruncateTestDb()

	userId, _, _ := seedMemberWithGroupRole(t, "gidempty", []string{permissions.GroupReceiptsRead})

	got, err := NewPermissionService(nil).GroupIdsWithPermission(userId, nil, permissions.GroupReceiptsRead)
	if err != nil {
		t.Fatalf("GroupIdsWithPermission: %v", err)
	}
	if got == nil {
		t.Fatalf("expected a non-nil empty slice so callers can return it directly")
	}
	if len(got) != 0 {
		t.Errorf("expected no groups, got %v", got)
	}
}

// sortedCopy returns a sorted copy of keys, leaving the input untouched.
func sortedCopy(keys []string) []string {
	out := append([]string(nil), keys...)
	sort.Strings(out)
	return out
}
