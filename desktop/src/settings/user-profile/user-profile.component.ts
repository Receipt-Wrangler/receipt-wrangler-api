import {Component, computed, OnInit, signal} from "@angular/core";
import {FormBuilder, FormGroup, Validators} from "@angular/forms";
import {MatDialog} from "@angular/material/dialog";
import {ActivatedRoute, Router} from "@angular/router";
import {Store} from "@ngxs/store";
import {catchError, of, switchMap, take, tap} from "rxjs";
import {DEFAULT_DIALOG_CONFIG} from "src/constants/dialog.constant";
import {FormMode} from "src/enums/form-mode.enum";
import {FormConfig} from "src/interfaces";
import {OidcConnectionView, OidcService, Permission, User, UserService} from "../../open-api";
import {ClaimsService, SnackbarService, TokenRefreshService} from "../../services";
import {AuthState, FeatureConfigState, Logout, UpdateUser} from "../../store";
import {ConfirmationDialogComponent} from "../../shared-ui/confirmation-dialog/confirmation-dialog.component";
import {DeleteAccountDialogComponent} from "../delete-account-dialog/delete-account-dialog.component";

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
    window.location.href = `/api/oidc/link/${encodeURIComponent(providerName)}`;
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
    const dialogRef = this.matDialog.open(
      DeleteAccountDialogComponent,
      DEFAULT_DIALOG_CONFIG,
    );

    dialogRef
      .afterClosed()
      .pipe(
        take(1),
        switchMap((password) => {
          if (password) {
            return this.userService.deleteAccount({ password }).pipe(
              switchMap(() => this.store.dispatch(new Logout())),
              tap(() => {
                this.snackbarService.success(
                  "Your account has been successfully deleted"
                );
                this.router.navigate(["/"]);
              }),
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
}
