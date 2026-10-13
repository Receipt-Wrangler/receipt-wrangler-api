import { FormGroup } from "@angular/forms";
import { FormMode } from "src/enums/form-mode.enum";
import { FormConfig } from "src/interfaces";
import { FormCommand } from "../form/index";

/**
 * Renders a server-side validator 400 on the form that produced it.
 *
 * The API writes errors in two shapes. `WriteErrorResponse` /
 * `WriteCustomErrorResponse` emit `{ errorMsg }`, which the HTTP interceptor
 * toasts at any status. `WriteValidatorErrorResponse` marshals a bare
 * `{ <control>: <message> }` map instead, and the interceptor has no branch for
 * it — so a 400 of that shape is completely silent unless a component reports
 * it. This is that report.
 *
 * `errorMsg` therefore means "the interceptor already owns this one" and is
 * skipped. That rule is what keeps the two mechanisms from both firing:
 * `MatSnackBar.open()` dismisses whatever is showing, so a double report costs
 * the user the message rather than merely repeating it.
 *
 * A key that names a control becomes an error on it, rendered verbatim by
 * `BaseInputComponent`'s `typeof value === "string"` branch — the same idiom
 * `durationValueValidator` uses, and the reason the key is deliberately absent
 * from that component's `errorMessages` map. A key that names nothing — a field
 * this form does not carry — goes to `onUnmatched` rather than being dropped,
 * since a silent 400 is the bug this exists to fix.
 */
export function applyApiErrors(
  form: FormGroup,
  body: unknown,
  onUnmatched?: (message: string) => void
): void {
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    return;
  }

  const fieldErrors = body as Record<string, unknown>;

  if (typeof fieldErrors["errorMsg"] === "string") {
    return;
  }

  const unmatched: string[] = [];

  Object.keys(fieldErrors).forEach((key) => {
    const message = fieldErrors[key];

    if (typeof message !== "string" || message.length === 0) {
      return;
    }

    const control = form.get(key);

    if (!control) {
      unmatched.push(message);
      return;
    }

    // Replace rather than merge: a client-side error still on the control
    // describes a value the server has now judged for itself.
    control.setErrors({ server: message });
    // Material's default ErrorStateMatcher shows a mat-error only for a control
    // that is invalid AND touched (or whose form was submitted), so without
    // this a control the user never focused stays silently valid-looking.
    control.markAsTouched();
  });

  if (unmatched.length > 0 && onUnmatched) {
    onUnmatched(unmatched.join(" "));
  }
}

export function setEntityHeaderText(
  entity: any,
  nameKey: string,
  formConfig: FormConfig,
  entityType?: string
): string {
  if (entity && entity[nameKey] && formConfig) {
    let entityName = entity[nameKey];
    let headerText = "";

    if (formConfig.mode === FormMode.view) {
      headerText = `View ${entityName}`;
    }

    if (formConfig.mode === FormMode.edit) {
      headerText = `Edit ${entityName}`;
    }

    if (formConfig.mode === FormMode.add) {
      headerText = `Create ${entityName}`;
    }

    if (entityType) {
      headerText += ` ${entityType}`;
    }

    return headerText;
  }

  return "";
}

export function applyFormCommand(form: FormGroup, formCommand: FormCommand) {
  if (formCommand.path) {
    (form.get(formCommand.path) as any)[formCommand.command](
      formCommand.payload
    );
  } else {
    (form as any)[formCommand.command](formCommand.payload);
  }
}
