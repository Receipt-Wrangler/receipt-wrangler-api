import { CurrencyPipe, DatePipe } from "@angular/common";
import { AfterViewInit, Component, computed, OnInit, signal, TemplateRef, ViewEncapsulation, viewChild } from "@angular/core";
import { MatDialog } from "@angular/material/dialog";
import { PageEvent } from "@angular/material/paginator";
import { Sort } from "@angular/material/sort";
import { MatTableDataSource } from "@angular/material/table";
import { ActivatedRoute, Router } from "@angular/router";
import { UntilDestroy, untilDestroyed } from "@ngneat/until-destroy";
import { Store } from "@ngxs/store";
import { map, take, tap } from "rxjs";
import { fadeInOut } from "src/animations";
import { ReceiptFilterService } from "src/services/receipt-filter.service";
import { ConfirmationDialogComponent } from "src/shared-ui/confirmation-dialog/confirmation-dialog.component";
import { ResetReceiptFilter, SetColumnConfig, SetPage, SetPageSize, SetReceiptFilterData, SetReceiptFilterField, } from "src/store/receipt-table.actions";
import { ReceiptTableState } from "src/store/receipt-table.state";
import { TableColumn } from "src/table/table-column.interface";
import { TableComponent } from "src/table/table/table.component";
import { DEFAULT_DIALOG_CONFIG, DEFAULT_HOST_CLASS } from "../../constants";
import { ReceiptTableColumnConfig } from "../../interfaces";
import {
  BulkStatusUpdateCommand,
  Category,
  FilterOperation,
  Group,
  GroupsService,
  PagedDataDataInner,
  Permission,
  Receipt,
  ReceiptPagedRequestFilter,
  ReceiptService,
  ReceiptStatus,
  Tag,
} from "../../open-api";
import { SnackbarService } from "../../services";
import { ReceiptExportService } from "../../services/receipt-export.service";
import { ReceiptFilterComponent } from "../../shared-ui/receipt-filter/receipt-filter.component";
import { AuthState, GroupState, UserState } from "../../store";
import { CustomCurrencyPipe } from "../../pipes/custom-currency.pipe";
import { applyFormCommand } from "../../utils/index";
import { FilterMonth, monthFilterEntry, monthFromFilterEntry } from "../../utils/receipt-date-filter";
import { buildReceiptFilterForm } from "../../utils/receipt-filter";
import { buildReceiptFilterChips, ReceiptFilterChip } from "../../utils/receipt-filter-chips";
import { isFilterEntryActive } from "../../utils/receipt-filter-entry";
import { openQuickScanDialog } from "../quick-scan-dialog/open-quick-scan-dialog";
import { BulkStatusUpdateComponent } from "../bulk-resolve-dialog/bulk-status-update-dialog.component";
import { ColumnConfigurationDialogComponent } from "../column-configuration-dialog/column-configuration-dialog.component";

@UntilDestroy()
@Component({
  selector: "app-receipts-table",
  templateUrl: "./receipts-table.component.html",
  styleUrls: ["./receipts-table.component.scss"],
  animations: [fadeInOut],
  encapsulation: ViewEncapsulation.None,
  host: DEFAULT_HOST_CLASS,
  // The chip labels are built in TS rather than the template, so the formatting
  // pipes are injected. CustomCurrencyPipe is declared in PipesModule and needs
  // CurrencyPipe, neither of which is providedIn: "root".
  providers: [CurrencyPipe, CustomCurrencyPipe, DatePipe],
  standalone: false
})
export class ReceiptsTableComponent implements OnInit, AfterViewInit {
  constructor(
    private activatedRoute: ActivatedRoute,
    private groupsService: GroupsService,
    private matDialog: MatDialog,
    private receiptExportService: ReceiptExportService,
    private receiptFilterService: ReceiptFilterService,
    private receiptService: ReceiptService,
    private router: Router,
    private snackbarService: SnackbarService,
    private store: Store,
    private customCurrencyPipe: CustomCurrencyPipe,
    private datePipe: DatePipe,
  ) {}

  readonly createdAtCell = viewChild.required<TemplateRef<any>>("createdAtCell");

