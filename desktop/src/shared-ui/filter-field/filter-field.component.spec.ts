import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { FormControl, FormGroup, ReactiveFormsModule } from "@angular/forms";
import { PipesModule } from "src/pipes/pipes.module";
import { FilterOperation } from "../../open-api";
import { OperationsPipe } from "../receipt-filter/operations.pipe";
import { FilterFieldComponent } from "./filter-field.component";

describe("FilterFieldComponent", () => {
  let fixture: ComponentFixture<FilterFieldComponent>;
  let component: FilterFieldComponent;

  const buildForm = (operation: FilterOperation | null, value: any = null): FormGroup =>
    new FormGroup({
      field: new FormGroup({
        operation: new FormControl(operation),
        value: Array.isArray(value)
          ? new FormGroup({
            0: new FormControl(value[0]),
            1: new FormControl(value[1]),
          })
          : new FormControl(value),
      }),
    });

  const render = (type: string, operation: FilterOperation | null, value: any = null) => {
    fixture = TestBed.createComponent(FilterFieldComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput("parentForm", buildForm(operation, value));
    fixture.componentRef.setInput("fieldName", "field");
    fixture.componentRef.setInput("type", type);
    fixture.componentRef.setInput("label", "Started At");
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [FilterFieldComponent, OperationsPipe],
      imports: [ReactiveFormsModule, PipesModule],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();
  });

  it("should create", () => {
    render("text", null);
    expect(component).toBeTruthy();
  });

  it("renders the control matching the field type, plus an Operation select", () => {
    const cases: { type: string; selector: string }[] = [
      { type: "text", selector: "app-input" },
      { type: "number", selector: "app-input" },
      { type: "date", selector: "app-datepicker" },
      { type: "list", selector: "app-autocomlete" },
      { type: "users", selector: "app-user-autocomplete" },
    ];

    cases.forEach(({ type, selector }) => {
      const element = render(type, null);

      expect(element.querySelectorAll(selector).length).toBe(1);
      expect(element.querySelectorAll("app-select").length).toBe(1);
    });
  });

  it("renders a two-slot range for BETWEEN", () => {
    const element = render("date", FilterOperation.Between, [null, null]);

    expect(element.querySelectorAll("app-datepicker").length).toBe(2);
    expect(element.textContent).toContain("and");
  });

  // WITHIN_CURRENT_MONTH carries no value, so the two datepickers only show the
  // range it implies and must never be editable.
  it("renders the implied range read-only for WITHIN_CURRENT_MONTH", () => {
    const element = render("date", FilterOperation.WithinCurrentMonth);

    expect(element.querySelectorAll("app-datepicker").length).toBe(2);
    expect(component.startOfMonthFormControl.disabled).toBe(true);
    expect(component.endOfTodayFormControl.disabled).toBe(true);
  });
});
