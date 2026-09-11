import { FilterOperation } from "../open-api";

/**
 * A single `{ operation, value }` pair out of a `ReceiptPagedRequestFilter`.
 * The generated model types every field as `object`, so callers hand us this.
 */
export interface ReceiptFilterEntry {
  operation?: FilterOperation | null;
  value?: unknown;
}

/**
 * Whether a filter field actually narrows the result set.
 *
 * `WITHIN_CURRENT_MONTH` is the one operation that carries no value — the API
 * pins the range to the current month itself — so it counts on the operation
 * alone. Everything else needs a non-empty value, and `"0"` is excluded so an
 * untouched numeric field doesn't read as a condition.
 */
export function isFilterEntryActive(entry: unknown): boolean {
  const typedEntry = entry as ReceiptFilterEntry | undefined | null;
  const stringValue = typedEntry?.value?.toString();

  if (stringValue !== undefined && stringValue.length > 0 && stringValue !== "0") {
    return true;
  }

  return typedEntry?.operation?.toString() === FilterOperation.WithinCurrentMonth;
}
