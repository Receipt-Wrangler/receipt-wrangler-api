import { CdkDragDrop, moveItemInArray } from '@angular/cdk/drag-drop';
import { Component, Inject, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { DEFAULT_RECEIPT_TABLE_COLUMNS, ReceiptTableColumnConfig } from '../../interfaces';
import { CUSTOM_FIELD_BADGE } from '../../shared-ui/badge/badge.component';
import { CustomField } from '../../open-api';
import { columnDisplayName, mergeCustomFieldColumns, parseCustomFieldColumnDef } from '../../utils';

interface ColumnConfigItem extends ReceiptTableColumnConfig {
  displayName: string;
  isCustom: boolean;
}

@Component({
  selector: 'app-column-configuration-dialog',
  templateUrl: './column-configuration-dialog.component.html',
  styleUrls: ['./column-configuration-dialog.component.scss'],
  standalone: false
})
export class ColumnConfigurationDialogComponent implements OnInit {
  public columns: ColumnConfigItem[] = [];

  protected readonly customFieldBadge = CUSTOM_FIELD_BADGE;

  constructor(
    private dialogRef: MatDialogRef<ColumnConfigurationDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: {
      currentColumns?: ReceiptTableColumnConfig[];
      customFields?: CustomField[];
    }
  ) {}

  ngOnInit(): void {
    this.initializeColumns();
  }

  private get customFields(): CustomField[] {
    return this.data.customFields ?? [];
  }

  private initializeColumns(): void {
    this.setColumns(this.data.currentColumns ?? DEFAULT_RECEIPT_TABLE_COLUMNS);
  }

  /**
   * Reconciles against the custom field catalog before display, so a field
   * deleted since the configuration was persisted disappears from the list and a
   * newly created one shows up (unchecked) without the user resetting anything.
   *
   * `mergeCustomFieldColumns` copies, which matters: the caller hands us the NGXS
   * snapshot, and that is deep-frozen in dev mode.
   */
  private setColumns(columns: ReceiptTableColumnConfig[]): void {
    this.columns = mergeCustomFieldColumns(columns, this.customFields).map(
      (col) => ({
        ...col,
        displayName: columnDisplayName(col.matColumnDef, this.customFields),
        isCustom: parseCustomFieldColumnDef(col.matColumnDef) !== undefined,
      })
    );
  }

  public toggleColumnVisibility(column: ColumnConfigItem): void {
    column.visible = !column.visible;
  }

  public drop(event: CdkDragDrop<ColumnConfigItem[]>): void {
    moveItemInArray(this.columns, event.previousIndex, event.currentIndex);

    this.columns.forEach((column, index) => {
      column.order = index;
    });
  }

  /**
   * Restores the built-in defaults. Custom fields stay listed — they are only
   * ever hidden by default, so dropping them here would just make them
   * reappear on the next open.
   */
  public resetToDefaults(): void {
    this.setColumns(DEFAULT_RECEIPT_TABLE_COLUMNS);
  }

  public saveConfiguration(): void {
    const result: ReceiptTableColumnConfig[] = this.columns.map(
      ({ displayName, isCustom, ...col }) => col
    );
    this.dialogRef.close(result);
  }

  public cancel(): void {
    this.dialogRef.close(null);
  }

  public get visibleColumnsCount(): number {
    return this.columns.filter(col => col.visible).length;
  }

  public canToggleOff(column: ColumnConfigItem): boolean {
    return this.visibleColumnsCount > 1 || !column.visible;
  }
}
