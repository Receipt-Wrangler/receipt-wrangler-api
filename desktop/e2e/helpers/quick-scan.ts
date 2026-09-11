import { openReceiptsOverflowMenu } from './receipts-table';
import { expect, type Locator, type Page, type Route } from '@playwright/test';

// Shared helpers for the Quick Scan dialog e2e specs. The dialog is only
// reachable behind the `aiPoweredReceipts` feature flag (off in dev/CI) and its
// per-image fields are driven by each group's quick-scan config + the caller's
// user preferences + the per-group category/tag catalogs — all delivered on
// AppData. Rather than mutate global server config, we intercept the AppData
// response client-side (per BrowserContext, like `stubTokenRefresh`) and inject
// exactly what a given test needs. This mirrors the original quick-scan-dialog
// spec and matches the desktop CLAUDE.md "AppData interception" pattern.

/** The fixture image fed into the dialog's file input (read client-side). */
export const RECEIPT_PNG = 'e2e/fixtures/receipt.png';

/** A partial `GroupReceiptSettings` quick-scan config to inject onto a group. */
export interface QuickScanConfig {
  quickScanPaidByEnabled?: boolean;
  quickScanPaidByRequired?: boolean;
  quickScanDefaultPaidByType?: string;
  quickScanStatusEnabled?: boolean;
  quickScanStatusRequired?: boolean;
  quickScanDefaultStatus?: string;
  quickScanCategoriesEnabled?: boolean;
  quickScanCategoriesRequired?: boolean;
  quickScanTagsEnabled?: boolean;
  quickScanTagsRequired?: boolean;
  quickScanCommentEnabled?: boolean;
  quickScanCommentRequired?: boolean;
  /** Hides comments group-wide, which also hides the quick-scan comment field. */
  hideComments?: boolean;
}

interface IdName {
  id: number;
  name: string;
}

export interface QuickScanAppData {
  /** Value for the `aiPoweredReceipts` feature flag (defaults to true). */
  aiPoweredReceipts?: boolean;
  /** Inject quick-scan config onto specific (real) groups by id. */
  groupConfigs?: { groupId: number; config: QuickScanConfig }[];
  /** Prefill values for the caller's `userPreferences` (e.g. quickScanDefault*). */
  userPreferences?: Record<string, unknown>;
  /** Per-group category catalog the pickers read (keyed by group id). */
  groupCategories?: Record<number, IdName[]>;
  /** Per-group tag catalog the pickers read (keyed by group id). */
  groupTags?: Record<number, IdName[]>;
  /**
   * Per-group effective permissions (keyed by group id). Use to withhold a
   * permission the real caller holds — e.g. `group.comments.create`, which gates
   * the quick-scan comment field — without provisioning a custom role.
   */
  groupPermissions?: Record<number, string[]>;
}

/**
 * Intercepts `GET /api/user/appData` to flip the AI flag on and inject the given
 * quick-scan config / user preferences / category+tag catalogs. The dialog reads
 * all of these from the NGXS store, so the injected values drive its behavior
 * without touching server state. Also stubs the pre-auth `/api/featureConfig`.
 */
export async function injectQuickScanAppData(
  page: Page,
  data: QuickScanAppData = {},
): Promise<void> {
  const aiPoweredReceipts = data.aiPoweredReceipts ?? true;

  await page.route('**/api/user/appData', async (route: Route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.featureConfig = { ...body.featureConfig, aiPoweredReceipts };

    for (const { groupId, config } of data.groupConfigs ?? []) {
      const target = (body.groups ?? []).find((g: any) => g.id === groupId);
      if (target) {
        target.groupReceiptSettings = { ...target.groupReceiptSettings, ...config };
      }
    }
    if (data.userPreferences) {
      body.userPreferences = { ...body.userPreferences, ...data.userPreferences };
    }
    // Object keys are strings at runtime, so a numeric-keyed record spreads into
    // the string-keyed AppData maps the group selectors index by id.
    if (data.groupCategories) {
      body.groupCategories = { ...body.groupCategories, ...data.groupCategories };
    }
    if (data.groupTags) {
      body.groupTags = { ...body.groupTags, ...data.groupTags };
    }
    if (data.groupPermissions) {
      body.groupPermissions = { ...body.groupPermissions, ...data.groupPermissions };
    }

    await route.fulfill({ response, json: body });
  });

  // Belt-and-suspenders for the pre-auth path (the logged-in path uses AppData).
  await page.route('**/api/featureConfig', (route: Route) =>
    route.fulfill({ json: { enableLocalSignUp: true, aiPoweredReceipts } }),
  );
}

