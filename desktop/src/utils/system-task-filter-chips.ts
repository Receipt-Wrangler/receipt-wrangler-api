import {
  SYSTEM_RAN_BY_OPTION_DISPLAY_VALUE,
  SYSTEM_RAN_BY_OPTION_ID,
  SYSTEM_TASK_FILTER_FIELDS,
  SystemTaskFilterFieldKey,
} from "../constants/system-task-filter-fields.constant";
import { SYSTEM_TASK_TYPE_OPTIONS } from "../constants/system-task-type-options";
import { SystemTaskPagedRequestFilter, User } from "../open-api";
import { buildFilterChips, FilterChip } from "./filter-chips";

export type SystemTaskFilterChip = FilterChip<SystemTaskFilterFieldKey>;

export interface SystemTaskFilterChipLookups {
  users: readonly User[];
  formatDate: (value: unknown) => string;
}

/**
 * One chip per active system task filter condition. See `buildFilterChips` for
 * the shared label rules; this wrapper only supplies the system-task id
 * resolution.
 */
export function buildSystemTaskFilterChips(
  filter: SystemTaskPagedRequestFilter | undefined | null,
  lookups: SystemTaskFilterChipLookups,
): SystemTaskFilterChip[] {
  return buildFilterChips(
    SYSTEM_TASK_FILTER_FIELDS,
    filter as Record<string, unknown> | undefined | null,
    {
      formatDate: lookups.formatDate,
      // No numeric system task field today; a currency label would be wrong for
      // one anyway, so this stays a plain stringify.
      formatCurrency: (value) => value?.toString() ?? "",
      resolveOptionName: (key, id) =>
        resolveOptionName(key as SystemTaskFilterFieldKey, id, lookups),
    },
  );
}

/**
 * An id the caller can no longer resolve — a deleted user — falls back to the
 * raw id rather than vanishing, so a condition that is actively removing rows
 * is never invisible.
 */
function resolveOptionName(
  key: SystemTaskFilterFieldKey,
  id: unknown,
  lookups: SystemTaskFilterChipLookups,
): string {
  const rawId = String(id);

  switch (key) {
    case "type":
      return SYSTEM_TASK_TYPE_OPTIONS.find((option) => option.value === rawId)?.displayValue ?? rawId;
    case "ranBy":
      if (rawId === String(SYSTEM_RAN_BY_OPTION_ID)) {
        return SYSTEM_RAN_BY_OPTION_DISPLAY_VALUE;
      }

      return lookups.users.find((user) => String(user.id) === rawId)?.displayName ?? rawId;
    default:
      return rawId;
  }
}
