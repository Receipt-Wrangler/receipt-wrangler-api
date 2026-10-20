import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { FormArray, FormControl, FormGroup, ReactiveFormsModule, } from "@angular/forms";
import { MatAutocompleteModule, MatAutocompleteSelectedEvent, } from "@angular/material/autocomplete";
import { NgxsModule, Store } from "@ngxs/store";
import { BaseInputComponent } from "../../base-input";
import { AuthState } from "../../store/auth.state";
import { AutocomleteComponent } from "./autocomlete.component";
import { OptionDisplayPipe } from "./option-display.pipe";

describe("AutocomleteComponent", () => {
  let component: AutocomleteComponent;
  let fixture: ComponentFixture<AutocomleteComponent>;
  jest.useFakeTimers();

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      // OptionDisplayPipe renders the chips. Only the panel-behavior cases
      // below reach it - they are the only ones that let change detection
      // run over the multiple-mode branch of the template.
      declarations: [AutocomleteComponent, BaseInputComponent, OptionDisplayPipe],
      imports: [
        NgxsModule.forRoot([AuthState]),
        MatAutocompleteModule,
        ReactiveFormsModule,
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(AutocomleteComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should filter the options when multiple is false", () => {
    fixture.componentRef.setInput('options', [
      { id: 1, name: "Option 1" },
      { id: 2, name: "Option 2" },
      { id: 3, name: "Option 3" },
    ]);
    fixture.componentRef.setInput('optionFilterKey', "name");

    component.multiple = false;
    const result = component._filter("Option 1");
    expect(result).toEqual([{ id: 1, name: "Option 1" }]);
  });

  // it('should filter the options when multiple is true and values are selected', () => {
  //   const service = TestBed.inject(FormBuilder);
  //   component.options = [
  //     { id: 1, name: 'Option 1' },
  //     { id: 2, name: 'Option 2' },
  //     { id: 3, name: 'Option 3' },
  //   ];
  //   component.optionFilterKey = 'name';

  //   component.multiple = true;
  //   component.inputFormControl = new FormArray([
  //     new FormGroup({
  //       id: new FormControl(2),
  //       name: new FormControl('Option 2'),
  //     }),
  //   ]) as any;

  //   const result = component._filter('Option');
  //   console.log(result);
  //   expect(result).toEqual([
  //     { id: 1, name: 'Option 1' },
  //     { id: 3, name: 'Option 3' },
  //   ]);
  // });

  it("should return an empty array when no options match the filter", () => {
    fixture.componentRef.setInput('options', [
      { id: 1, name: "Option 1" },
      { id: 2, name: "Option 2" },
      { id: 3, name: "Option 3" },
    ]);
    fixture.componentRef.setInput('optionFilterKey', "name");

    component.multiple = false;
    const result = component._filter("Non-existing option");
    expect(result).toEqual([]);
  });

  it("re-seeds the single-select display from inputFormControl on syncSingleDisplay", () => {
    component.multiple = false;
    component.inputFormControl.setValue(42);
    component.filterFormControl.setValue("");

    component.syncSingleDisplay();

    expect(component.filterFormControl.value).toEqual(42);
  });

  it("leaves the display untouched on syncSingleDisplay in multiple mode", () => {
    component.multiple = true;
    component.inputFormControl = new FormArray([new FormControl(1)]) as any;
    component.filterFormControl.setValue("stale");

    component.syncSingleDisplay();

    expect(component.filterFormControl.value).toEqual("stale");
  });

  it("should set the selected option value to the inputFormControl in single mode", () => {
    component.multiple = false;
    component.optionValueKey = "value";

    const event = {
      option: {
        id: "selected-option",
        value: "value",
      },
    } as MatAutocompleteSelectedEvent;

    component.optionSelected(event);

    expect(component.inputFormControl.value).toEqual("value");
  });

  it("should set the selected option value to the inputFormControl in single mode with full object as value", () => {
    component.multiple = false;

    const event = {
      option: {
        id: "selected-option",
        value: { id: "id1", name: "Groceries" },
      },
    } as MatAutocompleteSelectedEvent;

    component.optionSelected(event);

    expect(component.inputFormControl.value).toEqual({
      id: "id1",
      name: "Groceries",
    });
  });

  it("should add the selected option value to the inputFormControl as a FormControl instance in multiple mode", () => {
    component.multiple = true;
    component.optionValueKey = "value";
    component.inputFormControl = new FormArray([
      new FormControl("value 1"),
    ]) as any;

    const event = {
      option: {
        id: "selected-option",
        value: "value 2",
      },
    } as MatAutocompleteSelectedEvent;

    component.optionSelected(event);

    expect((component.inputFormControl as any as FormArray).value).toEqual([
      "value 1",
      "value 2",
    ]);
  });

  it("should add the selected option value to the inputFormControl as a FormControl instance in multiple mode with full object", () => {
    component.multiple = true;
    component.optionValueKey = "value";
    component.inputFormControl = new FormArray([
      new FormGroup({
        id: new FormControl("id0"),
        name: new FormControl("Utilities"),
      }),
    ]) as any;

    const event = {
      option: {
        id: "selected-option",
        value: {
          id: "id1",
          name: "Groceries",
        },
      },
    } as MatAutocompleteSelectedEvent;

    component.optionSelected(event);

    expect((component.inputFormControl as any as FormArray).value).toEqual([
      {
        id: "id0",
        name: "Utilities",
      },
      {
        id: "id1",
        name: "Groceries",
      },
    ]);
  });

  it("should add a custom option value to the inputFormControl as a FormControl instance in multiple mode with no option value key", () => {
    component.creatableOptionId = "create-option";
    fixture.componentRef.setInput('defaultCreatableObject', { name: "Custom Option" });
    fixture.componentRef.setInput('creatableValueKey', "name");
    component.multiple = true;
    component.inputFormControl = new FormArray([
      new FormControl("value 1"),
    ]) as any;

    const event = {
      option: {
        id: component.creatableOptionId,
        value: "Custom Option",
      },
    } as MatAutocompleteSelectedEvent;

    component.filterFormControl.setValue("new value");

    component.optionSelected(event);

    expect((component.inputFormControl as any as FormArray).value).toEqual([
      "value 1",
      {
        name: "new value",
      },
    ]);
  });

  it("should add a custom option value to the inputFormControl as a FormControl instance in multiple mode with option value key", () => {
    component.creatableOptionId = "create-option";
    component.optionValueKey = "name";
    component.multiple = true;
    component.inputFormControl = new FormArray([
      new FormControl("value 1"),
    ]) as any;

    const event = {
      option: {
        id: component.creatableOptionId,
        value: "Custom Option",
      },
    } as MatAutocompleteSelectedEvent;

    component.filterFormControl.setValue("new value");

    component.optionSelected(event);

    expect((component.inputFormControl as any as FormArray).value).toEqual([
      "value 1",
      "new value",
    ]);
  });

  // The panel handling runs in a setTimeout, so these are the only cases that
  // flush it - every other optionSelected case stops at the FormArray push.
  //
  // Both required viewChildren are stubbed rather than rendered: #inputMultiple
  // sits behind @if (multiple) and the real MatAutocompleteTrigger drives a
  // CDK overlay, neither of which this branch is about.
  describe("panel behavior after selecting in multiple mode", () => {
    function selectAnOption(): {
      openPanel: jest.Mock;
      closePanel: jest.Mock;
      blur: jest.Mock;
    } {
      const openPanel = jest.fn();
      const closePanel = jest.fn();
      const blur = jest.fn();

      component.multiple = true;
      component.inputFormControl = new FormArray([]) as any;
      (component as any).matAutocompleteTrigger = () => ({ openPanel, closePanel });
      (component as any).inputMultiple = () => ({ nativeElement: { blur } });

      // jest.useFakeTimers() is describe-scoped and none of the other
      // optionSelected cases flush, so the queue is carrying their timers -
      // bound to components TestBed has since destroyed. Drop them, or
      // runAllTimers() fires those instead and throws NG0205.
      jest.clearAllTimers();

      component.optionSelected({
        option: { id: "selected-option", value: "value 1" },
      } as MatAutocompleteSelectedEvent);
      jest.runAllTimers();

      return { openPanel, closePanel, blur };
    }

    it("should re-open the panel when the preference is off", () => {
      const { openPanel, closePanel, blur } = selectAnOption();

      expect(openPanel).toHaveBeenCalled();
      expect(closePanel).not.toHaveBeenCalled();
      expect(blur).not.toHaveBeenCalled();
      expect(component.filterFormControl.value).toEqual("");
    });

    it("should close the panel and blur the field when the preference is on", () => {
      const store = TestBed.inject(Store);
      store.reset({
        ...store.snapshot(),
        auth: {
          ...store.snapshot().auth,
          userPreferences: { closeChipSelectOnSelect: true },
        },
      });

      const { openPanel, closePanel, blur } = selectAnOption();

      // Material closes the panel on selection by itself, so the blur is what
      // actually distinguishes this path - without it the trigger re-opens on
      // the focus the input still holds.
      expect(closePanel).toHaveBeenCalled();
      expect(blur).toHaveBeenCalled();
      expect(openPanel).not.toHaveBeenCalled();
      expect(component.filterFormControl.value).toEqual("");
    });
  });
});
