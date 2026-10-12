import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { RouterTestingModule } from "@angular/router/testing";
import { NgxsModule } from "@ngxs/store";
import { of } from "rxjs";
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
});
