import { SortDirection } from "@angular/material/sort";
import { ReceiptDateFilterFieldKey } from "../constants/receipt-filter-fields.constant";
import { ReceiptPagedRequestFilter } from "../open-api";
import { ReceiptTableColumnConfig } from "./receipt-table-column-config.interface";

export interface ReceiptTableInterface {
  page: number;
  pageSize: number;
  orderBy: string;
  sortDirection: SortDirection;
  filter: ReceiptPagedRequestFilter;
  /**
   * The date field the quick date control writes to. Optional because state
   * persisted before it existed deserializes without it — read it through
   * `ReceiptTableState.quickDateField`, which supplies the fallback.
   */
  quickDateField?: ReceiptDateFilterFieldKey;
  columnConfig?: ReceiptTableColumnConfig[];
}
