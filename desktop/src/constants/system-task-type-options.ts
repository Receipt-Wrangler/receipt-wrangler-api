import { FormOption } from "src/interfaces/form-option.interface";
import { SystemTaskTypePipe } from "src/shared-ui/task-table/system-task-type.pipe";
import { SystemTaskType } from "../open-api";

/**
 * The task types that never appear as a top-level row: `GetPagedSystemTasks`
 * excludes them because they are only ever recorded as children of another
 * task, and the table shows them inside an expanded row.
 *
 * Offering them in the Type filter would be a picker that can only ever return
 * zero rows, so they are dropped from the options below. Keep in sync with
 * `filteredSystemTaskTypes` in `api/internal/repositories/system_task.go`
 * (pinned there by `TestGetPagedSystemTasksExcludesChildTaskTypes`).
 */
export const CHILD_ONLY_SYSTEM_TASK_TYPES: readonly SystemTaskType[] = [
  SystemTaskType.ReceiptUploaded,
  SystemTaskType.ChatCompletion,
  SystemTaskType.OcrProcessing,
];

const systemTaskTypePipe = new SystemTaskTypePipe();

export const SYSTEM_TASK_TYPE_OPTIONS: FormOption[] = Object.values(SystemTaskType)
  .filter((value) => !CHILD_ONLY_SYSTEM_TASK_TYPES.includes(value))
  .map((value) => ({
    value: value,
    displayValue: systemTaskTypePipe.transform(value),
  }));
