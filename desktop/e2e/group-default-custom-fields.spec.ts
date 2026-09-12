import { expect, test, type Page } from '@playwright/test';
import { creds, stubTokenRefresh } from './helpers/auth';
import {
  apiCreateCustomField,
  apiCreateGroup,
  apiCreateReceipt,
  apiDeleteCustomFieldById,
  apiDeleteGroupById,
  apiGetUserId,
  apiSetGroupDefaultCustomFields,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';

// Group default custom fields, end to end (see desktop/CLAUDE.md → "Per-group default custom
// fields"). The Jest specs cover the swap rules exhaustively against a mocked store; what only an
// e2e can prove is that the two delivery paths agree with the server:
//
//   - the settings page reads GET /group/{id} through groupResolverFn, so test 1 is a real
//     persistence round-trip;
//   - the receipt form reads GroupState, hydrated from AppData on every navigation, so test 2
//     proves the ids reach a different consumer over a different endpoint;
//   - a saved receipt picks its group's defaults up on load, so test 4 proves the retro-fit
//     survives a real save (enforceReceiptCustomFieldSelection would 403 a mismatched id set).
//
// Runs as admin: the section needs group.update + app.custom-fields.read, and seeding custom
// fields needs app.custom-fields.create/delete.

test.use({ storageState: 'e2e/.auth/admin.json' });

test.describe('Group default custom fields', () => {
  test.describe.configure({ mode: 'serial' });

  // Names deliberately end in distinct words: `getByLabel` matches substrings, so a name that is a
  // prefix of another would make every "this field is mounted" assertion ambiguous.
  const nameA = uniqueName('dcf-fielda');
  const nameB = uniqueName('dcf-fieldb');
  const nameC = uniqueName('dcf-fieldc');

  let fieldA: { id: number; name: string };
  let fieldB: { id: number; name: string };
  let fieldC: { id: number; name: string };
  let alpha: { id: number; name: string };
  let beta: { id: number; name: string };

  test.beforeAll(async () => {
    await withAdminApi(async (api) => {
      fieldA = await apiCreateCustomField(api, { name: nameA, type: 'TEXT' });
      fieldB = await apiCreateCustomField(api, { name: nameB, type: 'TEXT' });
      fieldC = await apiCreateCustomField(api, { name: nameC, type: 'TEXT' });
      alpha = await apiCreateGroup(api, uniqueName('dcf-alpha'));
      beta = await apiCreateGroup(api, uniqueName('dcf-beta'));
    });
  });

  test.afterAll(async () => {
    await withAdminApi(async (api) => {
      // Groups first: deleting a custom field destroys every value stored against it, so the
      // receipts holding those values have to be gone (cascaded with their group) first.
      await apiDeleteGroupById(api, String(alpha.id));
      await apiDeleteGroupById(api, String(beta.id));
      for (const field of [fieldA, fieldB, fieldC]) {
        if (field) {
          await apiDeleteCustomFieldById(api, field.id);
        }
      }
    });
  });

  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  /** The "Default Custom Fields" multi-select on the group receipt settings page. */
  function defaultFieldsPicker(page: Page) {
    return page.getByTestId('group-default-custom-fields');
  }

  /**
   * Adds [name] to the multi-select by filtering for it and picking the offered option.
   *
   * The trailing Escape closes the autocomplete panel. A multi-select keeps the input focused
   * after a pick, and the open overlay sits above the page's Save button — Playwright then waits
   * out the whole timeout on "subtree intercepts pointer events" instead of clicking it.
   */
  async function pickDefaultField(page: Page, name: string) {
    const input = defaultFieldsPicker(page).getByRole('combobox');
    await input.click();
    await input.fill(name);
    await page.getByRole('option', { name, exact: true }).click();
    await expect(chip(page, name)).toBeVisible();
    await page.keyboard.press('Escape');
  }

  /** Removes [name]'s chip from the multi-select. Escapes the panel for the same reason as above:
   *  mat-chip-row hands focus back to the input on removal, which re-opens the autocomplete. */
  async function removeDefaultField(page: Page, name: string) {
    await chip(page, name).getByRole('button').click();
    await expect(chip(page, name)).toHaveCount(0);
    await page.keyboard.press('Escape');
  }

  /** The selected chip for [name], if the picker currently holds it. */
  function chip(page: Page, name: string) {
    return defaultFieldsPicker(page).locator('mat-chip-row').filter({ hasText: name });
  }

  async function saveSettings(page: Page) {
    await page.getByRole('button', { name: 'Save', exact: true }).first().click();
    await page.waitForURL(/\/receipt-settings\/view/);
  }

  /** Asserts the receipt form's mounted custom-field set is exactly [names] — count included, so
   *  a duplicated field fails as loudly as a missing one. */
  async function expectFields(page: Page, names: string[]) {
    await expect(page.locator('app-custom-field')).toHaveCount(names.length);
    for (const name of names) {
      await expect(page.getByLabel(name)).toBeVisible();
    }
  }

  test('the picker persists a group default set and its ingest toggle', async ({ page }) => {
    await page.goto(`/groups/${alpha.id}/receipt-settings/edit`);
    await expect(defaultFieldsPicker(page)).toBeVisible();

    await pickDefaultField(page, nameA);
    await pickDefaultField(page, nameB);
    await page.getByTestId('apply-default-custom-fields-on-ingest').getByRole('checkbox').check();
    await saveSettings(page);

    // Re-navigate rather than assert in place: the value has to survive a fresh resolver fetch of
    // GET /group/{id}, which is the only thing that proves the ids were persisted and hydrated
    // back rather than just echoed out of the component's own form.
    await page.goto(`/groups/${alpha.id}/receipt-settings/edit`);
    await expect(defaultFieldsPicker(page).locator('mat-chip-row')).toHaveCount(2);
    await expect(chip(page, nameA)).toBeVisible();
    await expect(chip(page, nameB)).toBeVisible();
    await expect(
      page.getByTestId('apply-default-custom-fields-on-ingest').getByRole('checkbox'),
    ).toBeChecked();

    // Removing a chip persists too — the command carries the whole id set, so a shrunken set must
    // shrink the stored one rather than being treated as "leave unchanged".
    await removeDefaultField(page, nameB);
    await saveSettings(page);

    await page.goto(`/groups/${alpha.id}/receipt-settings/edit`);
    await expect(defaultFieldsPicker(page).locator('mat-chip-row')).toHaveCount(1);
    await expect(chip(page, nameA)).toBeVisible();
  });

  test('a deleted custom field drops out of every group default set', async ({ page }) => {
    const doomed = await withAdminApi((api) =>
      apiCreateCustomField(api, { name: uniqueName('dcf-doomed'), type: 'TEXT' }),
    );
    await withAdminApi((api) =>
      apiSetGroupDefaultCustomFields(api, alpha.id, [fieldA.id, doomed.id]),
    );

    await page.goto(`/groups/${alpha.id}/receipt-settings/edit`);
    await expect(defaultFieldsPicker(page).locator('mat-chip-row')).toHaveCount(2);

    await withAdminApi((api) => apiDeleteCustomFieldById(api, doomed.id));

    await page.goto(`/groups/${alpha.id}/receipt-settings/edit`);
    await expect(defaultFieldsPicker(page).locator('mat-chip-row')).toHaveCount(1);
    await expect(chip(page, fieldA.name)).toBeVisible();
    await expect(chip(page, doomed.name)).toHaveCount(0);
  });

  test('the receipt form applies a group default set and smart-swaps on a group change', async ({
    page,
  }) => {
    await withAdminApi(async (api) => {
      await apiSetGroupDefaultCustomFields(api, alpha.id, [fieldA.id, fieldB.id]);
      await apiSetGroupDefaultCustomFields(api, beta.id, [fieldC.id]);
    });

    /**
     * Selects [name] in the receipt form's Group autocomplete.
     *
     * The clear button first: a single-select `app-autocomlete` marks its input `readonly` once a
     * value is chosen, so typing into it or clicking it does nothing and the group would silently
     * stay put — every downstream assertion would then be made against the old group.
     */
    async function selectGroup(name: string) {
      const field = page.getByTestId('receipt-group');
      const clear = field.getByTestId('autocomplete-clear');
      if ((await clear.count()) > 0) {
        await clear.getByRole('button').click();
      }
      await field.getByRole('combobox').click();
      await page.getByRole('option', { name, exact: true }).click();
      await expect(field.getByTestId('autocomplete-clear')).toBeVisible();
    }

    /** Attaches [name] by hand through the "Manage custom fields" menu. */
    async function addFieldByHand(name: string) {
      // .locator('button'), not getByRole('button'): the cdkMenuTrigger puts role="button" on the
      // <app-button> host as well, so the role query is a strict-mode violation with two matches.
      await page.getByTestId('receipt-manage-custom-fields').locator('button').click();
      await page.getByLabel('Filter custom fields').fill(name);
      await page.getByRole('menuitem').filter({ hasText: name }).click();
      await page.keyboard.press('Escape');
      await expect(page.getByLabel(name)).toBeVisible();
    }

    const receiptName = uniqueName('dcf-receipt');
    await page.goto('/receipts/add');
    await expect(page.getByLabel('Name')).toBeVisible();
    await page.getByLabel('Name').fill(receiptName);
    await page.getByLabel('Amount').fill('42.00');

    // Alpha declares A + B, so both are pre-added on the create form.
    await selectGroup(alpha.name);
    await expectFields(page, [nameA, nameB]);

    // Typing into A makes it the user's data; C is added by hand, so it is never auto-managed.
    await page.getByLabel(nameA).fill('kept-A');
    await addFieldByHand(nameC);
    await expectFields(page, [nameA, nameB, nameC]);

    // Beta declares only C. B is the only field the swap owns AND is still empty, so it is the
    // only one dropped; C is already attached and must not be added a second time.
    await selectGroup(beta.name);
    await expectFields(page, [nameA, nameC]);
    await expect(page.getByLabel(nameA)).toHaveValue('kept-A');

    // Back to Alpha: B returns as a default and must come back BLANK (the form drops an
    // auto-applied field's stored value on removal), C stays because the user added it.
    await selectGroup(alpha.name);
    await expectFields(page, [nameA, nameB, nameC]);
    await expect(page.getByLabel(nameA)).toHaveValue('kept-A');
    await expect(page.getByLabel(nameB)).toHaveValue('');

    // Paid-by last: changing the group clears it.
    await page.getByLabel(nameB).fill('typed-B');
    await page.getByLabel('Paid By').click();
    await page.getByRole('option').first().click();

    await page.getByRole('button', { name: 'Save', exact: true }).first().click({ force: true });
    await page.waitForURL(/\/receipts\/\d+\/view/);

    // The save had to survive enforceReceiptCustomFieldSelection, which 403s a submitted id set
    // that doesn't match the stored one — so reaching /view at all proves the client sent every
    // attached field, empty ones included.
    await expect(page.getByLabel(nameA)).toHaveValue('kept-A');
    await expect(page.getByLabel(nameB)).toHaveValue('typed-B');
    await expect(page.getByLabel(nameC)).toBeVisible();
  });

  test('a saved receipt picks up a default its group gained afterwards', async ({ page }) => {
    // The receipt is created while beta declares nothing, so the field can only be on the form
    // because the form put it there on load.
    let receiptId: number;
    const receiptName = uniqueName('dcf-existing');
    await withAdminApi(async (api) => {
      await apiSetGroupDefaultCustomFields(api, beta.id, []);
      receiptId = await apiCreateReceipt(api, {
        groupId: beta.id,
        paidByUserId: await apiGetUserId(api, creds('admin').username),
        name: receiptName,
      });
      await apiSetGroupDefaultCustomFields(api, beta.id, [fieldA.id]);
    });

    // View mode: blank and read-only, exactly like a built-in field with no value.
    await page.goto(`/receipts/${receiptId!}/view`);
    await expect(page.getByLabel('Name')).toHaveValue(receiptName);
    await expectFields(page, [nameA]);
    await expect(page.getByLabel(nameA)).toHaveValue('');

    await page.goto(`/receipts/${receiptId!}/edit`);
    await expect(page.getByLabel('Name')).toHaveValue(receiptName);
    await expectFields(page, [nameA]);

    await page.getByLabel(nameA).fill('retro-fitted');
    await page.getByRole('button', { name: 'Save', exact: true }).first().click({ force: true });
    await page.waitForURL(/\/receipts\/\d+\/view/);

    // Reaching /view proves enforceReceiptCustomFieldSelection accepted the newly attached id;
    // re-navigating proves the value came back out of the server rather than the live form.
    await page.goto(`/receipts/${receiptId!}/view`);
    await expect(page.getByLabel(nameA)).toHaveValue('retro-fitted');
  });
});
