import { MatDialog } from "@angular/material/dialog";
import { Observable } from "rxjs";
import { DEFAULT_DIALOG_CONFIG } from "../../constants";
import { QuickScanDialogComponent } from "./quick-scan-dialog.component";

/**
 * Opens the Quick Scan dialog and emits once it closes.
 *
 * Shared by the `app-quick-scan-button` icon button and the receipts table's
 * overflow-menu item, which cannot reuse that component: `MatMenu` collects its
 * items with a **content** query, and a `mat-menu-item` rendered inside another
 * component's template is a view child of that component, so the menu would
 * never see it and keyboard navigation would skip the entry.
 */
export function openQuickScanDialog(matDialog: MatDialog): Observable<unknown> {
  return matDialog.open(QuickScanDialogComponent, DEFAULT_DIALOG_CONFIG).afterClosed();
}
