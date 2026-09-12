import { PagedTableInterface } from "./paged-table.interface";
import { SystemTaskPagedRequestFilter } from "../open-api";

export interface SystemTaskTableInterface extends PagedTableInterface {
  // Optional because the slice is persisted to localStorage: a session saved
  // before the filter shipped rehydrates without this key. Every read must
  // fall back to buildDefaultSystemTaskFilter().
  filter?: SystemTaskPagedRequestFilter;
}
