import { TestBed } from "@angular/core/testing";
import { NgxsModule, Store } from "@ngxs/store";
import { DEFAULT_RECEIPT_TABLE_COLUMNS } from "src/interfaces";
import { FilterOperation, Group } from "../open-api";
import { GroupState } from "./group.state";
import { RemoveGroup, SetSelectedGroupId } from "./group.state.actions";
import { defaultReceiptFilter, ReceiptTableState } from "./receipt-table.state";

describe("GroupState receipt-filter reset on group change", () => {
  let store: Store;

  const groups = [
    { id: 1, name: "Group One" },
    { id: 2, name: "Group Two" },
  ] as Group[];

  const nonDefaultFilter = {
    ...defaultReceiptFilter,
    name: { operation: FilterOperation.Contains, value: "coffee" },
  } as any;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NgxsModule.forRoot([GroupState, ReceiptTableState])],
    }).compileComponents();

    store = TestBed.inject(Store);
  });

  function seed(selectedGroupId: string): void {
    store.reset({
      groups: {
        groups,
        selectedGroupId,
        selectedDashboardId: "",
      },
      receiptTable: {
        page: 3,
        pageSize: 50,
        orderBy: "created_at",
        sortDirection: "desc",
        filter: nonDefaultFilter,
        columnConfig: DEFAULT_RECEIPT_TABLE_COLUMNS,
      },
    });
  }

  it("resets the receipt filter and page when switching to a different group", () => {
    seed("1");

    store.dispatch(new SetSelectedGroupId("2"));

    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      defaultReceiptFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
    expect(store.selectSnapshot(GroupState.selectedGroupId)).toEqual("2");
  });

  it("preserves the receipt filter and page when re-selecting the same group", () => {
    seed("1");

    store.dispatch(new SetSelectedGroupId("1"));

    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      nonDefaultFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(3);
  });

  it("resets when an empty payload resolves to a different group than the current", () => {
    // Current group is 2; an empty payload resolves to groups[0] (id 1), a change.
    seed("2");

    store.dispatch(new SetSelectedGroupId(""));

    expect(store.selectSnapshot(GroupState.selectedGroupId)).toEqual("1");
    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      defaultReceiptFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
  });

  it("resets the receipt filter and page when the active group is deleted", () => {
    // Selected group is 2; deleting it falls back to groups[0] (id 1), a change.
    seed("2");

    store.dispatch(new RemoveGroup("2"));

    expect(store.selectSnapshot(GroupState.selectedGroupId)).toEqual("1");
    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      defaultReceiptFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
  });

  it("preserves the receipt filter and page when a non-active group is deleted", () => {
    // Selected group is 1; deleting group 2 leaves the selection unchanged.
    seed("1");

    store.dispatch(new RemoveGroup("2"));

    expect(store.selectSnapshot(GroupState.selectedGroupId)).toEqual("1");
    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      nonDefaultFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(3);
  });

  it("falls back to the first remaining group when the active group at index 0 is deleted", () => {
    // groups are [{id:1},{id:2}]; the active group (id 1) is at index 0. The
    // fallback must come from the post-removal list ("2"), not the stale
    // pre-removal array (which would wrongly re-select the deleted "1").
    seed("1");

    store.dispatch(new RemoveGroup("1"));

    expect(store.selectSnapshot(GroupState.selectedGroupId)).toEqual("2");
    expect(store.selectSnapshot(ReceiptTableState.filterData).filter).toEqual(
      defaultReceiptFilter
    );
    expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
  });
});

describe("GroupState.soleGroupId", () => {
  let store: Store;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NgxsModule.forRoot([GroupState, ReceiptTableState])],
    }).compileComponents();

    store = TestBed.inject(Store);
  });

  function seedGroups(groups: Partial<Group>[]): void {
    store.reset({
      groups: {
        groups: groups as Group[],
        selectedGroupId: "",
        selectedDashboardId: "",
      },
    });
  }

  it("returns the id when the user belongs to exactly one group", () => {
    seedGroups([{ id: 7, name: "My Receipts", isAllGroup: false }]);

    expect(store.selectSnapshot(GroupState.soleGroupId)).toEqual(7);
  });

  it("ignores the synthetic All group when counting", () => {
    // The All group ships alongside every real group, so a single-group user
    // still has two entries in state — it must not make the count 2.
    seedGroups([
      { id: 1, name: "All Groups", isAllGroup: true },
      { id: 7, name: "My Receipts", isAllGroup: false },
    ]);

    expect(store.selectSnapshot(GroupState.soleGroupId)).toEqual(7);
  });

  it("returns undefined when the user belongs to more than one group", () => {
    seedGroups([
      { id: 7, name: "My Receipts", isAllGroup: false },
      { id: 8, name: "Household", isAllGroup: false },
    ]);

    expect(store.selectSnapshot(GroupState.soleGroupId)).toBeUndefined();
  });

  it("returns undefined when only the All group is present", () => {
    seedGroups([{ id: 1, name: "All Groups", isAllGroup: true }]);

    expect(store.selectSnapshot(GroupState.soleGroupId)).toBeUndefined();
  });

  it("returns undefined when there are no groups at all", () => {
    seedGroups([]);

    expect(store.selectSnapshot(GroupState.soleGroupId)).toBeUndefined();
  });
});

describe("GroupState.addTargetGroupId", () => {
  let store: Store;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NgxsModule.forRoot([GroupState, ReceiptTableState])],
    }).compileComponents();

    store = TestBed.inject(Store);
  });

  const allGroup = { id: 1, name: "All Groups", isAllGroup: true };

  function seed(groups: Partial<Group>[], selectedGroupId: string): void {
    store.reset({
      groups: {
        groups: groups as Group[],
        selectedGroupId,
        selectedDashboardId: "",
      },
    });
  }

  it("returns the actively selected group when it is a real one", () => {
    seed([allGroup, { id: 7 }, { id: 8 }], "8");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toEqual(8);
  });

  it("falls back to the only group when the All group is selected", () => {
    // Where a fresh single-group user lands: the API sorts All first.
    seed([allGroup, { id: 7 }], "1");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toEqual(7);
  });

  it("falls back to the only group when the selection is stale", () => {
    // selectedGroupId is persisted, so it outlives a group the user has left.
    seed([allGroup, { id: 7 }], "404");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toEqual(7);
  });

  it("falls back to the only group when nothing is selected", () => {
    seed([allGroup, { id: 7 }], "");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toEqual(7);
  });

  it("is undefined on the All group when the user has several groups", () => {
    seed([allGroup, { id: 7 }, { id: 8 }], "1");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toBeUndefined();
  });

  it("is undefined on a stale selection when the user has several groups", () => {
    seed([allGroup, { id: 7 }, { id: 8 }], "404");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toBeUndefined();
  });

  it("is undefined when the user has no selectable group at all", () => {
    seed([allGroup], "1");

    expect(store.selectSnapshot(GroupState.addTargetGroupId)).toBeUndefined();
  });
});
