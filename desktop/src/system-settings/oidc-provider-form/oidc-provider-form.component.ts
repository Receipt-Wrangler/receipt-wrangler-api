import { Component, Input, OnInit } from "@angular/core";
import { AbstractControl, FormBuilder, FormGroup, ValidationErrors, ValidatorFn, Validators } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { Observable, take, tap } from "rxjs";
import { FormMode } from "../../enums/form-mode.enum";
import { OidcProviderService, OidcProviderView, UpsertOidcProviderCommand } from "../../open-api/index";
import { SnackbarService } from "../../services/index";
import { absoluteUrlValidator, trimmedRequiredValidator } from "../../validators/index";

/**
 * Slugs that would shadow a route under /api/oidc/. Mirrors the backend's
 * ReservedOidcProviderNames so the user gets inline feedback rather than a 400.
 */
const RESERVED_PROVIDER_NAMES = ["login", "callback", "link", "exchange", "connections"];

/** Mirrors the backend's oidcProviderNameRegex. */
const PROVIDER_NAME_PATTERN = /^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$/;

/**
 * The name becomes a URL path segment and, through it, part of the redirect URI
 * registered at the identity provider. The message is carried as the error
 * *value* because BaseInputComponent only maps a handful of error keys and would
 * otherwise render an empty mat-error.
 */
function providerNameValidator(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const raw = String(control.value ?? "").trim();
    if (raw.length === 0) {
      return null;
    }

    if (!PROVIDER_NAME_PATTERN.test(raw)) {
      return {
        providerName:
          "Use lowercase letters, numbers and dashes, starting and ending with a letter or number.",
      };
    }

    if (RESERVED_PROVIDER_NAMES.includes(raw)) {
      return { providerName: `"${raw}" is reserved. Pick another name.` };
    }

    return null;
  };
}

/**
 * The scope must contain openid, or the provider returns no ID token and there
 * is nothing to verify an identity from.
 */
function openidScopeValidator(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const raw = String(control.value ?? "").trim();
    if (raw.length === 0) {
      return null;
    }

    return raw.split(/\s+/).includes("openid")
      ? null
      : { scope: "Scope must include openid." };
  };
}

@Component({
  selector: "app-oidc-provider-form",
  templateUrl: "./oidc-provider-form.component.html",
  styleUrl: "./oidc-provider-form.component.scss",
  standalone: false,
})
export class OidcProviderFormComponent implements OnInit {
  @Input() public headerText: string = "";

  @Input() public provider?: OidcProviderView;

  @Input() public mode: FormMode = FormMode.add;

  public form!: FormGroup;

  protected readonly FormMode = FormMode;

  constructor(
    private formBuilder: FormBuilder,
    private matDialogRef: MatDialogRef<OidcProviderFormComponent>,
    private oidcProviderService: OidcProviderService,
    private snackbarService: SnackbarService,
  ) {
  }

  /**
   * The name is part of the redirect URI already registered at the identity
   * provider, so renaming it would break every subsequent login with a mismatch
   * raised at the provider rather than here. The server rejects it too.
   */
  public get nameReadonly(): boolean {
    return this.mode !== FormMode.add;
  }

  public get redirectUri(): string {
    return this.provider?.redirectUri ?? "";
  }

  /**
   * Copy for the client-secret field. On edit the stored secret is deliberately
   * never sent to the browser, so a blank field means "keep it".
   */
  public get clientSecretHint(): string {
    if (this.mode === FormMode.add) {
      return "";
    }

    return this.provider?.hasClientSecret
      ? "Leave blank to keep the current secret."
      : "No secret is currently stored.";
  }

  public ngOnInit(): void {
    this.initForm();
  }

  public submit(): void {
    if (!this.form.valid) {
      return;
    }

    const command = this.buildCommand();

    const request$: Observable<OidcProviderView> =
      this.mode === FormMode.edit && this.provider
        ? this.oidcProviderService.updateOidcProvider(this.provider.id.toString(), command)
        : this.oidcProviderService.createOidcProvider(command);

    const successMessage =
      this.mode === FormMode.edit ? "OIDC provider updated" : "OIDC provider created";

    request$
      .pipe(
        take(1),
        tap(() => {
          this.snackbarService.success(successMessage);
          this.matDialogRef.close(true);
        })
      )
      .subscribe();
  }

  public closeDialog(): void {
    this.matDialogRef.close(false);
  }

  public async copyRedirectUri(): Promise<void> {
    if (!this.redirectUri) {
      return;
    }

    await navigator.clipboard.writeText(this.redirectUri);
    this.snackbarService.success("Redirect URI copied");
  }

  private buildCommand(): UpsertOidcProviderCommand {
    const { name, displayName, issuerUrl, clientId, clientSecret, scope, allowProvisioning, linkByUsername, enabled } =
      this.form.getRawValue();

    const command: UpsertOidcProviderCommand = {
      name,
      displayName,
      issuerUrl,
      clientId,
      scope,
      allowProvisioning,
      linkByUsername,
      enabled,
    };

    // Omitting the key is what tells the server to keep the stored secret. A
    // blank string would be rejected, and sending one on every save would be a
    // needless re-encrypt.
    const trimmedSecret = String(clientSecret ?? "").trim();
    if (trimmedSecret.length > 0) {
      command.clientSecret = trimmedSecret;
    }

    return command;
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      name: [
        this.provider?.name,
        [trimmedRequiredValidator(), providerNameValidator()],
      ],
      displayName: [this.provider?.displayName, [trimmedRequiredValidator()]],
      issuerUrl: [
        this.provider?.issuerUrl,
        [trimmedRequiredValidator(), absoluteUrlValidator()],
      ],
      clientId: [this.provider?.clientId, [trimmedRequiredValidator()]],
      // Required only on create: on edit a blank field means "keep the stored
      // secret", which is why it is never populated from the provider.
      clientSecret: [
        "",
        this.mode === FormMode.add ? [Validators.required] : [],
      ],
      scope: [
        this.provider?.scope ?? "openid profile email",
        [trimmedRequiredValidator(), openidScopeValidator()],
      ],
      allowProvisioning: [this.provider?.allowProvisioning ?? false],
      linkByUsername: [this.provider?.linkByUsername ?? false],
      enabled: [this.provider?.enabled ?? true],
    });

    if (this.mode === FormMode.view) {
      this.form.disable();
    }

    if (this.nameReadonly) {
      this.form.get("name")?.disable();
    }
  }
}
