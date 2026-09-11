import { expect, Route, test } from '@playwright/test';
import { stubTokenRefresh } from './helpers/auth';
import {
  apiCreateGroup,
  apiDeleteGroupById,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';
import { selectImageGroup } from './helpers/quick-scan';

// Verifies the quick-scan DIALOG responds to a group's quick-scan configuration — no scan is sent.
//
// The dialog is only reachable via a button gated behind the `aiPoweredReceipts` feature flag, which
// is off in dev/CI. Rather than mutate that global server config, we intercept the AppData response
// client-side (per-BrowserContext, like stubTokenRefresh) to (1) flip the flag on and (2) inject a
// distinctive quick-scan config onto the target group. The dialog reads config from
// GroupState.getGroupById(...).groupReceiptSettings, so the injected values drive its field visibility.

test.use({ storageState: 'e2e/.auth/admin.json' });

test.describe('Quick scan dialog field response', () => {
  let group: { id: number; name: string };

  test.beforeAll(async () => {
    await withAdminApi(async (api) => {
      group = await apiCreateGroup(api, uniqueName('qs-dlg'));
    });
  });

  test.afterAll(async () => {
    await withAdminApi(async (api) => {
      await apiDeleteGroupById(api, String(group.id));
    });
  });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);

    // Enable the AI feature flag and inject the group's quick-scan config into AppData.
    await page.route('**/api/user/appData', async (route: Route) => {
      const response = await route.fetch();
      const body = await response.json();
      body.featureConfig = { ...body.featureConfig, aiPoweredReceipts: true };
      const target = (body.groups ?? []).find((g: any) => g.id === group.id);
      if (target) {
        target.groupReceiptSettings = {
          ...target.groupReceiptSettings,
          quickScanPaidByEnabled: false, // hidden
          quickScanPaidByRequired: false,
          quickScanDefaultPaidByType: 'UPLOADER',
          quickScanStatusEnabled: true, // shown, optional
          quickScanStatusRequired: false,
          quickScanDefaultStatus: 'OPEN',
          quickScanCategoriesEnabled: true, // shown, required
          quickScanCategoriesRequired: true,
          quickScanTagsEnabled: false, // hidden
          quickScanTagsRequired: false,
        };
      }
      await route.fulfill({ response, json: body });
    });

    // Belt-and-suspenders for the pre-auth path (logged-in path uses AppData above).
    await page.route('**/api/featureConfig', (route: Route) =>
      route.fulfill({ json: { enableLocalSignUp: true, aiPoweredReceipts: true } }),
    );
  });

  test('shows/hides and requires fields per the selected group config', async ({ page }) => {
    await page.goto(`/receipts/group/${group.id}`);

    // The feature-flag-gated Quick Scan button now renders (flag stubbed + admin owns the group). It
    // is an icon-only button (tooltip is aria-describedby, not the a11y name), so target its testid.
    await page.getByTestId('receipts-quick-scan').getByRole('button').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Uploading an image adds the per-image fields (the file is read client-side).
    await dialog
      .locator('app-upload-image input[type="file"]')
      .setInputFiles('e2e/fixtures/receipt.png');
    const groupField = dialog.getByRole('combobox', { name: 'Group' });
    await expect(groupField).toBeVisible();

    // No group is selected yet (the admin belongs to several, so soleGroupId seeds nothing), and a
    // form with no group has no config to honour -- so ONLY the Group field renders rather than
    // guessing a field set it would have to flip the moment a group is chosen. The toHaveValue('')
    // is load-bearing: it pins that precondition, since an auto-filled group would silently turn
    // the five absence assertions below into vacuous truths.
    await expect(groupField).toHaveValue('');
    await expect(dialog.getByRole('combobox', { name: 'Paid By' })).toHaveCount(0);
    await expect(dialog.getByRole('combobox', { name: 'Status' })).toHaveCount(0);
    await expect(dialog.getByRole('combobox', { name: 'Categories' })).toHaveCount(0);
    await expect(dialog.getByRole('combobox', { name: 'Tags' })).toHaveCount(0);
    await expect(dialog.getByTestId('quick-scan-comment')).toHaveCount(0);

    // Select the injected-config group. Via the shared helper because the field
    // may already carry a value (it auto-fills for a single-group user), which
    // makes the input readonly until its X is clicked.
    await selectImageGroup(page, dialog, group.name);

    // Fields now reflect the injected config: paid-by + tags hidden, status + categories shown.
    await expect(dialog.getByRole('combobox', { name: 'Paid By' })).toHaveCount(0);
    await expect(dialog.getByRole('combobox', { name: 'Tags' })).toHaveCount(0);
    await expect(dialog.getByRole('combobox', { name: 'Status' })).toBeVisible();
    await expect(dialog.getByRole('combobox', { name: 'Categories' })).toBeVisible();

    // Categories is required and empty → submitting surfaces a validation error and queues nothing.
    await dialog.getByTestId('dialog-submit-button').click();
    await expect(
      page.getByText('Please fill in all required fields', { exact: false }),
    ).toBeVisible();
  });
});
