import { FilterOperation } from "../open-api";
import { monthFilterEntry, monthFromFilterEntry, monthOfDate, shiftMonth } from "./receipt-date-filter";

describe("receipt-date-filter", () => {
  describe("monthFilterEntry", () => {
    it("expresses a month as a BETWEEN over its first and last day", () => {
      const entry = monthFilterEntry({ year: 2026, month: 8 });
      const [start, end] = entry.value as Date[];

      expect(entry.operation).toBe(FilterOperation.Between);
      expect(start.getFullYear()).toBe(2026);
      expect(start.getMonth()).toBe(8);
      expect(start.getDate()).toBe(1);
      expect(end.getMonth()).toBe(8);
      expect(end.getDate()).toBe(30);
    });

    it("handles a 28-day February", () => {
      const [, end] = monthFilterEntry({ year: 2026, month: 1 }).value as Date[];

      expect(end.getDate()).toBe(28);
    });
  });

  describe("monthFromFilterEntry", () => {
    it("round-trips a month", () => {
      expect(monthFromFilterEntry(monthFilterEntry({ year: 2026, month: 8 }))).toEqual({
        year: 2026,
        month: 8,
      });
    });

    // The filter is persisted to localStorage, so every Date comes back as an
    // ISO string. Without this the stepper would silently read "Custom" after
    // every page reload.
    it("round-trips a month that has been through JSON serialization", () => {
      const persisted = JSON.parse(JSON.stringify(monthFilterEntry({ year: 2026, month: 11 })));

      expect(monthFromFilterEntry(persisted)).toEqual({ year: 2026, month: 11 });
    });

    // The dialog's datepickers write local midnight for both ends, with no
    // end-of-day component — that is still September.
    it("matches on calendar days, not on the exact instant", () => {
      expect(
        monthFromFilterEntry({
          operation: FilterOperation.Between,
          value: [new Date(2026, 8, 1), new Date(2026, 8, 30)],
        })
      ).toEqual({ year: 2026, month: 8 });
    });

    it("returns null for a range that is not exactly one whole month", () => {
      const partial = { operation: FilterOperation.Between, value: [new Date(2026, 8, 1), new Date(2026, 8, 15)] };
      const spanning = { operation: FilterOperation.Between, value: [new Date(2026, 8, 1), new Date(2026, 9, 31)] };
      const lateStart = { operation: FilterOperation.Between, value: [new Date(2026, 8, 2), new Date(2026, 8, 30)] };

      expect(monthFromFilterEntry(partial)).toBeNull();
      expect(monthFromFilterEntry(spanning)).toBeNull();
      expect(monthFromFilterEntry(lateStart)).toBeNull();
    });

    it("returns null for any other operation or a malformed value", () => {
      expect(monthFromFilterEntry({ operation: FilterOperation.GreaterThan, value: new Date(2026, 8, 1) })).toBeNull();
      expect(monthFromFilterEntry({ operation: FilterOperation.WithinCurrentMonth, value: null })).toBeNull();
      expect(monthFromFilterEntry({ operation: FilterOperation.Between, value: [new Date(2026, 8, 1)] })).toBeNull();
      expect(monthFromFilterEntry({ operation: FilterOperation.Between, value: ["nonsense", "junk"] })).toBeNull();
      expect(monthFromFilterEntry({ operation: null, value: null })).toBeNull();
      expect(monthFromFilterEntry(undefined)).toBeNull();
    });
  });

  describe("shiftMonth", () => {
    it("steps within a year", () => {
      expect(shiftMonth({ year: 2026, month: 8 }, 1)).toEqual({ year: 2026, month: 9 });
      expect(shiftMonth({ year: 2026, month: 8 }, -1)).toEqual({ year: 2026, month: 7 });
    });

    it("rolls the year at both boundaries", () => {
      expect(shiftMonth({ year: 2026, month: 11 }, 1)).toEqual({ year: 2027, month: 0 });
      expect(shiftMonth({ year: 2026, month: 0 }, -1)).toEqual({ year: 2025, month: 11 });
    });
  });

  describe("monthOfDate", () => {
    it("reads the calendar month off a date", () => {
      expect(monthOfDate(new Date(2026, 8, 11))).toEqual({ year: 2026, month: 8 });
    });
  });
});
