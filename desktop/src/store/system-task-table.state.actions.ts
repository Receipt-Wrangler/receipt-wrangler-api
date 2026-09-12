import { SortDirection } from "@angular/material/sort";
import { SystemTaskFilterFieldKey } from "src/constants";
import { SystemTaskPagedRequestFilter } from "src/open-api";

export class SetPage {
  static readonly type = "[SystemTaskTableComponent] Set Page";

  constructor(public page: number) {}
}

export class SetPageSize {
  static readonly type = "[SystemTaskTableComponent] Set Page Size";

  constructor(public pageSize: number) {}
}

export class SetOrderBy {
  static readonly type = "[SystemTaskTableComponent] Set Order By";

  constructor(public orderBy: string) {}
}

export class SetSortDirection {
  static readonly type = "[SystemTaskTableComponent] Set Sort Direction";

  constructor(public sortDirection: SortDirection) {}
}

export class SetSystemTaskFilter {
  static readonly type = "[SystemTaskTableComponent] Set System Task Filter";

  constructor(public data: SystemTaskPagedRequestFilter) {}
}

export class SetSystemTaskFilterField {
  static readonly type = "[SystemTaskTableComponent] Set System Task Filter Field";

  /** A null entry clears the field back to its default. */
  constructor(
    public field: SystemTaskFilterFieldKey,
    public entry: object | null
  ) {}
}

export class ResetSystemTaskFilter {
  static readonly type = "[SystemTaskTableComponent] Reset System Task Filter";
}
