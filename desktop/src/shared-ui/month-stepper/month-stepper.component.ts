import { Component, computed, input, linkedSignal, output } from "@angular/core";
import { MatMenuModule } from "@angular/material/menu";
import { ButtonModule } from "../../button";
import { FilterMonth, monthOfDate, shiftMonth } from "../../utils/receipt-date-filter";

const MONTH_LABELS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

const SHORT_MONTH_LABELS = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];

// Arbitrary, purely so the year pager is not an infinite hole.
const MIN_YEAR = 1970;
const MAX_YEAR_OFFSET = 5;

/**
 * A month picker shaped as a stepper: previous / next arrows either side of a
 * label that opens a year pager and a month grid, plus This month / Last month
 * / All time shortcuts.
 *
 * Deliberately presentational — it emits months and knows nothing about what a
 * caller does with them.
 */
@Component({
  selector: "app-month-stepper",
  templateUrl: "./month-stepper.component.html",
  styleUrls: ["./month-stepper.component.scss"],
  standalone: true,
  imports: [ButtonModule, MatMenuModule],
})
export class MonthStepperComponent {
  /** The month currently shown, or null when the caller is showing something else. */
  public readonly value = input<FilterMonth | null>(null);

  /**
   * Rendered on the label. The caller owns the wording because only it knows
   * whether "no month" means nothing is set or something it cannot express.
   */
  public readonly label = input.required<string>();

  public readonly monthSelected = output<FilterMonth>();

  public readonly allTimeSelected = output<void>();

  public readonly shortMonthLabels = SHORT_MONTH_LABELS;

  /**
   * The year the grid is paging through. Follows the bound month but stays
   * overridable, so paging the year does not change the filter and a reset
   * underneath the open panel still re-seeds it.
   */
  public readonly pagedYear = linkedSignal(
    () => this.value()?.year ?? new Date().getFullYear()
  );

  public readonly canPageBack = computed(() => this.pagedYear() > MIN_YEAR);

  public readonly canPageForward = computed(
    () => this.pagedYear() < new Date().getFullYear() + MAX_YEAR_OFFSET
  );

  public monthLabel(month: number): string {
    return MONTH_LABELS[month];
  }

  public isSelected(month: number): boolean {
    const value = this.value();

    return !!value && value.year === this.pagedYear() && value.month === month;
  }

  /**
   * Steps by whole months. With nothing selected it steps from the current
   * month, so the first press always lands somewhere useful rather than
   * no-oping.
   */
  public step(delta: number): void {
    this.monthSelected.emit(shiftMonth(this.value() ?? monthOfDate(new Date()), delta));
  }

  public pageYear(delta: number): void {
    this.pagedYear.update((year) => year + delta);
  }

  public pickMonth(month: number): void {
    this.monthSelected.emit({ year: this.pagedYear(), month });
  }

  /** `0` is this month, `-1` last month. */
  public pickRelativeMonth(delta: number): void {
    this.monthSelected.emit(shiftMonth(monthOfDate(new Date()), delta));
  }
}
