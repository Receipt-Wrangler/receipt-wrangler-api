import { HttpErrorResponse, provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { RouterTestingModule } from "@angular/router/testing";
import { NgxsModule } from "@ngxs/store";
import { of, Subject, throwError } from "rxjs";
import { ButtonModule } from "../../button";
import { CheckboxModule } from "../../checkbox/checkbox.module";
import { FormMode } from "../../enums/form-mode.enum";
import { InputModule } from "../../input";
import { ApiModule, OidcProviderService, OidcProviderView } from "../../open-api";
import { PipesModule } from "../../pipes/pipes.module";
import { SnackbarService } from "../../services/snackbar.service";
import { SharedUiModule } from "../../shared-ui/shared-ui.module";
import { OidcProviderFormComponent } from "./oidc-provider-form.component";

describe("OidcProviderFormComponent", () => {
  let component: OidcProviderFormComponent;
  let fixture: ComponentFixture<OidcProviderFormComponent>;
  let oidcProviderService: OidcProviderService;

  const provider: OidcProviderView = {
    id: 1,
    name: "google",
    displayName: "Google",
    issuerUrl: "https://accounts.google.com",
    clientId: "a-client-id",
    scope: "openid profile email",
    allowProvisioning: true,
    linkByUsername: false,
    enabled: true,
    hasClientSecret: true,
    redirectUri: "https://receipts.example.com/api/oidc/google/callback",
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [OidcProviderFormComponent],
      imports: [
        ApiModule,
        ButtonModule,
        CheckboxModule,
        InputModule,
        MatSnackBarModule,
        NgxsModule.forRoot([]),
        NoopAnimationsModule,
        PipesModule,
        ReactiveFormsModule,
        RouterTestingModule,
        SharedUiModule,
      ],
      providers: [
        SnackbarService,
        { provide: MatDialogRef, useValue: { close: jest.fn() } },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    oidcProviderService = TestBed.inject(OidcProviderService);
    fixture = TestBed.createComponent(OidcProviderFormComponent);
    component = fixture.componentInstance;
  });

  function initEdit(): void {
    component.provider = provider;
    component.mode = FormMode.edit;
    fixture.detectChanges();
  }

  function initAdd(): void {
    component.mode = FormMode.add;
    fixture.detectChanges();
  }

  it("disables the name in edit mode", () => {
    initEdit();

    // The name is part of the redirect URI already registered at the identity
    // provider, so a rename would break every subsequent login.
    expect(component.nameReadonly).toBe(true);
    expect(component.form.get("name")?.disabled).toBe(true);
  });

  it("leaves the name editable in add mode", () => {
    initAdd();

    expect(component.nameReadonly).toBe(false);
    expect(component.form.get("name")?.disabled).toBe(false);
  });

  it("requires a client secret on create", () => {
    initAdd();

    expect(component.form.get("clientSecret")?.hasError("required")).toBe(true);
  });

  // On edit a blank field means "keep the stored secret", so it must not be
  // required -- the secret is deliberately never sent to the browser.
  it("does not require a client secret on edit", () => {
    initEdit();

    expect(component.form.get("clientSecret")?.hasError("required")).toBe(false);
  });

  it("omits a blank client secret from the payload so the stored one survives", () => {
    initEdit();
    const updateSpy = jest
      .spyOn(oidcProviderService, "updateOidcProvider")
      .mockReturnValue(of(provider) as any);

    component.submit();

    expect(updateSpy).toHaveBeenCalled();
    const command = updateSpy.mock.calls[0][1] as any;
    expect(command.clientSecret).toBeUndefined();
    // And the immutable name still rides the payload, from getRawValue.
    expect(command.name).toBe("google");
  });

  it("includes the client secret when one is typed", () => {
    initEdit();
    component.form.get("clientSecret")?.setValue("  a-new-secret  ");

    const updateSpy = jest
      .spyOn(oidcProviderService, "updateOidcProvider")
      .mockReturnValue(of(provider) as any);

    component.submit();

    const command = updateSpy.mock.calls[0][1] as any;
    expect(command.clientSecret).toBe("a-new-secret");
  });

  it("rejects a scope that omits openid", () => {
    initAdd();
    const scope = component.form.get("scope");

    scope?.setValue("profile email");
    expect(scope?.invalid).toBe(true);

    scope?.setValue("openid profile");
    expect(scope?.valid).toBe(true);
  });

  it("rejects a reserved or malformed provider name", () => {
    initAdd();
    const name = component.form.get("name");

    name?.setValue("callback");
    expect(name?.invalid).toBe(true);

    name?.setValue("My Provider");
    expect(name?.invalid).toBe(true);

    name?.setValue("google");
    expect(name?.valid).toBe(true);
  });

  it("rejects an issuer URL that is not absolute", () => {
    initAdd();
    const issuerUrl = component.form.get("issuerUrl");

    issuerUrl?.setValue("accounts.google.com");
    expect(issuerUrl?.invalid).toBe(true);

    issuerUrl?.setValue("https://accounts.google.com");
    expect(issuerUrl?.valid).toBe(true);
  });

  it("renders the username-linking warning", () => {
    initAdd();

    expect(fixture.nativeElement.textContent).toContain(
      "Only enable this for a provider you control"
    );
  });

  it("shows the redirect URI to register with the provider", () => {
    initEdit();

    expect(component.redirectUri).toBe(provider.redirectUri);
    expect(fixture.nativeElement.textContent).toContain(provider.redirectUri);
  });

  describe("submit re-entrancy", () => {
    it("issues one request when submit runs again while the save is in flight", () => {
      initEdit();
      const pending = new Subject<OidcProviderView>();
      const updateSpy = jest
        .spyOn(oidcProviderService, "updateOidcProvider")
        .mockReturnValue(pending as any);

      component.submit();
      component.submit();

      expect(updateSpy).toHaveBeenCalledTimes(1);

      pending.next(provider);
      pending.complete();
    });

    it("accepts a new submit once the save has landed", () => {
      initEdit();
      const pending = new Subject<OidcProviderView>();
      const updateSpy = jest
        .spyOn(oidcProviderService, "updateOidcProvider")
        .mockReturnValue(pending as any);

      component.submit();
      pending.next(provider);
      pending.complete();
      component.submit();

      expect(updateSpy).toHaveBeenCalledTimes(2);
    });

    it("becomes submittable again after a failed save", () => {
      initEdit();
      const updateSpy = jest
        .spyOn(oidcProviderService, "updateOidcProvider")
        .mockReturnValue(throwError(() => new Error("boom")) as any);

      // The dialog stays open on a failure, so the form has to be usable again
      // -- which is why the flag clears in finalize, not in tap. submit()'s
      // catchError now handles the error, so nothing reaches RxJS's unhandled
      // channel and this needs no onUnhandledError scaffolding.
      component.submit();
      component.submit();

      expect(updateSpy).toHaveBeenCalledTimes(2);
    });
  });

  // A taken name comes back as a bare {"name": "..."} 400. The interceptor only
  // toasts errorMsg, so before this the dialog sat there having apparently done
  // nothing and the only evidence was the failed request in devtools.
  //
  // These drive ADD mode, which is the only way to reach that error: the name is
  // immutable after create and the server rejects a rename outright, so a
  // collision can only happen while the field is still editable -- which is also
  // the only state in which the message can render.
  describe("a rejected save", () => {
    const NAME_TAKEN = "an OIDC provider with this name already exists";

    function initValidAdd(): void {
      initAdd();
      component.form.patchValue({
        name: "google",
        displayName: "Google",
        issuerUrl: "https://accounts.google.com",
        clientId: "a-client-id",
        clientSecret: "a-client-secret",
        scope: "openid profile email",
      });

      // submit() early-returns on an invalid form, so a test that never reaches
      // the service would pass for the wrong reason.
      expect(component.form.valid).toBe(true);
    }

    function rejectCreate() {
      return jest.spyOn(oidcProviderService, "createOidcProvider").mockReturnValue(
        throwError(
          () =>
            new HttpErrorResponse({
              status: 400,
              statusText: "Bad Request",
              error: { name: NAME_TAKEN },
            })
        ) as any
      );
    }

    it("renders a field-keyed 400 on the control it names", () => {
      initValidAdd();
      rejectCreate();

      component.submit();

      const control = component.form.get("name")!;
      expect(control.errors).toEqual({ server: NAME_TAKEN });
      expect(control.touched).toBe(true);
    });

    it("neither toasts success nor closes the dialog", () => {
      initValidAdd();
      const snackbarService = TestBed.inject(SnackbarService);
      const dialogRef = TestBed.inject(MatDialogRef);
      rejectCreate();
      const successSpy = jest.spyOn(snackbarService, "success");
      const closeSpy = jest.spyOn(dialogRef, "close");

      component.submit();

      expect(successSpy).not.toHaveBeenCalled();
      expect(closeSpy).not.toHaveBeenCalled();
    });

    // The error clears as soon as the admin edits the field, because a value
    // change re-runs the validators and replaces the errors object -- so the
    // stale server verdict cannot outlive the value it judged.
    it("clears once the admin changes the name", () => {
      initValidAdd();
      rejectCreate();

      component.submit();
      expect(component.form.get("name")!.errors).toEqual({ server: NAME_TAKEN });

      component.form.get("name")!.setValue("google-workspace");

      expect(component.form.get("name")!.errors).toBeNull();
    });
  });

  describe("copying the redirect URI", () => {
    const originalClipboard = Object.getOwnPropertyDescriptor(
      window.navigator,
      "clipboard"
    );

    function setClipboard(value: unknown): void {
      Object.defineProperty(window.navigator, "clipboard", {
        value,
        configurable: true,
      });
    }

    afterEach(() => {
      if (originalClipboard) {
        Object.defineProperty(window.navigator, "clipboard", originalClipboard);
      } else {
        delete (window.navigator as any).clipboard;
      }
    });

    it("reports success", async () => {
      initEdit();
      const writeText = jest.fn().mockResolvedValue(undefined);
      setClipboard({ writeText });
      const successSpy = jest
        .spyOn(TestBed.inject(SnackbarService), "success")
        .mockImplementation(() => undefined as any);

      await component.copyRedirectUri();

      expect(writeText).toHaveBeenCalledWith(provider.redirectUri);
      expect(successSpy).toHaveBeenCalledWith("Redirect URI copied");
    });

    it("reports a rejected write instead of leaving it unhandled", async () => {
      initEdit();
      setClipboard({ writeText: jest.fn().mockRejectedValue(new Error("nope")) });
      const errorSpy = jest
        .spyOn(TestBed.inject(SnackbarService), "error")
        .mockImplementation(() => undefined as any);

      await expect(component.copyRedirectUri()).resolves.toBeUndefined();

      expect(errorSpy).toHaveBeenCalledWith(
        expect.stringContaining("Could not copy the redirect URI")
      );
    });

    it("reports an absent clipboard rather than throwing", async () => {
      // What a plain-http self-hosted install actually looks like: outside a
      // secure context navigator.clipboard is undefined, so the old code threw
      // synchronously and the button did nothing at all.
      initEdit();
      setClipboard(undefined);
      const errorSpy = jest
        .spyOn(TestBed.inject(SnackbarService), "error")
        .mockImplementation(() => undefined as any);

      await expect(component.copyRedirectUri()).resolves.toBeUndefined();

      expect(errorSpy).toHaveBeenCalledWith(
        expect.stringContaining("Could not copy the redirect URI")
      );
    });
  });
});
