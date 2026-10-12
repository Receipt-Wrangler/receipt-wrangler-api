import { Component, OnInit } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import { Store } from "@ngxs/store";
import { catchError, of, switchMap, tap } from "rxjs";
import { UserService } from "src/open-api";
import { setAppData } from "src/utils";
import { fadeInOut } from "../../animations";
import { GroupState } from "../../store";

/**
 * Lands the browser after a successful OIDC sign-in and bootstraps the app.
 *
 * This route exists because AppInitService only loads app data when
 * `isLoggedIn || hadSession`, both of which read persisted NGXS state out of
 * localStorage. A first-ever OIDC login has perfectly valid session cookies but
 * an EMPTY store, so the initializer takes its unauthenticated branch and the
 * app boots signed-out while holding a working session. Fetching app data here
 * is what closes that gap.
 *
 * The alternative -- making the initializer always attempt getAppData -- was
 * rejected: it would add a guaranteed 403 to every anonymous page load.
 */
@Component({
  selector: "app-oidc-callback",
  templateUrl: "./oidc-callback.component.html",
  styleUrls: ["./oidc-callback.component.scss"],
  animations: [fadeInOut],
  standalone: false,
})
export class OidcCallbackComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private store: Store,
    private userService: UserService,
  ) {
  }

  public ngOnInit(): void {
    // The backend redirects failures to /auth/login with a code, so anything
    // arriving here with one is a defensive fallback rather than the normal path.
    const errorCode = this.route.snapshot.queryParams?.["oidcError"];
    if (errorCode) {
      this.router.navigate(["/auth/login"], {
        queryParams: { oidcError: errorCode },
      });
      return;
    }

    this.userService
      .getAppData()
      .pipe(
        switchMap((appData) => setAppData(this.store, appData)),
        tap(() =>
          this.router.navigate([
            this.store.selectSnapshot(GroupState.dashboardLink),
          ]),
        ),
        catchError(() => {
          // The cookies did not survive, or the session was refused. Send the
          // user back to a login they can actually complete.
          this.router.navigate(["/auth/login"], {
            queryParams: { oidcError: "invalid_state" },
          });
          return of(undefined);
        }),
      )
      .subscribe();
  }
}
