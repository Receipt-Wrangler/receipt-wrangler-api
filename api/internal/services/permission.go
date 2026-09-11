package services

import (
	"errors"
	"fmt"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"

	"gorm.io/gorm"
)

// ErrNoRequiredPermissions is returned when a permission check is invoked
// without specifying any required permission.
var ErrNoRequiredPermissions = errors.New("at least one required permission must be provided")

// matcher is the matching strategy used by a check: permissions.HasAll (AND,
// the default) or permissions.HasAny (OR).
type matcher func(granted []string, required ...string) bool

// PermissionService resolves a user's effective permissions from the database
// at check time (never trusting JWT contents for authorization) and matches
// them against the required permissions. App-level and group-level checks are
// kept as separate entry points; both share the permissions package matcher.
type PermissionService struct {
	BaseService

	// grantMemo caches resolveEffectiveGrants per (user, group), so the category and
	// tag accessors — which each resolve the FULL set — share one membership lookup
	// instead of doing the work twice. See resolveEffectiveGrants.
	//
	// Entries carry the group-role grant cache's eviction generation and are
	// discarded once it advances, so a role update invalidates this memo too rather
	// than only the process-wide role cache.
	//
	// A map (not a pointer field) because every method has a value receiver: copies
	// of the service share the same memo, which is what lets a resolver struct
	// holding a copy (groupVisibilityResolver, receiptGrantFilter) benefit too.
	grantMemo map[grantMemoKey]grantMemoEntry
}

// grantMemoKey identifies one resolution. Grants are per membership, so the user
// alone is not enough — the same user resolves differently in each group.
type grantMemoKey struct {
	userId  uint
	groupId uint
}

// grantMemoEntry is a memoized resolution plus the eviction generation it was
// computed under. A nil grants field is a real answer ("not a member"), which is
// why presence is tested rather than the value.
type grantMemoEntry struct {
	grants     *effectiveGrantSet
	generation uint64
}

// NewPermissionService builds a request-scoped permission service.
//
// It memoizes grant resolution (see resolveEffectiveGrants). Role-grant changes
// invalidate that memo through the eviction generation, but a MEMBER-level grant
// write does not — so do not hold one instance across a call that reassigns a
// member's individual grants. No current flow does: the grants endpoint uses
// GroupService and constructs no PermissionService.
func NewPermissionService(tx *gorm.DB) PermissionService {
	service := PermissionService{
		BaseService: BaseService{
			DB: repositories.GetDB(),
			TX: tx,
		},
		grantMemo: make(map[grantMemoKey]grantMemoEntry),
	}
	return service
}

// HasAppPermissions reports whether the user's app role grants ALL of the
// required app-level permissions (logical AND — the default). A single-permission
// check is the common case: HasAppPermissions(userId, permissions.AppUsersRead).
func (service PermissionService) HasAppPermissions(userId uint, required ...string) (bool, error) {
	return service.checkApp(userId, permissions.HasAll, required...)
}

// HasAnyAppPermission reports whether the user's app role grants AT LEAST ONE of
// the required app-level permissions (logical OR).
func (service PermissionService) HasAnyAppPermission(userId uint, required ...string) (bool, error) {
	return service.checkApp(userId, permissions.HasAny, required...)
}

// HasGroupPermissions reports whether the user's role in the group grants ALL of
// the required group-level permissions (logical AND — the default).
func (service PermissionService) HasGroupPermissions(userId uint, groupId uint, required ...string) (bool, error) {
	return service.checkGroup(userId, groupId, permissions.HasAll, required...)
}

// HasAnyGroupPermission reports whether the user's role in the group grants AT
// LEAST ONE of the required group-level permissions (logical OR).
func (service PermissionService) HasAnyGroupPermission(userId uint, groupId uint, required ...string) (bool, error) {
	return service.checkGroup(userId, groupId, permissions.HasAny, required...)
}

// GetAppPermissionsForUser returns the user's effective app-level permissions,
// resolved from the database (the JWT is never trusted). A user with no app role
// — or a missing/deleted user — resolves to an empty slice (deny).
func (service PermissionService) GetAppPermissionsForUser(userId uint) ([]string, error) {
	return service.resolveAppPermissions(userId)
}

// GetGroupPermissionsForUser returns the user's effective group-level
// permissions for a group, resolved from the database. A non-member, or a
// member whose membership has no group role, resolves to an empty slice (deny).
func (service PermissionService) GetGroupPermissionsForUser(userId uint, groupId uint) ([]string, error) {
	return service.resolveGroupPermissions(userId, groupId)
}