  readonly dateCell = viewChild.required<TemplateRef<any>>("dateCell");

  readonly nameCell = viewChild.required<TemplateRef<any>>("nameCell");

  readonly paidByCell = viewChild.required<TemplateRef<any>>("paidByCell");

  readonly amountCell = viewChild.required<TemplateRef<any>>("amountCell");

  readonly categoryCell = viewChild.required<TemplateRef<any>>("categoryCell");

  readonly tagCell = viewChild.required<TemplateRef<any>>("tagCell");

  readonly statusCell = viewChild.required<TemplateRef<any>>("statusCell");

  readonly resolvedDateCell = viewChild.required<TemplateRef<any>>("resolvedDateCell");

  readonly actionsCell = viewChild.required<TemplateRef<any>>("actionsCell");

  readonly table = viewChild.required(TableComponent);

  public page = this.store.selectSignal(ReceiptTableState.page);

  public pageSize = this.store.selectSignal(ReceiptTableState.pageSize);

  public filter = this.store.selectSignal(ReceiptTableState.filterData);

  public columnConfig = this.store.selectSignal(ReceiptTableState.columnConfig);

  public selectedGroupId = this.store.selectSignal(GroupState.selectedGroupId);

  private groups = this.store.selectSignal(GroupState.groups);

  private users = this.store.selectSignal(UserState.users);

  private receiptFilter = computed(
    () => this.filter()?.filter as ReceiptPagedRequestFilter | undefined
  );

  /**
   * The month the quick date control is showing, or null when the Date filter
   * is unset or is something a month cannot express.
   */
  public stepperMonth = computed(() => monthFromFilterEntry(this.receiptFilter()?.date));

  public stepperLabel = computed(() => {
    const month = this.stepperMonth();
    if (month) {
      return this.datePipe.transform(new Date(month.year, month.month, 1), "LLLL y") ?? "";
    }

    // A Date filter the stepper cannot describe still has to be visible as a
    // filter — the chip beside it spells out what it actually is.
    return isFilterEntryActive(this.receiptFilter()?.date) ? "Custom" : "All time";
  });

  public filterChips = computed<ReceiptFilterChip[]>(() => {
    const categories = this.categories();
    const tags = this.tags();
    const groups = this.groups();
    const users = this.users();

    return buildReceiptFilterChips(
      this.receiptFilter(),
      {
        categories,
        tags,
        // Not groupsWithoutAll: a group filter set on the All Groups view is
        // persisted, so it can outlive the view that offers the control and
        // still has to name itself here.
        groups,
        users,
        formatDate: (value) => this.datePipe.transform(value as string) ?? "",
        formatCurrency: (value) => this.customCurrencyPipe.transform(value as number),
      },
      // The stepper already says "September 2026", so don't say it twice.
      this.stepperMonth() ? ["date"] : []
    );
  });

  private numFiltersAppliedRaw = this.store.selectSignal(ReceiptTableState.numFiltersApplied);

  public numFiltersApplied = computed(() => {
    const num = this.numFiltersAppliedRaw();
    return num > 0 ? num : undefined;
  });

  public categories = signal<Category[]>([]);

  public tags = signal<Tag[]>([]);

  public groupId: string = "0";

  public dataSource = signal(new MatTableDataSource<PagedDataDataInner>([]));

  public displayedColumns = signal<string[]>([]);

  public columns = signal<TableColumn[]>([]);

  public totalCount = signal(0);

  public selectedReceiptIds = signal<number[]>([]);

  public firstSort: boolean = true;

  public canEdit: boolean = false;

  public canCreate: boolean = false;

  public canQuickScan: boolean = false;

  public canPollEmail: boolean = false;

  public headerText: string = "";

  public group?: Group;

  protected readonly Permission = Permission;

