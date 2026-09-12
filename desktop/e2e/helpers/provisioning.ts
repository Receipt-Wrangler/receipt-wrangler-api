import {
  type APIRequestContext,
  expect,
  type Page,
  request as pwRequest,
} from '@playwright/test';
import { creds, type Role } from './auth';

// Shared provisioning for permission e2e specs. Roles/users/groups are CREATED
// through the real app UI (the realistic flow under test); TEARDOWN goes through
// the admin API. UI teardown can't reliably remove a custom role: the role-list
// delete button is disabled while the role is assigned, and the UI's *bulk*
// user-delete dummy-converts a group-owning user (keeping the role assigned), so
// the role is never freed. `DELETE /api/user/{id}` hard-deletes, and deleting a
// group frees its members' group-role assignment — so the API tears down cleanly.

export function uniqueName(tag: string): string {
  return `e2e-${tag}-${Date.now()}`;
}

export interface CreateRoleOptions {
  name: string;
  /** Role-type card to pick, e.g. 'Application role' | 'Group role'. */
  type: 'Application role' | 'Group role';
  /** Starting template button, matched by (sub)string, e.g. 'Start from scratch' | 'Viewer'. */
  preset: string;
  /** Resource categories to flip fully on, e.g. ['Account', 'Notifications', 'Groups']. */
  enableCategories?: string[];
  /**
   * Individual permissions to switch OFF after applying the preset. Identify the
   * accordion by its resource key (key minus last segment, e.g. 'group.dashboards')
   * and the row by its label (e.g. 'Read Dashboards').
   */
  disablePermissions?: { panelKey: string; label: string }[];
  /**
   * Group-role paid-by visibility (the "Visible paid-by users" picker). Pin
   * "Their own receipts" and/or specific users by display name. Leaving both
   * unset keeps the role unrestricted (members see every payer's receipts).
   */
  paidByOwn?: boolean;
  paidByUsers?: string[];
  /**
   * App-role only: tick "Don't create a personal group for new users with this
   * role" in the "Group creation" card, so accounts created with the role skip
   * the automatic "My Receipts" group (the virtual "All" group is still made).
   */
  skipDefaultGroup?: boolean;
  /**
   * Group-role category/tag grants (the shared app-grant-picker), by NAME. This
   * is the ceiling a member's individual assignment narrows within. Leaving both
   * unset keeps the role unrestricted (members see every category/tag).
   */
  categoryGrants?: string[];
  tagGrants?: string[];
  /**
   * Flips "Require an individual category/tag assignment for each member", so a
   * member with no assignment sees nothing instead of the role's set.
   */
  requiresIndividualCategories?: boolean;
  requiresIndividualTags?: boolean;
}

/**
 * Creates a role via the role form: pick type → name it → apply a preset →
 * optionally flip whole resource groups on → optionally switch individual
 * permissions off. Leaves the browser on `/roles`.
 */
