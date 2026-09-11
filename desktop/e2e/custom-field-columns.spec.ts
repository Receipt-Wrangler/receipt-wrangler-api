import { expect, Page, test } from '@playwright/test';
import { creds, stubTokenRefresh } from './helpers/auth';
import {
  apiCreateCustomField,
  apiCreateGroup,
  apiCreateReceipt,
  apiDeleteCustomFieldById,
  apiDeleteGroupById,
  apiGetUserId,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';

// Creating and deleting custom fields needs app.custom-fields.create/.delete, and
// the columns themselves are gated on app.custom-fields.read - all Legacy Admin.
test.use({ storageState: 'e2e/.auth/admin.json' });

// Serial: the tests share one seeded group, and the last one deletes a custom
// field the earlier ones assert on. Note the column configuration does NOT carry
// between them - it lives in localStorage and every test gets a fresh context -
// so each test enables the column it needs.
test.describe.configure({ mode: 'serial' });

/**
 * Custom fields as receipts-table columns.
 *
 * The Jest specs inject a column configuration into a mocked store, so they prove
 * nothing about the wire. What only an e2e can show: the field catalog reaching
 * the table through its resolver, a CURRENCY cell formatted by the configured
 * currency display, and - the part with a real backend behind it - the
 * `custom_<id>` sort key round-tripping to the API instead of being rejected.
 *
 * It seeds its own group so the rows are exactly these three, in a known order.
 */
test.describe('Receipts table — custom field columns', () => {
  const groupName = uniqueName('cf-col-grp');
  const tipName = uniqueName('Tip');
  const statusName = uniqueName('Approval');

  let groupId: number;
  let tipId: number;
  let statusId: number;

  test.beforeAll(async () => {
    await withAdminApi(async (api) => {
      const adminId = await apiGetUserId(api, creds('admin').username);
      groupId = (await apiCreateGroup(api, groupName)).id;

      tipId = (await apiCreateCustomField(api, { name: tipName, type: 'CURRENCY' })).id;
      const approval = await apiCreateCustomField(api, {
        name: statusName,
        type: 'SELECT',
        options: ['Zulu', 'Alpha'],
      });
      statusId = approval.id;

      const options = (await (
        await api.get(`/api/customField/${statusId}`)
      ).json()) as { options: { id: number; value: string }[] };
      const optionId = (value: string) =>
        options.options.find((option) => option.value === value)!.id;

      // Written so a text sort would order them 100, 20, 9 - the amounts are the
      // assertion, not decoration. The option ids run opposite to the alphabet for
      // the same reason.
      const receipts = [
        { name: 'cf-col-hundred', tip: '100.00', approval: 'Zulu' },
        { name: 'cf-col-twenty', tip: '20.00', approval: 'Alpha' },
        { name: 'cf-col-nine', tip: '9.00', approval: 'Zulu' },
      ];

      for (const receipt of receipts) {
        await apiCreateReceipt(api, {
          groupId,
          paidByUserId: adminId,
          name: receipt.name,
          customFields: [
            { customFieldId: tipId, currencyValue: receipt.tip },
            { customFieldId: statusId, selectValue: optionId(receipt.approval) },
          ],
        });
      }
    });
  });

  test.afterAll(async () => {
    try {
      await withAdminApi(async (api) => {
        // The group takes its receipts - and their custom field values - with it,
        // so the fields are only safe to drop afterwards.
        if (groupId) {
          await apiDeleteGroupById(api, String(groupId));
        }
        for (const id of [tipId, statusId]) {
          if (id) {
            await apiDeleteCustomFieldById(api, id);
          }
        }
      });
    } catch {
      // Best-effort cleanup: never mask a real failure with a teardown error.
    }
  });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  async function gotoTable(page: Page): Promise<void> {
    await page.goto(`/receipts/group/${groupId}`);
    await expect(page.getByTestId('configure-columns')).toBeVisible();
  }

  async function showColumn(page: Page, label: string): Promise<void> {
    await page.getByTestId('configure-columns').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByRole('checkbox', { name: label }).check();
    await dialog.getByTestId('dialog-submit-button').click();
    await expect(dialog).toBeHidden();
    await expect(page.getByRole('columnheader', { name: label })).toBeVisible();
  }

  /** The rendered row order, which is what every sort assertion compares. */
  async function receiptNames(page: Page): Promise<string[]> {
    const names = await page
      .locator('td a[href^="/receipts/"]')
      .allTextContents();
    return names.map((name) => name.trim());
  }

  test('lists every custom field after the built-in columns, unchecked', async ({ page }) => {
    await gotoTable(page);
    await page.getByTestId('configure-columns').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    const labels = await dialog.locator('mat-checkbox').allInnerTexts();
    const trimmed = labels.map((label) => label.trim());

    expect(trimmed).toContain(tipName);
    expect(trimmed).toContain(statusName);
    // After the built-ins, and after Resolved Date specifically.
    expect(trimmed.indexOf(tipName)).toBeGreaterThan(trimmed.indexOf('Resolved Date'));

    // Hidden by default: creating a custom field must not widen everyone's table.
    await expect(dialog.getByRole('checkbox', { name: tipName })).not.toBeChecked();
    await expect(page.getByRole('columnheader', { name: tipName })).toHaveCount(0);
  });

  test('badges the custom fields and nothing else', async ({ page }) => {
    await gotoTable(page);
    await page.getByTestId('configure-columns').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();

    // Custom fields are global, so other specs' fields are in this list too -
    // assert on the rows rather than on a total. Exactly the nine built-ins carry
    // no badge; everything past them is a custom field and is badged.
    const rows = dialog.locator('.column-item');
    const badged = dialog.getByTestId('column-config-custom');
    expect((await rows.count()) - (await badged.count())).toEqual(9);

    const tipRow = rows.filter({ hasText: tipName });
    await expect(tipRow.getByTestId('column-config-custom')).toHaveText('Custom');
    await expect(
      rows.filter({ hasText: 'Resolved Date' }).getByTestId('column-config-custom')
    ).toHaveCount(0);

    // The badge must stay out of the checkbox's accessible name, or every
    // locator that picks a column by its field name stops resolving.
    await expect(dialog.getByRole('checkbox', { name: tipName, exact: true })).toBeVisible();
  });

  test('renders a currency custom field through the configured currency display', async ({ page }) => {
    await gotoTable(page);
    await showColumn(page, tipName);

    await expect(page.getByRole('columnheader', { name: tipName })).toBeVisible();

    // Separator-tolerant: the currency configuration is a global System Setting on
    // a shared backend that this spec must not mutate. The symbol and the trailing
    // zero are what prove the value went through customCurrency rather than being
    // printed raw (the API sends the string "100").
    await expect(page.getByRole('cell', { name: /^\D?100[.,]00\D?$/ })).toBeVisible();
  });

  test('sorts numerically on a currency custom field', async ({ page }) => {
    await gotoTable(page);
    await showColumn(page, tipName);

    // A 500 here is the whole point: the sort key is `custom_<id>`, which the API
    // rejects outright unless it knows how to read it.
    await page.getByRole('columnheader', { name: tipName }).click();
    await expect
      .poll(() => receiptNames(page))
      .toEqual(['cf-col-nine', 'cf-col-twenty', 'cf-col-hundred']);

    await page.getByRole('columnheader', { name: tipName }).click();
    await expect
      .poll(() => receiptNames(page))
      .toEqual(['cf-col-hundred', 'cf-col-twenty', 'cf-col-nine']);
  });

  test('sorts a select custom field by the option text it shows', async ({ page }) => {
    await gotoTable(page);
    await showColumn(page, statusName);

    await expect(page.getByRole('cell', { name: 'Alpha', exact: true })).toBeVisible();

    // "Alpha" is the higher option id, so sorting on the id would put it last.
    // The two "Zulu" rows tie, and the API breaks ties by receipt id descending -
    // so the later-created cf-col-nine precedes cf-col-hundred. That ordering is
    // the tiebreaker, without which paging over a handful of distinct values
    // repeats and skips rows.
    await page.getByRole('columnheader', { name: statusName }).click();
    await expect.poll(() => receiptNames(page)).toEqual([
      'cf-col-twenty',
      'cf-col-nine',
      'cf-col-hundred',
    ]);
  });

  test('drops the column when its custom field is deleted', async ({ page }) => {
    await gotoTable(page);
    await showColumn(page, statusName);

    // Sorted by the column that is about to disappear, so the reload has to heal
    // the stored sort as well as the stored column.
    await page.getByRole('columnheader', { name: statusName }).click();
    await expect.poll(() => receiptNames(page)).toHaveLength(3);

    await withAdminApi(async (api) => {
      await apiDeleteCustomFieldById(api, statusId);
    });
    statusId = 0;

    // The configuration still names the deleted column in localStorage. It must
    // heal rather than leave mat-table displaying an id it cannot resolve - and the
    // table was sorted by it, so the request must not go out asking for it either.
    await gotoTable(page);

    await expect(page.getByRole('columnheader', { name: statusName })).toHaveCount(0);
    await expect.poll(() => receiptNames(page)).toHaveLength(3);
  });
});
