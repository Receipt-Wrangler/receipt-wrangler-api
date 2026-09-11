import { Injectable } from "@angular/core";
import { Action, Selector, State, StateContext } from "@ngxs/store";
import { ReceiptTableInterface } from "../interfaces";
import { DEFAULT_RECEIPT_TABLE_COLUMNS, ReceiptTableColumnConfig } from "../interfaces/receipt-table-column-config.interface";
import { ReceiptPagedRequestFilter } from "../open-api";
import { isFilterEntryActive } from "../utils/receipt-filter-entry";
import { ResetReceiptFilter, SetColumnConfig, SetPage, SetPageSize, SetReceiptFilter, SetReceiptFilterData, SetReceiptFilterField } from "./receipt-table.actions";

/**
 * A pristine filter. This is a factory rather than a shared constant because
 * the value is written straight into state on reset — handing out the same
 * object every time would let a later in-place edit corrupt the default for the
 * rest of the session.
 */
export function buildDefaultReceiptFilter(): ReceiptPagedRequestFilter {
  return {
    date: {
      operation: null,
      value: null
    },
    amount: {
      operation: null,
      value: null,
    },
    name: {
      operation: null,
      value: null,
    },
    paidBy: {
      operation: null,
      value: [],
    },
    categories: {
      operation: null,
      value: [],
    },
    tags: {
      operation: null,
      value: [],
    },
    status: {
      operation: null,
      value: [],
    },
    group: {
      operation: null,
      value: [],
    },
    resolvedDate: {
      operation: null,
      value: null,
    },
    createdAt: {
      operation: null,
      value: null,
    },
  } as ReceiptPagedRequestFilter;
}

export const defaultReceiptFilter = buildDefaultReceiptFilter();

// TODO: look into fixing date equals
@State<ReceiptTableInterface>({
  name: "receiptTable",
  defaults: {
    page: 1,
    pageSize: 50,
    orderBy: "created_at",
    sortDirection: "desc",
    filter: buildDefaultReceiptFilter(),
    columnConfig: DEFAULT_RECEIPT_TABLE_COLUMNS,
  },
})
@Injectable()
export class ReceiptTableState {
  @Selector()
  static page(state: ReceiptTableInterface): number {
    return state.page;
  }

  @Selector()
  static pageSize(state: ReceiptTableInterface): number {
    return state.pageSize;
  }

  @Selector()
  static filterData(state: ReceiptTableInterface): ReceiptTableInterface {
    return state;
  }

  @Selector()
  static numFiltersApplied(state: ReceiptTableInterface): number {
    const filter: any = state.filter;

    // Shares isFilterEntryActive with the filter chips, so the badge count and
    // the chips can never disagree about what counts as a condition.
    return Object.keys(filter).filter((key) => isFilterEntryActive(filter[key])).length;
  }

  @Selector()
  static columnConfig(state: ReceiptTableInterface): ReceiptTableColumnConfig[] {
    return state.columnConfig || DEFAULT_RECEIPT_TABLE_COLUMNS;
  }

  @Action(SetPage)
  setPage(
    { patchState }: StateContext<ReceiptTableInterface>,
    payload: SetPage
  ) {
    patchState({
      page: payload.page,
    });
  }

  @Action(SetPageSize)
  setPageSize(
    { patchState }: StateContext<ReceiptTableInterface>,
    payload: SetPageSize
  ) {
    patchState({
      pageSize: payload.pageSize,
    });
  }

  @Action(SetReceiptFilterData)
  setReceiptFilterData(
    { patchState }: StateContext<ReceiptTableInterface>,
    payload: SetReceiptFilterData
  ) {
    patchState(payload.data);
  }

  @Action(SetReceiptFilter)
  setReceiptFilter(
    { patchState }: StateContext<ReceiptTableInterface>,
    payload: SetReceiptFilter
  ) {
    patchState({
      filter: payload.data,
    });
  }

  @Action(SetReceiptFilterField)
  setReceiptFilterField(
    { getState, patchState }: StateContext<ReceiptTableInterface>,
    payload: SetReceiptFilterField
  ) {
    // A cleared field is rebuilt from a fresh default so the empty shape (list
    // fields go back to [], scalars to null) has one source of truth, and the
    // spread allocates a new container rather than writing through the current
    // filter — which may be shared with another reference.
    const clearedEntry = (buildDefaultReceiptFilter() as any)[payload.field];

    patchState({
      filter: {
        ...getState().filter,
        [payload.field]: payload.entry ?? clearedEntry,
      },
    });
  }

  @Action(ResetReceiptFilter)
  resetFilter({ patchState }: StateContext<ReceiptTableInterface>) {
    patchState({
      filter: buildDefaultReceiptFilter(),
    });
  }

  @Action(SetColumnConfig)
  setColumnConfig(
    { patchState }: StateContext<ReceiptTableInterface>,
    payload: SetColumnConfig
  ) {
    patchState({
      columnConfig: payload.columnConfig,
    });
  }
}
