package permissions

type Scope string

const (
	ScopeApp   Scope = "APP"
	ScopeGroup Scope = "GROUP"
)

type Descriptor struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Scope       Scope  `json:"scope"`
}

const (
	AppUsersCreate = "app.users.create"
	AppUsersRead   = "app.users.read"
	AppUsersUpdate = "app.users.update"
	AppUsersDelete = "app.users.delete"

	AppPromptsCreate = "app.prompts.create"
	AppPromptsRead   = "app.prompts.read"
	AppPromptsUpdate = "app.prompts.update"
	AppPromptsDelete = "app.prompts.delete"

	AppCategoriesCreate = "app.categories.create"
	AppCategoriesRead   = "app.categories.read"
	AppCategoriesUpdate = "app.categories.update"
	AppCategoriesDelete = "app.categories.delete"

	AppTagsCreate = "app.tags.create"
	AppTagsRead   = "app.tags.read"
	AppTagsUpdate = "app.tags.update"
	AppTagsDelete = "app.tags.delete"

	AppCustomFieldsCreate = "app.custom-fields.create"
	AppCustomFieldsRead   = "app.custom-fields.read"
	AppCustomFieldsUpdate = "app.custom-fields.update"
	AppCustomFieldsDelete = "app.custom-fields.delete"

	AppSystemSettingsRead              = "app.system-settings.read"
	AppSystemSettingsUpdate            = "app.system-settings.update"
	AppSystemSettingsRestartTaskServer = "app.system-settings.restart-task-server"

	AppReceiptProcessingSettingsCreate = "app.receipt-processing-settings.create"
	AppReceiptProcessingSettingsRead   = "app.receipt-processing-settings.read"
	AppReceiptProcessingSettingsUpdate = "app.receipt-processing-settings.update"
	AppReceiptProcessingSettingsDelete = "app.receipt-processing-settings.delete"

	AppSystemEmailsCreate = "app.system-emails.create"
	AppSystemEmailsRead   = "app.system-emails.read"
	AppSystemEmailsUpdate = "app.system-emails.update"
	AppSystemEmailsDelete = "app.system-emails.delete"

	AppSystemTasksRead = "app.system-tasks.read"

	AppImportsRun = "app.imports.run"

	AppGroupsCreate         = "app.groups.create"
	AppGroupsRead           = "app.groups.read"
	AppGroupsUpdateSettings = "app.groups.update-settings"
	AppGroupsDelete         = "app.groups.delete"

	AppApiKeysCreate    = "app.api-keys.create"
	AppApiKeysRead      = "app.api-keys.read"
	AppApiKeysReadAny   = "app.api-keys.read-any"
	AppApiKeysUpdate    = "app.api-keys.update"
	AppApiKeysDelete    = "app.api-keys.delete"
	AppApiKeysDeleteAny = "app.api-keys.delete-any"

	AppOidcProvidersCreate = "app.oidc-providers.create"
	AppOidcProvidersRead   = "app.oidc-providers.read"
	AppOidcProvidersUpdate = "app.oidc-providers.update"
	AppOidcProvidersDelete = "app.oidc-providers.delete"

	AppRolesCreate = "app.roles.create"
	AppRolesRead   = "app.roles.read"
	AppRolesUpdate = "app.roles.update"
	AppRolesDelete = "app.roles.delete"

	AppNotificationsRead   = "app.notifications.read"
	AppNotificationsDelete = "app.notifications.delete"

	AppUserPreferencesRead   = "app.user-preferences.read"
	AppUserPreferencesUpdate = "app.user-preferences.update"

	AppAccountRead   = "app.account.read"
	AppAccountUpdate = "app.account.update"
	AppAccountDelete = "app.account.delete"

	AppReceiptsSearch = "app.receipts.search"

	AppReportsRead      = "app.reports.read"
	AppReportsCreate    = "app.reports.create"
	AppReportsUpdate    = "app.reports.update"
	AppReportsDelete    = "app.reports.delete"
	AppReportsDuplicate = "app.reports.duplicate"
	AppReportsGenerate  = "app.reports.generate"

	// The "*All" report permissions bypass both the per-group access requirement and
	// the per-template grant matrix for that one action — a holder may perform it on
	// any template. They are the per-action opt-out of the default group-scoped model.
	AppReportsReadAll      = "app.reports.readAll"
	AppReportsCreateAll    = "app.reports.createAll"
	AppReportsUpdateAll    = "app.reports.updateAll"
	AppReportsDeleteAll    = "app.reports.deleteAll"
	AppReportsDuplicateAll = "app.reports.duplicateAll"
	AppReportsGenerateAll  = "app.reports.generateAll"
)

