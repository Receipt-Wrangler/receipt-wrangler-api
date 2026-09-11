import { FILTER_OPERATION_DISPLAY_VALUES } from "../constants/filter-operations-options.constant";
import {
  RECEIPT_FILTER_FIELDS,
  ReceiptFilterField,
  ReceiptFilterFieldKey,
} from "../constants/receipt-filter-fields.constant";
import { RECEIPT_STATUS_OPTIONS } from "../constants/receipt-status-options";
import { Category, FilterOperation, Group, ReceiptPagedRequestFilter, Tag, User } from "../open-api";
import { isFilterEntryActive, ReceiptFilterEntry } from "./receipt-filter-entry";

export interface ReceiptFilterChip {
  key: ReceiptFilterFieldKey;
  label: string;
}

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

const BETWEEN_SEPARATOR = " – ";

/**
 * One chip per filter field that actually narrows the result set, labelled
 * `"<Field> <operation> <value>"`.
 *
 * `omitKeys` lets a caller suppress a field it already renders another way —
 * the receipts table passes `["date"]` while the month stepper is displaying
 * that exact month, so the same condition never appears twice.
 */
export function buildReceiptFilterChips(
  filter: ReceiptPagedRequestFilter | undefined | null,
  lookups: ReceiptFilterChipLookups,
  omitKeys: readonly string[] = [],
): ReceiptFilterChip[] {
  if (!filter) {
    return [];
  }

  return RECEIPT_FILTER_FIELDS.filter(
    (field) =>
      !omitKeys.includes(field.key) &&
      isFilterEntryActive((filter as Record<string, unknown>)[field.key]),
  ).map((field) => ({
    key: field.key,
    label: buildLabel(field, (filter as Record<string, unknown>)[field.key] as ReceiptFilterEntry, lookups),
  }));
}

function buildLabel(
  field: ReceiptFilterField,
  entry: ReceiptFilterEntry,
  lookups: ReceiptFilterChipLookups,
): string {
  const operation = entry?.operation?.toString() ?? "";
  const operationLabel = (FILTER_OPERATION_DISPLAY_VALUES[operation] ?? "").toLowerCase();

  // WITHIN_CURRENT_MONTH carries no value, so the label stops at the operation.
  const valueLabel =
    operation === FilterOperation.WithinCurrentMonth
      ? ""
      : buildValueLabel(field, entry, operation, lookups);

  return [field.label, operationLabel, valueLabel].filter((part) => !!part).join(" ");
}

function buildValueLabel(
  field: ReceiptFilterField,
  entry: ReceiptFilterEntry,
  operation: string,
  lookups: ReceiptFilterChipLookups,
): string {
  const value = entry?.value;

  if (field.type === "date" || field.type === "number") {
    const format = field.type === "date" ? lookups.formatDate : lookups.formatCurrency;

    if (operation === FilterOperation.Between && Array.isArray(value)) {
      return [format(value[0]), format(value[1])].join(BETWEEN_SEPARATOR);
    }

    return format(value);
  }

  if (Array.isArray(value)) {
    return value.map((id) => resolveOptionName(field.key, id, lookups)).join(", ");
  }

  return value?.toString() ?? "";
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
