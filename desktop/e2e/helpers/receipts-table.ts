import { expect, type Page } from '@playwright/test';

/**
 * Navigates to the receipts table for the user's own group and waits for the
 * toolbar. The overflow menu is the readiness probe because it is the one
 * toolbar control no permission or feature flag can hide.
 */
export async function gotoReceiptsTable(page: Page): Promise<string> {
  // storageState means we're already authed; "/" redirects to the dashboard
  // for the user's group, from which we can recover the group id.
  await page.goto('/');
  await page.waitForURL(/\/dashboard\/group\/\d+/);
  const groupId = page.url().match(/\/dashboard\/group\/(\d+)/)![1];
  await page.goto(`/receipts/group/${groupId}`);
  await expect(page.getByTestId('receipts-overflow-menu')).toBeVisible();

  return groupId;
}

/**
 * Opens the receipts table's overflow menu, which holds Quick Scan, Export,
 * Configure Columns, Poll email(s) and the selection actions.
 *
 * The menu closes on any click inside it, so callers that trigger more than one
 * item have to reopen it each time.
 */
export async function openReceiptsOverflowMenu(page: Page): Promise<void> {
  await page.getByTestId('receipts-overflow-menu').click();
  await expect(page.getByRole('menu')).toBeVisible();
}