  public ngOnInit(): void {
    this.groupId = this.store
      .selectSnapshot(GroupState.selectedGroupId)
      ?.toString();
    this.setGroup();
    this.setCanEdit();

    this.setHeaderText();

    // Filter options come from the selected group's AppData catalog (filtered to
    // the user's grants), so a restricted user can't filter by a hidden one.
    const numericGroupId = Number(this.groupId);
    this.categories.set(
      Number.isNaN(numericGroupId)
        ? []
        : this.store.selectSnapshot(AuthState.groupCategories(numericGroupId))
    );
    this.tags.set(
      Number.isNaN(numericGroupId)
        ? []
        : this.store.selectSnapshot(AuthState.groupTags(numericGroupId))
    );
    this.getInitialData();
  }

  private setGroup(): void {
    this.group = this.store.selectSnapshot(GroupState.getGroupById(this.groupId));
  }

  private getInitialData(): void {
    this.receiptFilterService
      .getPagedReceiptsForGroups(this.groupId)
      .pipe(
        take(1),
        tap((pagedData) => {
          this.dataSource.set(new MatTableDataSource<PagedDataDataInner>(pagedData.data));
          this.totalCount.set(pagedData.totalCount);
          this.setColumns();
        })
      )
      .subscribe();
  }

  private setCanEdit(): void {
    const groupId = Number.parseInt(this.groupId);
    this.canEdit = this.store.selectSnapshot(
      AuthState.hasGroupPermission(groupId, Permission.GroupReceiptsUpdate)
    );
    this.canCreate = this.store.selectSnapshot(
      AuthState.hasGroupPermission(groupId, Permission.GroupReceiptsCreate)
    );
    this.canQuickScan = this.store.selectSnapshot(
      AuthState.hasGroupPermission(groupId, Permission.GroupReceiptsQuickScan)
    );
    this.canPollEmail = this.store.selectSnapshot(
      AuthState.hasGroupPermission(groupId, Permission.GroupEmailPoll)
    );
  }

  private setHeaderText(): void {
    const group = this.store.selectSnapshot(
      GroupState.getGroupById(this.groupId)
    );
    if (group) {
      if (group.name.toLowerCase().includes("receipt")) {
        this.headerText = group.name;
      } else {
        this.headerText = `${group.name} Receipts`;
      }
    }
  }

  public ngAfterViewInit(): void {
    this.setSelectedReceiptIdsObservable();
  }

  private setSelectedReceiptIdsObservable(): void {
    this.table()?.selection?.changed
      .pipe(
        untilDestroyed(this),
        map((event) => (event.source.selected as Receipt[]).map((r) => r.id)),
        tap((ids) => this.selectedReceiptIds.set(ids))
      )
      .subscribe();
  }

  private setColumns(): void {
    const currentColumnConfig = this.store.selectSnapshot(ReceiptTableState.columnConfig);

    const allColumns = [
      {
        columnHeader: "Added At",
        matColumnDef: "created_at",
        template: this.createdAtCell(),
        sortable: true,
      },
      {
        columnHeader: "Receipt Date",
        matColumnDef: "date",
        template: this.dateCell(),
        sortable: true,
      },
      {
        columnHeader: "Name",
        matColumnDef: "name",
        template: this.nameCell(),
        sortable: true,
      },
      {
        columnHeader: "Paid By",
        matColumnDef: "paid_by_user_id",
        template: this.paidByCell(),
        sortable: true,
      },
      {
        columnHeader: "Amount",
        matColumnDef: "amount",
        template: this.amountCell(),
        sortable: true,
      },
      {
        columnHeader: "Categories",
        matColumnDef: "categories",
        template: this.categoryCell(),
        sortable: false,
      },
      {
        columnHeader: "Tags",
        matColumnDef: "tags",
        template: this.tagCell(),
        sortable: false,
      },
      {
        columnHeader: "Status",
        matColumnDef: "status",
        template: this.statusCell(),
        sortable: true,
      },
      {
        columnHeader: "Resolved Date",
        matColumnDef: "resolved_date",
        template: this.resolvedDateCell(),
        sortable: true,
      },
    ] as TableColumn[];

    // Filter and order columns based on configuration
    const visibleColumnConfigs = currentColumnConfig
      .filter(config => config.visible)
      .sort((a, b) => a.order - b.order);

    const columns = visibleColumnConfigs
      .map(config => allColumns.find(col => col.matColumnDef === config.matColumnDef))
      .filter(col => col !== undefined) as TableColumn[];

    const displayColumns = ["select", ...visibleColumnConfigs.map(config => config.matColumnDef)];

    if (this.canEdit) {
      columns.push({
        columnHeader: "Actions",
        matColumnDef: "actions",
        template: this.actionsCell(),
        sortable: false,
      });
      displayColumns.push("actions");
    }

    const filter = this.store.selectSnapshot(ReceiptTableState.filterData);
    const orderByIndex = columns.findIndex(
      (c) => c.matColumnDef === filter.orderBy
    );

    if (orderByIndex >= 0) {
      columns[orderByIndex].defaultSortDirection = filter.sortDirection;
    } else if (columns.length > 0) {
      columns[0].defaultSortDirection = "desc";
    }

    this.columns.set(columns);
    this.displayedColumns.set(displayColumns);
  }

