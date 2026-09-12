import { provideZonelessChangeDetection } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { provideRouter } from "@angular/router";
import { FilterMonth } from "../../utils/receipt-date-filter";
import { MonthStepperComponent } from "./month-stepper.component";

/**
 * Expectations are derived from the real clock rather than a frozen one: jest's
 * fake timers stall `fixture.whenStable()` under zoneless change detection.
 */
function monthRelativeToToday(delta: number): FilterMonth {
  const today = new Date();
  const shifted = new Date(today.getFullYear(), today.getMonth() + delta, 1);

  return { year: shifted.getFullYear(), month: shifted.getMonth() };
}

describe("MonthStepperComponent", () => {
  let component: MonthStepperComponent;
  let fixture: ComponentFixture<MonthStepperComponent>;
  let emitted: FilterMonth[];
  let allTimeEmissions: number;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MonthStepperComponent, NoopAnimationsModule],
      // app-button binds [routerLink], so it needs a router context.
      providers: [provideZonelessChangeDetection(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(MonthStepperComponent);
    component = fixture.componentInstance;
    fixture.componentRef.setInput("label", "All time");

    emitted = [];
    allTimeEmissions = 0;
    component.monthSelected.subscribe((month) => emitted.push(month));
    component.allTimeSelected.subscribe(() => (allTimeEmissions += 1));

    await fixture.whenStable();
  });

  const setMonth = async (month: FilterMonth | null) => {
    fixture.componentRef.setInput("value", month);
    await fixture.whenStable();
  };

  it("renders the label it is given", async () => {
    fixture.componentRef.setInput("label", "September 2026");
    await fixture.whenStable();

    expect(
      fixture.nativeElement.querySelector('[data-testid="receipts-month-label"]').textContent
    ).toContain("September 2026");
  });

  describe("stepping", () => {
    it("steps from the bound month", async () => {
      await setMonth({ year: 2026, month: 8 });

      component.step(1);
      component.step(-1);

      expect(emitted).toEqual([
        { year: 2026, month: 9 },
        { year: 2026, month: 7 },
      ]);
    });

    // Without a month the arrows would otherwise no-op, so they step from today
    // and the first press always lands somewhere useful.
    it("steps from the current month when nothing is selected", () => {
      component.step(-1);
      component.step(1);

      expect(emitted).toEqual([monthRelativeToToday(-1), monthRelativeToToday(1)]);
    });

    it("rolls the year at both boundaries", async () => {
      await setMonth({ year: 2026, month: 11 });
      component.step(1);

      await setMonth({ year: 2026, month: 0 });
      component.step(-1);

      expect(emitted).toEqual([
        { year: 2027, month: 0 },
        { year: 2025, month: 11 },
      ]);
    });
  });

  describe("the panel", () => {
    it("seeds its year from the bound month and falls back to this year", async () => {
      expect(component.pagedYear()).toBe(new Date().getFullYear());

      await setMonth({ year: 2023, month: 4 });

      expect(component.pagedYear()).toBe(2023);
    });

    // Paging is a view concern — it must not write a filter.
    it("pages the year without emitting", async () => {
      await setMonth({ year: 2026, month: 8 });

      component.pageYear(-1);

      expect(component.pagedYear()).toBe(2025);
      expect(emitted).toEqual([]);
    });

    it("picks a month out of the paged year", async () => {
      await setMonth({ year: 2026, month: 8 });

      component.pageYear(1);
      component.pickMonth(0);

      expect(emitted).toEqual([{ year: 2027, month: 0 }]);
    });

    it("marks only the bound month as selected, and only in its own year", async () => {
      await setMonth({ year: 2026, month: 8 });

      expect(component.isSelected(8)).toBe(true);
      expect(component.isSelected(7)).toBe(false);

      component.pageYear(1);

      expect(component.isSelected(8)).toBe(false);
    });

    it("emits the right months for the shortcuts", () => {
      component.pickRelativeMonth(0);
      component.pickRelativeMonth(-1);

      expect(emitted).toEqual([monthRelativeToToday(0), monthRelativeToToday(-1)]);
    });

    it("emits all time separately", () => {
      component.pickAllTime();

      expect(allTimeEmissions).toBe(1);
      expect(emitted).toEqual([]);
    });

    it("opens and closes", () => {
      expect(component.panelOpen()).toBe(false);

      component.togglePanel();
      expect(component.panelOpen()).toBe(true);

      component.togglePanel();
      expect(component.panelOpen()).toBe(false);
    });

    it("closes once a month or a shortcut is picked", () => {
      for (const pick of [
        () => component.pickMonth(3),
        () => component.pickRelativeMonth(0),
        () => component.pickAllTime(),
      ]) {
        component.togglePanel();
        pick();
        expect(component.panelOpen()).toBe(false);
      }
    });

    // Paging is a view concern. Under the old mat-menu this needed a
    // stopPropagation hack because the menu closed on any click inside it.
    it("stays open while paging the year", () => {
      component.togglePanel();

      component.pageYear(-1);

      expect(component.panelOpen()).toBe(true);
      expect(emitted).toEqual([]);
    });

    it("bounds the year pager", async () => {
      await setMonth({ year: 1970, month: 0 });
      expect(component.canPageBack()).toBe(false);

      await setMonth({ year: new Date().getFullYear() + 5, month: 0 });
      expect(component.canPageForward()).toBe(false);

      await setMonth({ year: new Date().getFullYear(), month: 0 });
      expect(component.canPageBack()).toBe(true);
      expect(component.canPageForward()).toBe(true);
    });
  });
});
