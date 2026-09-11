import { BrowserContext, expect, Page, test } from '@playwright/test';
import { stubTokenRefresh } from './helpers/auth';
import {
  apiDeleteGroupById,
  apiDeleteRoleByName,
  createGroupWithMember,
  createRole,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';

// The group dashboard route is gated by groupDashboardReadGuard
// (group-dashboard-read.guard.ts): a member without group.dashboards.read who
// hits /dashboard/group/:id is redirected to /receipts/group/:id (a page they can
// use) instead of letting the dashboard fetch 403. The suite had no coverage of
// this guard; this spec adds the deny (redirect) path and an allow contrast.
//
// No seeded group role lacks group.dashboards.read, so an admin context
// provisions a custom group role — the "Viewer" preset minus "Read Dashboards"
// (keeping group.receipts.read so the redirect target loads) — and adds e2e-user
// to a fixture group with it. The admin who creates the group is its Owner
// (full perms incl. group.dashboards.read), giving the allow contrast for free.

test.describe('Group dashboard read gating', () => {
  test.describe.configure({ mode: 'serial' });

  let adminContext: BrowserContext;
  let adminPage: Page;
  let roleName: string;
  let groupName: string;
  let groupId: string;

  test.beforeAll(async ({ browser }) => {
    roleName = uniqueName('no-dash-role');
    groupName = uniqueName('no-dash-grp');

    adminContext = await browser.newContext({
      storageState: 'e2e/.auth/admin.json',
    });
    adminPage = await adminContext.newPage();
    await stubTokenRefresh(adminPage);

    await createRole(adminPage, {
      name: roleName,
      type: 'Group role',
      preset: 'Viewer',
      // Viewer grants group.dashboards.read + group.receipts.read; drop only the
      // dashboards read so the member keeps a loadable receipts list after the
      // redirect.
      disablePermissions: [
        { panelKey: 'group.dashboards', label: 'Read Dashboards' },
      ],
    });

    groupId = await createGroupWithMember(adminPage, {
      groupName,
      memberDisplayName: 'E2E User',
      roleName,
    });
  });

  test.afterAll(async () => {
    await adminContext?.close();
    try {
      await withAdminApi(async (api) => {
        // Delete the group first (frees the members' group-role assignment),
        // then the now-unassigned role.
        await apiDeleteGroupById(api, groupId);
        await apiDeleteRoleByName(api, roleName, 'GROUP');
      });
    } catch {
      // Best-effort cleanup — don't mask a test failure with a cleanup error.
    }
  });

  // Runs as the default project user (e2e-user = the no-dashboards member).
  test('a member without group.dashboards.read is redirected to receipts', async ({
    page,
  }) => {
    await stubTokenRefresh(page);
    await page.goto(`/dashboard/group/${groupId}`);
    // The guard redirects to the group's receipt list.
    await expect(page).toHaveURL(new RegExp(`/receipts/group/${groupId}`));
    await expect(page.getByTestId('receipts-overflow-menu')).toBeVisible();
  });

  test.describe('owner with group.dashboards.read', () => {
    // The admin created the group, so they are its Owner and hold the permission.
    test.use({ storageState: 'e2e/.auth/admin.json' });

    test('sees the dashboard (no redirect)', async ({ page }) => {
      await stubTokenRefresh(page);
      await page.goto(`/dashboard/group/${groupId}`);
      await expect(page).toHaveURL(new RegExp(`/dashboard/group/${groupId}`));
      await expect(
        page.getByRole('heading', { name: /Dashboards$/ }),
      ).toBeVisible();
    });
  });
});
