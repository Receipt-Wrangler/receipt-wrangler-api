import { Component, Input, OnInit, TemplateRef, input, output } from "@angular/core";
import { FormGroup, } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { Store } from "@ngxs/store";
import { take, tap } from "rxjs";
import { RECEIPT_FILTER_FIELDS, RECEIPT_STATUS_OPTIONS } from "src/constants";
import { SetReceiptFilter } from "src/store/receipt-table.actions";
import { setupAutoOperationSelection } from "src/utils/filter-form";
import { FormCommand } from "../../form/index";
import { Category, Tag } from "../../open-api";
import { GroupState } from "../../store";

@Component({
  selector: "app-receipt-filter",
  templateUrl: "./receipt-filter.component.html",
  styleUrls: ["./receipt-filter.component.scss"],
  standalone: false
})
export class ReceiptFilterComponent implements OnInit {
  @Input() public headerText: string = "";

  public readonly footerTemplate = input<TemplateRef<any>>();

  public readonly isOpen = input<boolean>(true);

  @Input() public previewTemplate?: TemplateRef<any>;

  public readonly previewTemplateContext = input<any>();

  public readonly inDialog = input<boolean>(true);

  @Input() public parentForm: FormGroup = new FormGroup({});

  @Input() public basePath: string = "";

  public readonly formCommand = output<FormCommand>();

  public readonly formInitialized = output<FormGroup>();

  public receiptStatusOptions = RECEIPT_STATUS_OPTIONS;

  // Sourced from the selected group's grant-filtered AppData catalog by the
  // caller (set imperatively for the dialog, bound for the inline dashboard use)
  // rather than the admin-only global GET /category and GET /tag endpoints.
  @Input() public categories: Category[] = [];

  @Input() public tags: Tag[] = [];

  // Only rendered on the "All groups" view (the caller sets it from
  // group.isAllGroup); hidden on single-group and dashboard-widget views, where
  // the view is already pinned to one group.
  @Input() public showGroupFilter: boolean = false;

  // The user's own groups (minus the synthetic "All" group) — the selectable
  // set for the group filter.
  public groups = this.store.selectSignal(GroupState.groupsWithoutAll);

  constructor(
    private store: Store,
    private dialogRef: MatDialogRef<ReceiptFilterComponent>
  ) {}

  public ngOnInit(): void {
    // RECEIPT_FILTER_FIELDS is the shared definition of every filter field's
    // label and operation type, so these rows and the filter chips that describe
    // them cannot drift apart.
    setupAutoOperationSelection(this.parentForm, this.basePath, RECEIPT_FILTER_FIELDS);
  }

  public resetFilter(): void {
    this.formCommand.emit({
      path: `${this.basePath}`,
      command: "reset",
    });
    this.formCommand.emit({
      path: `${this.basePath}paidBy.value`,
      command: "clear",
    });
    this.formCommand.emit({
      path: `${this.basePath}categories.value`,
      command: "clear",
    });
    this.formCommand.emit({
      path: `${this.basePath}tags.value`,
      command: "clear",
    });
    this.formCommand.emit({
      path: `${this.basePath}status.value`,
      command: "clear",
    });
    this.formCommand.emit({
      path: `${this.basePath}group.value`,
      command: "clear",
    });
  }

  public submitButtonClicked(): void {
    const filter = this.parentForm.value;

    if (this.parentForm.valid) {
      this.store
        .dispatch(new SetReceiptFilter(filter))
        .pipe(
          take(1),
          tap(() => {
            this.dialogRef.close(true);
          })
        )
        .subscribe();
    } else {
      this.parentForm.markAllAsTouched();
    }
  }

  public cancelButtonClicked(): void {
    this.dialogRef.close(false);
  }
}
