import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { Component, CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialogModule, MatDialogRef, } from "@angular/material/dialog";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { UntilDestroy } from "@ngneat/until-destroy";
import { Store } from "@ngxs/store";
import { of } from "rxjs";
import { PipesModule } from "src/pipes/pipes.module";
import { SetReceiptFilter } from "src/store/receipt-table.actions";
import { defaultReceiptFilter, } from "src/store/receipt-table.state";
import { InputModule } from "../../input";
import { Category, FilterOperation, ReceiptStatus, Tag } from "../../open-api";
import { StoreModule } from "../../store/store.module";
import { applyFormCommand } from "../../utils/index";
import { buildReceiptFilterForm } from "../../utils/receipt-filter";
import { OperationsPipe } from "./operations.pipe";
import { ReceiptFilterComponent } from "./receipt-filter.component";

@UntilDestroy()
@Component({
  selector: "app-noop",
  template: "",
  standalone: false
})
class NoopComponent {}

describe("ReceiptFilterComponent", () => {
  let component: ReceiptFilterComponent;
  let fixture: ComponentFixture<ReceiptFilterComponent>;
  let store: Store;

  const filledFilter = {
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
    group: {
      operation: FilterOperation.Contains,
      value: [1],
    },
    resolvedDate: {
      operation: FilterOperation.GreaterThan,
      value: "2023-01-06",
    },
    createdAt: {
      operation: FilterOperation.GreaterThan,
      value: "2023-01-06",
    },
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ReceiptFilterComponent, OperationsPipe],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [PipesModule,
        InputModule,
        MatDialogModule,
        StoreModule,
        NoopAnimationsModule,
        PipesModule,
        ReactiveFormsModule],
      providers: [
        {
          provide: MatDialogRef,
          useValue: {
            close: (value: any) => { },
          },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ]
    });

    store = TestBed.inject(Store);
    fixture = TestBed.createComponent(ReceiptFilterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("uses the category/tag options provided as inputs", () => {
    const categories = [{ id: 2, name: "Groceries" }] as Category[];
    const tags = [{ id: 3, name: "Reimbursable" }] as Tag[];

    component.categories = categories;
    component.tags = tags;
    component.ngOnInit();

    expect(component.categories).toBe(categories);
    expect(component.tags).toBe(tags);
  });

  it("should init form with no default initial data", () => {
    const noopComponent = TestBed.createComponent(NoopComponent).componentInstance;

    component.parentForm = buildReceiptFilterForm({}, noopComponent);
    component.ngOnInit();

    expect(component.parentForm.value).toEqual(defaultReceiptFilter);
  });

  it("should init form with initial data", () => {
    store.reset({
      receiptTable: {
        filter: filledFilter,
      },
    });

    const noopComponent = TestBed.createComponent(NoopComponent).componentInstance;

    component.parentForm = buildReceiptFilterForm(filledFilter, noopComponent);
    component.ngOnInit();

    expect(component.parentForm.value).toEqual(filledFilter);
  });

  it("should reset form", () => {
    store.reset({
      receiptTable: {
        filter: filledFilter,
      },
    });

    component.formCommand.subscribe((formCommand) => {
      applyFormCommand(component.parentForm, formCommand);
    });

    const noopComponent = TestBed.createComponent(NoopComponent).componentInstance;

    component.parentForm = buildReceiptFilterForm(filledFilter, noopComponent);
    component.ngOnInit();

    expect(component.parentForm.value).toEqual(filledFilter);

    component.resetFilter();
    expect(component.parentForm.value).toEqual(defaultReceiptFilter);
  });

  it("should set form in state and close dialog", () => {
    const dialogRefSpy = jest.spyOn(
      TestBed.inject(MatDialogRef<ReceiptFilterComponent>),
      "close"
    );
    const storeRefSpy = jest.spyOn(store, "dispatch").mockReturnValue(of(undefined));

    component.submitButtonClicked();

    expect(storeRefSpy).toHaveBeenCalledWith(
      new SetReceiptFilter(component.parentForm.value)
    );
    expect(dialogRefSpy).toHaveBeenCalledWith(true);
  });

  it("renders the group field only when showGroupFilter is true", () => {
    const noopComponent = TestBed.createComponent(NoopComponent).componentInstance;

    const countListRows = (showGroupFilter: boolean): number => {
      const localFixture = TestBed.createComponent(ReceiptFilterComponent);
      localFixture.componentInstance.parentForm = buildReceiptFilterForm({}, noopComponent);
      localFixture.componentInstance.showGroupFilter = showGroupFilter;
      localFixture.detectChanges();
      // The rows are the shared app-filter-field; CUSTOM_ELEMENTS_SCHEMA keeps
      // it unrendered here, so the inner controls it projects are not in the DOM.
      return localFixture.nativeElement.querySelectorAll("app-filter-field[type='list']").length;
    };

    // The group field is the only additional list row gated on the flag.
    expect(countListRows(true)).toBe(countListRows(false) + 1);
  });

  it("defaults showGroupFilter to false", () => {
    expect(component.showGroupFilter).toBe(false);
  });

  it("should close dialog on cancel", () => {
    const dialogRefSpy = jest.spyOn(
      TestBed.inject(MatDialogRef<ReceiptFilterComponent>),
      "close"
    );
    component.cancelButtonClicked();

    expect(dialogRefSpy).toHaveBeenCalledWith(false);
  });
});
