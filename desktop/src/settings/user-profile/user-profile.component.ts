import {Component, computed, OnInit, signal} from "@angular/core";
import {FormBuilder, FormGroup, Validators} from "@angular/forms";
import {MatDialog, MatDialogConfig} from "@angular/material/dialog";
import {ActivatedRoute, Router} from "@angular/router";
import {Store} from "@ngxs/store";
import {catchError, Observable, of, switchMap, take, tap} from "rxjs";
import {DEFAULT_DIALOG_CONFIG} from "src/constants/dialog.constant";
import {FormMode} from "src/enums/form-mode.enum";
import {FormConfig} from "src/interfaces";
import {DeleteAccountCommand, OidcConnectionView, OidcService, Permission, User, UserService} from "../../open-api";
import {ClaimsService, SnackbarService, TokenRefreshService} from "../../services";
import {AuthState, FeatureConfigState, Logout, UpdateUser} from "../../store";
import {ConfirmationDialogComponent} from "../../shared-ui/confirmation-dialog/confirmation-dialog.component";
import {
  DeleteAccountDialogComponent,
  DeleteAccountDialogData,
  DeleteAccountDialogResult,
} from "../delete-account-dialog/delete-account-dialog.component";

@Component({
    selector: "app-user-profile",
    templateUrl: "./user-profile.component.html",
    styleUrls: ["./user-profile.component.scss"],
    standalone: false
})
export class UserProfileComponent implements OnInit {
  public form: FormGroup = new FormGroup({});

  public user!: User;

  public formConfig!: FormConfig;

  public formMode = FormMode;

  public usernameTooltip: string =
    "Only system admin may change your username.";

  protected readonly Permission = Permission;

  protected readonly canEdit = this.store.selectSignal(
    AuthState.hasAppPermission(Permission.AppAccountUpdate)
  );

  /** The caller's linked identity providers, loaded on init. */
  public readonly connections = signal<OidcConnectionView[]>([]);

  private readonly configuredProviders = this.store.selectSignal(
    FeatureConfigState.oidcProviders
  );

  /**
   * Providers this account has not connected yet. Computed from the public
   * feature config rather than a second request -- the login page already needs
   * that list, so it is always loaded.
   */
  public readonly availableProviders = computed(() => {
    const connected = new Set(this.connections().map((c) => c.providerName));

    return this.configuredProviders().filter((p) => !connected.has(p.name));
  });

  /** Shown only when there is something to say about connected accounts. */
  public readonly showConnectedAccounts = computed(
    () => this.connections().length > 0 || this.availableProviders().length > 0
  );

  /**
   * An account created by a provider has no password of its own, so it confirms
   * a self-delete by signing in at one of its providers again. Every connection
   * of such an account carries the flag. A failed load leaves the list empty,
   * which falls back to the password dialog -- the behaviour before this existed.
   */
  public readonly isPasswordless = computed(() =>
    this.connections().some((c) => c.provisionedUser)
  );

  constructor(
    private claimsService: ClaimsService,
    private formBuilder: FormBuilder,
    private matDialog: MatDialog,
    private route: ActivatedRoute,
    private router: Router,
    private snackbarService: SnackbarService,
    private store: Store,
    private oidcService: OidcService,
    private tokenRefreshService: TokenRefreshService,
    private userService: UserService,
  ) {
  }

  public ngOnInit(): void {
    this.user = this.store.selectSnapshot(AuthState.loggedInUser);
    this.formConfig = this.route?.snapshot?.data?.["formConfig"];
    this.initForm();
    this.loadConnections();
    this.reportLinkOutcome();
    this.completeDeleteConfirmation();
  }

  /**
   * Starts a "connect account" flow.
   *
   * A full-page navigation, not an HttpClient call: the browser has to follow
   * the redirect chain to the identity provider. The session cookie authorizes
   * the request, and the backend links the identity to this account directly --
   * it never has to infer who the caller is from a claim.
   */
  public connectProvider(providerName: string): void {
    this.navigateToUrl(`/api/oidc/link/${encodeURIComponent(providerName)}`);
  }

  /** Seam for the full-page navigation, so a spec can assert where it goes. */
  protected navigateToUrl(url: string): void {
    window.location.href = url;
  }

