import {
  RECEIPT_FILTER_FIELDS,
  ReceiptFilterFieldKey,
} from "../constants/receipt-filter-fields.constant";
import { RECEIPT_STATUS_OPTIONS } from "../constants/receipt-status-options";
import { Category, Group, ReceiptPagedRequestFilter, Tag, User } from "../open-api";
import { buildFilterChips, FilterChip } from "./filter-chips";

export type ReceiptFilterChip = FilterChip<ReceiptFilterFieldKey>;

/**
 * Everything needed to turn stored ids into names. Passed in rather than looked
 * up here so the builder stays pure and testable — the same shape
 * `report-template-summary.ts` uses.
 */
export interface ReceiptFilterChipLookups {
  categories: readonly Category[];
  tags: readonly Tag[];
  groups: readonly Group[];
  users: readonly User[];
  formatDate: (value: unknown) => string;
  formatCurrency: (value: unknown) => string;
}

/**
 * One chip per receipt filter field that actually narrows the result set. See
 * `buildFilterChips` for the shared label rules; this wrapper only supplies the
 * receipt-specific id resolution.
 */
export function buildReceiptFilterChips(
  filter: ReceiptPagedRequestFilter | undefined | null,
  lookups: ReceiptFilterChipLookups,
  omitKeys: readonly string[] = [],
): ReceiptFilterChip[] {
  return buildFilterChips(
    RECEIPT_FILTER_FIELDS,
    filter as Record<string, unknown> | undefined | null,
    {
      formatDate: lookups.formatDate,
      formatCurrency: lookups.formatCurrency,
      resolveOptionName: (key, id) => resolveOptionName(key as ReceiptFilterFieldKey, id, lookups),
    },
    omitKeys,
  );
}

/**
 * An id the caller can no longer resolve — a category outside their grants, a
 * group they have left — falls back to the raw id rather than vanishing, so the
 * condition stays visible and clearable.
 */
function resolveOptionName(
  key: ReceiptFilterFieldKey,
  id: unknown,
  lookups: ReceiptFilterChipLookups,
): string {
  const rawId = String(id);

  switch (key) {
    case "categories":
      return findName(lookups.categories, rawId) ?? rawId;
    case "tags":
      return findName(lookups.tags, rawId) ?? rawId;
    case "group":
      return findName(lookups.groups, rawId) ?? rawId;
    case "paidBy":
      return lookups.users.find((user) => String(user.id) === rawId)?.displayName ?? rawId;
    case "status":
      return RECEIPT_STATUS_OPTIONS.find((option) => option.value === rawId)?.displayValue ?? rawId;
    default:
      return rawId;
  }
}

function findName(
  options: readonly { id?: number; name?: string }[],
  rawId: string,
): string | undefined {
  return options.find((option) => String(option.id) === rawId)?.name;
}
