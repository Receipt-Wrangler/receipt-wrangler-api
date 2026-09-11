import { expect, test, type Page } from '@playwright/test';
import { creds, stubTokenRefresh } from './helpers/auth';
import {
  apiCreateGroup,
  apiCreateReceipt,
  apiDeleteGroupById,
  apiGetUserId,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';
import { gotoReceiptsTable } from './helpers/receipts-table';

// The receipts table's quick date filter and its filter chips, end to end.
//
// The Jest specs cover the month<->filter translation and the chip labels
// exhaustively, but they all run against a mocked store, so what only an e2e
// can prove is the wire:
//
//   - a month written by the stepper is a BETWEEN the SERVER understands, so
//     the table actually narrows to that month;
//   - the filter survives a reload. ReceiptTableState is persisted to
//     localStorage, which serializes every Date to an ISO string, so the
//     stepper has to read a month back out of strings it never wrote. Get that
//     wrong and the label silently degrades to "Custom" after every refresh —
//     the single most likely bug in this feature;
//   - the dialog and the chips share one filter, so a condition added in the
//     dialog is clearable from a chip without disturbing the month.
//
// Runs as admin and seeds its own group so the month buckets contain exactly
// this spec's receipts, whatever else the shared backend holds.

test.use({ storageState: 'e2e/.auth/admin.json' });

test.describe('Receipts quick date filter', () => {
  test.describe.configure({ mode: 'serial' });

  const thisMonthReceipt = uniqueName('qdf-this');
  const lastMonthReceipt = uniqueName('qdf-last');
  const olderReceipt = uniqueName('qdf-older');

  let group: { id: number; name: string };

  /** An ISO date in the middle of the month [delta] months from now. */
  function isoInMonth(delta: number): string {
    const now = new Date();
    return new Date(Date.UTC(now.getFullYear(), now.getMonth() + delta, 15)).toISOString();
  }

  function monthLabel(delta: number): string {
    const now = new Date();
    return new Date(now.getFullYear(), now.getMonth() + delta, 1).toLocaleString('en-US', {
      month: 'long',
      year: 'numeric',
    });
  }

  const stepperLabel = (page: Page) => page.getByTestId('receipts-month-label');
  const rowLink = (page: Page, name: string) => page.getByRole('link', { name });

  test.beforeAll(async () => {
    await withAdminApi(async (api) => {
      const adminId = await apiGetUserId(api, creds('admin').username);
      group = await apiCreateGroup(api, uniqueName('qdf-group'));

      for (const [name, delta] of [
        [thisMonthReceipt, 0],
        [lastMonthReceipt, -1],
        [olderReceipt, -2],
      ] as const) {
        await apiCreateReceipt(api, {
          groupId: group.id,
          paidByUserId: adminId,
          name,
          date: isoInMonth(delta),
        });
      }
    });
  });

  test.afterAll(async () => {
    await withAdminApi(async (api) => {
      try {
        await apiDeleteGroupById(api, String(group.id));
      } catch {
        // Best effort — a cleanup failure must not mask a real result.
      }
    });
  });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  async function gotoSeededGroup(page: Page): Promise<void> {
    await gotoReceiptsTable(page);
    await page.goto(`/receipts/group/${group.id}`);
    await expect(page.getByTestId('receipts-overflow-menu')).toBeVisible();
  }

  test('starts on All time with every receipt and no chips', async ({ page }) => {
    await gotoSeededGroup(page);

    await expect(stepperLabel(page)).toContainText('All time');
    await expect(page.locator('[data-testid^="receipt-filter-chip-"]')).toHaveCount(0);

    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
    await expect(rowLink(page, lastMonthReceipt)).toBeVisible();
    await expect(rowLink(page, olderReceipt)).toBeVisible();
  });

  test('narrows the table to a month and shows no duplicate date chip', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();

    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
    await expect(rowLink(page, lastMonthReceipt)).toHaveCount(0);
    await expect(rowLink(page, olderReceipt)).toHaveCount(0);

    // The stepper already names the month, so the Date chip stays suppressed…
    await expect(page.getByTestId('receipt-filter-chip-date')).toHaveCount(0);
    // …but it is still a filter, so the badge and the reset control show it.
    await expect(page.getByTestId('receipts-filter-reset')).toBeVisible();
  });

  test('steps a month at a time', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();
    await expect(stepperLabel(page)).toContainText(monthLabel(0));

    await page.getByTestId('receipts-month-prev').click();

    await expect(stepperLabel(page)).toContainText(monthLabel(-1));
    await expect(rowLink(page, lastMonthReceipt)).toBeVisible();
    await expect(rowLink(page, thisMonthReceipt)).toHaveCount(0);

    await page.getByTestId('receipts-month-next').click();
    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
  });

  // The filter is persisted, and JSON turns both Dates into ISO strings. If the
  // month is only recognised on Date instances this reads "Custom" after the
  // reload and the table silently keeps filtering by a month nothing names.
  test('keeps the month across a reload', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-last-month').click();
    await expect(stepperLabel(page)).toContainText(monthLabel(-1));

    await page.reload();

    await expect(stepperLabel(page)).toContainText(monthLabel(-1));
    await expect(rowLink(page, lastMonthReceipt)).toBeVisible();
    await expect(rowLink(page, thisMonthReceipt)).toHaveCount(0);
  });

  test('clears one dialog-set condition from its chip without losing the month', async ({
    page,
  }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();
    await expect(stepperLabel(page)).toContainText(monthLabel(0));

    // Add a second condition through the real filter dialog.
    await page.getByTestId('receipts-filter').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByLabel('Name').fill(thisMonthReceipt);
    await dialog.getByTestId('dialog-submit-button').click();
    await expect(dialog).toBeHidden();

    const nameChip = page.getByTestId('receipt-filter-chip-name');
    await expect(nameChip).toContainText(`Name contains ${thisMonthReceipt}`);
    // Still only the one chip — the month is the stepper's to show.
    await expect(page.getByTestId('receipt-filter-chip-date')).toHaveCount(0);

    await page.getByTestId('receipt-filter-chip-clear-name').click();

    await expect(nameChip).toHaveCount(0);
    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
  });

  // A Date filter the stepper cannot express has to stay visible and clearable.
  test('falls back to Custom with a date chip for a range that is not a month', async ({
    page,
  }) => {
    await gotoSeededGroup(page);

    await page.getByTestId('receipts-filter').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByLabel('Operation').first().click();
    await page.getByRole('option', { name: 'Within current month' }).click();
    await dialog.getByTestId('dialog-submit-button').click();
    await expect(dialog).toBeHidden();

    await expect(stepperLabel(page)).toContainText('Custom');
    await expect(page.getByTestId('receipt-filter-chip-date')).toContainText(
      'Date within current month',
    );

    await page.getByTestId('receipt-filter-chip-clear-date').click();
    await expect(stepperLabel(page)).toContainText('All time');
  });

  test('returns to all time from the panel', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();
    await expect(rowLink(page, olderReceipt)).toHaveCount(0);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-all-time').click();

    await expect(stepperLabel(page)).toContainText('All time');
    await expect(page.locator('[data-testid^="receipt-filter-chip-"]')).toHaveCount(0);
    await expect(rowLink(page, olderReceipt)).toBeVisible();
  });
});