  public sort(sortState: Sort): void {
    if (!this.firstSort) {
      const filterData = this.store.selectSnapshot(
        ReceiptTableState.filterData
      );

      this.store.dispatch(
        new SetReceiptFilterData({
          page: filterData.page,
          pageSize: filterData.pageSize,
          orderBy: sortState.active,
          sortDirection: sortState.direction,
          filter: filterData.filter,
        })
      );

      this.getFilteredReceipts();
    }
    this.firstSort = false;
  }

  public filterButtonClicked(): void {
    const filter = this.store.selectSnapshot(ReceiptTableState.filterData).filter as any;

    const dialogRef = this.matDialog.open(ReceiptFilterComponent, {
      minWidth: "75%",
      maxWidth: "100%",
    });

    dialogRef.componentInstance.categories = this.categories();
    dialogRef.componentInstance.tags = this.tags();
    dialogRef.componentInstance.parentForm = buildReceiptFilterForm(filter, this);
    dialogRef.componentInstance.headerText = "Filter Receipts";
    // The group filter is only meaningful on the "All groups" view; a
    // single-group view is already scoped to one group.
    dialogRef.componentInstance.showGroupFilter = this.group?.isAllGroup ?? false;
    const formCommandSubscription = dialogRef.componentInstance.formCommand.subscribe((formCommand) => {
      applyFormCommand(dialogRef.componentInstance.parentForm, formCommand);
    });

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        tap((applyFilter) => {
          if (applyFilter) {
            this.store.dispatch(new SetPage(1));
            this.getFilteredReceipts();
          }

          formCommandSubscription.unsubscribe();
        })
      )
      .subscribe();
  }

  public monthSelected(month: FilterMonth): void {
    // The quick control IS the Date filter, so it overwrites whatever was there.
    this.applyFilterField("date", monthFilterEntry(month) as any);
  }

  public allTimeSelected(): void {
    this.applyFilterField("date", null);
  }

  public filterChipCleared(field: keyof ReceiptPagedRequestFilter): void {
    this.applyFilterField(field, null);
  }

  /**
   * The one write path for a single-field filter change, so the page reset and
   * the refetch can never be forgotten — narrowing a filter while on page 7
   * would otherwise land on an empty page.
   */
  private applyFilterField(
    field: keyof ReceiptPagedRequestFilter,
    entry: { operation: FilterOperation | null; value: unknown } | null
  ): void {
    this.store.dispatch(new SetReceiptFilterField(field, entry));
    this.store.dispatch(new SetPage(1));
    this.getFilteredReceipts();
  }

  public quickScanClicked(): void {
    openQuickScanDialog(this.matDialog)
      .pipe(
        take(1),
        tap(() => this.getFilteredReceipts())
      )
      .subscribe();
  }

  public exportAllReceipts(): void {
    this.receiptExportService.exportReceiptsFromFilter(this.groupId, this.filter());
  }

  public resetFilterButtonClicked(): void {
    this.store.dispatch(new ResetReceiptFilter());
    this.getFilteredReceipts();
  }

  public configureColumnsButtonClicked(): void {
    const currentColumnConfig = this.store.selectSnapshot(ReceiptTableState.columnConfig);

    const dialogRef = this.matDialog.open(ColumnConfigurationDialogComponent, {
      ...DEFAULT_DIALOG_CONFIG,
      data: { currentColumns: currentColumnConfig }
    });

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        tap((result: ReceiptTableColumnConfig[] | null) => {
          if (result) {
            this.store.dispatch(new SetColumnConfig(result));
            this.setColumns();
          }
        })
      )
      .subscribe();
  }

  public getFilteredReceipts(): void {
    this.receiptFilterService
      .getPagedReceiptsForGroups(this.groupId.toString())
      .pipe(
        take(1),
        tap((pagedData) => {
          this.dataSource.set(new MatTableDataSource(pagedData.data));
          this.totalCount.set(pagedData.totalCount);
        })
      )
      .subscribe();
  }

  public deleteReceipt(row: Receipt): void {
    const dialogRef = this.matDialog.open(ConfirmationDialogComponent);

    dialogRef.componentInstance.headerText = "Delete Receipt";
    dialogRef.componentInstance.dialogContent = `Are you sure you would like to delete the receipt ${row.name}? This action is irreversible.`;

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        tap((r) => {
          if (r) {
            this.receiptService
              .deleteReceiptById(row.id as number)
              .pipe(
                take(1),
                tap(() => {
                  this.dataSource.update(ds => new MatTableDataSource(ds.data.filter(
                    (r) => r.id !== row.id
                  )));
                  this.snackbarService.success("Receipt successfully deleted");
                })
              )
              .subscribe();
          }
        })
      )
      .subscribe();
  }

  public duplicateReceipt(id: string): void {
    this.receiptService
      .duplicateReceipt(Number.parseInt(id))
      .pipe(
        tap((r: Receipt) => {
          this.snackbarService.success("Receipt successfully duplicated");
          this.router.navigateByUrl(`/receipts/${r.id}/view`);
        })
      )
      .subscribe();
  }

  public updatePageData(pageEvent: PageEvent): void {
    const newPage = pageEvent.pageIndex + 1;
    this.store.dispatch(new SetPage(newPage));
    this.store.dispatch(new SetPageSize(pageEvent.pageSize));

    this.getFilteredReceipts();
  }

  public showStatusUpdateDialog(): void {
    const ref = this.matDialog.open(
      BulkStatusUpdateComponent,
      DEFAULT_DIALOG_CONFIG
    );

    ref
      .afterClosed()
      .pipe(
        take(1),
        tap(
          (
            commentForm:
              | {
              comment: string;
              status: ReceiptStatus;
            }
              | undefined
          ) => {
            const table = this.table();
            if (table.selection.hasValue() && commentForm) {
              const receiptIds = (
                table.selection.selected as Receipt[]
              ).map((r) => r.id as number);

              const bulkResolve: BulkStatusUpdateCommand = {
                comment: commentForm?.comment ?? "",
                status: commentForm?.status,
                receiptIds: receiptIds,
              };
              this.receiptService
                .bulkReceiptStatusUpdate(bulkResolve)
                .pipe(
                  take(1),
                  tap((receipts) => {
                    let newReceipts = Array.from(this.dataSource().data);
                    receipts.forEach((r) => {
                      const receiptInTable = newReceipts.find(
                        (nr) => r.id === nr.id
                      ) as any as Receipt;
                      if (receiptInTable) {
                        receiptInTable.status = r.status;
                        receiptInTable.resolvedDate = r.resolvedDate;
                      }
                    });
                    this.dataSource.set(new MatTableDataSource(newReceipts));
                  })
                )
                .subscribe();
            }
          }
        )
      )
      .subscribe();
  }

  public pollEmail(): void {
    const groupId = this.store.selectSnapshot(GroupState.selectedGroupId);

    this.groupsService
      .pollGroupEmail(groupId as any)
      .pipe(
        take(1),
        tap(() => {
          this.snackbarService.success("Email successfully poll successfully queued");
        }),
      )
      .subscribe();
  }

  public exportSelectedReceipts(): void {
    const receiptIds = this.dataSource().data.map(data => data.id);
    this.receiptExportService.exportReceiptsById(receiptIds);
  }
}
