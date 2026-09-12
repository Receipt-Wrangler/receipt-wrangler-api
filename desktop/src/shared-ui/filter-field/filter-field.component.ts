import { Component, OnInit, input } from "@angular/core";
import { FormControl, FormGroup } from "@angular/forms";
import { endOfDay, startOfMonth } from "date-fns";
import { FilterFieldType } from "src/constants";
import { FilterOperation } from "../../open-api";

/**
 * One row of a filter dialog: the field's value editor beside its Operation
 * select, switching shape with the selected operation (a two-slot range for
 * BETWEEN, a read-only implied range for WITHIN_CURRENT_MONTH, a single editor
 * otherwise).
 *
 * Deliberately presentational and form-agnostic — it reaches into the caller's
 * `parentForm` by path, exactly as the receipt filter's template did before it
 * was extracted here, so `app-receipt-filter` and `app-system-task-filter`
 * render identical rows.
 */
@Component({
  selector: "app-filter-field",
  templateUrl: "./filter-field.component.html",
  styleUrls: ["./filter-field.component.scss"],
  standalone: false
})
export class FilterFieldComponent implements OnInit {
  public readonly parentForm = input<FormGroup>(new FormGroup({}));

  public readonly basePath = input<string>("");

  public readonly label = input<string>("");

  public readonly fieldName = input<string>("");

  public readonly type = input<FilterFieldType>("text");

  public readonly isCurrency = input<boolean>(false);

  public readonly options = input<any[]>([]);

  public readonly optionFilterKey = input<string>("");

  public readonly optionDisplayKey = input<string>("");

  public readonly optionValueKey = input<string>("");

  public readonly multiple = input<boolean>(false);

  // Display-only: WITHIN_CURRENT_MONTH carries no value, so the two datepickers
  // that show the range it implies are disabled and never read back.
  public startOfMonthFormControl = new FormControl(startOfMonth(new Date()));

  public endOfTodayFormControl = new FormControl(endOfDay(new Date()));

  public ngOnInit(): void {
    this.startOfMonthFormControl.disable();
    this.endOfTodayFormControl.disable();
  }

  protected readonly FilterOperation = FilterOperation;
}
