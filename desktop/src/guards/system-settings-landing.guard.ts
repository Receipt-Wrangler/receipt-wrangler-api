import { inject } from "@angular/core";
import { CanActivateFn, Router, UrlTree } from "@angular/router";
import { Store } from "@ngxs/store";
import { Permission } from "../open-api";
import { AuthState, GroupState } from "../store";

/**
 * Landing guard for the System Settings shell. The sidebar links to
 * `/system-settings`; this redirects to the first tab the user can read, in
 * the same order the tab bar renders them. Falls back to the group dashboard
 * when the user can read none of them.
 */
export const systemSettingsLandingGuard: CanActivateFn = (): UrlTree => {
  const store: Store = inject(Store);
  const router: Router = inject(Router);

  const tabs: { path: string; permission: Permission }[] = [
    { path: "settings/view", permission: Permission.AppSystemSettingsRead },
    {
      path: "receipt-processing-settings",
      permission: Permission.AppReceiptProcessingSettingsRead,
    },
    { path: "prompts", permission: Permission.AppPromptsRead },
    { path: "system-emails", permission: Permission.AppSystemEmailsRead },
    { path: "oidc-providers", permission: Permission.AppOidcProvidersRead },
    { path: "system-tasks", permission: Permission.AppSystemTasksRead },
  ];

  for (const tab of tabs) {
    if (store.selectSnapshot(AuthState.hasAppPermission(tab.permission))) {
      return router.parseUrl(`/system-settings/${tab.path}`);
    }
  }

  return router.parseUrl(store.selectSnapshot(GroupState.dashboardLink));
};
