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
  // Created RESOLVED, so the server stamps resolved_date to now: its Resolved
  // Date is in the current month while its Date is two months back. That split
  // is what tells the two columns apart on the wire.
  const resolvedReceipt = uniqueName('qdf-resolved');

  let group: { id: number; name: string };

  /**
   * One instant for the whole suite. The seeded receipt dates, the expected
   * labels and the browser's own clock all derive from it, so a run that
   * crosses a month boundary cannot leave the app a month ahead of its fixture.
   */
  const REFERENCE_TIME = new Date();

  /** An ISO date in the middle of the month [delta] months from the reference. */
  function isoInMonth(delta: number): string {
    return new Date(
      Date.UTC(REFERENCE_TIME.getFullYear(), REFERENCE_TIME.getMonth() + delta, 15),
    ).toISOString();
  }

  function monthLabel(delta: number): string {
    return new Date(
      REFERENCE_TIME.getFullYear(),
      REFERENCE_TIME.getMonth() + delta,
      1,
    ).toLocaleString('en-US', { month: 'long', year: 'numeric' });
  }

  const stepperLabel = (page: Page) => page.getByTestId('receipts-month-label');
  const fieldPicker = (page: Page) => page.getByTestId('receipts-quick-date-field');
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

      await apiCreateReceipt(api, {
        groupId: group.id,
        paidByUserId: adminId,
        name: resolvedReceipt,
        date: isoInMonth(-2),
        status: 'RESOLVED',
      });
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
    // Pin the app's clock to the same instant. The component reads its own
    // new Date() for This month / Last month and for an arrow press with
    // nothing selected, so a spec-side reference alone would still disagree
    // with it across a month boundary. setFixedTime only changes what Date
    // returns — timers keep running, so animations are unaffected.
    await page.clock.setFixedTime(REFERENCE_TIME);
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

  test('narrows the table to a month and chips the date condition', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();

    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
    await expect(rowLink(page, lastMonthReceipt)).toHaveCount(0);
    await expect(rowLink(page, olderReceipt)).toHaveCount(0);

    // The chip row names which date column is filtered — the stepper's label
    // only says the month, and the target field is selectable.
    await expect(page.getByTestId('receipt-filter-chip-date')).toContainText('Receipt Date');
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
    // Both conditions are chipped; clearing one must not touch the other.
    await expect(page.getByTestId('receipt-filter-chip-date')).toBeVisible();

    await page.getByTestId('receipt-filter-chip-clear-name').click();

    await expect(nameChip).toHaveCount(0);
    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
  });

  // The stepper's target field is selectable, and only an e2e proves a chosen
  // field reaches the SERVER as a different column — the Jest specs assert
  // against a mocked store. It also pins the pointer's persistence, the same
  // reload trap the month itself has.
  test('filters on the chosen date field and keeps it across a reload', async ({ page }) => {
    await gotoSeededGroup(page);

    // Start on Date, this month: the resolved receipt is dated two months back,
    // so it is excluded.
    await expect(fieldPicker(page)).toContainText('Receipt Date');
    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();
    await expect(rowLink(page, thisMonthReceipt)).toBeVisible();
    await expect(rowLink(page, resolvedReceipt)).toHaveCount(0);
    await expect(page.getByTestId('receipt-filter-chip-date')).toBeVisible();

    // Switching the field changes no condition, only which one the stepper
    // describes — so the Date filter is still applied and still chipped.
    await fieldPicker(page).click();
    await page.getByTestId('receipts-quick-date-field-resolvedDate').click();

    await expect(fieldPicker(page)).toContainText('Resolved Date');
    await expect(stepperLabel(page)).toContainText('All time');
    await expect(page.getByTestId('receipt-filter-chip-date')).toBeVisible();

    await page.getByTestId('receipt-filter-chip-clear-date').click();
    await expect(page.getByTestId('receipt-filter-chip-date')).toHaveCount(0);

    // Now the month lands on resolved_date, which only the RESOLVED receipt has.
    await stepperLabel(page).click();
    await page.getByTestId('month-stepper-this-month').click();

    await expect(rowLink(page, resolvedReceipt)).toBeVisible();
    await expect(rowLink(page, thisMonthReceipt)).toHaveCount(0);
    await expect(page.getByTestId('receipt-filter-chip-resolvedDate')).toBeVisible();

    await page.reload();

    await expect(fieldPicker(page)).toContainText('Resolved Date');
    await expect(stepperLabel(page)).toContainText(monthLabel(0));
    await expect(rowLink(page, resolvedReceipt)).toBeVisible();
    await expect(rowLink(page, thisMonthReceipt)).toHaveCount(0);
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
      'Receipt Date within current month',
    );

    await page.getByTestId('receipt-filter-chip-clear-date').click();
    await expect(stepperLabel(page)).toContainText('All time');
  });

  // The panel is a dialog, not a menu. Under the mat-menu it was mouse-only:
  // with no mat-menu-items the key manager was empty so arrows did nothing, and
  // ListKeyManager turns Tab into tabOut, which MatMenu wires to close — so Tab
  // dismissed the panel instead of entering it.
  test('is operable by keyboard and returns focus to the trigger', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).locator('button').focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeVisible();

    // Tab must move INTO the panel, not close it.
    await page.keyboard.press('Tab');
    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeVisible();
    expect(
      await page.evaluate(() => !!document.activeElement?.closest('[role="dialog"]')),
    ).toBe(true);

    // Reach a month button by keyboard and pick it.
    await page.getByTestId('month-stepper-month-0').focus();
    await page.keyboard.press('Enter');

    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeHidden();
    await expect(stepperLabel(page)).toContainText('January');

    // The focus trap hands focus back to the control that opened it.
    expect(
      await page.evaluate(
        () => document.activeElement?.closest('[data-testid]')?.getAttribute('data-testid'),
      ),
    ).toEqual('receipts-month-label');
  });

  test('closes on Escape without changing the filter', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeVisible();

    await page.keyboard.press('Escape');

    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeHidden();
    await expect(stepperLabel(page)).toContainText('All time');
  });

  // Paging the year used to need a stopPropagation hack, because a mat-menu
  // closes on any click inside it.
  test('stays open while paging the year', async ({ page }) => {
    await gotoSeededGroup(page);

    await stepperLabel(page).click();
    const year = await page.getByTestId('month-stepper-year').textContent();

    await page.getByTestId('month-stepper-year-prev').click();

    await expect(page.getByRole('dialog', { name: 'Filter by month' })).toBeVisible();
    await expect(page.getByTestId('month-stepper-year')).toHaveText(
      String(Number(year) - 1),
    );
    // Paging is a view concern — it must not have written a filter.
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
