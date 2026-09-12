import { expect, test, type Page } from '@playwright/test';
import { creds, stubTokenRefresh } from './helpers/auth';
import {
  apiGetUserId,
  apiPagedSystemTasks,
  apiRecordApiKeyDeletedSystemTask,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';

// The System Tasks table's filter, end to end.
//
// The Jest specs cover the chip labels, the state actions and the dialog
// against a mocked store, so what only an e2e can prove is the wire:
//
//   - the filter reaches the API and the SERVER narrows (totalCount included,
//     which is what proves the predicates land before the count);
//   - the "System" option really matches the rows with no ranByUserId — the
//     majority of the table, and unreachable without the sentinel;
//   - a Started At of the task's own day matches it while the previous day does
//     not, which is the whole-day widening the timestamp columns need;
//   - the filter survives a reload, since SystemTaskTableState is persisted to
//     localStorage and serializes every Date to an ISO string.
//
// Runs as admin (the page is gated on app.system-tasks.read) and seeds its own
// task, so the assertions do not depend on what else the shared backend holds.

test.use({ storageState: 'e2e/.auth/admin.json' });

test.describe('System tasks filter', () => {
  test.describe.configure({ mode: 'serial' });

  const apiKeyName = uniqueName('sys-task-filter');
  let adminUserId: number;

  test.beforeAll(async () => {
    await withAdminApi(async (api) => {
      adminUserId = await apiGetUserId(api, creds('admin').username);
      // Records exactly one API_KEY_DELETED task, ran by the admin, right now.
      await apiRecordApiKeyDeletedSystemTask(api, apiKeyName);
    });
  });

  const gotoSystemTasks = async (page: Page): Promise<void> => {
    await stubTokenRefresh(page);
    await page.goto('/system-settings/system-tasks');
    await expect(page.getByTestId('system-tasks-filter')).toBeVisible();
  };

  const isoDay = (offsetDays: number): string => {
    const date = new Date();
    date.setDate(date.getDate() + offsetDays);
    date.setHours(0, 0, 0, 0);
    return date.toISOString();
  };

  test('the server narrows by type, and only by types it can return', async () => {
    await withAdminApi(async (api) => {
      const unfiltered = await apiPagedSystemTasks(api);
      const byType = await apiPagedSystemTasks(api, {
        type: { operation: 'CONTAINS', value: ['API_KEY_DELETED'] },
      });

      expect(byType.totalCount).toBeGreaterThan(0);
      expect(byType.totalCount).toBeLessThanOrEqual(unfiltered.totalCount);
      expect(byType.data.every((task) => task.type === 'API_KEY_DELETED')).toBe(true);

      // The three child-only types never appear as top-level rows, which is why
      // the Type picker omits them.
      const childOnly = await apiPagedSystemTasks(api, {
        type: { operation: 'CONTAINS', value: ['RECEIPT_UPLOADED'] },
      });
      expect(childOnly.totalCount).toBe(0);
    });
  });

  test('the System option matches unattributed tasks and excludes the admin\'s', async () => {
    await withAdminApi(async (api) => {
      const byAdmin = await apiPagedSystemTasks(api, {
        ranBy: { operation: 'CONTAINS', value: [adminUserId] },
      });
      expect(byAdmin.totalCount).toBeGreaterThan(0);
      expect(byAdmin.data.every((task) => task.ranByUserId === adminUserId)).toBe(true);

      // -1 is SYSTEM_RAN_BY_OPTION_ID: the rows the table labels "System".
      const bySystem = await apiPagedSystemTasks(api, {
        ranBy: { operation: 'CONTAINS', value: [-1] },
      });
      expect(bySystem.data.every((task) => !task.ranByUserId)).toBe(true);

      const both = await apiPagedSystemTasks(api, {
        ranBy: { operation: 'CONTAINS', value: [-1, adminUserId] },
      });
      expect(both.totalCount).toBe(byAdmin.totalCount + bySystem.totalCount);
    });
  });

  test('a Started At of today matches the seeded task but yesterday does not', async () => {
    await withAdminApi(async (api) => {
      const seeded = { type: { operation: 'CONTAINS', value: ['API_KEY_DELETED'] } };

      // EQUALS on a timestamp column only ever matches if the server widens the
      // picked date to the whole calendar day.
      const today = await apiPagedSystemTasks(api, {
        ...seeded,
        startedAt: { operation: 'EQUALS', value: isoDay(0) },
      });
      expect(today.totalCount).toBeGreaterThan(0);

      const yesterday = await apiPagedSystemTasks(api, {
        ...seeded,
        startedAt: { operation: 'EQUALS', value: isoDay(-1) },
      });
      expect(yesterday.data.some((task) => task.resultDescription?.includes(apiKeyName))).toBe(false);
    });
  });

  test('applying a filter in the dialog narrows the table, renders a chip and survives a reload', async ({
    page,
  }) => {
    await gotoSystemTasks(page);

    // No conditions yet, so no chips and no reset control.
    await expect(page.locator('mat-chip[data-testid^="system-task-filter-chip-"]')).toHaveCount(0);
    await expect(page.getByTestId('system-tasks-filter-reset')).toBeHidden();

    await page.getByTestId('system-tasks-filter').click();
    await expect(page.getByRole('heading', { name: 'Filter System Tasks' })).toBeVisible();

    // getByLabel matches the autocomplete's listbox as well as its input, and a
    // multi-select panel stays open over the footer after a pick — so the row
    // is driven by role and the panel is dismissed by clicking the heading.
    await page.getByRole('combobox', { name: 'Type' }).click();
    await page.getByRole('option', { name: 'API Key Deleted', exact: true }).click();
    await page.getByRole('heading', { name: 'Filter System Tasks' }).click();

    const request = page.waitForRequest(
      (req) => req.url().includes('/api/systemTask/getPagedSystemTasks') && req.method() === 'POST',
    );
    await page.getByTestId('dialog-submit-button').click();

    // The condition reaches the server, not just the store.
    const body = (await (await request).postDataJSON()) as any;
    expect(body.filter.type.value).toEqual(['API_KEY_DELETED']);
    expect(body.page).toBe(1);

    const chip = page.getByTestId('system-task-filter-chip-type');
    await expect(chip).toBeVisible();
    await expect(chip).toContainText('Type contains API Key Deleted');

    // Every visible row is the filtered type.
    const typeCells = page.locator('table tbody tr:not(.example-detail-row) td:nth-child(3)');
    await expect(typeCells.first()).toBeVisible();
    for (const cell of await typeCells.all()) {
      await expect(cell).toContainText('API Key Deleted');
    }

    // Persisted to localStorage, so it has to come back after a reload.
    await page.reload();
    await expect(page.getByTestId('system-task-filter-chip-type')).toContainText(
      'Type contains API Key Deleted',
    );

    // A chip clears its own condition and nothing else.
    await page.getByTestId('system-task-filter-chip-clear-type').click();
    await expect(page.locator('mat-chip[data-testid^="system-task-filter-chip-"]')).toHaveCount(0);
    await expect(page.getByTestId('system-tasks-filter-reset')).toBeHidden();
  });
});
