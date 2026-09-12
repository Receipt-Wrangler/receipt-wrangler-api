import { ReceiptPagedRequestFilter } from "../open-api";
import { FilterField, FilterFieldType } from "./filter-fields.constant";

export type ReceiptFilterFieldKey = keyof ReceiptPagedRequestFilter;

export type ReceiptFilterFieldType = FilterFieldType;

export type ReceiptFilterField = FilterField<ReceiptFilterFieldKey>;

/**
 * The ten filterable receipt fields, in the order the filter dialog renders
 * them. This is the single definition of a field's label and type — the dialog's
 * auto-operation wiring and the filter chips both read it, so a chip can never
 * disagree with the row that produced it.
 */
export const RECEIPT_FILTER_FIELDS: readonly ReceiptFilterField[] = [
  { key: "date", label: "Date", type: "date" },
  { key: "name", label: "Name", type: "text" },
  { key: "paidBy", label: "Paid by", type: "users" },
  { key: "group", label: "Group", type: "list" },
  { key: "amount", label: "Amount", type: "number" },
  { key: "categories", label: "Categories", type: "list" },
  { key: "tags", label: "Tags", type: "list" },
  { key: "status", label: "Status", type: "list" },
  { key: "resolvedDate", label: "Resolved Date", type: "date" },
  { key: "createdAt", label: "Added At", type: "date" },
];
