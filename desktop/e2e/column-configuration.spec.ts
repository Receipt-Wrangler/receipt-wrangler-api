import { expect, test } from '@playwright/test';
import { stubTokenRefresh } from './helpers/auth';
import { gotoReceiptsTable, openReceiptsOverflowMenu } from './helpers/receipts-table';

// Configuring the receipts-table columns regressed with
// "TypeError: 0 is read-only": the dialog sorted the (frozen, dev-mode) NGXS
// snapshot in place, so ngOnInit threw and the dialog never rendered. These
// tests open the dialog (the core regression) and round-trip a couple of
// configurations. The default chromium project uses e2e/.auth/user.json
// (e2e-user = Legacy User), which can read its own group's receipts.

test.describe('Receipt table column configuration', () => {
  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  test('opens the dialog without crashing', async ({ page }) => {
    await gotoReceiptsTable(page);

    await openReceiptsOverflowMenu(page);
    await page.getByTestId('configure-columns').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    // Pre-fix, ngOnInit threw on the frozen snapshot and this content never
    // rendered.
    await expect(
      dialog.getByText('Select which columns to display'),
    ).toBeVisible();
    await expect(dialog.locator('mat-checkbox').first()).toBeVisible();
  });

  test('hides a column on save and restores it via reset', async ({ page }) => {
    await gotoReceiptsTable(page);

    // "Added At" (created_at) is visible by default and maps to the first
    // dialog checkbox.
    const addedAtHeader = page.getByRole('columnheader', { name: 'Added At' });
    await expect(addedAtHeader).toBeVisible();

    // Open the dialog and toggle "Added At" off, then save.
    await openReceiptsOverflowMenu(page);
    await page.getByTestId('configure-columns').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    const addedAtCheckbox = dialog.getByRole('checkbox', { name: 'Added At' });
    await addedAtCheckbox.uncheck();
    await expect(addedAtCheckbox).not.toBeChecked();
    await dialog.getByTestId('dialog-submit-button').click();

    await expect(dialog).toBeHidden();
    await expect(addedAtHeader).toHaveCount(0);

    // Reopen and reset back to the defaults — the column returns.
    await openReceiptsOverflowMenu(page);
    await page.getByTestId('configure-columns').click();
    await expect(dialog).toBeVisible();
    await dialog.getByTestId('reset-columns').click();
    await dialog.getByTestId('dialog-submit-button').click();

    await expect(dialog).toBeHidden();
    await expect(addedAtHeader).toBeVisible();
  });
});
