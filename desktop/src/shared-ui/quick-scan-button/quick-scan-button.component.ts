import { Component, output } from "@angular/core";
import { MatDialog } from "@angular/material/dialog";
import { take, tap } from "rxjs";
import { openQuickScanDialog } from "../../receipts/quick-scan-dialog/open-quick-scan-dialog";

@Component({
    selector: "app-quick-scan-button",
    templateUrl: "./quick-scan-button.component.html",
    styleUrls: ["./quick-scan-button.component.scss"],
    standalone: false
})
export class QuickScanButtonComponent {
  public readonly afterClosed = output<void>();

  constructor(private matDialog: MatDialog) {}

  public showQuickScanDialog(): void {
    openQuickScanDialog(this.matDialog)
      .pipe(
        take(1),
        tap(() => {
          this.afterClosed.emit();
        })
      )
      .subscribe();
  }
}
