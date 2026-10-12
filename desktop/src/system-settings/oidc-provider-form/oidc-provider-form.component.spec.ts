import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { RouterTestingModule } from "@angular/router/testing";
import { NgxsModule } from "@ngxs/store";
import { config, of, Subject, throwError } from "rxjs";
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

      // subscribe() takes no error handler -- the interceptor is what reports
      // a failed save -- so RxJS routes the error here rather than throwing
      // out of submit(). Swallow it so it does not surface as an unhandled
      // error in a later test.
      const originalOnUnhandledError = config.onUnhandledError;
      config.onUnhandledError = () => undefined;

      try {
        // The dialog stays open on a failure, so the form has to be usable
        // again -- which is why the flag clears in finalize, not in tap.
        component.submit();
        component.submit();
      } finally {
        config.onUnhandledError = originalOnUnhandledError;
      }

      expect(updateSpy).toHaveBeenCalledTimes(2);
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
