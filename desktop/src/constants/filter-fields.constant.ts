/**
 * The operation-option bucket a filter field belongs to. Matches the `type`
 * argument `OperationsPipe` switches on, so a field's operations and its label
 * come from the same place.
 */
export type FilterFieldType = "date" | "text" | "number" | "list" | "users";

/**
 * One filterable field: the key it occupies on its filter object, the label
 * every surface names it by, and the bucket its operations come from.
 *
 * Shared by the receipt filter and the system task filter — the dialog rows,
 * the auto-operation wiring and the chip builder all read the same array, so a
 * chip can never describe a condition differently from the row that produced it.
 */
export interface FilterField<TKey extends string = string> {
  key: TKey;
  label: string;
  type: FilterFieldType;
}