/**
 * Opens the Quick Scan dialog from a group's receipts page (the header button is
 * icon-only, so it's targeted by testid, not accessible name). Returns the
 * dialog locator.
 */
export async function openQuickScanDialog(page: Page, groupId: number): Promise<Locator> {
  await page.goto(`/receipts/group/${groupId}`);
  await openReceiptsOverflowMenu(page);
  // A mat-menu-item is the button itself, so there is no inner button to reach.
  await page.getByTestId('receipts-quick-scan').click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  return dialog;
}

/**
 * Uploads [count] copies of the fixture image into the dialog — one carousel
 * slide per image (upload-image emits `fileLoaded` once per file). Waits until
 * every slide has mounted.
 */
export async function uploadQuickScanImages(dialog: Locator, count = 1): Promise<void> {
  await dialog
    .locator('app-upload-image input[type="file"]')
    .setInputFiles(Array.from({ length: count }, () => RECEIPT_PNG));
  await expect(dialog.locator('slide')).toHaveCount(count);
}

/**
 * Selects a group on the given per-image scope (the whole dialog for a single
 * image, or the visible `slide` for the image being configured). If a group is
 * already selected the text input goes readonly, so its X (clear) button is
 * clicked first to make the field editable again — this is how a user switches
 * an image's group.
 */
export async function selectImageGroup(
  page: Page,
  scope: Locator,
  groupName: string,
): Promise<void> {
  // The single-select autocomplete only renders a button (the X clear control)
  // once a value is selected; clear it so the input becomes editable.
  const clear = scope.locator('app-group-autocomplete').getByRole('button');
  if (await clear.count()) {
    await clear.first().click();
  }
  const groupField = scope.getByRole('combobox', { name: 'Group' });
  await groupField.click();
  await groupField.fill(groupName);
  await page.getByRole('option', { name: groupName, exact: true }).click();
}

/**
 * Extracts the text form-fields (name → values) from a `multipart/form-data`
 * request body, tolerating a binary file part (parts with a `filename=` are
 * skipped). Used to assert the exact values the quick-scan submit sends — the
 * one thing the mobile suite can't observe (its queued receipt has no id).
 */
export function parseMultipartFields(
  body: Buffer,
  contentType: string,
): Map<string, string[]> {
  const match = /boundary=(.+)$/.exec(contentType);
  if (!match) {
    throw new Error(`no multipart boundary in content-type: ${contentType}`);
  }
  const boundary = `--${match[1]}`;
  // latin1 is a 1:1 byte→char decode: it never corrupts the ASCII field parts
  // the way utf8 would when the body also carries the raw PNG bytes.
  const text = body.toString('latin1');
  const fields = new Map<string, string[]>();

  for (const segment of text.split(boundary)) {
    const headerEnd = segment.indexOf('\r\n\r\n');
    if (headerEnd === -1) {
      continue;
    }
    const headers = segment.slice(0, headerEnd);
    const nameMatch = /name="([^"]+)"/.exec(headers);
    if (!nameMatch || /filename="/.test(headers)) {
      continue; // not a field, or a (binary) file part
    }
    const value = segment.slice(headerEnd + 4).replace(/\r\n$/, '');
    const values = fields.get(nameMatch[1]) ?? [];
    values.push(value);
    fields.set(nameMatch[1], values);
  }

  return fields;
}
