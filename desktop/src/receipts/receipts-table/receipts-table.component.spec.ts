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
import { of, Subject, throwError } from "rxjs";
import { PipesModule } from "src/pipes/pipes.module";
import { ReceiptTableState } from "src/store/receipt-table.state";
import { MonthStepperComponent } from "../../shared-ui/month-stepper/month-stepper.component";
import { ApiModule, FilterOperation, Permission, Receipt, ReceiptStatus } from "../../open-api";
import { ReceiptFilterService } from "../../services/receipt-filter.service";
import { AuthState, GroupState, UserState } from "../../store";
import { SetPermissions } from "../../store/auth.state.actions";
import { SetQuickDateField, SetReceiptFilter } from "../../store/receipt-table.actions";
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
        { key: "date", label: "Receipt Date within current month" },
      ]);
    });

    // The chip row used to omit the field the stepper was naming. It is the only
    // place that says WHICH date column is filtered now that the target is
    // selectable, so it no longer makes exceptions.
    it("chips the month the stepper is showing", () => {
      component.monthSelected({ year: 2026, month: 8 });

      expect(component.filterChips().map((chip) => chip.key)).toEqual(["date"]);
    });

    describe("choosing which date field to filter on", () => {
      it("defaults to the receipt date", () => {
        expect(component.quickDateField()).toEqual("date");
        expect(component.quickDateFieldLabel()).toEqual("Receipt Date");
      });

      it("offers exactly the date fields, labelled as the dialog labels them", () => {
        expect(component.dateFilterFields.map((field) => [field.key, field.label])).toEqual([
          ["date", "Receipt Date"],
          ["resolvedDate", "Resolved Date"],
          ["createdAt", "Added At"],
        ]);
      });

      it("writes the picked month to the selected field, not to date", () => {
        component.quickDateFieldSelected("resolvedDate");
        component.monthSelected({ year: 2026, month: 8 });

        const filter = store.selectSnapshot(ReceiptTableState.filterData).filter as any;
        expect(filter.resolvedDate.operation).toEqual(FilterOperation.Between);
        expect(filter.date).toEqual({ operation: null, value: null });

        expect(component.quickDateFieldLabel()).toEqual("Resolved Date");
        expect(component.stepperLabel()).toEqual("September 2026");
      });

      it("reads the stepper label and Custom fallback off the selected field", () => {
        setFilter({
          date: { operation: FilterOperation.Between, value: [new Date(2026, 8, 1), new Date(2026, 8, 30)] },
          createdAt: { operation: FilterOperation.WithinCurrentMonth, value: null },
        });

        expect(component.stepperLabel()).toEqual("September 2026");

        component.quickDateFieldSelected("createdAt");

        expect(component.stepperMonth()).toBeNull();
        expect(component.stepperLabel()).toEqual("Custom");
      });

      // Switching the target changes no condition, only which one the stepper
      // describes — so nothing the user set elsewhere is destroyed, and the
      // abandoned condition stays visible and clearable as its own chip.
      it("leaves the previous field's condition applied, with its chip", () => {
        component.monthSelected({ year: 2026, month: 8 });
        refetch.mockClear();

        component.quickDateFieldSelected("resolvedDate");

        const filter = store.selectSnapshot(ReceiptTableState.filterData).filter as any;
        expect(filter.date.operation).toEqual(FilterOperation.Between);
        expect(component.stepperLabel()).toEqual("All time");
        expect(component.filterChips().map((chip) => chip.key)).toEqual(["date"]);
        expect(refetch).not.toHaveBeenCalled();
      });

      it("follows a field chosen outside the component", () => {
        store.dispatch(new SetQuickDateField("createdAt"));

        expect(component.quickDateFieldLabel()).toEqual("Added At");
      });
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

    // Each refresh used to be its own subscription, so the last RESPONSE won
    // rather than the last REQUEST — and the quick date arrows are one click
    // apart, which is what makes this reachable.
    it("lets a newer refresh supersede an in-flight one", () => {
      const superseded = new Subject<any>();
      const latest = new Subject<any>();
      refetch.mockReturnValueOnce(superseded).mockReturnValueOnce(latest);

      component.monthSelected({ year: 2026, month: 8 });
      component.monthSelected({ year: 2026, month: 9 });

      // The first request resolving late must not repaint the table.
      superseded.next({ data: [{ id: 1 }], totalCount: 1 });
      expect(component.totalCount()).toEqual(0);

      latest.next({ data: [{ id: 2 }], totalCount: 2 });
      expect(component.totalCount()).toEqual(2);
    });

    // switchMap completes the outer stream on an error unless the inner one
    // swallows it, which would silently kill every refresh after the first
    // failure.
    it("keeps refreshing after a failed request", () => {
      refetch.mockReturnValueOnce(throwError(() => new Error("boom")));

      component.monthSelected({ year: 2026, month: 8 });

      refetch.mockReturnValue(of({ data: [{ id: 3 }], totalCount: 3 }));
      component.monthSelected({ year: 2026, month: 9 });

      expect(component.totalCount()).toEqual(3);
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
