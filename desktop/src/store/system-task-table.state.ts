import { Injectable } from "@angular/core";
import { Action, Selector, State, StateContext } from "@ngxs/store";
import { SystemTaskTableInterface } from "src/interfaces";
import { SortDirection, SystemTaskPagedRequestFilter } from "../open-api";
import { isFilterEntryActive } from "../utils/receipt-filter-entry";
import { PagedTableState } from "./paged-table.state";
import {
  ResetSystemTaskFilter,
  SetOrderBy,
  SetPage,
  SetPageSize,
  SetSortDirection,
  SetSystemTaskFilter,
  SetSystemTaskFilterField
} from "./system-task-table.state.actions";

/**
 * A pristine filter. This is a factory rather than a shared constant because
 * the value is written straight into state on reset — handing out the same
 * object every time would let a later in-place edit corrupt the default for the
 * rest of the session. Same reasoning as `buildDefaultReceiptFilter`.
 */
export function buildDefaultSystemTaskFilter(): SystemTaskPagedRequestFilter {
  return {
    type: {
      operation: null,
      value: [],
    },
    ranBy: {
      operation: null,
      value: [],
    },
    startedAt: {
      operation: null,
      value: null,
    },
    endedAt: {
      operation: null,
      value: null,
    },
  } as SystemTaskPagedRequestFilter;
}

@State<SystemTaskTableInterface>({
  name: "systemTaskTable",
  defaults: {
    page: 1,
    pageSize: 50,
    orderBy: "started_at",
    sortDirection: SortDirection.Desc,
    filter: buildDefaultSystemTaskFilter(),
  },
})
@Injectable()
export class SystemTaskTableState extends PagedTableState {
  /**
   * The slice is persisted to localStorage, so a session saved before the
   * filter shipped rehydrates with no `filter` key at all. Every read goes
   * through here rather than touching `state.filter` directly.
   */
  @Selector()
  static filter(state: SystemTaskTableInterface): SystemTaskPagedRequestFilter {
    return state.filter ?? buildDefaultSystemTaskFilter();
  }

  @Selector()
  static numFiltersApplied(state: SystemTaskTableInterface): number {
    const filter: any = state.filter ?? buildDefaultSystemTaskFilter();

    // Shares isFilterEntryActive with the filter chips, so the badge count and
    // the chips can never disagree about what counts as a condition.
    return Object.keys(filter).filter((key) => isFilterEntryActive(filter[key])).length;
  }

  @Action(SetPage)
  setPage({ patchState }: StateContext<SystemTaskTableInterface>, payload: SetPage) {
    patchState({
      page: payload.page,
    });
  }

  @Action(SetPageSize)
  setPageSize(
    { patchState }: StateContext<SystemTaskTableInterface>,
    payload: SetPageSize
  ) {
    patchState({
      pageSize: payload.pageSize,
    });
  }

  @Action(SetOrderBy)
  setOrderBy(
    { patchState }: StateContext<SystemTaskTableInterface>,
    payload: SetOrderBy
  ) {
    patchState({
      orderBy: payload.orderBy,
    });
  }

  @Action(SetSortDirection)
  setSortDirection(
    { patchState }: StateContext<SystemTaskTableInterface>,
    payload: SetSortDirection
  ) {
    patchState({
      sortDirection: payload.sortDirection,
    });
  }

  @Action(SetSystemTaskFilter)
  setSystemTaskFilter(
    { patchState }: StateContext<SystemTaskTableInterface>,
    payload: SetSystemTaskFilter
  ) {
    patchState({
      filter: payload.data,
    });
  }

  @Action(SetSystemTaskFilterField)
  setSystemTaskFilterField(
    { getState, patchState }: StateContext<SystemTaskTableInterface>,
    payload: SetSystemTaskFilterField
  ) {
    // A cleared field is rebuilt from a fresh default so the empty shape (list
    // fields go back to [], scalars to null) has one source of truth, and the
    // spread allocates a new container rather than writing through the current
    // filter — which may be shared with another reference.
    const clearedEntry = (buildDefaultSystemTaskFilter() as any)[payload.field];

    patchState({
      filter: {
        ...(getState().filter ?? buildDefaultSystemTaskFilter()),
        [payload.field]: payload.entry ?? clearedEntry,
      },
    });
  }

  @Action(ResetSystemTaskFilter)
  resetSystemTaskFilter({ patchState }: StateContext<SystemTaskTableInterface>) {
    patchState({
      filter: buildDefaultSystemTaskFilter(),
    });
  }
}
