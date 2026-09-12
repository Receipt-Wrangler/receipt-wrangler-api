import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { Component, CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialogModule, MatDialogRef } from "@angular/material/dialog";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { UntilDestroy } from "@ngneat/until-destroy";
import { Store } from "@ngxs/store";
import { of } from "rxjs";
import { PipesModule } from "src/pipes/pipes.module";
import { SYSTEM_RAN_BY_OPTION_ID } from "src/constants";
import { buildDefaultSystemTaskFilter } from "src/store/system-task-table.state";
import { SetSystemTaskFilter } from "src/store/system-task-table.state.actions";
import { InputModule } from "../../input";
import { FilterOperation } from "../../open-api";
import { StoreModule } from "../../store/store.module";
import { applyFormCommand } from "../../utils/index";
import { buildSystemTaskFilterForm } from "../../utils/system-task-filter";
import { SystemTaskFilterComponent } from "./system-task-filter.component";

@UntilDestroy()
@Component({
  selector: "app-noop",
  template: "",
  standalone: false
})
class NoopComponent {}

describe("SystemTaskFilterComponent", () => {
  let component: SystemTaskFilterComponent;
  let fixture: ComponentFixture<SystemTaskFilterComponent>;
  let store: Store;
  let noopComponent: NoopComponent;

  const filledFilter = {
    type: {
      operation: FilterOperation.Contains,
      value: ["QUICK_SCAN"],
    },
    ranBy: {
      operation: FilterOperation.Contains,
      value: [SYSTEM_RAN_BY_OPTION_ID, 4],
    },
    startedAt: {
      operation: FilterOperation.Equals,
      value: "2026-03-11T00:00:00Z",
    },
    endedAt: {
      operation: FilterOperation.LessThan,
      value: "2026-03-12T00:00:00Z",
    },
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [SystemTaskFilterComponent, NoopComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [
        InputModule,
        MatDialogModule,
        NoopAnimationsModule,
        PipesModule,
        ReactiveFormsModule,
        StoreModule,
      ],
      providers: [
        { provide: MatDialogRef, useValue: { close: () => {} } },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    store = TestBed.inject(Store);
    noopComponent = TestBed.createComponent(NoopComponent).componentInstance;

    fixture = TestBed.createComponent(SystemTaskFilterComponent);
    component = fixture.componentInstance;
    component.parentForm = buildSystemTaskFilterForm({}, noopComponent);
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("builds an empty form matching the default filter", () => {
    expect(component.parentForm.value).toEqual(buildDefaultSystemTaskFilter());
  });

  it("round-trips a seeded filter", () => {
    const localFixture = TestBed.createComponent(SystemTaskFilterComponent);
    localFixture.componentInstance.parentForm = buildSystemTaskFilterForm(filledFilter, noopComponent);
    localFixture.detectChanges();

    expect(localFixture.componentInstance.parentForm.value).toEqual(filledFilter);
  });

  it("pins System above the real users in the Ran By options", () => {
    expect(component.ranByOptions()[0]).toEqual({
      id: SYSTEM_RAN_BY_OPTION_ID,
      displayName: "System",
    });
  });

  it("renders one row per filter field", () => {
    expect(fixture.nativeElement.querySelectorAll("app-filter-field").length).toBe(4);
  });

  // reset() leaves a FormArray populated, so the two list fields need an
  // explicit clear command on top of it.
  it("resets every field, clearing the list ones", () => {
    component.parentForm = buildSystemTaskFilterForm(filledFilter, noopComponent);
    component.formCommand.subscribe((formCommand) =>
      applyFormCommand(component.parentForm, formCommand)
    );

    component.resetFilter();

    expect(component.parentForm.value).toEqual(buildDefaultSystemTaskFilter());
  });

  it("dispatches the filter and closes with true on submit", () => {
    const dispatchSpy = jest.spyOn(store, "dispatch").mockReturnValue(of({}));
    const closeSpy = jest.spyOn(TestBed.inject(MatDialogRef<SystemTaskFilterComponent>), "close");

    component.parentForm = buildSystemTaskFilterForm(filledFilter, noopComponent);
    component.submitButtonClicked();

    expect(dispatchSpy).toHaveBeenCalledWith(new SetSystemTaskFilter(filledFilter as any));
    expect(closeSpy).toHaveBeenCalledWith(true);
  });

  it("closes with false on cancel", () => {
    const closeSpy = jest.spyOn(TestBed.inject(MatDialogRef<SystemTaskFilterComponent>), "close");

    component.cancelButtonClicked();

    expect(closeSpy).toHaveBeenCalledWith(false);
  });
});
