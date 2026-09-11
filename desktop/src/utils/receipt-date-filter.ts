import { endOfMonth, getDaysInMonth, isValid, startOfMonth } from "date-fns";
import { FilterOperation } from "../open-api";
import { ReceiptFilterEntry } from "./receipt-filter-entry";

/** A calendar month. `month` is zero-based, matching `Date.getMonth()`. */
export interface FilterMonth {
  year: number;
  month: number;
}

/**
 * The filter entry for a whole calendar month.
 *
 * A month is expressed as `BETWEEN [first day, last day]` — the one operation
 * that can describe *any* month, so the quick date control needs no API change.
 * (`WITHIN_CURRENT_MONTH` is not equivalent: the server pins it to
 * month-start through *today*, so it can only ever mean the current month.)
 */
export function monthFilterEntry(month: FilterMonth): ReceiptFilterEntry {
  const start = startOfMonth(new Date(month.year, month.month, 1));

  return {
    operation: FilterOperation.Between,
    value: [start, endOfMonth(start)],
  };
}

/**
 * The month a filter entry describes, or `null` when it describes anything else
 * (a partial range, a `GREATER_THAN`, `WITHIN_CURRENT_MONTH`, nothing at all).
 *
 * Values arrive as `Date`s from the datepickers but as ISO **strings** once the
 * filter has been through the persisted NGXS state, so both are accepted.
 * Comparison is on calendar fields rather than on the serialized text, which
 * differs between the two forms.
 */
export function monthFromFilterEntry(entry: unknown): FilterMonth | null {
  const typedEntry = entry as ReceiptFilterEntry | undefined | null;

  if (typedEntry?.operation?.toString() !== FilterOperation.Between) {
    return null;
  }

  const value = typedEntry.value;
  if (!Array.isArray(value) || value.length !== 2) {
    return null;
  }

  const start = toDate(value[0]);
  const end = toDate(value[1]);

  if (!start || !end) {
    return null;
  }

  const spansWholeMonth =
    start.getDate() === 1 &&
    start.getFullYear() === end.getFullYear() &&
    start.getMonth() === end.getMonth() &&
    end.getDate() === getDaysInMonth(start);

  return spansWholeMonth
    ? { year: start.getFullYear(), month: start.getMonth() }
    : null;
}

/** Steps a month by `delta` months, rolling the year at the boundaries. */
export function shiftMonth(month: FilterMonth, delta: number): FilterMonth {
  const shifted = new Date(month.year, month.month + delta, 1);

  return { year: shifted.getFullYear(), month: shifted.getMonth() };
}

/** The calendar month a date falls in. */
export function monthOfDate(date: Date): FilterMonth {
  return { year: date.getFullYear(), month: date.getMonth() };
}

function toDate(value: unknown): Date | null {
  if (value instanceof Date) {
    return isValid(value) ? value : null;
  }

  if (typeof value === "string" && value.trim().length > 0) {
    const parsed = new Date(value);
    return isValid(parsed) ? parsed : null;
  }

  return null;
}
