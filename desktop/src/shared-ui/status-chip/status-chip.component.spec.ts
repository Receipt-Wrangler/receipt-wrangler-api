import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StatusChipComponent } from './status-chip.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { MatChipsModule } from '@angular/material/chips';
import { PipesModule } from 'src/pipes/pipes.module';
import { ReceiptStatus } from 'src/open-api';

describe('StatusChipComponent', () => {
  let component: StatusChipComponent;
  let fixture: ComponentFixture<StatusChipComponent>;

  const chipClasses = (): string[] =>
    Array.from(
      fixture.nativeElement.querySelector('mat-chip')?.classList ?? []
    );

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ StatusChipComponent ],
      imports: [MatChipsModule, PipesModule],
      schemas: [CUSTOM_ELEMENTS_SCHEMA]
    })
    .compileComponents();

    fixture = TestBed.createComponent(StatusChipComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  // The status -> class map is the entire color layer, so it is asserted here rather
  // than left to the stylesheet. DECLINED owns red; NEEDS_ATTENTION is amber.
  const statusClassCases: { status: ReceiptStatus; expected: string }[] = [
    { status: ReceiptStatus.NeedsAttention, expected: 'needs-attention' },
    { status: ReceiptStatus.Declined, expected: 'declined' },
    { status: ReceiptStatus.Open, expected: 'open' },
    { status: ReceiptStatus.Resolved, expected: 'resolved' },
  ];

  statusClassCases.forEach(({ status, expected }) => {
    it(`should apply the ${expected} class for ${status}`, async () => {
      fixture.componentRef.setInput('status', status);
      await fixture.whenStable();

      const classes = chipClasses();
      expect(classes).toContain(expected);

      statusClassCases
        .map((c) => c.expected)
        .filter((c) => c !== expected)
        .forEach((other) => expect(classes).not.toContain(other));
    });
  });

  it('should apply no status class for DRAFT', async () => {
    fixture.componentRef.setInput('status', ReceiptStatus.Draft);
    await fixture.whenStable();

    const classes = chipClasses();
    statusClassCases.forEach(({ expected }) =>
      expect(classes).not.toContain(expected)
    );
  });

  // The activity list drives the chip with customStatusColor for system-task status.
  // 'red' must keep meaning red now that NEEDS_ATTENTION is amber.
  it('should map the red custom status color onto the declined class', async () => {
    fixture.componentRef.setInput('customStatusColor', 'red');
    await fixture.whenStable();

    const classes = chipClasses();
    expect(classes).toContain('declined');
    expect(classes).not.toContain('needs-attention');
  });

  it('should map the green custom status color onto the resolved class', async () => {
    fixture.componentRef.setInput('customStatusColor', 'green');
    await fixture.whenStable();

    expect(chipClasses()).toContain('resolved');
  });
});
