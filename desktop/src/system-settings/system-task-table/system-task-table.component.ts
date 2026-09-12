import { DatePipe } from "@angular/common";
import { Component, OnInit, TemplateRef, computed, viewChild } from "@angular/core";
import { MatDialog } from "@angular/material/dialog";
import { ActivatedRoute } from "@angular/router";
import { UntilDestroy } from "@ngneat/until-destroy";
import { Store } from "@ngxs/store";
import { take, tap } from "rxjs";
import { DEFAULT_DIALOG_CONFIG } from "../../constants/dialog.constant";
import { SystemTaskFilterFieldKey } from "../../constants";
import { AssociatedEntityType, Prompt, ReceiptProcessingSettings, SystemTaskPagedRequestFilter } from "../../open-api";
import { TABLE_SERVICE_INJECTION_TOKEN } from "../../services/injection-tokens/table-service";
import { SystemTaskTableService } from "../../services/system-task-table.service";
import { TaskTableComponent } from "../../shared-ui/task-table/task-table.component";
import { UserState } from "../../store";
import { SystemTaskTableState } from "../../store/system-task-table.state";
import {
  ResetSystemTaskFilter,
  SetPage,
  SetSystemTaskFilterField
} from "../../store/system-task-table.state.actions";
import { applyFormCommand } from "../../utils/form.utils";
import { buildSystemTaskFilterChips, SystemTaskFilterChip } from "../../utils/system-task-filter-chips";
import { buildSystemTaskFilterForm } from "../../utils/system-task-filter";
import { SystemTaskFilterComponent } from "../system-task-filter/system-task-filter.component";

@UntilDestroy()
@Component({
  selector: "app-system-task-table",
  templateUrl: "./system-task-table.component.html",
  styleUrl: "./system-task-table.component.scss",
  providers: [
    {
      provide: TABLE_SERVICE_INJECTION_TOKEN,
      useClass: SystemTaskTableService
    },
    DatePipe,
  ],
  standalone: false
})
export class SystemTaskTableComponent implements OnInit {
  public readonly expandedRowTemplate = viewChild.required<TemplateRef<any>>("expandedRowTemplate");
  public readonly taskTableComponent = viewChild.required(TaskTableComponent);

  public prompts: Prompt[] = [];
  public allReceiptProcessingSettings: ReceiptProcessingSettings[] = [];
  protected readonly AssociatedEntityType = AssociatedEntityType;

  public filter = this.store.selectSignal(SystemTaskTableState.filter);

  public numFiltersApplied = this.store.selectSignal(SystemTaskTableState.numFiltersApplied);

  /**
   * Handed to `app-task-table` so it reads the filter when it builds the
   * request. A stable reference, so the binding's identity never changes; see
   * `TaskTableComponent.filterProvider` for why it cannot be the value.
   */
  public readonly filterProvider = (): SystemTaskPagedRequestFilter => this.filter();

  private users = this.store.selectSignal(UserState.users);

  public filterChips = computed<SystemTaskFilterChip[]>(() =>
    buildSystemTaskFilterChips(this.filter(), {
      users: this.users(),
      formatDate: (value) => this.datePipe.transform(value as string) ?? "",
    })
  );

  constructor(
    private activatedRoute: ActivatedRoute,
    private store: Store,
    private matDialog: MatDialog,
    private datePipe: DatePipe
  ) {}

  public ngOnInit(): void {
    this.prompts = this.activatedRoute.snapshot.data["prompts"] || [];
    this.allReceiptProcessingSettings = this.activatedRoute.snapshot.data["allReceiptProcessingSettings"] || [];
  }

  public refresh(): void {
    this.taskTableComponent().getTableData();
  }

  public filterButtonClicked(): void {
    const dialogRef = this.matDialog.open(SystemTaskFilterComponent, {
      ...DEFAULT_DIALOG_CONFIG,
      minWidth: "75%",
      maxWidth: "100%",
    });
    dialogRef.componentInstance.headerText = "Filter System Tasks";
    dialogRef.componentInstance.parentForm = buildSystemTaskFilterForm(this.filter(), this);

    const formCommandSubscription = dialogRef.componentInstance.formCommand.subscribe(
      (formCommand) => applyFormCommand(dialogRef.componentInstance.parentForm, formCommand)
    );

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        tap((applyFilter) => {
          if (applyFilter) {
            this.store.dispatch(new SetPage(1));
            this.refresh();
          }

          formCommandSubscription.unsubscribe();
        })
      )
      .subscribe();
  }

  public resetFilterButtonClicked(): void {
    this.applyFilterChange(new ResetSystemTaskFilter());
  }

  public filterChipCleared(field: SystemTaskFilterFieldKey): void {
    this.applyFilterChange(new SetSystemTaskFilterField(field, null));
  }

  /**
   * Every filter write outside the dialog goes through here so the page reset
   * cannot be forgotten — narrowing from page 7 would otherwise land the user
   * on an empty page.
   */
  private applyFilterChange(action: object): void {
    this.store.dispatch(action);
    this.store.dispatch(new SetPage(1));
    this.refresh();
  }
}
