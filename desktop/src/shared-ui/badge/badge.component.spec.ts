import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BadgeComponent, BadgeTone } from "./badge.component";

describe("BadgeComponent", () => {
  let component: BadgeComponent;
  let fixture: ComponentFixture<BadgeComponent>;

  const render = (text: string, tone?: BadgeTone): HTMLElement => {
    fixture.componentRef.setInput("text", text);
    if (tone) {
      fixture.componentRef.setInput("tone", tone);
    }
    fixture.detectChanges();

    return fixture.nativeElement.querySelector(".rw-badge") as HTMLElement;
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BadgeComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(BadgeComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("renders its text", () => {
    expect(render("Custom").textContent?.trim()).toEqual("Custom");
  });

  // Purple marks a custom field, which is what the badge is for nearly everywhere
  // it appears - hard-coding it per call site is how the two implementations this
  // replaced ended up different colours.
  it("defaults to the custom-field tone", () => {
    expect(render("Custom").classList).toContain("rw-badge--purple");
  });

  it("applies an explicit tone", () => {
    for (const tone of ["slate", "blue", "green"] as BadgeTone[]) {
      expect(render("Agg", tone).classList).toContain(`rw-badge--${tone}`);
    }
  });

  it("renders exactly one element, with no surrounding text", () => {
    // Both the mat-option and the report row place the badge inside content whose
    // accessible name is its text, so a wrapper or a stray text node would change
    // what those elements are named.
    fixture.componentRef.setInput("text", "Custom");
    fixture.detectChanges();

    expect(fixture.nativeElement.children.length).toEqual(1);
    expect((fixture.nativeElement as HTMLElement).textContent?.trim()).toEqual("Custom");
  });
});
