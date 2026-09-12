import { FormGroup } from "@angular/forms";
import { FilterOperation } from "../open-api/index";
import { buildFieldFormGroup, listenForBetweenOperation } from "./filter-form";

/**
 * The system task filter's reactive form — the same `{ operation, value }`
 * shape as `buildReceiptFilterForm`, built from the same shared helpers, so the
 * two dialogs behave identically.
 *
 * `thisContext` must be an `@UntilDestroy()` component: the helpers tie their
 * subscriptions to its lifetime.
 */
export function buildSystemTaskFilterForm(filter: any, thisContext: any): FormGroup {
  const formGroup = new FormGroup({
    type: buildFieldFormGroup(
      filter?.type?.value ?? [],
      filter?.type?.operation,
      thisContext,
      true
    ),
    ranBy: buildFieldFormGroup(
      filter?.ranBy?.value ?? [],
      filter?.ranBy?.operation,
      thisContext,
      true
    ),
    startedAt: buildFieldFormGroup(
      filter?.startedAt?.value,
      filter?.startedAt?.operation,
      thisContext,
      filter?.startedAt?.operation === FilterOperation.Between
    ),
    endedAt: buildFieldFormGroup(
      filter?.endedAt?.value,
      filter?.endedAt?.operation,
      thisContext,
      filter?.endedAt?.operation === FilterOperation.Between
    ),
  });

  listenForBetweenOperation(formGroup, "startedAt", thisContext);
  listenForBetweenOperation(formGroup, "endedAt", thisContext);

  return formGroup;
}