// GroupIdsWithPermission narrows groupIds to those where the user's group role
// grants ALL of the required group-level permissions, preserving input order.
// It serves the read surfaces that FILTER rather than reject: a group the caller
// cannot read is dropped silently instead of failing the whole request. Contrast
// CanReportOverGroups, which is all-must-pass.
//
// It exists because membership is deliberately NOT access. A surface that derives
// its group set from the caller's memberships (rather than from a group id in the
// request, which the declarative HandleRequest gate covers) must still apply the
// group-scoped permission itself, exactly as that gate would for a named group.
//
// An empty result is a real answer ("no groups qualify"), never a denial — callers
// must map it to "no results", not to a permission error. The returned slice is
// always non-nil so it can be returned directly.
//
// Cost is one membership lookup per group (each role's permission list is cached
// process-wide by role id, the per-user assignment is not), so callers doing other
// per-group work should filter FIRST to shorten those later loops.
func (service PermissionService) GroupIdsWithPermission(userId uint, groupIds []uint, required ...string) ([]uint, error) {
	allowed := make([]uint, 0, len(groupIds))

	for _, groupId := range groupIds {
		hasPermission, err := service.HasGroupPermissions(userId, groupId, required...)
		if err != nil {
			return nil, err
		}
		if hasPermission {
			allowed = append(allowed, groupId)
		}
	}

	return allowed, nil
}

func (service PermissionService) checkApp(userId uint, match matcher, required ...string) (bool, error) {
	if err := validateRequiredPermissions(permissions.ScopeApp, required); err != nil {
		return false, err
	}

	granted, err := service.resolveAppPermissions(userId)
	if err != nil {
		return false, err
	}

	return match(granted, required...), nil
}

func (service PermissionService) checkGroup(userId uint, groupId uint, match matcher, required ...string) (bool, error) {
	if err := validateRequiredPermissions(permissions.ScopeGroup, required); err != nil {
		return false, err
	}

	granted, err := service.resolveGroupPermissions(userId, groupId)
	if err != nil {
		return false, err
	}

	return match(granted, required...), nil
}

// validateRequiredPermissions guards against programmer error: an empty
// requirement, an unknown permission key (typo), or a key whose scope does not
// match the check being performed (e.g. an app permission passed to a group
// check).
func validateRequiredPermissions(scope permissions.Scope, required []string) error {
	if len(required) == 0 {
		return ErrNoRequiredPermissions
	}

	for _, key := range required {
		descriptor, ok := permissions.Get(key)
		if !ok {
			return fmt.Errorf("unknown permission %q", key)
		}
		if descriptor.Scope != scope {
			return fmt.Errorf("permission %q is %s-scoped, expected %s", key, descriptor.Scope, scope)
		}
	}

	return nil
}

// resolveAppPermissions loads the user's current app role permissions. A user
// with no app role assigned — or a missing/deleted user — resolves to no
// permissions (deny) rather than an error.
func (service PermissionService) resolveAppPermissions(userId uint) ([]string, error) {
	roleRepository := repositories.NewRoleRepository(service.TX)

	appRoleId, err := roleRepository.GetUserAppRoleId(userId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if appRoleId == nil {
		return []string{}, nil
	}

	return loadRolePermissions(roleRepository, permissions.ScopeApp, *appRoleId)
}

// resolveGroupPermissions loads the user's current group role permissions for a
// group. A user who is not a member, or whose membership has no group role,
// resolves to no permissions (deny).
func (service PermissionService) resolveGroupPermissions(userId uint, groupId uint) ([]string, error) {
	roleRepository := repositories.NewRoleRepository(service.TX)

	groupRoleId, err := roleRepository.GetGroupMemberRoleId(userId, groupId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if groupRoleId == nil {
		return []string{}, nil
	}

	return loadRolePermissions(roleRepository, permissions.ScopeGroup, *groupRoleId)
}

// loadRolePermissions returns a role's permission strings, consulting the cache
// first and populating it on a miss. Callers must not mutate the returned slice.
func loadRolePermissions(roleRepository repositories.RoleRepository, scope permissions.Scope, roleId uint) ([]string, error) {
	if cached, ok := getCachedRolePermissions(scope, roleId); ok {
		return cached, nil
	}

	// Capture the eviction generation before reading so a concurrent role
	// update/delete invalidates this write instead of being undone by it.
	observedGen := rolePermissionCacheGen()

	var perms []string
	var err error
	if scope == permissions.ScopeApp {
		perms, err = roleRepository.GetAppRolePermissions(roleId)
	} else {
		perms, err = roleRepository.GetGroupRolePermissions(roleId)
	}
	if err != nil {
		return nil, err
	}

	setCachedRolePermissions(scope, roleId, perms, observedGen)
	return perms, nil
}