const (
	GroupView   = "group.view"
	GroupUpdate = "group.update"
	GroupDelete = "group.delete"

	GroupMembersCreate = "group.members.create"
	GroupMembersUpdate = "group.members.update"
	GroupMembersDelete = "group.members.delete"
	// GroupMembersGrantsUpdate is deliberately SEPARATE from GroupMembersUpdate.
	// Per-member category/tag grants are a privacy boundary, so the ability to edit
	// them must not ride along with ordinary member management — otherwise a
	// restricted member holding group.members.update could widen their own grants
	// and lift the very restriction the feature exists to enforce.
	GroupMembersGrantsUpdate = "group.members.grants.update"

	GroupReceiptsCreate    = "group.receipts.create"
	GroupReceiptsRead      = "group.receipts.read"
	GroupReceiptsUpdate    = "group.receipts.update"
	GroupReceiptsDelete    = "group.receipts.delete"
	GroupReceiptsDuplicate = "group.receipts.duplicate"
	GroupReceiptsMagicFill = "group.receipts.magic-fill"
	GroupReceiptsQuickScan = "group.receipts.quick-scan"

	GroupCommentsCreate = "group.comments.create"
	GroupCommentsDelete = "group.comments.delete"

	GroupDashboardsCreate = "group.dashboards.create"
	GroupDashboardsRead   = "group.dashboards.read"
	GroupDashboardsUpdate = "group.dashboards.update"
	GroupDashboardsDelete = "group.dashboards.delete"

	GroupWidgetsRead = "group.widgets.read"

	GroupReportsRead = "group.reports.read"

	GroupActivitiesRead  = "group.activities.read"
	GroupActivitiesRerun = "group.activities.rerun"

	GroupEmailPoll = "group.email.poll"
)

