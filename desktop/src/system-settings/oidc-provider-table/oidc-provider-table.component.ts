import { AfterViewInit, Component, OnInit, signal, TemplateRef, viewChild } from "@angular/core";
import { MatDialog } from "@angular/material/dialog";
import { PageEvent } from "@angular/material/paginator";
import { Sort } from "@angular/material/sort";
import { MatTableDataSource } from "@angular/material/table";
import { Store } from "@ngxs/store";
import { of, switchMap, take, tap } from "rxjs";
import {
  OidcProviderService,
  OidcProviderView,
  PagedDataDataInner,
  PagedRequestCommand,
  Permission,
} from "src/open-api";
import { ConfirmationDialogComponent } from "src/shared-ui/confirmation-dialog/confirmation-dialog.component";
import { TableComponent } from "src/table/table/table.component";
import { DEFAULT_DIALOG_CONFIG } from "../../constants/index";
import { FormMode } from "../../enums/form-mode.enum";
import { SnackbarService } from "../../services/index";
import { AuthState } from "../../store/index";
import { OidcProviderTableState } from "../../store/oidc-provider-table.state";
import {
  SetOrderBy,
  SetPage,
  SetPageSize,
  SetSortDirection,
} from "../../store/oidc-provider-table.state.actions";
import { TableColumn } from "../../table/table-column.interface";
import { OidcProviderFormComponent } from "../oidc-provider-form/oidc-provider-form.component";

@Component({
  selector: "app-oidc-provider-table",
  templateUrl: "./oidc-provider-table.component.html",
  styleUrl: "./oidc-provider-table.component.scss",
  standalone: false,
})
export class OidcProviderTableComponent implements OnInit, AfterViewInit {
  public readonly displayNameCell = viewChild.required<TemplateRef<any>>("displayNameCell");

  public readonly issuerCell = viewChild.required<TemplateRef<any>>("issuerCell");

  public readonly statusCell = viewChild.required<TemplateRef<any>>("statusCell");

  public readonly actionsCell = viewChild.required<TemplateRef<any>>("actionsCell");

  public readonly table = viewChild.required(TableComponent);

  public state = this.store.selectSignal(OidcProviderTableState.state);

  public dataSource = signal(new MatTableDataSource<PagedDataDataInner>([]));

  public displayedColumns: string[] = [];

  public columns: TableColumn[] = [];

  public totalCount = signal(0);

  protected readonly Permission = Permission;

  protected readonly FormMode = FormMode;

  private readonly canUpdate = this.store.selectSignal(
    AuthState.hasAppPermission(Permission.AppOidcProvidersUpdate)
  );

  constructor(
    private matDialog: MatDialog,
    private oidcProviderService: OidcProviderService,
    private snackbarService: SnackbarService,
    private store: Store,
  ) {
  }

  public ngOnInit(): void {
    this.getProviders();
  }

  public ngAfterViewInit(): void {
    this.setColumns();
  }

  public updatePageData(pageEvent: PageEvent): void {
    this.store.dispatch(new SetPage(pageEvent.pageIndex + 1));
    this.store.dispatch(new SetPageSize(pageEvent.pageSize));

    this.getProviders();
  }

  public sorted({ sortState }: { sortState: Sort }): void {
    this.store.dispatch(new SetOrderBy(sortState.active));
    this.store.dispatch(new SetSortDirection(sortState.direction));

    this.getProviders();
  }

  public openProviderDialog(provider?: OidcProviderView, mode?: FormMode): void {
    const dialogRef = this.matDialog.open(OidcProviderFormComponent, DEFAULT_DIALOG_CONFIG);
    const resolvedMode = mode ?? this.resolveDialogMode(provider);

    dialogRef.componentInstance.headerText = this.buildDialogHeaderText(provider, resolvedMode);
    dialogRef.componentInstance.provider = provider;
    dialogRef.componentInstance.mode = resolvedMode;

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        tap((refreshData) => {
          if (refreshData) {
            this.getProviders();
          }
        })
      )
      .subscribe();
  }

  public openDeleteConfirmationDialog(provider: OidcProviderView): void {
    const dialogRef = this.matDialog.open(ConfirmationDialogComponent, DEFAULT_DIALOG_CONFIG);

    dialogRef.componentInstance.headerText = `Delete ${provider.displayName}`;
    dialogRef.componentInstance.dialogContent =
      `Are you sure you want to delete ${provider.displayName}? Anyone who signs in with it will lose that ability, and accounts created by it may be left without a way to sign in.`;

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        switchMap((confirmed) => {
          if (!confirmed) {
            return of(undefined);
          }

          return this.oidcProviderService.deleteOidcProvider(provider.id.toString()).pipe(
            tap(() => {
              this.snackbarService.success("OIDC provider successfully deleted");
              this.getProviders();
            })
          );
        })
      )
      .subscribe();
  }

  // Clicking a provider opens the editor for anyone who may edit it, and the
  // read-only view for everyone else.
  private resolveDialogMode(provider?: OidcProviderView): FormMode {
    if (!provider) {
      return FormMode.add;
    }

    return this.canUpdate() ? FormMode.edit : FormMode.view;
  }

  private buildDialogHeaderText(provider: OidcProviderView | undefined, mode: FormMode): string {
    if (!provider) {
      return "Add OIDC Provider";
    }

    return mode === FormMode.edit ? `Edit ${provider.displayName}` : "View OIDC Provider";
  }

  private getProviders(): void {
    const command: PagedRequestCommand = this.store.selectSnapshot(
      OidcProviderTableState.state
    );

    this.oidcProviderService
      .getPagedOidcProviders(command)
      .pipe(
        take(1),
        tap((pagedData) => {
          this.dataSource.set(new MatTableDataSource<PagedDataDataInner>(pagedData.data));
          this.totalCount.set(pagedData.totalCount);
        })
      )
      .subscribe();
  }

  private setColumns(): void {
    const columns = [
      {
        columnHeader: "Name",
        matColumnDef: "display_name",
        template: this.displayNameCell(),
        sortable: true,
      },
      {
        columnHeader: "Issuer",
        matColumnDef: "issuer_url",
        template: this.issuerCell(),
        sortable: true,
      },
      {
        columnHeader: "Status",
        matColumnDef: "enabled",
        template: this.statusCell(),
        sortable: true,
      },
      {
        columnHeader: "Actions",
        matColumnDef: "actions",
        template: this.actionsCell(),
        sortable: false,
      },
    ] as TableColumn[];

    const tableState = this.store.selectSnapshot(OidcProviderTableState.state);
    if (tableState.orderBy) {
      const column = columns.find((c) => c.matColumnDef === tableState.orderBy);
      if (column) {
        column.defaultSortDirection = tableState.sortDirection;
      }
    }

    this.columns = columns;
    this.displayedColumns = ["display_name", "issuer_url", "enabled"];

    // Gated on either action, not just delete -- gating on delete alone would
    // hide the whole column, and with it the edit button, from an update-only
    // holder.
    const canUseRowActions = this.store.selectSnapshot(
      AuthState.hasAnyAppPermission([
        Permission.AppOidcProvidersUpdate,
        Permission.AppOidcProvidersDelete,
      ])
    );

    if (canUseRowActions) {
      this.displayedColumns.push("actions");
    }
  }
}
