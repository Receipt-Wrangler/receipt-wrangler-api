import { FILTER_OPERATION_DISPLAY_VALUES } from "../constants/filter-operations-options.constant";
import { FilterField } from "../constants/filter-fields.constant";
import { FilterOperation } from "../open-api";
import { isFilterEntryActive, ReceiptFilterEntry } from "./receipt-filter-entry";

export interface FilterChip<TKey extends string = string> {
  key: TKey;
  label: string;
}

/**
 * Everything the label builder needs that is not on the filter itself. Passed
 * in rather than looked up here so the builder stays pure and testable — the
 * same shape `report-template-summary.ts` uses.
 *
 * `resolveOptionName` turns one stored id into the name a human reads. It must
 * fall back to the raw id rather than dropping the value, so a filter that is
 * actively removing rows is never invisible.
 */
export interface FilterChipLookups {
  formatDate: (value: unknown) => string;
  formatCurrency: (value: unknown) => string;
  resolveOptionName: (key: string, id: unknown) => string;
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
export function buildFilterChips<TKey extends string>(
  fields: readonly FilterField<TKey>[],
  filter: Record<string, unknown> | undefined | null,
  lookups: FilterChipLookups,
  omitKeys: readonly string[] = [],
): FilterChip<TKey>[] {
  if (!filter) {
    return [];
  }

  return fields
    .filter((field) => !omitKeys.includes(field.key) && isFilterEntryActive(filter[field.key]))
    .map((field) => ({
      key: field.key,
      label: buildLabel(field, filter[field.key] as ReceiptFilterEntry, lookups),
    }));
}

function buildLabel(
  field: FilterField,
  entry: ReceiptFilterEntry,
  lookups: FilterChipLookups,
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
  field: FilterField,
  entry: ReceiptFilterEntry,
  operation: string,
  lookups: FilterChipLookups,
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
    return value.map((id) => lookups.resolveOptionName(field.key, id)).join(", ");
  }

  return value?.toString() ?? "";
}