export async function createRole(page: Page, opts: CreateRoleOptions): Promise<void> {
  await page.goto('/roles');
  await page.getByRole('button', { name: 'Add Role' }).first().click();
  await expect(page).toHaveURL(/\/roles\/new$/);
  await expect(page.getByLabel('Role Name')).toBeVisible();

  // Pick the scope first — it swaps the available presets.
  await page.getByRole('button', { name: new RegExp(opts.type) }).click();
  await page.getByLabel('Role Name').fill(opts.name);
  await page.getByRole('button', { name: opts.preset }).click();

  for (const category of opts.enableCategories ?? []) {
    await page
      .getByRole('button', { name: `Toggle all ${category} permissions` })
      .click();
  }

  // Expand each accordion at most once (clicking its sub-text toggles it), then
  // switch off every listed permission under it. Opening a panel per entry would
  // re-close it when two permissions share a panel (e.g. two app.reports perms).
  const openedPanels = new Set<string>();
  for (const { panelKey, label } of opts.disablePermissions ?? []) {
    if (!openedPanels.has(panelKey)) {
      // The resource key is unique on the page, so this won't collide with e.g. a
      // sidebar "Dashboards" link.
      const subText = new RegExp(`${panelKey.replace(/\./g, '\\.')} \\u00b7`);
      await page.getByText(subText).click();
      openedPanels.add(panelKey);
    }
    await page.getByRole('button', { name: `Toggle ${label}` }).click();
  }

  if (opts.skipDefaultGroup) {
    await page
      .getByTestId('skip-default-group')
      .getByRole('checkbox')
      .check();
  }

  if (opts.paidByOwn || (opts.paidByUsers?.length ?? 0) > 0) {
    // The picker is a multi-select autocomplete; while its panel is open the
    // listbox shares the field's label, so target the combobox role.
    const paidByField = page.getByRole('combobox', {
      name: 'Visible paid-by users',
    });
    if (opts.paidByOwn) {
      await paidByField.click();
      await page
        .getByRole('option', { name: 'Their own receipts', exact: true })
        .click();
    }
    for (const displayName of opts.paidByUsers ?? []) {
      await paidByField.click();
      await paidByField.fill(displayName);
      await page.getByRole('option', { name: displayName, exact: true }).click();
    }
    // Selecting an option re-opens the panel; close it so Save is clickable.
    await page.keyboard.press('Escape');
  }

  await selectGrants(page, 'grant-picker-categories', 'Categories', opts.categoryGrants);
  await selectGrants(page, 'grant-picker-tags', 'Tags', opts.tagGrants);

  if (opts.requiresIndividualCategories) {
    await page
      .getByText('Require an individual category assignment for each member')
      .click();
  }
  if (opts.requiresIndividualTags) {
    await page
      .getByText('Require an individual tag assignment for each member')
      .click();
  }

  await page.getByRole('button', { name: 'Save Role' }).click();
  await expect(page).toHaveURL(/\/roles$/);
}

/**
 * Picks [names] in one of the grant picker's multi-select autocompletes. While
 * the option panel is open the listbox shares the field's label, so the field is
 * targeted by its combobox role within the testid-scoped picker.
 */
async function selectGrants(
  page: Page,
  testId: string,
  label: string,
  names?: string[],
): Promise<void> {
  if (!names?.length) {
    return;
  }

  const field = page.getByTestId(testId).getByRole('combobox', { name: label });
  for (const name of names) {
    await field.click();
    await field.fill(name);
    await page.getByRole('option', { name, exact: true }).click();
  }
  // Selecting an option re-opens the panel; close it so Save stays clickable.
  await page.keyboard.press('Escape');
}

/** Creates a user assigned [opts.role] via the Add User dialog. */
export async function createUserWithRole(
  page: Page,
  opts: { username: string; password: string; role: string },
): Promise<void> {
  await page.goto('/users');
  await page.getByTestId('user-add').click();
  const dialog = page.getByRole('dialog').filter({ hasText: 'Create User' });
  await expect(dialog).toBeVisible();

  // The password input is *ngIf-ed in (create mode only) and carries the
  // generate/visibility buttons, so it settles a tick after the dialog opens.
  // Wait for it BEFORE filling anything: typing into a half-rendered dialog can
  // land the password in whichever field currently owns focus (seen as a
  // username of "<name><password>" and an empty, still-required password).
  const password = dialog.getByLabel('Password', { exact: true });
  await expect(password).toBeVisible();

  // Waiting for the password field narrows that race but does not close it — a
  // fill can still land in a neighbouring input while the dialog settles, and
  // because Username and Displayname carry the SAME value the damage reads as a
  // doubled username ("<name><name>") rather than as an obvious mis-type. So
  // fill and verify as one retryable unit: a misdirected round re-fills from
  // scratch instead of failing the whole spec's beforeAll. Same idiom as
  // openComboboxAndPick in helpers/reports.ts.
  await expect(async () => {
    await dialog.getByLabel('Username').fill(opts.username);
    await dialog.getByLabel('Displayname').fill(opts.username);
    await password.fill(opts.password);

    await expect(dialog.getByLabel('Username')).toHaveValue(opts.username, { timeout: 2000 });
    await expect(dialog.getByLabel('Displayname')).toHaveValue(opts.username, { timeout: 2000 });
    await expect(password).toHaveValue(opts.password, { timeout: 2000 });
  }).toPass({ timeout: 15_000 });
  await dialog.getByRole('combobox', { name: 'App Role' }).click();
  // The option panel is a floating overlay rendered on the page, not the dialog.
  await page.getByRole('option', { name: opts.role, exact: true }).click();
  await dialog.locator('app-submit-button button').click();
  await expect(dialog).toBeHidden();
}

