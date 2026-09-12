import { AbstractControl, FormArray, FormBuilder, FormGroup, Validators } from "@angular/forms";
import { untilDestroyed } from "@ngneat/until-destroy";
import { distinctUntilChanged, pairwise, startWith, tap } from "rxjs";
import { FilterField, FilterFieldType } from "../constants/filter-fields.constant";
import { OperationsPipe } from "../shared-ui/receipt-filter/operations.pipe";
import { FilterOperation } from "../open-api/index";

/**
 * The machinery every `{ operation, value }` filter form is built from —
 * shared by `buildReceiptFilterForm` and `buildSystemTaskFilterForm` so the two
 * cannot drift on how a BETWEEN range or an auto-selected operation behaves.
 *
 * Every entry point takes a `thisContext`: `untilDestroyed` needs the calling
 * component, which must therefore carry `@UntilDestroy()`.
 */

/**
 * One `{ operation, value }` group. `isArray` makes the value a `FormArray`,
 * which the multi-select autocompletes require — the shared `app-autocomlete`
 * `push()`es the picked option onto the control.
 */
export function buildFieldFormGroup(
  value: string | string[] | number | any,
  operation: string | undefined,
  thisContext: any,
  isArray?: boolean,
): FormGroup {
  const formBuilder = new FormBuilder();
  let valueControl: AbstractControl;
  const operationControl = formBuilder.control(operation);

  if (operation === FilterOperation.Between) {
    valueControl = formBuilder.array([value?.[0], value?.[1]]);
  } else if (isArray) {
    valueControl = formBuilder.array(value);
  } else {
    valueControl = formBuilder.control(value);
  }

  valueControl.valueChanges.pipe(
    untilDestroyed(thisContext),
    startWith(value),
    tap((value) => {
      if (!!value && value?.length > 0) {
        operationControl.addValidators(Validators.required);
      } else {
        operationControl.removeValidators(Validators.required);
      }

      operationControl.updateValueAndValidity();
    })).subscribe();

  return formBuilder.group({
    operation: operationControl,
    value: valueControl
  });
}

/**
 * Swaps a field's value control between a scalar and a two-slot range as the
 * operation flips in and out of BETWEEN.
 */
export function listenForBetweenOperation(form: FormGroup, key: string, thisContext: any): void {
  updateValidatorsOnOperationChange(true, form, key, null, form.get(key)?.get("operation")?.value);

  form
    .get(key)
    ?.get("operation")
    ?.valueChanges
    .pipe(
      startWith(form.get(key)?.get("operation")?.value),
      distinctUntilChanged(),
      pairwise(),
      untilDestroyed(thisContext),
      tap(([prev, curr]: [FilterOperation | null, FilterOperation | null]) => {
        updateValidatorsOnOperationChange(false, form, key, prev, curr);
      }),
    ).subscribe();
}

function updateValidatorsOnOperationChange(
  firstRun: boolean,
  form: FormGroup,
  key: string,
  prev: FilterOperation | null,
  curr: FilterOperation | null
)
  : void {
  const formBuilder = new FormBuilder();
  if (curr === FilterOperation.Between) {
    if (!firstRun) {
      (form.get(key) as FormGroup).removeControl("value");
      (form.get(key) as FormGroup).addControl("value", formBuilder.array([null, null]));
    }

    (form.get(key) as FormGroup).get("value.0")?.addValidators(Validators.required);
    (form.get(key) as FormGroup).get("value.1")?.addValidators(Validators.required);

    (form.get(key) as FormGroup).get("value")?.addValidators(betweenValidator);
  } else if (prev === FilterOperation.Between && curr !== FilterOperation.Between) {
    if (!firstRun) {
      (form.get(key) as FormGroup).removeControl("value");
      (form.get(key) as FormGroup).addControl("value", formBuilder.control(null));
    }
  }
}

function betweenValidator(control: AbstractControl): { [key: string]: any } | null {
  const formArray = control as FormArray;

  if (formArray.value[0] > formArray.value[1] && formArray.value[1] !== null) {
    formArray.at(0).setErrors({ invalidValue: true });
  }

  if (formArray.value[0] < formArray.value[1]) {
    formArray.at(0).setErrors(null);
  }

  return null;
}

/**
 * Picks the first operation for a field's type as soon as it gains a value, and
 * clears the operation again when the value empties — so a user who types a
 * name never has to also pick "Contains".
 */
export function setupAutoOperationSelection(
  parentForm: FormGroup,
  basePath: string,
  fields: readonly FilterField[],
): void {
  const operationsPipe = new OperationsPipe();

  fields.forEach(({ key, type }) => {
    const valueControl = parentForm.get(`${basePath}${key}.value`);
    const operationControl = parentForm.get(`${basePath}${key}.operation`);

    if (!valueControl || !operationControl) {
      return;
    }

    valueControl.valueChanges.subscribe((value) => {
      if (hasFieldValue(value, type)) {
        // Set first operation if none is selected
        if (!operationControl.value) {
          const operations = operationsPipe.transform(type, false);
          if (operations.length > 0) {
            operationControl.setValue(operations[0]);
          }
        }
      } else {
        // Clear operation if field is empty
        operationControl.setValue(null);
      }
    });
  });
}

function hasFieldValue(value: any, type: FilterFieldType): boolean {
  if (value === null || value === undefined) {
    return false;
  }

  if (type === "list" || type === "users") {
    return Array.isArray(value) && value.length > 0;
  }

  if (typeof value === "string") {
    return value.trim().length > 0;
  }

  return value !== "";
}
