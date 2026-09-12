import { FormGroup } from "@angular/forms";
import { FilterOperation } from "../open-api/index";
import { buildFieldFormGroup, listenForBetweenOperation } from "./filter-form";

export function buildReceiptFilterForm(filter: any, thisContext: any): FormGroup {
  const formGroup = new FormGroup({
    date: buildFieldFormGroup(
      filter?.date?.value,
      filter?.date?.operation,
      thisContext,
      filter?.date?.operation === FilterOperation.Between
    ),
    amount: buildFieldFormGroup(
      filter?.amount?.value,
      filter?.amount?.operation,
      thisContext,
      filter?.amount?.operation === FilterOperation.Between
    ),
    name: buildFieldFormGroup(
      filter?.name?.value,
      filter?.name?.operation,
      thisContext
    ),
    paidBy: buildFieldFormGroup(
      filter?.paidBy?.value ?? [],
      filter?.paidBy?.operation,
      thisContext,
      true
    ),
    categories: buildFieldFormGroup(
      filter?.categories?.value ?? [],
      filter?.categories?.operation,
      thisContext,
      true
    ),
    tags: buildFieldFormGroup(
      filter?.tags?.value ?? [],
      filter?.tags?.operation,
      thisContext,
      true
    ),
    status: buildFieldFormGroup(
      filter?.status?.value ?? [],
      filter?.status?.operation,
      thisContext,
      true
    ),
    group: buildFieldFormGroup(
      filter?.group?.value ?? [],
      filter?.group?.operation,
      thisContext,
      true
    ),
    resolvedDate: buildFieldFormGroup(
      filter?.resolvedDate?.value,
      filter?.resolvedDate?.operation,
      thisContext,
      filter?.resolvedDate?.operation === FilterOperation.Between
    ),
    createdAt: buildFieldFormGroup(
      filter?.createdAt?.value,
      filter?.createdAt?.operation,
      thisContext,
      filter?.createdAt?.operation === FilterOperation.Between
    ),
  });

  listenForBetweenOperation(formGroup, "amount", thisContext);
  listenForBetweenOperation(formGroup, "date", thisContext);
  listenForBetweenOperation(formGroup, "resolvedDate", thisContext);
  listenForBetweenOperation(formGroup, "createdAt", thisContext);


  return formGroup;
}
