import { expect, Locator, Page, test } from '@playwright/test';
import { stubTokenRefresh } from './helpers/auth';

// An API key's scope (r / w / rw) is enforced on every request, so the owner
// has to be able to pick it — on create AND on edit, since a key created before
// the picker was offered can only be fixed by editing it. The Jest specs drive
// the form against a mocked ApiKeyService; this spec proves the chosen scope
// goes out on the wire, is stored, and comes back into the table and the form.

const API_KEYS_URL = '/settings/api-keys/view';

function keyRow(page: Page, name: string): Locator {
  return page.getByRole('row').filter({ hasText: name }).first();
}

// The option panel is an overlay attached to the page root, not the dialog.
async function pickScope(page: Page, dialog: Locator, label: string) {
  await dialog.getByRole('combobox', { name: 'Scope' }).click();
  await page.getByRole('option', { name: label, exact: true }).click();
}

test.describe('API key scope', () => {
  // Default project storageState = e2e-user, who owns the key and holds
  // app.api-keys.update.
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  const keyName = `e2e-apikey-scope-${Date.now()}`;

  test('a key is created with the scope picked in the form', async ({
    page,
  }) => {
    await page.goto(API_KEYS_URL);
    await page.getByTestId('api-key-add').click();
    const dialog = page.getByRole('dialog').filter({ hasText: 'Create API Key' });
    await expect(dialog).toBeVisible();

    await expect(dialog.getByTestId('api-key-scope-hint')).toBeVisible();
    await dialog.getByLabel('Name').fill(keyName);
    await pickScope(page, dialog, 'Read/Write');
    await dialog.getByTestId('dialog-submit-button').click();
    await dialog.getByRole('button', { name: 'Close' }).click();
    await expect(dialog).toBeHidden();

    await expect(keyRow(page, keyName).getByTestId('api-key-scope')).toHaveText(
      'Read/Write',
    );
  });

  test('editing a key changes its scope', async ({ page }) => {
    await page.goto(API_KEYS_URL);
    await keyRow(page, keyName).getByTestId('api-key-edit').click();
    const dialog = page.getByRole('dialog').filter({ hasText: `Edit ${keyName}` });
    await expect(dialog).toBeVisible();
    await expect(dialog.getByRole('combobox', { name: 'Scope' })).toHaveText(
      'Read/Write',
    );

    await pickScope(page, dialog, 'Write');
    await dialog.getByTestId('dialog-submit-button').click();
    await expect(dialog).toBeHidden();

    // Reload so the table and the form both have to read the scope back out
    // of the server rather than out of the dialog that just closed.
    await page.reload();
    await expect(keyRow(page, keyName).getByTestId('api-key-scope')).toHaveText(
      'Write',
    );

    await keyRow(page, keyName).getByTestId('api-key-edit').click();
    await expect(dialog.getByRole('combobox', { name: 'Scope' })).toHaveText(
      'Write',
    );
  });

  test.afterAll(async ({ browser }) => {
    const context = await browser.newContext({
      storageState: 'e2e/.auth/user.json',
    });
    const page = await context.newPage();
    try {
      await stubTokenRefresh(page);
      await page.goto(API_KEYS_URL);
      // Wait for the row rather than counting rows: the table loads after
      // navigation, so an immediate count reads 0 and skips the cleanup.
      const row = keyRow(page, keyName);
      await row.waitFor({ timeout: 10_000 });
      await row.getByTestId('api-key-delete').click();
      // Wait on the request itself: while the confirmation is open MatDialog
      // hides the table from the accessibility tree, so "no matching row"
      // would pass whether or not anything was deleted.
      const deleted = page.waitForResponse(
        (response) =>
          response.request().method() === 'DELETE' &&
          response.url().includes('/api/apiKey/'),
      );
      await page.getByRole('dialog').getByTestId('dialog-submit-button').click();
      await deleted;
    } catch {
      // Best-effort cleanup — don't mask a test failure.
    } finally {
      await context.close();
    }
  });
});
