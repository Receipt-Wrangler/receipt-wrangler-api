import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA, provideZonelessChangeDetection } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatChipsModule } from "@angular/material/chips";
import { MatDialogModule } from "@angular/material/dialog";
import { MatMenuModule } from "@angular/material/menu";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { MatTooltipModule } from "@angular/material/tooltip";
import { ActivatedRoute, provideRouter } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { of } from "rxjs";
import { PipesModule } from "src/pipes/pipes.module";
import { ReceiptTableState } from "src/store/receipt-table.state";
import { MonthStepperComponent } from "../../shared-ui/month-stepper/month-stepper.component";
import { ApiModule, FilterOperation, Permission, Receipt, ReceiptStatus } from "../../open-api";
import { ReceiptFilterService } from "../../services/receipt-filter.service";
import { AuthState, GroupState, UserState } from "../../store";
import { SetPermissions } from "../../store/auth.state.actions";
import { SetReceiptFilter } from "../../store/receipt-table.actions";
import { ReceiptsTableComponent } from "./receipts-table.component";
import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";

describe("ReceiptsTableComponent", () => {
  let component: ReceiptsTableComponent;
  let fixture: ComponentFixture<ReceiptsTableComponent>;
  let store: Store;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [ReceiptsTableComponent],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    imports: [ApiModule,
        NgxsModule.forRoot([ReceiptTableState, AuthState, GroupState, UserState]),
        ReactiveFormsModule,
        MatSnackBarModule,
        MatTooltipModule,
        MatDialogModule,
        MatMenuModule,
        MatChipsModule,
        MonthStepperComponent,
        PipesModule],
    providers: [
        {
            provide: ActivatedRoute,
            useValue: {
                snapshot: {
                    data: {
                        categories: [],
                        tags: [],
                    },
                },
            },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
        provideZonelessChangeDetection(),
        provideRouter([]),
    ]
}).compileComponents();

    store = TestBed.inject(Store);
    fixture = TestBed.createComponent(ReceiptsTableComponent);
    component = fixture.componentInstance;
    Object.defineProperty(component, 'table', {
      value: () => ({
        selection: {},
        changed: of(undefined),
      }),
    });
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("gates each header action on its own group permission", () => {
    store.dispatch(
      new SetPermissions([], {
        5: [
          Permission.GroupReceiptsCreate,
          Permission.GroupReceiptsQuickScan,
          Permission.GroupEmailPoll,
        ],
      })
    );
    component.groupId = "5";

    (component as any).setCanEdit();

    expect(component.canCreate).toEqual(true);
    expect(component.canQuickScan).toEqual(true);
    expect(component.canPollEmail).toEqual(true);
    // None of those imply update, which gates Edit / Bulk Status Update.
    expect(component.canEdit).toEqual(false);
  });

  it("keeps canEdit tied to group.receipts.update only", () => {
    store.dispatch(
      new SetPermissions([], { 5: [Permission.GroupReceiptsUpdate] })
    );
    component.groupId = "5";

    (component as any).setCanEdit();

    expect(component.canEdit).toEqual(true);
    expect(component.canCreate).toEqual(false);
    expect(component.canQuickScan).toEqual(false);
    expect(component.canPollEmail).toEqual(false);
  });

  it("should map selected ids from selecton", () => {
    const selectedReceipts: Receipt[] = [
      {
        id: 1,
      } as Receipt,
      {
        id: 2,
      } as Receipt,
    ];
    Object.defineProperty(component, 'table', {
      value: () => ({
        selection: {
          changed: of({
            source: {
              selected: selectedReceipts,
            },
          }),
        },
      }),
    });
    component.ngAfterViewInit();

    expect(component.selectedReceiptIds()).toEqual([1, 2]);
  });
  describe("quick date filtering and filter chips", () => {
    let refetch: jest.SpyInstance;

    const setFilter = (filter: Record<string, unknown>) =>
      store.dispatch(new SetReceiptFilter(filter as any));

    beforeEach(() => {
      refetch = jest
        .spyOn(TestBed.inject(ReceiptFilterService), "getPagedReceiptsForGroups")
        .mockReturnValue(of({ data: [], totalCount: 0 } as any));
    });

    it("reads All time with no date filter", () => {
      expect(component.stepperMonth()).toBeNull();
      expect(component.stepperLabel()).toEqual("All time");
    });

    it("writes the picked month as a BETWEEN over the whole month", () => {
      component.monthSelected({ year: 2026, month: 8 });

      const date = (store.selectSnapshot(ReceiptTableState.filterData).filter as any).date;
      const [start, end] = date.value as Date[];
      expect(date.operation).toEqual(FilterOperation.Between);
      expect(start.getMonth()).toEqual(8);
      expect(start.getDate()).toEqual(1);
      expect(end.getDate()).toEqual(30);

      expect(component.stepperLabel()).toEqual("September 2026");
      expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
      expect(refetch).toHaveBeenCalled();
    });

    // The quick control IS the date filter, so it replaces whatever was there.
    it("overrides an existing date filter rather than sitting beside it", () => {
      setFilter({ date: { operation: FilterOperation.GreaterThan, value: new Date(2020, 0, 1) } });

      component.monthSelected({ year: 2026, month: 8 });

      const date = (store.selectSnapshot(ReceiptTableState.filterData).filter as any).date;
      expect(date.operation).toEqual(FilterOperation.Between);
    });

    it("clears the date filter for all time", () => {
      component.monthSelected({ year: 2026, month: 8 });
      component.allTimeSelected();

      expect(component.stepperLabel()).toEqual("All time");
      expect(component.filterChips()).toEqual([]);
    });

    // A date filter the stepper cannot express stays visible as a chip so it can
    // still be seen and cleared.
    it("reads Custom and shows a date chip for a filter it cannot express", () => {
      setFilter({ date: { operation: FilterOperation.WithinCurrentMonth, value: null } });

      expect(component.stepperMonth()).toBeNull();
      expect(component.stepperLabel()).toEqual("Custom");
      expect(component.filterChips()).toEqual([
        { key: "date", label: "Date within current month" },
      ]);
    });

    it("never shows a date chip for the month the stepper is already showing", () => {
      component.monthSelected({ year: 2026, month: 8 });

      expect(component.filterChips()).toEqual([]);
    });

    it("builds a chip per active field, resolving ids to names", () => {
      component.categories.set([{ id: 3, name: "Groceries" }] as any);
      setFilter({
        name: { operation: FilterOperation.Contains, value: "whole" },
        categories: { operation: FilterOperation.Contains, value: [3] },
        status: { operation: FilterOperation.Contains, value: [ReceiptStatus.Open] },
      });

      expect(component.filterChips()).toEqual([
        { key: "name", label: "Name contains whole" },
        { key: "categories", label: "Categories contains Groceries" },
        { key: "status", label: "Status contains Open" },
      ]);
    });

    it("clears exactly the field whose chip was dismissed", () => {
      setFilter({
        name: { operation: FilterOperation.Contains, value: "whole" },
        status: { operation: FilterOperation.Contains, value: [ReceiptStatus.Open] },
      });

      component.filterChipCleared("status");

      const filter = store.selectSnapshot(ReceiptTableState.filterData).filter as any;
      expect(filter.status).toEqual({ operation: null, value: [] });
      expect(filter.name).toEqual({ operation: FilterOperation.Contains, value: "whole" });
      expect(component.filterChips().map((chip) => chip.key)).toEqual(["name"]);
      expect(store.selectSnapshot(ReceiptTableState.page)).toEqual(1);
      expect(refetch).toHaveBeenCalled();
    });
  });
});
