import { SystemTaskPagedRequestFilter } from "../open-api";
import { FilterField } from "./filter-fields.constant";

export type SystemTaskFilterFieldKey = keyof SystemTaskPagedRequestFilter;

export type SystemTaskFilterField = FilterField<SystemTaskFilterFieldKey>;

/**
 * The filterable system task fields, in the order the filter dialog renders
 * them. Single definition of each field's label and type — the dialog rows, the
 * auto-operation wiring and the filter chips all read it.
 *
 * "Ran By" is a `list`, not `users`, so the picker can prepend the synthetic
 * "System" option for tasks with no user (see SYSTEM_RAN_BY_OPTION_ID). Both
 * types offer the same CONTAINS-only operation, so the row is identical either
 * way. Same reason the report builder's paid-by picker is a plain
 * `app-autocomlete` rather than `app-user-autocomplete`.
 */
export const SYSTEM_TASK_FILTER_FIELDS: readonly SystemTaskFilterField[] = [
  { key: "type", label: "Type", type: "list" },
  { key: "ranBy", label: "Ran By", type: "list" },
  { key: "startedAt", label: "Started At", type: "date" },
  { key: "endedAt", label: "Ended At", type: "date" },
];

/**
 * The "System" entry pinned above the real users in the Ran By picker. It
 * matches tasks whose `ranByUserId` is null — the ones the table renders as
 * "System", which is most of them — and the API turns it into an
 * `IS NULL` disjunct (`repositories.SystemRanByUserId`).
 *
 * Negative so it can never collide with a real user id, the convention
 * `OWN_PAID_RECEIPTS_OPTION_ID` and `REPORT_GENERATOR_PAID_BY_ID` already use.
 */
export const SYSTEM_RAN_BY_OPTION_ID = -1;

export const SYSTEM_RAN_BY_OPTION_DISPLAY_VALUE = "System";
