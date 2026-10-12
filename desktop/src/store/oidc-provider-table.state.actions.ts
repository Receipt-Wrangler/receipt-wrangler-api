import { SortDirection } from "@angular/material/sort";

export class SetPage {
  static readonly type = "[OidcProviderTableComponent] Set Page";

  constructor(public page: number) {}
}

export class SetPageSize {
  static readonly type = "[OidcProviderTableComponent] Set Page Size";

  constructor(public pageSize: number) {}
}

export class SetOrderBy {
  static readonly type = "[OidcProviderTableComponent] Set Order By";

  constructor(public orderBy: string) {}
}

export class SetSortDirection {
  static readonly type = "[OidcProviderTableComponent] Set Sort Direction";

  constructor(public sortDirection: SortDirection) {}
}