  public disconnectProvider(connection: OidcConnectionView): void {
    const dialogRef = this.matDialog.open(
      ConfirmationDialogComponent,
      DEFAULT_DIALOG_CONFIG,
    );

    dialogRef.componentInstance.headerText = `Disconnect ${connection.providerDisplayName}`;
    dialogRef.componentInstance.dialogContent =
      `Are you sure you want to disconnect ${connection.providerDisplayName}? You will no longer be able to sign in with it.`;

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        switchMap((confirmed) => {
          if (!confirmed) {
            return of(undefined);
          }

          return this.oidcService.deleteOidcConnection(connection.providerName).pipe(
            tap(() => {
              this.snackbarService.success(
                `${connection.providerDisplayName} disconnected`
              );
              this.loadConnections();
            }),
            // The server refuses to strand an account that has no password to
            // fall back on; the interceptor already surfaced the reason.
            catchError(() => of(undefined)),
          );
        })
      )
      .subscribe();
  }

  /**
   * An account created by a provider has only an unusable password, so removing
   * its last connection would lock it out. The server refuses either way; this
   * just keeps the button from dangling.
   */
  public canDisconnect(connection: OidcConnectionView): boolean {
    return !connection.provisionedUser || this.connections().length > 1;
  }

  private loadConnections(): void {
    this.oidcService
      .getOidcConnections()
      .pipe(
        take(1),
        tap((connections) => this.connections.set(connections ?? [])),
        // A user without the account-read permission, or an older server, simply
        // gets no section rather than an error.
        catchError(() => of(undefined)),
      )
      .subscribe();
  }

  /** Reports the result of a link flow the backend redirected back from. */
  private reportLinkOutcome(): void {
    const params = this.route?.snapshot?.queryParams ?? {};

    if (params["oidcLinked"]) {
      this.snackbarService.success("Account connected");
      return;
    }

    if (params["oidcError"] === "already_linked") {
      this.snackbarService.error(
        "That identity is already connected to an account."
      );
      return;
    }

    if (params["oidcError"]) {
      this.snackbarService.error("Could not connect that account.");
    }
  }

  /**
   * Finishes a provider-confirmed account deletion the backend redirected back
   * from. The parameters are stripped from the address bar first, so a reload
   * cannot resubmit a spent code or repeat a stale error.
   */
  private completeDeleteConfirmation(): void {
    const params = this.route?.snapshot?.queryParams ?? {};
    const confirmationCode = params["deleteConfirmation"];
    const errorCode = params["deleteError"];

    if (!confirmationCode && !errorCode) {
      return;
    }

    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { deleteConfirmation: null, deleteError: null },
      queryParamsHandling: "merge",
      replaceUrl: true,
    });

    if (confirmationCode) {
      this.deleteAccountAndSignOut({ confirmationCode })
        .pipe(
          take(1),
          // A spent or expired code is a 400; the interceptor already reported it.
          catchError(() => of(undefined)),
        )
        .subscribe();
      return;
    }

    // Own-property check, not a bare lookup: the query param is whatever the
    // address bar holds, and an inherited key like "constructor" would resolve
    // to a function.
    this.snackbarService.error(
      Object.hasOwn(DELETE_CONFIRMATION_ERROR_MESSAGES, errorCode)
        ? DELETE_CONFIRMATION_ERROR_MESSAGES[errorCode]
        : DELETE_CONFIRMATION_ERROR_MESSAGES["default"]
    );
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      username: this.user?.username ?? "",
      displayName: [this.user?.displayName ?? "", Validators.required],
      defaultAvatarColor: [
        this.user?.defaultAvatarColor ?? "",
        Validators.pattern("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$"),
      ],
    });

    if (this.formConfig.mode === FormMode.edit) {
      this.form.get("username")?.disable();
    }
  }

  public submit(): void {
    if (this.form.valid) {
      this.userService
        .updateUserProfile(this.form.value)
        .pipe(
          take(1),
          switchMap(() => this.tokenRefreshService.refreshToken()),
          switchMap(() => this.claimsService.getAndSetClaimsForLoggedInUser()),
          switchMap(() => {
            const loggedInUser = this.store.selectSnapshot(
              AuthState.loggedInUser
            );
            return this.store.dispatch(
              new UpdateUser(loggedInUser.id.toString(), loggedInUser)
            );
          }),
          tap(() => {
            this.snackbarService.success("User profile successfully updated");
            this.router.navigate(["/settings/user-profile/view"],
              {
                queryParams: {
                  tab: "user-profile",
                }
              });
          })
        )
        .subscribe();
    }
  }

  public deleteAccount(): void {
    const dialogRef = this.matDialog.open<
      DeleteAccountDialogComponent,
      DeleteAccountDialogData,
      DeleteAccountDialogResult
    >(DeleteAccountDialogComponent, this.deleteAccountDialogConfig());

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        switchMap((result) => {
          if (result && "providerName" in result) {
            // The provider sends the browser back here with a single-use code,
            // which completeDeleteConfirmation() then spends.
            this.navigateToUrl(buildOidcDeleteConfirmationUrl(result.providerName));
            return of(undefined);
          }

          if (result && "password" in result) {
            return this.deleteAccountAndSignOut({ password: result.password }).pipe(
              catchError((err) => {
                if (err.status === 401) {
                  this.deleteAccount();
                }
                return of(undefined);
              })
            );
          }
          return of(undefined);
        })
      )
      .subscribe();
  }

  /**
   * A passwordless account gets one confirm button per connected provider;
   * every other account gets the password form.
   */
  private deleteAccountDialogConfig(): MatDialogConfig<DeleteAccountDialogData> {
    if (!this.isPasswordless()) {
      return DEFAULT_DIALOG_CONFIG;
    }

    return {
      ...DEFAULT_DIALOG_CONFIG,
      data: {
        providers: this.connections().map((c) => ({
          name: c.providerName,
          displayName: c.providerDisplayName,
        })),
      },
    };
  }

  private deleteAccountAndSignOut(command: DeleteAccountCommand): Observable<void> {
    return this.userService.deleteAccount(command).pipe(
      switchMap(() => this.store.dispatch(new Logout())),
      tap(() => {
        this.snackbarService.success(
          "Your account has been successfully deleted"
        );
        this.router.navigate(["/"]);
      }),
    );
  }
}

/**
 * Builds the URL that starts a provider-confirmed account deletion. Pure, so it
 * can be asserted directly.
 */
export function buildOidcDeleteConfirmationUrl(providerName: string): string {
  return `/api/oidc/connections/${encodeURIComponent(providerName)}/delete-confirmation`;
}

/**
 * Copy for the error codes a delete confirmation can redirect back with, keyed
 * by the backend's fixed vocabulary so an unknown code still gets a message.
 */
const DELETE_CONFIRMATION_ERROR_MESSAGES: Record<string, string> = {
  identity_mismatch:
    "You signed in with a different account than the one connected here.",
  stale_login:
    "Your provider didn't ask you to sign in again. Sign out of it and try again.",
  not_linked: "That provider is no longer connected to your account.",
  forbidden: "Your account can't be deleted this way.",
  default: "Could not confirm the account deletion. Please try again.",
};
