import { expect, test } from '@playwright/test';
import { rmSync } from 'node:fs';
import { stubTokenRefresh } from './helpers/auth';
import {
  apiDeleteUserByName,
  createUserWithRole,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';
import {
  injectQuickScanAppData,
  openQuickScanDialog,
  uploadQuickScanImages,
} from './helpers/quick-scan';

// A picker with one option is not a choice: when a user belongs to exactly one
// group, the receipt form and the Quick Scan dialog seed it instead of making
// them pick it (GroupState.soleGroupId).
//
// The seeded e2e accounts accumulate groups as other specs run, so the single-
// group state has to be provisioned: a freshly created user gets exactly one
// real group ("My Receipts", api/internal/repositories/users.go) plus the
// synthetic "All" group, which is never a receipt target. That makes this the
// real wire — the group list comes from AppData, not from a client-side stub,
// which is the half the unit tests (mocked store) cannot prove.

const NEW_USER_PASSWORD = `${uniqueName('pw')}-Aa1!`;
const AUTH_FILE = 'e2e/.auth/single-group-user.json';

test.describe('Single-group users get the Group field pre-filled', () => {
  test.use({ storageState: AUTH_FILE });

  let username: string;
  let soleGroupId: number;
  let allGroupId: number;

  test.beforeAll(async ({ browser }) => {
    username = uniqueName('single-group-user');

    const admin = await browser.newContext({ storageState: 'e2e/.auth/admin.json' });
    const adminPage = await admin.newPage();
    await stubTokenRefresh(adminPage);
    await createUserWithRole(adminPage, {
      username,
      password: NEW_USER_PASSWORD,
      role: 'Legacy User',
    });
    await admin.close();

    // Log in once and persist the session (the describe's AUTH_FILE does not
    // exist yet, so this context opts out of it).
    const userContext = await browser.newContext({ storageState: undefined });
    const userPage = await userContext.newPage();
    await userPage.goto('/auth/login');
    await userPage.getByLabel('Username').fill(username);
    await userPage.getByLabel('Password').fill(NEW_USER_PASSWORD);
    await userPage.getByRole('button', { name: 'Login' }).click();
    await expect(userPage).toHaveURL(/\/dashboard\/group\/\d+/, { timeout: 15_000 });
    await userContext.storageState({ path: AUTH_FILE });

    // The one real group, read back off the wire — login lands on whichever
    // group the API sorts first, which is the "All" group, so it can't be taken
    // from the URL.
    const groups = await userPage.evaluate(async () => {
      const res = await fetch('/api/group/', { credentials: 'include' });
      return (await res.json()) as { id: number; name: string; isAllGroup: boolean }[];
    });
    const real = groups.filter((g) => !g.isAllGroup);
    expect(real, 'a freshly created user has exactly one real group').toHaveLength(1);
    soleGroupId = real[0].id;
    allGroupId = groups.find((g) => g.isAllGroup)!.id;

    await userContext.close();
  });

  test.afterAll(async () => {
    try {
      await withAdminApi(async (api) => {
        await apiDeleteUserByName(api, username);
      });
    } catch {
      // Best-effort cleanup — don't mask a test failure with a teardown error.
    }
    rmSync(AUTH_FILE, { force: true });
  });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  test('the manual receipt form opens on the user\'s only group and saves', async ({ page }) => {
    await page.goto('/receipts/add');
    await expect(page.getByLabel('Name')).toBeVisible();

    // Pre-filled without a pick — and the group-derived Paid By picker is usable
    // straight away, which is the point of seeding it.
    await expect(page.getByTestId('receipt-group').getByRole('combobox')).toHaveValue(
      'My Receipts',
    );

    const name = uniqueName('single-group-receipt');
    await page.getByLabel('Name').fill(name);
    await page.getByLabel('Amount').fill('12.34');
    await page.getByLabel('Paid By').click();
    await page.getByRole('option').first().click();

    await page.getByRole('button', { name: 'Save', exact: true }).first().click();
    await expect(page).toHaveURL(/\/receipts\/\d+\/view/);

    // The seed is a real selection, not just a label: the receipt is listed
    // under the group it was seeded with.
    await page.goto(`/receipts/group/${soleGroupId}`);
    await expect(
      page.getByRole('row').filter({ hasText: name }).first(),
    ).toBeVisible();
  });

  test('the Quick Scan dialog seeds each image with the only group', async ({ page }) => {
    // Only the AI flag is injected — the single-group state is genuinely the
    // provisioned user's, so this exercises the real AppData group list.
    await injectQuickScanAppData(page, {});

    // Deliberately opened from the "All" group's receipts list, not the sole
    // group's: the browsed group and the expected group must differ, or the
    // assertion below would also pass for an implementation that simply used
    // whichever group the user happened to be looking at.
    const dialog = await openQuickScanDialog(page, allGroupId);
    await uploadQuickScanImages(dialog, 1);

    await expect(dialog.getByRole('combobox', { name: 'Group' })).toHaveValue(
      'My Receipts',
    );
  });
});
