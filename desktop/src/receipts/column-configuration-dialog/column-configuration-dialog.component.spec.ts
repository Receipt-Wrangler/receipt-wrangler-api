import { CdkDragDrop } from "@angular/cdk/drag-drop";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MAT_DIALOG_DATA, MatDialogRef } from "@angular/material/dialog";
import { provideRouter } from "@angular/router";
import { NgxsModule } from "@ngxs/store";
import { DEFAULT_RECEIPT_TABLE_COLUMNS, ReceiptTableColumnConfig } from "../../interfaces";
import { CustomField, CustomFieldType } from "../../open-api";
import { SharedUiModule } from "../../shared-ui/shared-ui.module";
import { ReceiptTableState } from "../../store/receipt-table.state";
import { RECEIPT_COLUMN_DISPLAY_NAMES } from "../../utils";
import { ColumnConfigurationDialogComponent } from "./column-configuration-dialog.component";

describe("ColumnConfigurationDialogComponent", () => {
  let component: ColumnConfigurationDialogComponent;
  let fixture: ComponentFixture<ColumnConfigurationDialogComponent>;
  let mockDialogRef: jest.Mocked<MatDialogRef<ColumnConfigurationDialogComponent>>;
  let mockDialogData: {
    currentColumns?: ReceiptTableColumnConfig[];
    customFields?: CustomField[];
  };

  const customField = (id: number, name: string): CustomField =>
    ({ id, name, type: CustomFieldType.Text } as CustomField);

  /** The configuration a real session persists: every built-in column, in order. */
  const allBuiltInColumns = (): ReceiptTableColumnConfig[] =>
    DEFAULT_RECEIPT_TABLE_COLUMNS.map((column) => ({ ...column }));

  beforeEach(async () => {
    // Create spy for MatDialogRef
    mockDialogRef = { close: jest.fn() } as unknown as jest.Mocked<MatDialogRef<ColumnConfigurationDialogComponent>>;

    // Initialize mock dialog data
    mockDialogData = {
      currentColumns: [
        { matColumnDef: "created_at", visible: true, order: 0 },
        { matColumnDef: "name", visible: true, order: 1 },
        { matColumnDef: "amount", visible: false, order: 2 }
      ]
    };

    await TestBed.configureTestingModule({
      declarations: [ColumnConfigurationDialogComponent],
      imports: [SharedUiModule, NgxsModule.forRoot([ReceiptTableState])],
      providers: [
        { provide: MatDialogRef, useValue: mockDialogRef },
        { provide: MAT_DIALOG_DATA, useValue: mockDialogData },
        // The shared dialog shell pulls in router-aware buttons, so rendering the
        // template (rather than only exercising the class) needs a router.
        provideRouter([])
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ColumnConfigurationDialogComponent);
    component = fixture.componentInstance;
  });

  describe("Component Initialization", () => {
    it("should create", () => {
      expect(component).toBeTruthy();
    });

    it("should have correct column display name mapping", () => {
      const expectedMapping = {
        "created_at": "Added At",
        "date": "Receipt Date",
        "name": "Name",
        "paid_by_user_id": "Paid By",
        "amount": "Amount",
        "categories": "Categories",
        "tags": "Tags",
        "status": "Status",
        "resolved_date": "Resolved Date"
      };

      expect(RECEIPT_COLUMN_DISPLAY_NAMES).toEqual(expectedMapping);
    });

    it("should initialize columns on ngOnInit", () => {
      jest.spyOn(component as any, "initializeColumns");

      component.ngOnInit();

      expect(component["initializeColumns"]).toHaveBeenCalled();
    });
  });

  describe("initializeColumns", () => {
    it("should initialize columns from provided data", () => {
      component.ngOnInit();

      expect(component.columns[0]).toEqual({
        matColumnDef: "created_at",
        visible: true,
        order: 0,
        displayName: "Added At",
        isCustom: false
      });
      expect(component.columns[1]).toEqual({
        matColumnDef: "name",
        visible: true,
        order: 1,
        displayName: "Name",
        isCustom: false
      });
      expect(component.columns[2]).toEqual({
        matColumnDef: "amount",
        visible: false,
        order: 2,
        displayName: "Amount",
        isCustom: false
      });
    });

    it("should sort columns by order before mapping", () => {
      mockDialogData.currentColumns = [
        { matColumnDef: "amount", visible: true, order: 2 },
        { matColumnDef: "created_at", visible: true, order: 0 },
        { matColumnDef: "name", visible: true, order: 1 }
      ];

      component.ngOnInit();

      expect(component.columns[0].matColumnDef).toBe("created_at");
      expect(component.columns[1].matColumnDef).toBe("name");
      expect(component.columns[2].matColumnDef).toBe("amount");
    });

    it("does not mutate a frozen (store-sourced) currentColumns array", () => {
      // The opener passes the NGXS snapshot, which is deep-frozen in dev mode.
      // Sorting it in place threw "TypeError: 0 is read-only"; we must copy first.
      mockDialogData.currentColumns = Object.freeze([
        Object.freeze({ matColumnDef: "amount", visible: true, order: 2 }),
        Object.freeze({ matColumnDef: "created_at", visible: true, order: 0 }),
        Object.freeze({ matColumnDef: "name", visible: true, order: 1 })
      ]) as ReceiptTableColumnConfig[];

      expect(() => component.ngOnInit()).not.toThrow();
      expect(component.columns.slice(0, 3).map(col => col.matColumnDef)).toEqual([
        "created_at",
        "name",
        "amount"
      ]);
    });

    it("should use DEFAULT_RECEIPT_TABLE_COLUMNS when no currentColumns provided", () => {
      mockDialogData.currentColumns = undefined;

      component.ngOnInit();

      expect(component.columns.length).toEqual(DEFAULT_RECEIPT_TABLE_COLUMNS.length);
      expect(component.columns[0].matColumnDef).toBe(DEFAULT_RECEIPT_TABLE_COLUMNS[0].matColumnDef);
    });

    it("drops a persisted column that no longer resolves", () => {
      // A column can outlive its definition: the configuration is persisted per
      // browser, so a released build that removes a column - or a deleted custom
      // field - leaves an id behind that mat-table would reject.
      mockDialogData.currentColumns = [
        ...allBuiltInColumns(),
        { matColumnDef: "unknown_column", visible: true, order: 9 }
      ];

      component.ngOnInit();

      expect(
        component.columns.some(col => col.matColumnDef === "unknown_column")
      ).toBe(false);
    });

    it("lists every custom field after the built-in columns, unchecked", () => {
      mockDialogData.currentColumns = allBuiltInColumns();
      mockDialogData.customFields = [customField(7, "Vendor"), customField(2, "Approver")];

      component.ngOnInit();

      // Sorted by name, so the ids are deliberately out of order above.
      expect(component.columns.slice(DEFAULT_RECEIPT_TABLE_COLUMNS.length)).toEqual([
        { matColumnDef: "custom_2", visible: false, order: 9, displayName: "Approver", isCustom: true },
        { matColumnDef: "custom_7", visible: false, order: 10, displayName: "Vendor", isCustom: true }
      ]);
    });

    it("drops a custom field column whose field no longer exists", () => {
      mockDialogData.currentColumns = [
        ...allBuiltInColumns(),
        { matColumnDef: "custom_99", visible: true, order: 9 }
      ];
      mockDialogData.customFields = [customField(7, "Vendor")];

      component.ngOnInit();

      expect(component.columns.map(col => col.matColumnDef)).not.toContain("custom_99");
      expect(component.columns.map(col => col.matColumnDef)).toContain("custom_7");
    });

    it("keeps a custom field column the user moved above a built-in one", () => {
      mockDialogData.currentColumns = [
        { matColumnDef: "custom_7", visible: true, order: 0 },
        ...allBuiltInColumns().map((column, index) => ({ ...column, order: index + 1 }))
      ];
      mockDialogData.customFields = [customField(7, "Vendor")];

      component.ngOnInit();

      expect(component.columns[0]).toEqual({
        matColumnDef: "custom_7",
        visible: true,
        order: 0,
        displayName: "Vendor",
        isCustom: true
      });
    });
  });

  describe("custom field badge", () => {
    it("badges a custom field row and no built-in one", () => {
      mockDialogData.customFields = [customField(7, "Vendor")];
      component.ngOnInit();
      fixture.detectChanges();

      const rows = fixture.nativeElement.querySelectorAll(".column-item");
      const badged = Array.from(rows as ArrayLike<HTMLElement>).filter((row) =>
        row.querySelector("app-badge")
      );

      expect(badged.length).toEqual(1);
      expect(badged[0].textContent).toContain("Vendor");
      expect(badged[0].querySelector("app-badge")?.textContent?.trim()).toEqual("Custom");
    });

    // Inside the label the badge would join the checkbox's accessible name, and
    // every locator picking a column by its field name would stop resolving.
    it("keeps the badge out of the checkbox label", () => {
      mockDialogData.customFields = [customField(7, "Vendor")];
      component.ngOnInit();
      fixture.detectChanges();

      const checkbox = fixture.nativeElement.querySelector(
        ".column-item:last-child mat-checkbox"
      ) as HTMLElement;

      expect(checkbox.textContent?.trim()).toEqual("Vendor");
    });
  });

  describe("toggleColumnVisibility", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should toggle visible column to invisible", () => {
      const visibleColumn = component.columns.find(col => col.visible)!;
      const originalVisibility = visibleColumn.visible;

      component.toggleColumnVisibility(visibleColumn);

      expect(visibleColumn.visible).toBe(!originalVisibility);
    });

    it("should toggle invisible column to visible", () => {
      const invisibleColumn = component.columns.find(col => !col.visible)!;
      const originalVisibility = invisibleColumn.visible;

      component.toggleColumnVisibility(invisibleColumn);

      expect(invisibleColumn.visible).toBe(!originalVisibility);
    });

    it("should not affect other columns when toggling one column", () => {
      const targetColumn = component.columns[0];
      const otherColumns = component.columns.slice(1);
      const originalStates = otherColumns.map(col => col.visible);

      component.toggleColumnVisibility(targetColumn);

      otherColumns.forEach((col, index) => {
        expect(col.visible).toBe(originalStates[index]);
      });
    });
  });

  describe("drop", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should reorder columns when dropped", () => {
      const originalOrder = [...component.columns];
      const mockEvent: CdkDragDrop<any[]> = {
        previousIndex: 0,
        currentIndex: 2,
        item: null as any,
        container: null as any,
        previousContainer: null as any,
        isPointerOverContainer: false,
        distance: { x: 0, y: 0 },
        dropPoint: { x: 0, y: 0 },
        event: null as any
      };

      component.drop(mockEvent);

      expect(component.columns[0]).toEqual({ ...originalOrder[1], order: 0 });
      expect(component.columns[1]).toEqual({ ...originalOrder[2], order: 1 });
      expect(component.columns[2]).toEqual({ ...originalOrder[0], order: 2 });
    });

    it("should update order property for all columns after drop", () => {
      const mockEvent: CdkDragDrop<any[]> = {
        previousIndex: 1,
        currentIndex: 0,
        item: null as any,
        container: null as any,
        previousContainer: null as any,
        isPointerOverContainer: false,
        distance: { x: 0, y: 0 },
        dropPoint: { x: 0, y: 0 },
        event: null as any
      };

      component.drop(mockEvent);

      component.columns.forEach((column, index) => {
        expect(column.order).toBe(index);
      });
    });

    it("should handle drop from last to first position", () => {
      const originalLastColumn = component.columns[component.columns.length - 1];
      const mockEvent: CdkDragDrop<any[]> = {
        previousIndex: component.columns.length - 1,
        currentIndex: 0,
        item: null as any,
        container: null as any,
        previousContainer: null as any,
        isPointerOverContainer: false,
        distance: { x: 0, y: 0 },
        dropPoint: { x: 0, y: 0 },
        event: null as any
      };

      component.drop(mockEvent);

      expect(component.columns[0].matColumnDef).toBe(originalLastColumn.matColumnDef);
      expect(component.columns[0].order).toBe(0);
    });
  });

  describe("resetToDefaults", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should reset columns to DEFAULT_RECEIPT_TABLE_COLUMNS", () => {
      // Modify columns first
      component.columns[0].visible = false;
      component.columns[1].order = 999;

      component.resetToDefaults();

      expect(component.columns.length).toEqual(DEFAULT_RECEIPT_TABLE_COLUMNS.length);

      DEFAULT_RECEIPT_TABLE_COLUMNS.forEach((defaultCol, index) => {
        expect(component.columns[index].matColumnDef).toBe(defaultCol.matColumnDef);
        expect(component.columns[index].visible).toBe(defaultCol.visible);
        expect(component.columns[index].order).toBe(defaultCol.order);
      });
    });

    it("should add display names to default columns", () => {
      component.resetToDefaults();

      component.columns.forEach(column => {
        const expectedDisplayName = RECEIPT_COLUMN_DISPLAY_NAMES[column.matColumnDef] || column.matColumnDef;
        expect(column.displayName).toBe(expectedDisplayName);
      });
    });

    it("keeps custom fields listed and unchecked after a reset", () => {
      mockDialogData.customFields = [customField(7, "Vendor")];
      component.ngOnInit();
      component.columns.find(col => col.matColumnDef === "custom_7")!.visible = true;

      component.resetToDefaults();

      // Dropping them instead would just make them reappear on the next open.
      expect(component.columns[component.columns.length - 1]).toEqual({
        matColumnDef: "custom_7",
        visible: false,
        order: DEFAULT_RECEIPT_TABLE_COLUMNS.length,
        displayName: "Vendor",
        isCustom: true
      });
    });

    it("should preserve displayName mapping when resetting", () => {
      component.resetToDefaults();

      const createdAtColumn = component.columns.find(col => col.matColumnDef === "created_at");
      expect(createdAtColumn?.displayName).toBe("Added At");

      const nameColumn = component.columns.find(col => col.matColumnDef === "name");
      expect(nameColumn?.displayName).toBe("Name");
    });
  });

  describe("saveConfiguration", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should close dialog with configuration result", () => {
      component.saveConfiguration();

      expect(mockDialogRef.close).toHaveBeenCalledWith(
        expect.arrayContaining([
          expect.objectContaining({
            matColumnDef: expect.any(String),
            visible: expect.any(Boolean),
            order: expect.any(Number)
          })
        ])
      );
    });

    it("should remove view-only properties from result", () => {
      // The saved payload is persisted to localStorage and read back as the
      // column config, so a display-only field riding along would be stored and
      // then compared against a freshly derived one.
      component.saveConfiguration();

      const result = mockDialogRef.close.mock.calls[mockDialogRef.close.mock.calls.length - 1][0];
      result.forEach((col: any) => {
        expect(col.displayName).toBeUndefined();
        expect(col.isCustom).toBeUndefined();
        expect(col.matColumnDef).toBeDefined();
        expect(col.visible).toBeDefined();
        expect(col.order).toBeDefined();
      });
    });

    it("should preserve all other column properties in result", () => {
      component.saveConfiguration();

      const result = mockDialogRef.close.mock.calls[mockDialogRef.close.mock.calls.length - 1][0];
      expect(result.length).toEqual(component.columns.length);

      result.forEach((resultCol: ReceiptTableColumnConfig, index: number) => {
        const originalCol = component.columns[index];
        expect(resultCol.matColumnDef).toBe(originalCol.matColumnDef);
        expect(resultCol.visible).toBe(originalCol.visible);
        expect(resultCol.order).toBe(originalCol.order);
      });
    });
  });

  describe("cancel", () => {
    it("should close dialog with null result", () => {
      component.cancel();

      expect(mockDialogRef.close).toHaveBeenCalledWith(null);
    });
  });

  describe("visibleColumnsCount", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should return correct count of visible columns", () => {
      const visibleCount = component.columns.filter(col => col.visible).length;

      expect(component.visibleColumnsCount).toBe(visibleCount);
    });

    it("should update when column visibility changes", () => {
      const initialCount = component.visibleColumnsCount;
      const invisibleColumn = component.columns.find(col => !col.visible);

      if (invisibleColumn) {
        component.toggleColumnVisibility(invisibleColumn);
        expect(component.visibleColumnsCount).toBe(initialCount + 1);
      }
    });

    it("should return 0 when all columns are invisible", () => {
      component.columns.forEach(col => col.visible = false);

      expect(component.visibleColumnsCount).toBe(0);
    });

    it("should return total count when all columns are visible", () => {
      component.columns.forEach(col => col.visible = true);

      expect(component.visibleColumnsCount).toBe(component.columns.length);
    });
  });

  describe("canToggleOff", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("should return true when more than one column is visible", () => {
      // Ensure at least 2 columns are visible
      component.columns[0].visible = true;
      component.columns[1].visible = true;
      component.columns[2].visible = false;

      expect(component.canToggleOff(component.columns[0])).toBe(true);
    });

    it("should return false when only one column is visible and trying to toggle it off", () => {
      // Set only one column as visible
      component.columns.forEach((col, index) => {
        col.visible = index === 0;
      });

      expect(component.canToggleOff(component.columns[0])).toBe(false);
    });

    it("should return true for invisible columns regardless of visible count", () => {
      // Set only one column as visible
      component.columns.forEach((col, index) => {
        col.visible = index === 0;
      });

      const invisibleColumn = component.columns.find(col => !col.visible)!;
      expect(component.canToggleOff(invisibleColumn)).toBe(true);
    });

    it("should return true when trying to toggle off visible column with multiple visible columns", () => {
      component.columns.forEach(col => col.visible = true);

      expect(component.canToggleOff(component.columns[0])).toBe(true);
    });

    it("should handle edge case with no visible columns", () => {
      component.columns.forEach(col => col.visible = false);

      expect(component.canToggleOff(component.columns[0])).toBe(true);
    });
  });

  describe("Component Integration", () => {
    it("should maintain column integrity during multiple operations", () => {
      component.ngOnInit();

      // Toggle visibility
      component.toggleColumnVisibility(component.columns[0]);

      // Drop operation
      const mockEvent: CdkDragDrop<any[]> = {
        previousIndex: 0,
        currentIndex: 1,
        item: null as any,
        container: null as any,
        previousContainer: null as any,
        isPointerOverContainer: false,
        distance: { x: 0, y: 0 },
        dropPoint: { x: 0, y: 0 },
        event: null as any
      };
      component.drop(mockEvent);

      // Verify columns still have required properties
      component.columns.forEach((column, index) => {
        expect(column.matColumnDef).toBeDefined();
        expect(typeof column.visible).toBe("boolean");
        expect(column.order).toBe(index);
        expect(column.displayName).toBeDefined();
      });
    });

    it("heals an empty persisted configuration back to the defaults", () => {
      mockDialogData.currentColumns = [];

      component.ngOnInit();

      expect(component.columns.map(col => col.matColumnDef)).toEqual(
        DEFAULT_RECEIPT_TABLE_COLUMNS.map(col => col.matColumnDef)
      );
    });
  });
});
