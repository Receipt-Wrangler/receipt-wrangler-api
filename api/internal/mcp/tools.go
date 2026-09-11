package mcp

import (
	"context"
	"errors"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/permissions"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxSearchResults bounds how many receipts a single search returns, matching
// the limit used by the REST search handler.
const maxSearchResults = 100

// emptyInput is the parameter type for tools that take no arguments. It yields
// an empty JSON object schema.
type emptyInput struct{}

type searchReceiptsInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"maxResults,omitempty"`
}

type getReceiptInput struct {
	Id string `json:"id"`
}

type listDashboardsInput struct {
	GroupId string `json:"groupId"`
}

// registerTools wires the read-only MCP tools onto the server. Every handler
// derives the acting user from the verified token and enforces the same
// group-scoped authorization the REST handlers do.
func registerTools(server *mcpsdk.Server) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "search_receipts",
		Description: "Search the authenticated user's receipts by name. Returns receipts from groups the user belongs to, most recent first.",
	}, handleSearchReceipts)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_receipt",
		Description: "Get a single receipt by id, including its items, categories, and tags. Only receipts in the user's groups are accessible.",
	}, handleGetReceipt)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_groups",
		Description: "List the groups the authenticated user belongs to.",
	}, handleListGroups)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_categories",
		Description: "List the receipt categories available to the authenticated user (filtered by the user's group role grants).",
	}, handleListCategories)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_tags",
		Description: "List the receipt tags available to the authenticated user (filtered by the user's group role grants).",
	}, handleListTags)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_dashboards",
		Description: "List the authenticated user's dashboards for a given group. Requires a groupId the user can access.",
	}, handleListDashboards)
}

func handleSearchReceipts(ctx context.Context, req *mcpsdk.CallToolRequest, in searchReceiptsInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	limit := in.MaxResults
	if limit <= 0 || limit > maxSearchResults {
		limit = maxSearchResults
	}

	// Delegate to the shared enforced read so REST and MCP can't drift: it
	// enforces app.receipts.search, narrows the user's groups to those granting
	// group.receipts.read, and applies paid-by visibility in SQL before the limit.
	results, err := services.NewReceiptService(nil).SearchReceiptsForUser(claims.UserId, in.Query, limit)
	if err != nil {
		if errors.Is(err, services.ErrSearchForbidden) {
			return nil, nil, errors.New("unauthorized")
		}
		return nil, nil, err
	}

	return nil, results, nil
}

func handleGetReceipt(ctx context.Context, req *mcpsdk.CallToolRequest, in getReceiptInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	if len(in.Id) == 0 {
		return nil, nil, errors.New("id is required")
	}

	// Delegate to the shared enforced read (permission + paid-by visibility +
	// category/tag stripping). Collapse every access failure to a non-leaking
	// "receipt not found" so we don't disclose the existence of receipts in
	// other users' groups.
	receipt, err := services.NewReceiptService(nil).GetReceiptForUser(claims.UserId, in.Id)
	if err != nil {
		if errors.Is(err, services.ErrReceiptAccessDenied) {
			return nil, nil, errors.New("receipt not found")
		}
		return nil, nil, err
	}

	return nil, receipt, nil
}

func handleListGroups(ctx context.Context, req *mcpsdk.CallToolRequest, _ emptyInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	groupService := services.NewGroupService(nil)
	groups, err := groupService.GetGroupsForUser(utils.UintToString(claims.UserId))
	if err != nil {
		return nil, nil, err
	}

	// Member-presence isolation: strip each roster to the members the caller may see.
	permissionService := services.NewPermissionService(nil)
	if err := permissionService.FilterGroupMembersForGroups(claims.UserId, groups); err != nil {
		return nil, nil, err
	}

	return nil, groups, nil
}

func handleListCategories(ctx context.Context, req *mcpsdk.CallToolRequest, _ emptyInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	categories, err := repositories.NewCategoryRepository(nil).GetAllCategories("*")
	if err != nil {
		return nil, nil, err
	}

	permissionService := services.NewPermissionService(nil)
	visible, err := visibleByGrants(
		claims.UserId,
		categories,
		func(category models.Category) uint { return category.ID },
		permissions.AppCategoriesRead,
		func(groupId uint) (map[uint]struct{}, bool, error) {
			return permissionService.GetGroupCategoryIdsForUser(claims.UserId, groupId)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return nil, visible, nil
}

func handleListTags(ctx context.Context, req *mcpsdk.CallToolRequest, _ emptyInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	tags, err := repositories.NewTagsRepository(nil).GetAllTags("*")
	if err != nil {
		return nil, nil, err
	}

	permissionService := services.NewPermissionService(nil)
	visible, err := visibleByGrants(
		claims.UserId,
		tags,
		func(tag models.Tag) uint { return tag.ID },
		permissions.AppTagsRead,
		func(groupId uint) (map[uint]struct{}, bool, error) {
			return permissionService.GetGroupTagIdsForUser(claims.UserId, groupId)
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return nil, visible, nil
}

// visibleByGrants filters a global catalog (categories or tags) to what the user
// may actually see, mirroring how GetAppData builds the per-group catalogs.
// Holders of the app-level read permission see the whole pool (consistent with
// the admin-only REST global list); everyone else sees the union of their group
// roles' grants across their groups — being unrestricted in any group means the
// full pool, and a user with no groups or no grants sees nothing.
func visibleByGrants[T any](
	userId uint,
	all []T,
	idOf func(T) uint,
	appReadPermission string,
	resolve func(groupId uint) (map[uint]struct{}, bool, error),
) ([]T, error) {
	bypass, err := services.NewPermissionService(nil).HasAppPermissions(userId, appReadPermission)
	if err != nil {
		return nil, err
	}
	if bypass {
		return all, nil
	}

	groupIds, err := repositories.NewGroupMemberRepository(nil).GetGroupIdsByUserId(utils.UintToString(userId))
	if err != nil {
		return nil, err
	}

	allowed := make(map[uint]struct{})
	for _, groupId := range groupIds {
		ids, unrestricted, err := resolve(groupId)
		if err != nil {
			return nil, err
		}
		if unrestricted {
			return all, nil
		}
		for id := range ids {
			allowed[id] = struct{}{}
		}
	}

	visible := make([]T, 0, len(allowed))
	for _, item := range all {
		if _, ok := allowed[idOf(item)]; ok {
			visible = append(visible, item)
		}
	}
	return visible, nil
}

func handleListDashboards(ctx context.Context, req *mcpsdk.CallToolRequest, in listDashboardsInput) (*mcpsdk.CallToolResult, any, error) {
	claims, err := claimsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	if len(in.GroupId) == 0 {
		return nil, nil, errors.New("groupId is required")
	}

	groupId, err := utils.StringToUint(in.GroupId)
	if err != nil {
		return nil, nil, errors.New("invalid groupId")
	}

	permissionService := services.NewPermissionService(nil)
	hasAccess, err := permissionService.HasGroupPermissions(claims.UserId, groupId, permissions.GroupDashboardsRead)
	if err != nil || !hasAccess {
		return nil, nil, errors.New("unauthorized to access this group")
	}

	dashboardRepository := repositories.NewDashboardRepository(nil)
	dashboards, err := dashboardRepository.GetDashboardsForUserByGroup(claims.UserId, groupId)
	if err != nil {
		return nil, nil, err
	}

	return nil, dashboards, nil
}

// claimsFromRequest retrieves the verified user claims that verifyToken stashed
// on the bearer TokenInfo.
func claimsFromRequest(req *mcpsdk.CallToolRequest) (*structs.Claims, error) {
	if req.Extra == nil || req.Extra.TokenInfo == nil {
		return nil, errors.New("missing authentication context")
	}

	claims, ok := req.Extra.TokenInfo.Extra[claimsKey].(*structs.Claims)
	if !ok || claims == nil {
		return nil, errors.New("invalid authentication context")
	}

	return claims, nil
}
