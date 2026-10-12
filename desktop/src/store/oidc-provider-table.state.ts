import { Injectable } from "@angular/core";
import { Action, State, StateContext } from "@ngxs/store";
import { PagedTableInterface } from "src/interfaces/paged-table.interface";
import { SetOrderBy, SetPage, SetPageSize, SetSortDirection } from "./oidc-provider-table.state.actions";
import { PagedTableState } from "./paged-table.state";

@State<PagedTableInterface>({
  name: "oidcProviderTable",
  defaults: {
    page: 1,
    pageSize: 50,
    // Sorted by display name because that is the label an administrator
    // recognizes; the slug is an implementation detail of the login URL.
    orderBy: "display_name",
    sortDirection: "asc",
  },
})
@Injectable()
export class OidcProviderTableState extends PagedTableState {
  @Action(SetPage)
  setPage({ patchState }: StateContext<PagedTableInterface>, payload: SetPage) {
    patchState({
      page: payload.page,
    });
  }

  @Action(SetPageSize)
  setPageSize(
    { patchState }: StateContext<PagedTableInterface>,
    payload: SetPageSize
  ) {
    patchState({
      pageSize: payload.pageSize,
    });
  }

  @Action(SetOrderBy)
  setOrderBy(
    { patchState }: StateContext<PagedTableInterface>,
    payload: SetOrderBy
  ) {
    patchState({
      orderBy: payload.orderBy,
    });
  }

  @Action(SetSortDirection)
  setSortDirection(
    { patchState }: StateContext<PagedTableInterface>,
    payload: SetSortDirection
  ) {
    patchState({
      sortDirection: payload.sortDirection,
    });
  }
}
