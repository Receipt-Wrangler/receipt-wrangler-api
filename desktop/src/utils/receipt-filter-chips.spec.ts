import { Category, FilterOperation, Group, ReceiptPagedRequestFilter, Tag, User } from "../open-api";
import { buildReceiptFilterChips, ReceiptFilterChipLookups } from "./receipt-filter-chips";

const lookups: ReceiptFilterChipLookups = {
  categories: [{ id: 3, name: "Groceries" }, { id: 7, name: "Dining" }] as Category[],
  tags: [{ id: 1, name: "Reimbursable" }] as Tag[],
  groups: [{ id: 4, name: "Household" }] as Group[],
  users: [{ id: 2, displayName: "Dana Kim" }] as User[],
  formatDate: (value) => (value ? new Date(value as string).toDateString() : ""),
  formatCurrency: (value) => `$${Number(value).toFixed(2)}`,
};

function filterWith(partial: Record<string, unknown>): ReceiptPagedRequestFilter {
  return partial as ReceiptPagedRequestFilter;
}

describe("buildReceiptFilterChips", () => {
  it("returns nothing for an empty or missing filter", () => {
    expect(buildReceiptFilterChips(undefined, lookups)).toEqual([]);
    expect(buildReceiptFilterChips(filterWith({ name: { operation: null, value: "" } }), lookups)).toEqual([]);
  });

  it("emits one chip per active field, in dialog order", () => {
    const chips = buildReceiptFilterChips(
      filterWith({
        status: { operation: FilterOperation.Contains, value: ["OPEN"] },
        name: { operation: FilterOperation.Contains, value: "whole" },
        tags: { operation: FilterOperation.Contains, value: [] },
      }),
      lookups
    );

    expect(chips.map((chip) => chip.key)).toEqual(["name", "status"]);
  });

  it("labels a BETWEEN date range", () => {
    const [chip] = buildReceiptFilterChips(
      filterWith({
        date: {
          operation: FilterOperation.Between,
          value: [new Date(2026, 8, 1), new Date(2026, 8, 30)],
        },
      }),
      lookups
    );

    expect(chip.label).toBe("Date between Tue Sep 01 2026 – Wed Sep 30 2026");
  });

  it("labels a currency comparison", () => {
    const [chip] = buildReceiptFilterChips(
      filterWith({ amount: { operation: FilterOperation.GreaterThan, value: 25 } }),
      lookups
    );

    expect(chip.label).toBe("Amount greater than $25.00");
  });

  it("resolves ids to names for every list-shaped field", () => {
    const chips = buildReceiptFilterChips(
      filterWith({
        paidBy: { operation: FilterOperation.Contains, value: [2] },
        group: { operation: FilterOperation.Contains, value: [4] },
        categories: { operation: FilterOperation.Contains, value: [3, 7] },
        status: { operation: FilterOperation.Contains, value: ["OPEN"] },
      }),
      lookups
    );

    expect(chips.map((chip) => chip.label)).toEqual([
      "Paid by contains Dana Kim",
      "Group contains Household",
      "Categories contains Groceries, Dining",
      "Status contains Open",
    ]);
  });

  // A category outside the user's grants, or a group they have left, still has
  // to be visible so it can be cleared.
  it("falls back to the raw id when a lookup misses", () => {
    const [chip] = buildReceiptFilterChips(
      filterWith({ categories: { operation: FilterOperation.Contains, value: [99] } }),
      lookups
    );

    expect(chip.label).toBe("Categories contains 99");
  });

  // WITHIN_CURRENT_MONTH carries no value, so the label stops at the operation.
  it("labels WITHIN_CURRENT_MONTH without a value segment", () => {
    const [chip] = buildReceiptFilterChips(
      filterWith({ date: { operation: FilterOperation.WithinCurrentMonth, value: null } }),
      lookups
    );

    expect(chip.label).toBe("Date within current month");
  });

  // The quick date control used to suppress its own field's chip while the
  // month stepper named that month. Its target field is now selectable, so the
  // chip row is the only place that says WHICH date column is filtered — and a
  // condition with no chip is one the user cannot clear from here.
  it("chips a whole-month date condition like any other, alongside the rest", () => {
    const filter = filterWith({
      date: { operation: FilterOperation.Between, value: [new Date(2026, 8, 1), new Date(2026, 8, 30)] },
      name: { operation: FilterOperation.Contains, value: "whole" },
    });

    expect(buildReceiptFilterChips(filter, lookups).map((chip) => chip.key)).toEqual(["date", "name"]);
  });
});
