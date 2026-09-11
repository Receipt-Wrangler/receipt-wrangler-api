import { CUSTOM_ELEMENTS_SCHEMA, provideZonelessChangeDetection } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatSelectModule } from '@angular/material/select';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { BadgeComponent } from '../../shared-ui/badge/badge.component';
import { OptionDisplayPipe } from '../pipes/option-display.pipe';
import { ReadonlyValuePipe } from '../pipes/readonly-value.pipe';
import { SelectComponent } from './select.component';

describe('SelectComponent', () => {
  let component: SelectComponent;
  let fixture: ComponentFixture<SelectComponent>;

  const options = [
    { value: 'category', displayValue: 'Category', badge: '' },
    { value: 'custom_7', displayValue: 'HST', badge: 'Custom' },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [SelectComponent, OptionDisplayPipe, ReadonlyValuePipe],
      imports: [MatSelectModule, ReactiveFormsModule, NoopAnimationsModule, BadgeComponent],
      providers: [provideZonelessChangeDetection()],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(SelectComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // --- optionBadgeKey -----------------------------------------------------

  function configure(optionBadgeKey: string, value: string | null): void {
    component.inputFormControl = new FormControl(value);
    component.optionValueKey = 'value';
    fixture.componentRef.setInput('options', options);
    fixture.componentRef.setInput('optionDisplayKey', 'displayValue');
    fixture.componentRef.setInput('optionBadgeKey', optionBadgeKey);
  }

  it('reads an option badge only when optionBadgeKey names one', () => {
    fixture.componentRef.setInput('optionBadgeKey', 'badge');
    expect(component.badgeFor(options[1])).toBe('Custom');
    expect(component.badgeFor(options[0])).toBe('');
    expect(component.badgeFor(undefined)).toBe('');

    // Unset (every existing call site) means no badge at all.
    fixture.componentRef.setInput('optionBadgeKey', '');
    expect(component.badgeFor(options[1])).toBe('');
  });

  it('resolves the selected option by its value key, or -1 when absent', () => {
    configure('badge', 'custom_7');
    expect(component.selectedIndex()).toBe(1);

    component.inputFormControl.setValue('nope');
    expect(component.selectedIndex()).toBe(-1);
  });

  // Without an optionValueKey the control holds the option OBJECT itself, so the
  // match is by reference — the other half of selectedIndex()'s branch.
  it('resolves the selection by identity when there is no value key', () => {
    component.inputFormControl = new FormControl(options[1]);
    component.optionValueKey = '';
    fixture.componentRef.setInput('options', options);
    fixture.componentRef.setInput('optionDisplayKey', 'displayValue');
    fixture.componentRef.setInput('optionBadgeKey', 'badge');

    expect(component.selectedIndex()).toBe(1);

    // An equal-looking copy is a different reference, so it does not match —
    // which is the same rule mat-select itself applies without a compareWith.
    component.inputFormControl.setValue({ ...options[1] });
    expect(component.selectedIndex()).toBe(-1);

    component.inputFormControl.setValue('nope');
    expect(component.selectedIndex()).toBe(-1);
  });

  // The closed select renders the selected option's text content, so a badged
  // option would otherwise read "HSTCustom". A badged select draws its own
  // trigger to keep the two apart.
  it('draws the badge in the trigger for a badged selection', async () => {
    configure('badge', 'custom_7');
    await fixture.whenStable();

    const badges = fixture.nativeElement.querySelectorAll('app-badge');
    expect(badges.length).toBe(1);
    expect(badges[0].textContent.trim()).toBe('Custom');
    expect(fixture.nativeElement.textContent).toContain('HST');
  });

  it('draws no badge for an unbadged selection, or when badges are off', async () => {
    configure('badge', 'category');
    await fixture.whenStable();
    expect(fixture.nativeElement.querySelectorAll('app-badge').length).toBe(0);

    configure('', 'custom_7');
    await fixture.whenStable();
    expect(fixture.nativeElement.querySelectorAll('app-badge').length).toBe(0);
  });
});
