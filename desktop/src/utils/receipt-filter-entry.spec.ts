import { FilterOperation } from "../open-api";
import { isFilterEntryActive } from "./receipt-filter-entry";

describe("isFilterEntryActive", () => {
  it("treats a populated value as active", () => {
    expect(isFilterEntryActive({ operation: FilterOperation.Contains, value: "whole" })).toBe(true);
    expect(isFilterEntryActive({ operation: FilterOperation.Contains, value: [1, 2] })).toBe(true);
    expect(isFilterEntryActive({ operation: FilterOperation.Equals, value: 12.5 })).toBe(true);
  });

  it("treats an empty value as inactive", () => {
    expect(isFilterEntryActive({ operation: null, value: null })).toBe(false);
    expect(isFilterEntryActive({ operation: null, value: [] })).toBe(false);
    expect(isFilterEntryActive({ operation: null, value: "" })).toBe(false);
    expect(isFilterEntryActive(undefined)).toBe(false);
    expect(isFilterEntryActive(null)).toBe(false);
  });

  // An untouched numeric field stringifies to "0", which must not read as a
  // condition — this is the rule the filter-count badge has always applied.
  it("treats a zero value as inactive", () => {
    expect(isFilterEntryActive({ operation: FilterOperation.Equals, value: 0 })).toBe(false);
  });

  // The one operation that carries no value: the API pins the range itself.
  it("treats WITHIN_CURRENT_MONTH as active despite having no value", () => {
    expect(
      isFilterEntryActive({ operation: FilterOperation.WithinCurrentMonth, value: null })
    ).toBe(true);
  });
});
