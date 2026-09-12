import { SYSTEM_RAN_BY_OPTION_ID } from "../constants";
import { FilterOperation, SystemTaskPagedRequestFilter, User } from "../open-api";
import { buildSystemTaskFilterChips } from "./system-task-filter-chips";

describe("buildSystemTaskFilterChips", () => {
  const users = [
    { id: 4, displayName: "Alice" },
    { id: 7, displayName: "Bob" },
  ] as User[];

  const lookups = {
    users,
    formatDate: (value: unknown) => `date(${value})`,
  };

  const build = (filter: any): { key: string; label: string }[] =>
    buildSystemTaskFilterChips(filter as SystemTaskPagedRequestFilter, lookups);

  it("returns nothing for an absent or pristine filter", () => {
    expect(build(null)).toEqual([]);
    expect(
      build({
        type: { operation: null, value: [] },
        ranBy: { operation: null, value: [] },
        startedAt: { operation: null, value: null },
        endedAt: { operation: null, value: null },
      })
    ).toEqual([]);
  });

  it("names task types by their display value", () => {
    const chips = build({
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN", "EMAIL_READ"] },
    });

    expect(chips).toEqual([{ key: "type", label: "Type contains Quick Scan, Email Read" }]);
  });

  it("names the System sentinel and resolves real users", () => {
    const chips = build({
      ranBy: { operation: FilterOperation.Contains, value: [SYSTEM_RAN_BY_OPTION_ID, 4] },
    });

    expect(chips).toEqual([{ key: "ranBy", label: "Ran By contains System, Alice" }]);
  });

  // A user who has since been deleted must still leave a clearable chip, or a
  // filter that is actively removing rows becomes invisible.
  it("falls back to the raw id for an unresolvable user", () => {
    const chips = build({
      ranBy: { operation: FilterOperation.Contains, value: [999] },
    });

    expect(chips).toEqual([{ key: "ranBy", label: "Ran By contains 999" }]);
  });

  it("formats both bounds of a date range", () => {
    const chips = build({
      startedAt: {
        operation: FilterOperation.Between,
        value: ["2026-03-10T00:00:00Z", "2026-03-12T00:00:00Z"],
      },
    });

    expect(chips).toEqual([
      {
        key: "startedAt",
        label: "Started At between date(2026-03-10T00:00:00Z) – date(2026-03-12T00:00:00Z)",
      },
    ]);
  });

  it("renders one chip per active condition, in field order", () => {
    const chips = build({
      endedAt: { operation: FilterOperation.LessThan, value: "2026-03-12T00:00:00Z" },
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
    });

    expect(chips.map((chip) => chip.key)).toEqual(["type", "endedAt"]);
  });
});
