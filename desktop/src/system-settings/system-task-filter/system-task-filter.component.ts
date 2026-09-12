import { Component, Input, OnInit, computed, output } from "@angular/core";
import { FormGroup } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { Store } from "@ngxs/store";
import {
  SYSTEM_RAN_BY_OPTION_DISPLAY_VALUE,
  SYSTEM_RAN_BY_OPTION_ID,
  SYSTEM_TASK_FILTER_FIELDS,
  SYSTEM_TASK_TYPE_OPTIONS
} from "src/constants";
import { SetSystemTaskFilter } from "src/store/system-task-table.state.actions";
import { setupAutoOperationSelection } from "src/utils/filter-form";
import { take, tap } from "rxjs";
import { FormCommand } from "../../form/index";
import { UserState } from "../../store";

/**
 * The System Tasks table's filter dialog. Deliberately the same shape as
 * `app-receipt-filter` — the caller owns the form (`buildSystemTaskFilterForm`)
 * and applies the reset `FormCommand`s, and each row is the shared
 * `app-filter-field`.
 */
@Component({
  selector: "app-system-task-filter",
  templateUrl: "./system-task-filter.component.html",
  styleUrls: ["./system-task-filter.component.scss"],
  standalone: false
})
export class SystemTaskFilterComponent implements OnInit {
  @Input() public headerText: string = "";

  @Input() public parentForm: FormGroup = new FormGroup({});

  public readonly formCommand = output<FormCommand>();

  public systemTaskTypeOptions = SYSTEM_TASK_TYPE_OPTIONS;

  private users = this.store.selectSignal(UserState.users);

  /**
   * "System" is pinned above the real users so tasks with no `ranByUserId` —
   * the ones the table labels "System", and most of the table — are selectable
   * at all. The API resolves the sentinel to an IS NULL disjunct.
   */
  public ranByOptions = computed(() => [
    { id: SYSTEM_RAN_BY_OPTION_ID, displayName: SYSTEM_RAN_BY_OPTION_DISPLAY_VALUE },
    ...this.users().map((user) => ({ id: user.id, displayName: user.displayName })),
  ]);

  constructor(
    private store: Store,
    private dialogRef: MatDialogRef<SystemTaskFilterComponent>
  ) {}

  public ngOnInit(): void {
    setupAutoOperationSelection(this.parentForm, "", SYSTEM_TASK_FILTER_FIELDS);
  }

  public resetFilter(): void {
    this.formCommand.emit({
      path: "",
      command: "reset",
    });
    // The two list fields are FormArrays, which reset() leaves populated.
    this.formCommand.emit({
      path: "type.value",
      command: "clear",
    });
    this.formCommand.emit({
      path: "ranBy.value",
      command: "clear",
    });
  }

  public submitButtonClicked(): void {
    if (!this.parentForm.valid) {
      this.parentForm.markAllAsTouched();
      return;
    }

    this.store
      .dispatch(new SetSystemTaskFilter(this.parentForm.value))
      .pipe(
        take(1),
        tap(() => {
          this.dialogRef.close(true);
        })
      )
      .subscribe();
  }

  public cancelButtonClicked(): void {
    this.dialogRef.close(false);
  }
}
