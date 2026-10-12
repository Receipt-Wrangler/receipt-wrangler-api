import { HttpErrorResponse } from "@angular/common/http";
import { Component, OnInit, ViewEncapsulation, signal } from "@angular/core";
import { FormBuilder, FormControl, FormGroup, Validators, } from "@angular/forms";
import { ActivatedRoute, Router } from "@angular/router";
import { Store } from "@ngxs/store";
import { BehaviorSubject, catchError, EMPTY, finalize, switchMap, tap, } from "rxjs";
import { AppData, AuthService, OidcProviderSummary } from "src/open-api";
import { SnackbarService } from "src/services";
import { setAppData } from "src/utils";
import { fadeIn, fadeInOut } from "../../animations";
import { FeatureConfigState, GroupState } from "../../store";
import { UserValidators } from "../../validators";

@Component({
    selector: "app-auth-form",
    templateUrl: "./auth-form.component.html",
    styleUrls: ["./auth-form.component.scss"],
    encapsulation: ViewEncapsulation.None,
    providers: [UserValidators],
    animations: [fadeInOut, fadeIn],
    standalone: false
})
export class AuthForm implements OnInit {
  public form: FormGroup = new FormGroup({});
  public isSignUp: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(
    false
  );
  public headerText: string = "";
  public primaryButtonText: string = "";
  public secondaryButtonText: string = "";
  public secondaryButtonRouterLink: string[] = [];
  public readonly isLoading = signal<boolean>(false);
  public readonly oidcProviders = this.store.selectSignal(
    FeatureConfigState.oidcProviders
  );

  constructor(
    private authSerivce: AuthService,
    private snackbarService: SnackbarService,
    protected formBuilder: FormBuilder,
    protected route: ActivatedRoute,
    protected router: Router,
    protected store: Store,
    protected userValidators: UserValidators
  ) {
  }

  public ngOnInit(): void {
    this.initForm();
    this.listenForRouteChanges();
    this.listenForIsSignUpChanges();
    this.showOidcErrorIfPresent();
  }

  /**
   * Starts an OIDC login.
   *
   * This is a full-page navigation, not an HttpClient call: the browser has to
   * follow the redirect chain to the identity provider itself, and an XHR would
   * just fetch the provider's HTML into JavaScript.
   */
  public loginWithOidc(provider: OidcProviderSummary): void {
    this.navigateToUrl(buildOidcLoginUrl(provider.name));
  }

  /** Seam for the full-page navigation, so a spec can assert where it goes. */
  protected navigateToUrl(url: string): void {
    window.location.href = url;
  }

  /**
   * Surfaces a failure the backend redirected back with. The backend only ever
   * sends a code from a small fixed vocabulary -- an identity provider's own
   * error text is never echoed through -- so this maps it to our own copy.
   */
  private showOidcErrorIfPresent(): void {
    const code = this.route.snapshot?.queryParams?.["oidcError"];
    if (!code) {
      return;
    }

    this.snackbarService.error(OIDC_ERROR_MESSAGES[code] ?? OIDC_ERROR_MESSAGES["default"]);
  }

  private listenForRouteChanges(): void {
    this.route.data
      .pipe(
        tap((data) => {
          this.isSignUp.next(!!data?.["isSignUp"]);
        })
      )
      .subscribe();
  }

  private listenForIsSignUpChanges(): void {
    this.isSignUp
      .pipe(
        tap((isSignUp) => {
          if (isSignUp) {
            this.headerText = "Sign Up";
            this.primaryButtonText = "Sign Up";
            this.secondaryButtonRouterLink = ["/auth/login"];
            this.secondaryButtonText = "Back to Login";
            this.form
              .get("username")
              ?.addAsyncValidators(this.userValidators.uniqueUsername(0, ""));
            this.form.addControl(
              "displayname",
              new FormControl("", Validators.required)
            );
          } else {
            this.headerText = "Login";
            this.primaryButtonText = "Login";
            this.secondaryButtonRouterLink = ["/auth/sign-up"];
            this.secondaryButtonText = "Sign Up";
            this.form
              .get("username")
              ?.removeAsyncValidators(
                this.userValidators.uniqueUsername(0, "")
              );
            this.form.removeControl("displayname");
          }
        })
      )
      .subscribe();
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      username: ["", [Validators.required]],
      password: ["", Validators.required],
    });
  }

  public submit(): void {
    const isSignUp = this.isSignUp.getValue();
    const isValid = this.form.valid;

    if (isValid && isSignUp) {
      this.signUp();
    } else if (isValid && !isSignUp) {
      this.login();
    }
  }

  private signUp(): void {
    this.authSerivce
      .signUp(this.form.value)
      .pipe(
        tap(() => {
          this.login();
        }),
        catchError((err: HttpErrorResponse) => {
          // A sign-up validation failure is a flat field map with no errorMsg
          // (api/internal/structs/utils.go), so the interceptor stays silent and
          // this is the only report. Anything carrying errorMsg has already been
          // toasted by the interceptor — toasting again would dismiss it.
          const usernameError = err.error?.["username"];

          if (usernameError) {
            this.snackbarService.error(usernameError);
          }

          return EMPTY;
        })
      )
      .subscribe();
  }

  private login(): void {
    this.isLoading.set(true);
    this.authSerivce
      .login(this.form.value)
      .pipe(
        switchMap((appData: AppData) => setAppData(this.store, appData)),
        tap(() =>
          this.router.navigate([
            this.store.selectSnapshot(GroupState.dashboardLink),
          ]),
        ),
        catchError(() => EMPTY), // the HTTP interceptor surfaces the failure toast
        finalize(() => this.isLoading.set(false))
      )
      .subscribe();
  }
}

/** Builds a provider's sign-in URL. Pure, so it can be asserted directly. */
export function buildOidcLoginUrl(providerName: string): string {
  return `/api/oidc/${encodeURIComponent(providerName)}/login`;
}

/**
 * Copy for the error codes the OIDC callback can redirect back with. Keyed by
 * the backend's fixed vocabulary so an unknown code still gets a sensible
 * message rather than nothing.
 */
const OIDC_ERROR_MESSAGES: Record<string, string> = {
  unknown_provider: "That sign-in provider is not available.",
  no_account:
    "No Receipt Wrangler account is linked to that identity. Sign in and connect it from your profile, or ask an administrator.",
  account_exists:
    "An account with that username already exists. Sign in with your password and connect the provider from your profile.",
  already_linked: "That identity is already connected to another account.",
  invalid_state: "That sign-in attempt expired. Please try again.",
  nonce_mismatch: "That sign-in attempt could not be verified. Please try again.",
  no_id_token: "The provider did not return an identity token.",
  provider_error: "The sign-in provider reported an error.",
  invalid_request: "That sign-in request was not valid.",
  default: "Sign in failed. Please try again.",
};
