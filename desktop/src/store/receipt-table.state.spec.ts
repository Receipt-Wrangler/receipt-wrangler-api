import { TestBed } from "@angular/core/testing";
import { NgxsModule, Store } from "@ngxs/store";
import { DEFAULT_RECEIPT_TABLE_COLUMNS, ReceiptTableInterface } from "src/interfaces";
import { FilterOperation, ReceiptPagedRequestFilter, ReceiptStatus } from "../open-api";
import { ResetReceiptFilter, SetPage, SetPageSize, SetQuickDateField, SetReceiptFilter, SetReceiptFilterData, SetReceiptFilterField, } from "./receipt-table.actions";
import { buildDefaultReceiptFilter, defaultReceiptFilter, ReceiptTableState } from "./receipt-table.state";

describe("ReceiptTableState", () => {
  let store: Store;
  let filledFilter: ReceiptPagedRequestFilter;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NgxsModule.forRoot([ReceiptTableState])],
    }).compileComponents();

    filledFilter = {
      date: {
        operation: FilterOperation.Equals,
        value: "2023-01-06",
      },
      name: {
        operation: FilterOperation.Equals,
        value: "hello world",
      },
      amount: {
        operation: FilterOperation.GreaterThan,
        value: 12.05,
      },
      paidBy: {
        operation: FilterOperation.Contains,
        value: [1],
      },
      categories: {
        operation: FilterOperation.Contains,
        value: [2],
      },
      tags: {
        operation: FilterOperation.Contains,
        value: [3, 4],
      },
      status: {
        operation: FilterOperation.Contains,
        value: [ReceiptStatus.Open],
      },
      resolvedDate: {
        operation: FilterOperation.GreaterThan,
        value: "2023-01-06",
      },
    } as any;

    store = TestBed.inject(Store);
  });

  it("should return page number", () => {
    const result = store.selectSnapshot(ReceiptTableState.page);

    expect(result).toEqual(1);
  });

  it("should return pageSize", () => {
    const result = store.selectSnapshot(ReceiptTableState.pageSize);

    expect(result).toEqual(50);
  });

  it("should return the full state", () => {
    const result = store.selectSnapshot(ReceiptTableState.filterData);

    expect(result).toEqual({
      page: 1,
      pageSize: 50,
      orderBy: "created_at",
      sortDirection: "desc",
      filter: defaultReceiptFilter,
      quickDateField: "date",
      columnConfig: DEFAULT_RECEIPT_TABLE_COLUMNS,
    });
  });

  it("should return 0 since no filters are applied", () => {
    const result = store.selectSnapshot(ReceiptTableState.numFiltersApplied);

    expect(result).toEqual(0);
  });

  it("should return 8 since all filters are applied", () => {
    store.reset({
      receiptTable: {
        filter: filledFilter,
      },
    });
    const result = store.selectSnapshot(ReceiptTableState.numFiltersApplied);

    expect(result).toEqual(8);
  });

  it("should return 8 since all filters are applied, with date set as within current month", () => {
    (filledFilter.date as any).operation = FilterOperation.WithinCurrentMonth;
    (filledFilter.date as any).value = undefined;

    store.reset({
      receiptTable: {
        filter: filledFilter,
      },
    });
    const result = store.selectSnapshot(ReceiptTableState.numFiltersApplied);

    expect(result).toEqual(8);
  });

  it("should set page", () => {
    store.dispatch(new SetPage(40));

    const result = store.selectSnapshot(ReceiptTableState.page);
    expect(result).toEqual(40);
  });

  it("should set page size", () => {
    store.dispatch(new SetPageSize(100));

    const result = store.selectSnapshot(ReceiptTableState.pageSize);
    expect(result).toEqual(100);
  });

  it("should set filter data", () => {
    const filterData: ReceiptTableInterface = {
      page: 20,
      pageSize: 40,
      orderBy: "amount",
      sortDirection: "asc",
      filter: {} as any,
    };
    store.dispatch(new SetReceiptFilterData(filterData));

    // The payload omits columnConfig and quickDateField, and patchState only
    // touches the keys it is handed — so sorting cannot reset either.
    const result = store.selectSnapshot(ReceiptTableState.filterData);
    expect(result).toEqual({
      ...filterData,
      quickDateField: "date",
      columnConfig: DEFAULT_RECEIPT_TABLE_COLUMNS,
    });
  });

  it("should set filter receipt filter", () => {
    store.dispatch(new SetReceiptFilter(filledFilter));

    const result = store.selectSnapshot(ReceiptTableState.filterData);
    expect(result.filter).toEqual(filledFilter);
  });

  it("should reset filter", () => {
    store.reset({
      receiptTable: {
        filter: filledFilter,
      },
    });

    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      filledFilter
    );

    store.dispatch(new ResetReceiptFilter());

    const result = store.selectSnapshot(ReceiptTableState.filterData).filter;
    expect(result).toEqual(defaultReceiptFilter);
  });
  it("should count a zero-valued filter field", () => {
    store.reset({
      receiptTable: {
        filter: { ...defaultReceiptFilter, amount: { operation: FilterOperation.Equals, value: 0 } },
      },
    });

    expect(store.selectSnapshot(ReceiptTableState.numFiltersApplied)).toEqual(1);
  });

  it("should set a single filter field, leaving the others alone", () => {
    store.dispatch(
      new SetReceiptFilterField("status", {
        operation: FilterOperation.Contains,
        value: [ReceiptStatus.Open],
      })
    );

    const result = store.selectSnapshot(ReceiptTableState.filterData).filter as any;
    expect(result.status).toEqual({
      operation: FilterOperation.Contains,
      value: [ReceiptStatus.Open],
    });
    expect(result.categories).toEqual({ operation: null, value: [] });
    expect(result.date).toEqual({ operation: null, value: null });
  });

  it("should clear a single filter field back to its default empty shape", () => {
    store.reset({ receiptTable: { filter: filledFilter } });

    store.dispatch(new SetReceiptFilterField("categories", null));
    store.dispatch(new SetReceiptFilterField("date", null));

    const result = store.selectSnapshot(ReceiptTableState.filterData).filter as any;
    // A list field clears to [], a scalar to null.
    expect(result.categories).toEqual({ operation: null, value: [] });
    expect(result.date).toEqual({ operation: null, value: null });
    expect(result.name).toEqual(filledFilter.name);
  });

  it("should set the quick date field", () => {
    store.dispatch(new SetQuickDateField("resolvedDate"));

    expect(store.selectSnapshot(ReceiptTableState.quickDateField)).toEqual("resolvedDate");
  });

  // The quick date field arrived after this slice was already being persisted to
  // localStorage, so every existing install hydrates a state without the key.
  // @State defaults do not run for a hydrated state, which is why the fallback
  // has to live in the selector.
  it("should fall back to date for a state persisted before the key existed", () => {
    store.reset({ receiptTable: { filter: filledFilter } });

    expect(store.selectSnapshot(ReceiptTableState.quickDateField)).toEqual("date");
  });

  it("should reset the quick date field along with the filter", () => {
    store.dispatch(new SetQuickDateField("createdAt"));
    store.dispatch(new ResetReceiptFilter());

    expect(store.selectSnapshot(ReceiptTableState.quickDateField)).toEqual("date");
  });

  // ResetReceiptFilter writes a default filter straight into state, so without a
  // fresh object per write a later field update would corrupt the module-level
  // default for the rest of the session.
  it("should never mutate the exported default filter", () => {
    store.dispatch(new ResetReceiptFilter());
    store.dispatch(
      new SetReceiptFilterField("name", { operation: FilterOperation.Contains, value: "whole" })
    );

    expect(defaultReceiptFilter).toEqual(buildDefaultReceiptFilter());
    expect((defaultReceiptFilter as any).name).toEqual({ operation: null, value: null });
  });
});
