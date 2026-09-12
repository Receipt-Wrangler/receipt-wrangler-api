import { FilterOperation } from "../open-api";

export const listOperationOptions = Object.values(FilterOperation).filter(
  (k) => k === "CONTAINS" && k !== FilterOperation.WithinCurrentMonth && !!k
);

export const dateOperationOptions = Object.values(FilterOperation).filter(
  (k) => !k.includes("CONTAINS") && !!k
);

export const numberOperationOptions = Object.values(FilterOperation).filter(
  (k) => !k.includes("CONTAINS") && k !== FilterOperation.WithinCurrentMonth && !!k
);

export const textOperationOptions = Object.values(FilterOperation).filter(
  (k) => !k.includes("THAN") && k !== FilterOperation.WithinCurrentMonth && !k.includes("BETWEEN") && !!k
);

export const usersOperationOptions = Object.values(FilterOperation).filter(
  (k) => k === "CONTAINS" && k !== FilterOperation.WithinCurrentMonth && !!k
);

/**
 * Human-readable label per operation. Shared by the filter dialog's Operation
 * select (via `OperationsPipe`) and the filter chips, so both read a condition
 * the same way.
 */
export const FILTER_OPERATION_DISPLAY_VALUES: { [key: string]: string } = {
  [FilterOperation.Contains]: "Contains",
  [FilterOperation.Equals]: "Equals",
  [FilterOperation.GreaterThan]: "Greater than",
  [FilterOperation.LessThan]: "Less than",
  [FilterOperation.Between]: "Between",
  [FilterOperation.WithinCurrentMonth]: "Within current month",
};
