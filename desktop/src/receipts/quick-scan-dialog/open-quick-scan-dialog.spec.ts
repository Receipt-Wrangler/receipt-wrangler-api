import { MatDialog } from "@angular/material/dialog";
import { of } from "rxjs";
import { DEFAULT_DIALOG_CONFIG } from "../../constants";
import { openQuickScanDialog } from "./open-quick-scan-dialog";
import { QuickScanDialogComponent } from "./quick-scan-dialog.component";

describe("openQuickScanDialog", () => {
  // A plain function over MatDialog — no TestBed needed, and a double keeps the
  // real dialog (and its whole component tree) out of the test.
  function dialogDouble(afterClosed = of(undefined)) {
    return {
      open: jest.fn().mockReturnValue({ afterClosed: () => afterClosed }),
    } as unknown as MatDialog;
  }

  it("opens the quick scan dialog with the shared dialog config", () => {
    const matDialog = dialogDouble();

    openQuickScanDialog(matDialog);

    expect(matDialog.open).toHaveBeenCalledTimes(1);
    expect(matDialog.open).toHaveBeenCalledWith(
      QuickScanDialogComponent,
      DEFAULT_DIALOG_CONFIG
    );
  });

  // Both call sites (the icon button and the receipts-table overflow item)
  // refetch off this observable, so it has to be the dialog's own afterClosed.
  it("returns the dialog's afterClosed observable", (done) => {
    const afterClosed = of("closed");
    const matDialog = dialogDouble(afterClosed as any);

    const result = openQuickScanDialog(matDialog);

    expect(result).toBe(afterClosed);
    result.subscribe((value) => {
      expect(value).toEqual("closed");
      done();
    });
  });
});
