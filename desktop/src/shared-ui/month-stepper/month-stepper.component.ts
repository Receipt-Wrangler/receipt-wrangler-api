import { A11yModule } from "@angular/cdk/a11y";
import { ConnectedPosition, OverlayModule } from "@angular/cdk/overlay";
import { Component, computed, input, linkedSignal, output, signal } from "@angular/core";
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
 *
 * The panel is a CDK overlay rather than a `mat-menu`. A menu's key manager
 * only tracks `mat-menu-item`s, and this panel has none — so arrow keys did
 * nothing and, worse, `ListKeyManager` turns Tab into `tabOut`, which `MatMenu`
 * wires to close. That left the grid, the pager and the shortcuts reachable by
 * mouse only. An overlay lets this be what it actually is: a small dialog.
 */
@Component({
  selector: "app-month-stepper",
  templateUrl: "./month-stepper.component.html",
  styleUrls: ["./month-stepper.component.scss"],
  standalone: true,
  imports: [A11yModule, ButtonModule, OverlayModule],
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

  public readonly panelOpen = signal(false);

  /** Anchored under the label, flipping above it when there is no room below. */
  public readonly panelPositions: ConnectedPosition[] = [
    { originX: "end", originY: "bottom", overlayX: "end", overlayY: "top", offsetY: 4 },
    { originX: "end", originY: "top", overlayX: "end", overlayY: "bottom", offsetY: -4 },
  ];

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

  public togglePanel(): void {
    this.panelOpen.update((open) => !open);
  }

  /**
   * Focus returns to the trigger on its own: the focus trap captures the
   * previously focused element and restores it when the overlay is destroyed.
   */
  public closePanel(): void {
    this.panelOpen.set(false);
  }

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

  /** Leaves the panel open — paging is a view concern, not a selection. */
  public pageYear(delta: number): void {
    this.pagedYear.update((year) => year + delta);
  }

  public pickMonth(month: number): void {
    this.monthSelected.emit({ year: this.pagedYear(), month });
    this.closePanel();
  }

  /** `0` is this month, `-1` last month. */
  public pickRelativeMonth(delta: number): void {
    this.monthSelected.emit(shiftMonth(monthOfDate(new Date()), delta));
    this.closePanel();
  }

  public pickAllTime(): void {
    this.allTimeSelected.emit();
    this.closePanel();
  }
}