var registry = []Descriptor{
	{AppUsersCreate, "Create Users", "Create new user accounts.", "User Management", ScopeApp},
	{AppUsersRead, "Read Users", "List and look up user accounts.", "User Management", ScopeApp},
	{AppUsersUpdate, "Update Users", "Edit user profiles, reset passwords, and convert dummy users.", "User Management", ScopeApp},
	{AppUsersDelete, "Delete Users", "Delete users individually or in bulk.", "User Management", ScopeApp},

	{AppPromptsCreate, "Create AI Prompts", "Create new AI prompt definitions.", "AI", ScopeApp},
	{AppPromptsRead, "Read AI Prompts", "View the AI prompt library.", "AI", ScopeApp},
	{AppPromptsUpdate, "Update AI Prompts", "Edit existing AI prompts.", "AI", ScopeApp},
	{AppPromptsDelete, "Delete AI Prompts", "Remove AI prompts.", "AI", ScopeApp},

	{AppCategoriesCreate, "Create Categories", "Create new categories.", "Catalog", ScopeApp},
	{AppCategoriesRead, "Read Categories", "List and look up categories.", "Catalog", ScopeApp},
	{AppCategoriesUpdate, "Update Categories", "Edit existing categories.", "Catalog", ScopeApp},
	{AppCategoriesDelete, "Delete Categories", "Remove categories.", "Catalog", ScopeApp},

	{AppTagsCreate, "Create Tags", "Create new tags.", "Catalog", ScopeApp},
	{AppTagsRead, "Read Tags", "List and look up tags.", "Catalog", ScopeApp},
	{AppTagsUpdate, "Update Tags", "Edit existing tags.", "Catalog", ScopeApp},
	{AppTagsDelete, "Delete Tags", "Remove tags.", "Catalog", ScopeApp},

	{AppCustomFieldsCreate, "Create Custom Fields", "Create new custom field definitions.", "Catalog", ScopeApp},
	{AppCustomFieldsRead, "Read Custom Fields", "List and look up custom fields.", "Catalog", ScopeApp},
	{AppCustomFieldsUpdate, "Update Custom Fields", "Edit an existing custom field's name, description, and select options. The field's type can never be changed.", "Catalog", ScopeApp},
	{AppCustomFieldsDelete, "Delete Custom Fields", "Remove custom field definitions.", "Catalog", ScopeApp},

	{AppSystemSettingsRead, "Read System Settings", "View system settings.", "System", ScopeApp},
	{AppSystemSettingsUpdate, "Update System Settings", "Edit system settings.", "System", ScopeApp},
	{AppSystemSettingsRestartTaskServer, "Restart Task Server", "Restart the background task worker.", "System", ScopeApp},

	{AppReceiptProcessingSettingsCreate, "Create Receipt Processing Configs", "Add new OCR/AI processing configurations.", "System", ScopeApp},
	{AppReceiptProcessingSettingsRead, "Read Receipt Processing Configs", "View and test OCR/AI processing configurations.", "System", ScopeApp},
	{AppReceiptProcessingSettingsUpdate, "Update Receipt Processing Configs", "Edit OCR/AI processing configurations.", "System", ScopeApp},
	{AppReceiptProcessingSettingsDelete, "Delete Receipt Processing Configs", "Remove OCR/AI processing configurations.", "System", ScopeApp},

	{AppSystemEmailsCreate, "Create System Email", "Add inbound email integrations.", "System", ScopeApp},
	{AppSystemEmailsRead, "Read System Email", "View and test inbound email integrations.", "System", ScopeApp},
	{AppSystemEmailsUpdate, "Update System Email", "Edit inbound email integrations.", "System", ScopeApp},
	{AppSystemEmailsDelete, "Delete System Email", "Remove inbound email integrations.", "System", ScopeApp},

	{AppSystemTasksRead, "Read System Tasks", "Inspect the system-wide activity log.", "System", ScopeApp},
	{AppImportsRun, "Import Configuration", "Restore or seed the system from a configuration export.", "System", ScopeApp},

	{AppGroupsCreate, "Create Groups", "Create new groups.", "Group Management", ScopeApp},
	{AppGroupsRead, "Read All Groups", "List and look up groups across the system, including ones the calling user is not a member of.", "Group Management", ScopeApp},
	{AppGroupsUpdateSettings, "Update Group System Settings", "Edit system-level settings on any group (separate from per-group ownership editing).", "Group Management", ScopeApp},
	{AppGroupsDelete, "Delete Any Group", "Permanently delete any group in the system, including ones the calling user is not a member of. Intended for cleaning up abandoned or accidentally created groups; pairs with Read All Groups.", "Group Management", ScopeApp},

	{AppApiKeysCreate, "Create API Keys", "Issue API keys for the calling user.", "Security", ScopeApp},
	{AppApiKeysRead, "Read API Keys", "List the calling user's API keys.", "Security", ScopeApp},
	{AppApiKeysReadAny, "Read Any API Key", "List API keys belonging to other users.", "Security", ScopeApp},
	{AppApiKeysUpdate, "Update API Keys", "Edit the calling user's API keys.", "Security", ScopeApp},
	{AppApiKeysDelete, "Delete API Keys", "Revoke the calling user's API keys.", "Security", ScopeApp},
	{AppApiKeysDeleteAny, "Delete Any API Key", "Revoke API keys belonging to other users.", "Security", ScopeApp},

	{AppOidcProvidersCreate, "Create OIDC Providers", "Register a new OpenID Connect identity provider users can sign in with.", "Security", ScopeApp},
	{AppOidcProvidersRead, "Read OIDC Providers", "View configured OpenID Connect identity providers. Client secrets are never returned.", "Security", ScopeApp},
	{AppOidcProvidersUpdate, "Update OIDC Providers", "Edit a configured OpenID Connect identity provider, including its provisioning and account-linking behavior.", "Security", ScopeApp},
	{AppOidcProvidersDelete, "Delete OIDC Providers", "Remove an OpenID Connect identity provider. Accounts linked to it lose the ability to sign in with it.", "Security", ScopeApp},

	{AppRolesCreate, "Create Roles", "Create new app or group roles.", "Access Control", ScopeApp},
	{AppRolesRead, "Read Roles", "List roles and view the permission catalog.", "Access Control", ScopeApp},
	{AppRolesUpdate, "Update Roles", "Edit existing roles.", "Access Control", ScopeApp},
	{AppRolesDelete, "Delete Roles", "Remove roles.", "Access Control", ScopeApp},

	{AppNotificationsRead, "Read Notifications", "View your own notifications and unread count.", "Account", ScopeApp},
	{AppNotificationsDelete, "Delete Notifications", "Dismiss your own notifications.", "Account", ScopeApp},

	{AppUserPreferencesRead, "Read User Preferences", "View your own user preferences.", "Account", ScopeApp},
	{AppUserPreferencesUpdate, "Update User Preferences", "Edit your own user preferences.", "Account", ScopeApp},

	{AppAccountRead, "Read Own Account", "Read your own profile, claims, groups, and app bootstrap data.", "Account", ScopeApp},
	{AppAccountUpdate, "Update Own Account", "Edit your own profile.", "Account", ScopeApp},
	{AppAccountDelete, "Delete Own Account", "Delete your own account.", "Account", ScopeApp},

	{AppReceiptsSearch, "Search Receipts", "Search across receipts you can access.", "Receipts", ScopeApp},

	{AppReportsRead, "Access Reports", "Access the report builder and saved report templates.", "Reports", ScopeApp},
	{AppReportsCreate, "Save Report Templates", "Save a report configuration as a reusable template.", "Reports", ScopeApp},
	{AppReportsUpdate, "Update Report Templates", "Update a saved report template in place.", "Reports", ScopeApp},
	{AppReportsDelete, "Delete Report Templates", "Delete a saved report template.", "Reports", ScopeApp},
	{AppReportsDuplicate, "Duplicate Report Templates", "Duplicate a saved report template.", "Reports", ScopeApp},
	{AppReportsGenerate, "Generate Reports", "Generate and download reports.", "Reports", ScopeApp},

	{AppReportsReadAll, "Read All Report Templates", "View and act on every report template, bypassing per-group access and per-template restrictions.", "Reports", ScopeApp},
	{AppReportsCreateAll, "Create Reports For Any Group", "Save report templates covering any group, bypassing the group-access requirement on create.", "Reports", ScopeApp},
	{AppReportsUpdateAll, "Update All Report Templates", "Update any report template, bypassing per-group access and per-template restrictions.", "Reports", ScopeApp},
	{AppReportsDeleteAll, "Delete All Report Templates", "Delete any report template, bypassing per-group access and per-template restrictions.", "Reports", ScopeApp},
	{AppReportsDuplicateAll, "Duplicate All Report Templates", "Duplicate any report template, bypassing per-group access and per-template restrictions.", "Reports", ScopeApp},
	{AppReportsGenerateAll, "Generate All Reports", "Generate and download any saved report template, bypassing per-group access and per-template restrictions.", "Reports", ScopeApp},

	{GroupView, "View Group", "See the group, its members, and metadata.", "Group", ScopeGroup},
	{GroupUpdate, "Update Group", "Edit group name, settings, and receipt-handling configuration.", "Group", ScopeGroup},
	{GroupDelete, "Delete Group", "Permanently delete the group.", "Group", ScopeGroup},

	{GroupMembersCreate, "Add Group Members", "Add members to the group.", "Group", ScopeGroup},
	{GroupMembersUpdate, "Update Group Members", "Change a member's group role.", "Group", ScopeGroup},
	{GroupMembersDelete, "Remove Group Members", "Remove members from the group.", "Group", ScopeGroup},
	{GroupMembersGrantsUpdate, "Assign Member Categories & Tags", "Assign which categories and tags an individual member can see, within the limits of their group role.", "Group", ScopeGroup},

	{GroupReceiptsCreate, "Create Receipts", "Upload images and create receipts.", "Receipts", ScopeGroup},
	{GroupReceiptsRead, "Read Receipts", "Read, list, and export receipts.", "Receipts", ScopeGroup},
	{GroupReceiptsUpdate, "Update Receipts", "Edit receipts and update status in bulk.", "Receipts", ScopeGroup},
	{GroupReceiptsDelete, "Delete Receipts", "Remove receipts.", "Receipts", ScopeGroup},
	{GroupReceiptsDuplicate, "Duplicate Receipts", "Create a copy of an existing receipt.", "Receipts", ScopeGroup},
	{GroupReceiptsMagicFill, "Magic Fill Receipts", "Run AI-powered data extraction on a receipt image to pre-fill receipt fields.", "Receipts", ScopeGroup},
	{GroupReceiptsQuickScan, "Quick Scan Receipts", "Use the Quick Scan flow to rapidly capture and create receipts.", "Receipts", ScopeGroup},

	{GroupCommentsCreate, "Create Comments", "Add comments to receipts.", "Receipts", ScopeGroup},
	{GroupCommentsDelete, "Delete Comments", "Remove comments.", "Receipts", ScopeGroup},

	{GroupDashboardsCreate, "Create Dashboards", "Create dashboards for the group.", "Dashboards", ScopeGroup},
	{GroupDashboardsRead, "Read Dashboards", "View dashboards.", "Dashboards", ScopeGroup},
	{GroupDashboardsUpdate, "Update Dashboards", "Edit dashboards.", "Dashboards", ScopeGroup},
	{GroupDashboardsDelete, "Delete Dashboards", "Remove dashboards.", "Dashboards", ScopeGroup},

	{GroupWidgetsRead, "Read Widgets", "Read widget data (charts, summaries).", "Dashboards", ScopeGroup},

	{GroupReportsRead, "Read Reports", "Generate and download reports over the group's receipts.", "Reports", ScopeGroup},

	{GroupActivitiesRead, "Read Activities", "View the activity feed for the group.", "Activity", ScopeGroup},
	{GroupActivitiesRerun, "Rerun Activities", "Re-execute a failed or stale activity.", "Activity", ScopeGroup},

	{GroupEmailPoll, "Poll Inbound Email", "Trigger an inbound email poll for the group's inbox.", "Group", ScopeGroup},
}

func All() []Descriptor {
	out := make([]Descriptor, len(registry))
	copy(out, registry)
	return out
}

func Get(key string) (Descriptor, bool) {
	for _, d := range registry {
		if d.Key == key {
			return d, true
		}
	}
	return Descriptor{}, false
}

func Exists(key string) bool {
	_, ok := Get(key)
	return ok
}
