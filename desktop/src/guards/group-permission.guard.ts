import { inject } from "@angular/core";
import { CanActivateFn, Router } from "@angular/router";
import { Store } from "@ngxs/store";
import { AuthState, GroupState } from "../store";

/**
 * Allows activation when the user holds the group-scoped permission declared on
 * `route.data['groupPermission']` for the resolved group (or one of the
 * optional `route.data['orAppPermissions']` app-scoped fallbacks). On deny,
 * redirects to the group dashboard.
 *
 * Group id resolution: when `route.data['useRouteGroupId']` is set the id comes
 * from the `:id` route param (own or parent); when `route.data['useAddTargetGroupId']`
 * is set it comes from `GroupState.addTargetGroupId` (the group a new record
 * would be created in, which is not always the one being browsed - see the
 * selector); otherwise from the currently selected group in `GroupState`.
 */
export const groupPermissionGuard: CanActivateFn = (route) => {
  const store: Store = inject(Store);
  const router: Router = inject(Router);

  const permission = route.data["groupPermission"] as string;
  const orApp = (route.data["orAppPermissions"] as string[]) ?? [];

  const selectedGroupId = () =>
    Number.parseInt(store.selectSnapshot(GroupState.selectedGroupId));

  let groupId: number;
  if (route.data["useRouteGroupId"]) {
    groupId = Number.parseInt(
      route?.params?.["id"] ?? route?.parent?.params["id"]
    );
  } else if (route.data["useAddTargetGroupId"]) {
    // Gate on the group the record would actually be created in, so a member of
    // exactly one group is not turned away because the group they happen to be
    // browsing is the synthetic "All" one. Falls back to the selected group when
    // there is no single add target, which keeps multi-group users unchanged.
    groupId =
      store.selectSnapshot(GroupState.addTargetGroupId) ?? selectedGroupId();
  } else {
    groupId = selectedGroupId();
  }

  if (
    store.selectSnapshot(
      AuthState.hasGroupPermission(groupId, permission, orApp)
    )
  ) {
    return true;
  }

  router.navigate([store.selectSnapshot(GroupState.dashboardLink)]);
  return false;
};
