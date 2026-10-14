import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MAT_DIALOG_DATA, MatDialogRef } from "@angular/material/dialog";
import { provideRouter } from "@angular/router";
import {
  DeleteAccountDialogComponent,
  DeleteAccountDialogData,
} from "./delete-account-dialog.component";
import { ButtonModule } from "src/button";
import { PipesModule } from "src/pipes/pipes.module";
import { CancelButtonComponent } from "src/shared-ui/cancel-button/cancel-button.component";

describe("DeleteAccountDialogComponent", () => {
  let component: DeleteAccountDialogComponent;
  let fixture: ComponentFixture<DeleteAccountDialogComponent>;
  let dialogRefSpy: jest.Mocked<MatDialogRef<DeleteAccountDialogComponent>>;

  async function setUp(data?: DeleteAccountDialogData): Promise<void> {
    dialogRefSpy = {
      close: jest.fn(),
    } as any;

    // app-button carries a routerLink, so the real buttons need a router.
    const providers: any[] = [
      provideRouter([]),
      { provide: MatDialogRef, useValue: dialogRefSpy },
    ];
    // Existing callers open the dialog with no data at all, so the token is
    // left unprovided rather than provided as undefined.
    if (data) {
      providers.push({ provide: MAT_DIALOG_DATA, useValue: data });
    }

    await TestBed.configureTestingModule({
      declarations: [CancelButtonComponent, DeleteAccountDialogComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [ButtonModule, ReactiveFormsModule, PipesModule],
      providers,
    }).compileComponents();

    fixture = TestBed.createComponent(DeleteAccountDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  describe("password mode", () => {
    beforeEach(() => setUp());

    it("should create", () => {
      expect(component).toBeTruthy();
    });

    it("renders the password field and no provider buttons", () => {
      const element: HTMLElement = fixture.nativeElement;

      expect(element.querySelector("app-input")).toBeTruthy();
      expect(element.querySelector("app-dialog-footer")).toBeTruthy();
      expect(element.querySelectorAll("[data-testid^='delete-confirm-']").length).toBe(0);
    });

    it("should close with the password when submit is clicked and form is valid", () => {
      component.form.patchValue({ password: "mypassword" });
      component.submitButtonClicked();

      expect(dialogRefSpy.close).toHaveBeenCalledWith({ password: "mypassword" });
    });

    it("should not close when submit is clicked and form is invalid", () => {
      component.submitButtonClicked();

      expect(dialogRefSpy.close).not.toHaveBeenCalled();
    });

    it("should close with false when cancel is clicked", () => {
      component.cancelButtonClicked();

      expect(dialogRefSpy.close).toHaveBeenCalledWith(false);
    });
  });

  describe("provider mode", () => {
    beforeEach(() =>
      setUp({
        providers: [
          { name: "google", displayName: "Google" },
          { name: "keycloak", displayName: "Keycloak" },
        ],
      })
    );

    it("renders one confirm button per provider and no password field", () => {
      const element: HTMLElement = fixture.nativeElement;

      expect(element.querySelector("[data-testid='delete-confirm-google']")).toBeTruthy();
      expect(element.querySelector("[data-testid='delete-confirm-keycloak']")).toBeTruthy();
      expect(element.querySelectorAll("[data-testid^='delete-confirm-']").length).toBe(2);
      expect(element.querySelector("app-input")).toBeNull();
      expect(element.textContent).toContain("no password of its own");
    });

    it("labels each button with the provider's display name", () => {
      const button: HTMLElement = fixture.nativeElement.querySelector(
        "[data-testid='delete-confirm-keycloak']"
      );

      expect(button.textContent?.trim()).toBe("Confirm with Keycloak");
    });

    it("closes with the provider's name when its button is clicked", () => {
      fixture.nativeElement
        .querySelector("[data-testid='delete-confirm-keycloak'] button")
        .click();

      expect(dialogRefSpy.close).toHaveBeenCalledWith({ providerName: "keycloak" });
    });

    it("closes with false when cancel is clicked", () => {
      fixture.nativeElement.querySelector("app-cancel-button button").click();

      expect(dialogRefSpy.close).toHaveBeenCalledWith(false);
    });
  });

  it("falls back to the password form for an empty provider list", async () => {
    await setUp({ providers: [] });

    expect(fixture.nativeElement.querySelector("app-input")).toBeTruthy();
    expect(component.providers).toEqual([]);
  });
});
