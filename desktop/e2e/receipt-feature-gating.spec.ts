import { BrowserContext, expect, Page, test } from '@playwright/test';
import { creds, stubTokenRefresh } from './helpers/auth';
import { openReceiptsOverflowMenu } from './helpers/receipts-table';
import {
  apiCreateReceipt,
  apiDeleteGroupById,
  apiDeleteRoleByName,
  apiGetUserId,
  createGroupWithMember,
  createRole,
  uniqueName,
  withAdminApi,
} from './helpers/provisioning';
import { injectQuickScanAppData } from './helpers/quick-scan';

// The AI-powered receipt features are permission-gated in the UI:
//   - Quick Scan  -> group.receipts.quick-scan (sidebar + receipts-table header),
//   - Poll Email  -> group.email.poll (receipts-table header),
//   - Magic Fill  -> group.receipts.magic-fill (receipt form image toolbar).
// A "Viewer" group role holds none of these, so an admin context provisions a
// Viewer group with e2e-user + a seeded receipt and asserts (as e2e-user) that
// none of the three controls render.
//
// IMPORTANT CONFOUND: all three controls ALSO sit behind the `aiPoweredReceipts`
// feature flag (`*appFeature="'aiPoweredReceipts'"` for Quick Scan/Poll Email,
// and `aiPoweredReceipts()` for Magic Fill). That flag is server-config
// (FeatureConfig) and is `false` in the dev/CI API, so the negative tests below
// hold regardless of the flag — a permission-denied member sees no control.
//
// The Quick Scan POSITIVE contrast IS now covered: rather than mutate server
// config, the last test injects `aiPoweredReceipts: true` client-side (the same
// AppData interception the quick-scan-dialog spec uses). With the flag held on,
// the button then hinges purely on the permission — a Legacy Editor member sees
// it, the Viewer does not. (Poll Email / Magic Fill positives are still deferred;
// they need extra server state — email integration / a magic-fill-capable role.)

test.describe('Receipt feature gating (quick-scan / poll-email / magic-fill)', () => {
  test.describe.configure({ mode: 'serial' });

  let adminContext: BrowserContext;
  let adminPage: Page;
  const roleName = uniqueName('no-ai-role');
  const groupName = uniqueName('no-ai-grp');
  const editorGroupName = uniqueName('ai-editor-grp');
  let groupId: string;
  let editorGroupId: string;
  let receiptId: number;

  test.beforeAll(async ({ browser }) => {
    adminContext = await browser.newContext({
      storageState: 'e2e/.auth/admin.json',
    });
    adminPage = await adminContext.newPage();
    await stubTokenRefresh(adminPage);

    // Viewer: holds group.receipts.read (so the list/form load) but none of
    // quick-scan / magic-fill / email.poll.
    await createRole(adminPage, {
      name: roleName,
      type: 'Group role',
      preset: 'Viewer',
    });

    groupId = await createGroupWithMember(adminPage, {
      groupName,
      memberDisplayName: 'E2E User',
      roleName,
    });

    // A second group where e2e-user holds the Legacy Editor system role, which
    // includes group.receipts.quick-scan — the positive contrast for the gate.
    editorGroupId = await createGroupWithMember(adminPage, {
      groupName: editorGroupName,
      memberDisplayName: 'E2E User',
      roleName: 'Legacy Editor',
    });

    await withAdminApi(async (api) => {
      const userId = await apiGetUserId(api, creds('user').username);
      receiptId = await apiCreateReceipt(api, {
        groupId,
        paidByUserId: userId,
        name: uniqueName('no-ai-rcpt'),
      });
    });
  });

  test.afterAll(async () => {
    try {
      await withAdminApi(async (api) => {
        await apiDeleteGroupById(api, groupId);
        await apiDeleteGroupById(api, editorGroupId);
        await apiDeleteRoleByName(api, roleName, 'GROUP');
      });
    } catch {
      // Best-effort cleanup — don't mask a test failure with a cleanup error.
    }
    await adminContext?.close();
  });

  // Assertions run as the default project user (e2e-user = the Viewer).
  test.beforeEach(async ({ page }) => {
    await stubTokenRefresh(page);
  });

  // Both controls live in the toolbar's overflow menu, so the menu has to be
  // open for their absence to mean anything. They are asserted by testid rather
  // than by name because a mat-menu-item's role is "menuitem", so a
  // getByRole('button', ...) negative would pass whether or not they rendered.
  test('the receipts overflow menu shows no Quick Scan entry', async ({ page }) => {
    await page.goto(`/receipts/group/${groupId}`);
    // The table renders for a member who holds group.receipts.read.
    await openReceiptsOverflowMenu(page);
    // Configure Columns is ungated, so the menu is genuinely populated.
    await expect(page.getByTestId('configure-columns')).toBeVisible();
    // Quick Scan (group.receipts.quick-scan, also feature-flagged) is absent.
    await expect(page.getByTestId('receipts-quick-scan')).toHaveCount(0);
  });

  test('the receipts overflow menu shows no Poll Email entry', async ({ page }) => {
    await page.goto(`/receipts/group/${groupId}`);
    await openReceiptsOverflowMenu(page);
    await expect(page.getByTestId('configure-columns')).toBeVisible();
    // Poll Email (group.email.poll, also feature-flagged + needs email
    // integration enabled) is absent.
    await expect(page.getByTestId('receipts-poll-email')).toHaveCount(0);
  });

  test('the receipt form shows no Magic Fill button', async ({ page }) => {
    await page.goto(`/receipts/${receiptId}/view`);
    // The receipt form renders for a member who holds group.receipts.read.
    await expect(page.getByLabel('Name')).toBeVisible();
    // Magic Fill (group.receipts.magic-fill, also feature-flagged) is absent.
    await expect(page.getByRole('button', { name: 'Magic fill' })).toHaveCount(
      0,
    );
  });

  // Positive contrast for Quick Scan: inject aiPoweredReceipts client-side (the
  // harness must not mutate server FeatureConfig), holding the flag on so the
  // button hinges purely on the permission. The Legacy Editor group holds
  // group.receipts.quick-scan and shows it; the Viewer group — same user, same
  // flag on — does not, proving the permission is the discriminator.
  test('positive contrast: a Legacy Editor member sees the Quick Scan button', async ({
    page,
  }) => {
    await injectQuickScanAppData(page);

    // Editor group (holds quick-scan) -> the entry renders.
    await page.goto(`/receipts/group/${editorGroupId}`);
    await openReceiptsOverflowMenu(page);
    await expect(page.getByTestId('receipts-quick-scan')).toBeVisible();

    // Viewer group (lacks quick-scan), flag still on -> the entry is absent.
    await page.goto(`/receipts/group/${groupId}`);
    await openReceiptsOverflowMenu(page);
    await expect(page.getByTestId('configure-columns')).toBeVisible();
    await expect(page.getByTestId('receipts-quick-scan')).toHaveCount(0);
  });
});