/**
 * Creates a group with [opts.memberDisplayName] added as [opts.roleName], and
 * returns the new group's id. Mirrors the group-member form flow.
 */
export async function createGroupWithMember(
  page: Page,
  opts: { groupName: string; memberDisplayName: string; roleName: string },
): Promise<string> {
  await page.goto('/groups/create');
  await expect(page.getByLabel('Group Name')).toBeVisible();
  await page.getByLabel('Group Name').fill(opts.groupName);

  await page.getByTestId('add-group-member').click();
  const memberDialog = page
    .getByRole('dialog')
    .filter({ hasText: 'Add Group Member' });
  await expect(memberDialog).toBeVisible();

  // The user autocomplete filters/displays by display name; while the option
  // panel is open the listbox shares the field's aria label, so target the combobox.
  const userField = memberDialog.getByRole('combobox', { name: 'User' });
  await userField.fill(opts.memberDisplayName);
  await page
    .getByRole('option', { name: opts.memberDisplayName, exact: true })
    .click();

  // Role defaults to "Legacy Owner" — switch it to the chosen role.
  const roleSelect = memberDialog.getByRole('combobox', { name: 'Role' });
  await roleSelect.click();
  await page.getByRole('option', { name: opts.roleName, exact: true }).click();

  await memberDialog.getByTestId('dialog-submit-button').click();
  await expect(memberDialog).toBeHidden();

  // The page-level save is icon-only; the app-form submits on Enter from a field.
  await page.getByLabel('Group Name').press('Enter');
  await page.waitForURL(/\/groups\/\d+\/details\/view/);
  return page.url().match(/\/groups\/(\d+)\//)![1];
}

// --- API teardown -----------------------------------------------------------
//
// Cleanup runs over the admin API (see the file header for why the UI can't do
// it). `request.newContext` against E2E_BASE_URL reuses the dev-server proxy to
// `/api`; the admin login cookie is stored on the context for later DELETEs.

const apiBaseUrl = (): string =>
  process.env.E2E_BASE_URL ?? 'http://localhost:4200';

/**
 * Opens an `APIRequestContext` authenticated with the given credentials, runs
 * [fn], and disposes it. Use for a user the test provisioned itself, which has
 * no `E2E_*` env credentials — for a fixture account use `withApiAs` instead.
 */
export async function withApiAsCreds<T>(
  username: string,
  password: string,
  fn: (api: APIRequestContext) => Promise<T>,
): Promise<T> {
  const api = await pwRequest.newContext({ baseURL: apiBaseUrl() });
  try {
    const res = await api.post('/api/login', { data: { username, password } });
    if (!res.ok()) {
      throw new Error(`${username} API login failed: HTTP ${res.status()}`);
    }
    return await fn(api);
  } finally {
    await api.dispose();
  }
}

/**
 * Opens an `APIRequestContext` authenticated as the given e2e [role], runs [fn],
 * and disposes it. Use to drive API calls AS a specific user — e.g. asserting a
 * restricted user's request is denied, or admin teardown.
 */
export async function withApiAs<T>(
  role: Role,
  fn: (api: APIRequestContext) => Promise<T>,
): Promise<T> {
  const { username, password } = creds(role);
  return withApiAsCreds(username, password, fn);
}

/**
 * Opens an admin-authenticated `APIRequestContext`, runs [fn], and disposes it.
 * Use in `afterAll` for best-effort teardown of provisioned roles/users/groups.
 */
export async function withAdminApi<T>(
  fn: (api: APIRequestContext) => Promise<T>,
): Promise<T> {
  return withApiAs('admin', fn);
}

/** Returns the id of the user with [username] (via the admin user list). */
export async function apiGetUserId(
  api: APIRequestContext,
  username: string,
): Promise<number> {
  const users = (await (await api.get('/api/user/')).json()) as {
    id: number;
    username: string;
  }[];
  const user = users.find((u) => u.username === username);
  if (!user) {
    throw new Error(`user ${username} not found`);
  }
  return user.id;
}

/**
 * Returns the names of the groups the authenticated caller belongs to, including
 * the virtual "All" group. `GET /api/group/` lists the caller's own groups and is
 * gated on app.account.read, so the caller's role must grant it.
 */
export async function apiGroupNames(api: APIRequestContext): Promise<string[]> {
  const res = await api.get('/api/group/');
  if (!res.ok()) {
    throw new Error(
      `list groups failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const groups = (await res.json()) as { name: string }[];
  return groups.map((group) => group.name);
}

/**
 * One custom field value on a receipt. Exactly one of the typed columns is
 * populated, chosen by the field's own type — a currency field reads
 * currencyValue, a date field dateValue, and so on (the server stores whatever
 * it is given, so a mismatched column simply resolves to no value in a report).
 */
export interface ReceiptCustomFieldValue {
  customFieldId: number;
  stringValue?: string;
  dateValue?: string;
  selectValue?: number;
  currencyValue?: string;
  booleanValue?: boolean;
}

/** Creates a minimal OPEN receipt and returns its id. */
export async function apiCreateReceipt(
  api: APIRequestContext,
  opts: {
    groupId: number | string;
    paidByUserId: number;
    name: string;
    /** Receipt date as an ISO string. Defaults to a fixed date in 2024. */
    date?: string;
    /**
     * Receipt status. Defaults to OPEN. Pass RESOLVED to have the server stamp
     * `resolvedDate` — the only way to seed a receipt the Resolved Date filter
     * can match.
     */
    status?: string;
    customFields?: ReceiptCustomFieldValue[];
  },
): Promise<number> {
  const res = await api.post('/api/receipt', {
    data: {
      name: opts.name,
      amount: '10.00',
      date: opts.date ?? '2024-01-01T00:00:00Z',
      groupId: Number(opts.groupId),
      paidByUserId: opts.paidByUserId,
      status: opts.status ?? 'OPEN',
      ...(opts.customFields ? { customFields: opts.customFields } : {}),
    },
  });
  if (!res.ok()) {
    throw new Error(
      `create receipt failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const receipt = (await res.json()) as { id: number };
  return receipt.id;
}

/**
 * Creates a comment on [receiptId] and returns its id. The comment is owned by
 * whoever the [api] context is authenticated as (the per-comment delete control
 * only renders for the logged-in user's own comments).
 */
export async function apiCreateComment(
  api: APIRequestContext,
  opts: { receiptId: number; comment: string },
): Promise<number> {
  const res = await api.post('/api/comment/', {
    data: { comment: opts.comment, receiptId: opts.receiptId },
  });
  if (!res.ok()) {
    throw new Error(
      `create comment failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const comment = (await res.json()) as { id: number };
  return comment.id;
}

/**
 * Creates a group via the admin API (the caller is auto-added as owner) and
 * returns its id + name. Create only needs name+status; members are required
 * on update, not create.
 */
export async function apiCreateGroup(
  api: APIRequestContext,
  name: string,
): Promise<{ id: number; name: string }> {
  const res = await api.post('/api/group', {
    data: { name, status: 'ACTIVE' },
  });
  if (!res.ok()) {
    throw new Error(
      `create group failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const group = (await res.json()) as { id: number; name: string };
  return { id: group.id, name: group.name };
}

/** Hard-deletes the user with [username] (frees any app-role assignment). */
export async function apiDeleteUserByName(
  api: APIRequestContext,
  username: string,
): Promise<void> {
  const users = (await (await api.get('/api/user/')).json()) as {
    id: number;
    username: string;
  }[];
  const user = users.find((u) => u.username === username);
  if (user) {
    await api.delete(`/api/user/${user.id}`);
  }
}

/** Deletes the group [groupId] (frees its members' group-role assignment). */
export async function apiDeleteGroupById(
  api: APIRequestContext,
  groupId: string,
): Promise<void> {
  await api.delete(`/api/group/${groupId}`);
}

/** Returns a group id (as a string) the caller can generate reports over. */
export async function apiFirstReportGroupId(api: APIRequestContext): Promise<string> {
  const appData = (await (await api.get('/api/user/appData')).json()) as {
    groupPermissions?: Record<string, string[]>;
  };
  const groupPermissions = appData.groupPermissions ?? {};
  const groupId = Object.keys(groupPermissions).find((id) =>
    (groupPermissions[id] ?? []).includes('group.reports.read'),
  );
  if (!groupId) {
    throw new Error('no group with group.reports.read for the caller');
  }
  return groupId;
}

/**
 * Resolves a category the caller can both report on and see in the builder: a group
 * with group.reports.read whose appData catalog carries at least one category. Read
 * from the same appData.groupCategories the builder's picker unions, so a filter
 * seeded with the returned id resolves to its name chip on open-in-builder.
 */
export async function apiFirstReportCategory(
  api: APIRequestContext,
): Promise<{ groupId: string; categoryId: number; categoryName: string }> {
  const appData = (await (await api.get('/api/user/appData')).json()) as {
    groupPermissions?: Record<string, string[]>;
    groupCategories?: Record<string, { id?: number; name?: string }[]>;
  };
  const groupPermissions = appData.groupPermissions ?? {};
  const groupCategories = appData.groupCategories ?? {};
  for (const groupId of Object.keys(groupPermissions)) {
    if (!(groupPermissions[groupId] ?? []).includes('group.reports.read')) {
      continue;
    }
    const category = (groupCategories[groupId] ?? []).find((c) => c.id != null && !!c.name);
    if (category) {
      return { groupId, categoryId: category.id!, categoryName: category.name! };
    }
  }
  throw new Error('no category in a group with group.reports.read for the caller');
}

/**
 * Creates a report template via the API and returns its id + name (requires
 * app.reports.create on the caller). The body is a complete, buildable
 * ReportRequestCommand — records mode over a group the caller can report on.
 */
export async function apiCreateReportTemplate(
  api: APIRequestContext,
  opts?: {
    name?: string;
    groupIds?: string[];
    formats?: string[];
    filter?: unknown;
    // The shape of the report itself. Defaulted to the records-mode config every
    // caller wanted before; override to store a grouped/aggregated template, e.g.
    // one grouped by a custom field's `custom_<id>` key.
    groupBy?: string[];
    detail?: { mode: 'records' | 'aggregate'; by?: string };
    columns?: unknown[];
  },
): Promise<{ id: number; name: string }> {
  const name = opts?.name ?? uniqueName('report-template');
  // Scope a real group the caller can report on, so generating over it isn't
  // group-gated out (the seeded admin's group ids are not necessarily [1]).
  const groupIds = opts?.groupIds ?? [await apiFirstReportGroupId(api)];
  const res = await api.post('/api/report/template', {
    data: {
      name,
      groupIds,
      period: { preset: 'this_month' },
      detail: opts?.detail ?? { mode: 'records' },
      columns: opts?.columns ?? [
        { kind: 'dimension', name: 'Name', label: 'Name', field: 'name' },
      ],
      formats: opts?.formats ?? ['csv'],
      ...(opts?.groupBy ? { groupBy: opts.groupBy } : {}),
      ...(opts?.filter ? { filter: opts.filter } : {}),
    },
  });
  if (!res.ok()) {
    throw new Error(
      `create report template failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const template = (await res.json()) as { id: number; name: string };
  return { id: template.id, name: template.name };
}

export interface UpsertRolePayload {
  name: string;
  scope: 'APP' | 'GROUP';
  permissions: string[];
  reportTemplateGrants?: { reportTemplateId: number; permissions: string[] }[];
  /**
   * Category/tag ids a GROUP role's members may see. This is the CEILING that a
   * member's individual assignment narrows within — the two intersect. Empty (or
   * omitted) means unrestricted, matching the backend's grant convention.
   */
  categoryGrants?: number[];
  tagGrants?: number[];
  /**
   * When true, a member holding this role with NO individual assignment sees
   * nothing at all rather than falling back to the role's set — so forgetting to
   * assign a new member fails closed.
   */
  requiresIndividualCategoryGrants?: boolean;
  requiresIndividualTagGrants?: boolean;
}

/**
 * Creates a role via the admin API and returns its id. Unlike createRole (which
 * drives the UI), this is for provisioning a role whose exact permission set /
 * report-template matrix a test needs precisely — e.g. a group role granting
 * group.reports.read plus a specific reportTemplateGrants matrix.
 */
export async function apiCreateRole(
  api: APIRequestContext,
  opts: UpsertRolePayload,
): Promise<{ id: number }> {
  const res = await api.post('/api/role', {
    data: { description: '', ...opts },
  });
  if (!res.ok()) {
    throw new Error(`create role failed: HTTP ${res.status()} ${await res.text()}`);
  }
  const role = (await res.json()) as { id: number };
  return { id: role.id };
}

/**
 * Updates the role [id] in place (scope carried in the body). Used to set a group
 * role's reportTemplateGrants matrix AFTER the templates it references exist (the
 * matrix references template ids, which need the group, which needs the role — so
 * the role is created first and its matrix filled here).
 */
export async function apiUpdateRole(
  api: APIRequestContext,
  id: number,
  opts: UpsertRolePayload,
): Promise<void> {
  const res = await api.put(`/api/role/${id}`, {
    data: { description: '', ...opts },
  });
  if (!res.ok()) {
    throw new Error(`update role failed: HTTP ${res.status()} ${await res.text()}`);
  }
}

/** Deletes the report template [id] (requires app.reports.delete on the caller). */
export async function apiDeleteReportTemplateById(
  api: APIRequestContext,
  id: number | string,
): Promise<void> {
  await api.delete(`/api/report/template/${id}`);
}

// --- Categories / tags -------------------------------------------------------
//
// Categories and tags are GLOBAL with a uniquely-indexed name, so every spec must
// mint uniqueName-suffixed ones and delete them in teardown; a leaked name makes
// the next run's create fail.

/** Creates a category and returns its id + name. */
export async function apiCreateCategory(
  api: APIRequestContext,
  name: string,
): Promise<{ id: number; name: string }> {
  const res = await api.post('/api/category', { data: { name, description: '' } });
  if (!res.ok()) {
    throw new Error(
      `create category failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const category = (await res.json()) as { id: number; name: string };
  return { id: category.id, name: category.name };
}

/** Creates a tag and returns its id + name. */
export async function apiCreateTag(
  api: APIRequestContext,
  name: string,
): Promise<{ id: number; name: string }> {
  const res = await api.post('/api/tag', { data: { name, description: '' } });
  if (!res.ok()) {
    throw new Error(`create tag failed: HTTP ${res.status()} ${await res.text()}`);
  }
  const tag = (await res.json()) as { id: number; name: string };
  return { id: tag.id, name: tag.name };
}

/**
 * Warns rather than throws on a failed delete.
 *
 * Every caller runs inside an `afterAll` whose body is wrapped in `try/catch {}`
 * so cleanup can never mask a test failure. Throwing here would be swallowed by
 * that catch AND abort the remaining deletes in the same block — leaking more
 * names than it reports. A warning keeps every delete running and still puts the
 * failure in the test output, which is what makes a leaked unique name (the thing
 * that breaks the next run's create) diagnosable.
 */
async function warnOnFailedDelete(
  api: APIRequestContext,
  path: string,
  what: string,
): Promise<void> {
  const res = await api.delete(path);
  if (!res.ok()) {
    console.warn(
      `e2e cleanup: failed to delete ${what} — HTTP ${res.status()} ${await res.text()}`,
    );
  }
}

export async function apiDeleteCategoryById(
  api: APIRequestContext,
  id: number,
): Promise<void> {
  await warnOnFailedDelete(api, `/api/category/${id}`, `category ${id}`);
}

export async function apiDeleteTagById(
  api: APIRequestContext,
  id: number,
): Promise<void> {
  await warnOnFailedDelete(api, `/api/tag/${id}`, `tag ${id}`);
}

// --- Custom fields -----------------------------------------------------------
//
// Like categories and tags the pool is GLOBAL, so mint uniqueName-suffixed fields
// and delete them in teardown — a leaked one shows up in every other spec's
// pickers. Deleting a field destroys every value stored against it, so a spec
// that seeds receipts with values must delete those receipts (or their group)
// too.

export type CustomFieldType = 'TEXT' | 'DATE' | 'SELECT' | 'CURRENCY' | 'BOOLEAN';

/** Creates a custom field and returns its id + name. */
export async function apiCreateCustomField(
  api: APIRequestContext,
  opts: { name: string; type: CustomFieldType; options?: string[] },
): Promise<{ id: number; name: string }> {
  const res = await api.post('/api/customField/', {
    data: {
      name: opts.name,
      type: opts.type,
      description: '',
      options: (opts.options ?? []).map((value) => ({ value })),
    },
  });
  if (!res.ok()) {
    throw new Error(
      `create custom field failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  const customField = (await res.json()) as { id: number; name: string };
  return { id: customField.id, name: customField.name };
}

/** Reads one custom field back, so a spec can assert what an edit persisted. */
export async function apiGetCustomFieldById(
  api: APIRequestContext,
  id: number,
): Promise<{
  id: number;
  name: string;
  type: string;
  description?: string;
  options?: { id: number; value: string }[];
}> {
  const res = await api.get(`/api/customField/${id}`);
  if (!res.ok()) {
    throw new Error(
      `get custom field failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
  return res.json();
}

/** Deletes a custom field. Requires app.custom-fields.delete (admin-only). */
export async function apiDeleteCustomFieldById(
  api: APIRequestContext,
  id: number,
): Promise<void> {
  await warnOnFailedDelete(api, `/api/customField/${id}`, `custom field ${id}`);
}

// --- Group default custom fields ---------------------------------------------

/**
 * Persists [customFieldIds] as [groupId]'s default custom fields, optionally
 * flipping the "apply on ingest" toggle with [applyOnIngest].
 *
 * `PUT /group/{id}/groupReceiptSettings` is an UPSERT over the whole settings
 * object, so this reads the group's current settings back and echoes them
 * unchanged around the override — sending a partial body would reset every
 * quick-scan flag to false. Only the boolean flags are echoed: the enum
 * defaults (`quickScanDefaultPaidByType` / `...Status`) are omitted because the
 * backend keeps its stored value and rejects an empty enum, which is also why
 * paid-by and status stay shown+required here (making either optional would
 * need a configured default). Mirrors the mobile suite's
 * `setGroupDefaultCustomFields` fixture.
 *
 * Pass `[]` to clear the group's set.
 */
export async function apiSetGroupDefaultCustomFields(
  api: APIRequestContext,
  groupId: number | string,
  customFieldIds: number[],
  applyOnIngest?: boolean,
): Promise<void> {
  const groupsRes = await api.get('/api/group/');
  if (!groupsRes.ok()) {
    throw new Error(
      `set default custom fields: GET /api/group/ failed: HTTP ${groupsRes.status()} ${await groupsRes.text()}`,
    );
  }
  const groups = (await groupsRes.json()) as {
    id: number;
    groupReceiptSettings?: Record<string, unknown>;
  }[];
  const group = groups.find((g) => String(g.id) === String(groupId));
  if (!group) {
    throw new Error(`set default custom fields: group ${groupId} not visible`);
  }

  const settings = group.groupReceiptSettings ?? {};
  const command: Record<string, unknown> = {
    defaultCustomFieldIds: customFieldIds,
  };
  for (const key of [
    'hideImages', 'hideReceiptCategories', 'hideReceiptTags',
    'hideItemCategories', 'hideItemTags', 'hideComments',
    'hideShareCategories', 'hideShareTags',
    'quickScanPaidByEnabled', 'quickScanPaidByRequired',
    'quickScanStatusEnabled', 'quickScanStatusRequired',
    'quickScanCategoriesEnabled', 'quickScanCategoriesRequired',
    'quickScanTagsEnabled', 'quickScanTagsRequired',
    'quickScanCommentEnabled', 'quickScanCommentRequired',
    'applyDefaultCustomFieldsOnIngest',
  ]) {
    command[key] = settings[key] ?? false;
  }
  if (applyOnIngest !== undefined) {
    command['applyDefaultCustomFieldsOnIngest'] = applyOnIngest;
  }

  const res = await api.put(`/api/group/${groupId}/groupReceiptSettings`, {
    data: command,
  });
  if (!res.ok()) {
    throw new Error(
      `set default custom fields failed: HTTP ${res.status()} ${await res.text()}`,
    );
  }
}

// --- Per-member category/tag grants ------------------------------------------

/**
 * Replaces one group member's individual category/tag assignment, returning the
 * RAW response so a spec can assert the status (200 / 400 out-of-ceiling / 403
 * missing group.members.grants.update / 404 non-member) rather than only the
 * happy-path effect.
 *
 * [groupId]/[userId] go in the URL because that is what the endpoint authorizes
 * against; [body] may carry contradicting ids to prove the URL wins.
 */
export async function apiSetMemberGrants(
  api: APIRequestContext,
  groupId: number | string,
  userId: number,
  body: { categoryGrants?: number[]; tagGrants?: number[] } & Record<string, unknown>,
) {
  return api.put(`/api/group/${groupId}/member/${userId}/grants`, { data: body });
}

/**
 * The category/tag names the calling user can actually SEE in [groupId], read
 * from appData's per-group catalogs.
 *
 * This is deliberately the assertion surface for effective visibility: it is the
 * very array the desktop's receipt-form pickers render from, so it tests the real
 * delivery path rather than a parallel one. Names (not ids) so failures read
 * clearly.
 */
export async function apiMemberCatalog(
  api: APIRequestContext,
  groupId: number | string,
): Promise<{ categories: string[]; tags: string[] }> {
  const appData = (await (await api.get('/api/user/appData')).json()) as {
    groupCategories?: Record<string, { name?: string }[]>;
    groupTags?: Record<string, { name?: string }[]>;
  };
  const key = String(groupId);
  const names = (entries?: { name?: string }[]) =>
    (entries ?? []).map((e) => e.name ?? '').sort();

  return {
    categories: names(appData.groupCategories?.[key]),
    tags: names(appData.groupTags?.[key]),
  };
}

/** Replaces a group's whole member roster (the wholesale UpdateGroup path). */
export async function apiSetGroupRoster(
  api: APIRequestContext,
  groupId: number | string,
  opts: {
    name: string;
    members: { userId: number; groupRoleId: number }[];
  },
) {
  return api.put(`/api/group/${groupId}`, {
    data: {
      name: opts.name,
      status: 'ACTIVE',
      isAllGroup: false,
      groupMembers: opts.members.map((m) => ({
        userId: m.userId,
        groupId: Number(groupId),
        groupRoleId: m.groupRoleId,
      })),
    },
  });
}

/** Returns a group's members (including their individual grant ids). */
export async function apiGetGroupMembers(
  api: APIRequestContext,
  groupId: number | string,
): Promise<
  {
    userId: number;
    groupRoleId: number;
    categoryGrants?: number[] | null;
    tagGrants?: number[] | null;
  }[]
> {
  const group = (await (await api.get(`/api/group/${groupId}`)).json()) as {
    groupMembers: {
      userId: number;
      groupRoleId: number;
      categoryGrants?: number[] | null;
      tagGrants?: number[] | null;
    }[];
  };
  return group.groupMembers;
}

/**
 * Deletes the [scope] role named [name]. Only succeeds once it's unassigned, so
 * call after the user/group that referenced it is gone.
 */
export async function apiDeleteRoleByName(
  api: APIRequestContext,
  name: string,
  scope: 'APP' | 'GROUP',
): Promise<void> {
  const roles = (await (await api.get('/api/role')).json()) as {
    id: number;
    name: string;
    scope: string;
  }[];
  const role = roles.find((r) => r.name === name && r.scope === scope);
  if (role) {
    await api.delete(`/api/role/${role.id}?scope=${scope}`);
  }
}
