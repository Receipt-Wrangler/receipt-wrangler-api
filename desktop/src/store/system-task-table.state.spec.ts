import { TestBed } from "@angular/core/testing";
import { NgxsModule, Store } from "@ngxs/store";
import { FilterOperation, SortDirection } from "../open-api";
import { buildDefaultSystemTaskFilter, SystemTaskTableState } from "./system-task-table.state";
import {
  ResetSystemTaskFilter,
  SetOrderBy,
  SetPage,
  SetPageSize,
  SetSortDirection,
  SetSystemTaskFilter,
  SetSystemTaskFilterField
} from "./system-task-table.state.actions";

describe("SystemTaskTableState", () => {
  let store: Store;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [NgxsModule.forRoot([SystemTaskTableState])]
    });

    store = TestBed.inject(Store);
  });

  it("should set page", () => {
    store.dispatch(new SetPage(2));
    const state = store.selectSnapshot(SystemTaskTableState.state);
    expect(state.page).toBe(2);
  });

  it("should set page size", () => {
    store.dispatch(new SetPageSize(100));
    const state = store.selectSnapshot(SystemTaskTableState.state);
    expect(state.pageSize).toBe(100);
  });

  it("should set order by", () => {
    store.dispatch(new SetOrderBy("type"));
    const state = store.selectSnapshot(SystemTaskTableState.state);
    expect(state.orderBy).toBe("type");
  });

  it("should set sort direction", () => {
    store.dispatch(new SetSortDirection(SortDirection.Asc));
    const state = store.selectSnapshot(SystemTaskTableState.state);
    expect(state.sortDirection).toBe(SortDirection.Asc);
  });

  it("defaults to a pristine filter", () => {
    expect(store.selectSnapshot(SystemTaskTableState.filter)).toEqual(buildDefaultSystemTaskFilter());
    expect(store.selectSnapshot(SystemTaskTableState.numFiltersApplied)).toBe(0);
  });

  it("hands out a fresh default each time, so an in-place edit cannot corrupt it", () => {
    const first: any = buildDefaultSystemTaskFilter();
    first.type.value.push("QUICK_SCAN");

    expect((buildDefaultSystemTaskFilter() as any).type.value).toEqual([]);
  });

  it("should set the whole filter and count the applied conditions", () => {
    store.dispatch(new SetSystemTaskFilter({
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
      ranBy: { operation: FilterOperation.Contains, value: [-1] },
      startedAt: { operation: null, value: null },
      endedAt: { operation: null, value: null },
    } as any));

    expect(store.selectSnapshot(SystemTaskTableState.numFiltersApplied)).toBe(2);
  });

  it("should clear a single field back to its default without touching the others", () => {
    store.dispatch(new SetSystemTaskFilter({
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
      ranBy: { operation: FilterOperation.Contains, value: [-1] },
      startedAt: { operation: null, value: null },
      endedAt: { operation: null, value: null },
    } as any));

    store.dispatch(new SetSystemTaskFilterField("type", null));

    const filter: any = store.selectSnapshot(SystemTaskTableState.filter);
    expect(filter.type).toEqual({ operation: null, value: [] });
    expect(filter.ranBy.value).toEqual([-1]);
    expect(store.selectSnapshot(SystemTaskTableState.numFiltersApplied)).toBe(1);
  });

  it("should reset the filter", () => {
    store.dispatch(new SetSystemTaskFilterField("ranBy", { operation: FilterOperation.Contains, value: [4] }));
    store.dispatch(new ResetSystemTaskFilter());

    expect(store.selectSnapshot(SystemTaskTableState.filter)).toEqual(buildDefaultSystemTaskFilter());
  });

  // The slice is persisted, so a session saved before the filter shipped
  // rehydrates with no `filter` key at all. Every read has to survive that.
  it("falls back to a default filter when the persisted state carries none", () => {
    store.reset({ systemTaskTable: { page: 1, pageSize: 50, orderBy: "started_at", sortDirection: "desc" } });

    expect(store.selectSnapshot(SystemTaskTableState.filter)).toEqual(buildDefaultSystemTaskFilter());
    expect(store.selectSnapshot(SystemTaskTableState.numFiltersApplied)).toBe(0);

    store.dispatch(new SetSystemTaskFilterField("type", { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] }));

    const filter: any = store.selectSnapshot(SystemTaskTableState.filter);
    expect(filter.type.value).toEqual(["QUICK_SCAN"]);
    expect(filter.ranBy).toEqual({ operation: null, value: [] });
  });
});
