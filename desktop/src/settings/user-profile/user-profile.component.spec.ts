import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule } from "@angular/forms";
import { MatDialog, MatDialogModule, MatDialogRef } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { ActivatedRoute, Router } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { of, throwError } from "rxjs";
import { PipesModule } from "src/pipes/pipes.module";
import { DirectivesModule } from "../../directives";
import { ApiModule, OidcConnectionView, OidcService, Permission, UserService } from "../../open-api";
import { SnackbarService, TokenRefreshService } from "../../services";
import { AuthState, FeatureConfigState, Logout, UserState } from "../../store";
import { SetPermissions } from "../../store/auth.state.actions";
import { buildOidcDeleteConfirmationUrl, UserProfileComponent } from "./user-profile.component";
import { DeleteAccountDialogComponent } from "../delete-account-dialog/delete-account-dialog.component";

function connection(overrides: Partial<OidcConnectionView> = {}): OidcConnectionView {
  return {
    providerName: "google",
    providerDisplayName: "Google",
    provisionedUser: true,
    linkedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("UserProfileComponent", () => {
  let component: UserProfileComponent;
  let fixture: ComponentFixture<UserProfileComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [UserProfileComponent],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    imports: [ApiModule,
        DirectivesModule,
        PipesModule,
        MatDialogModule,
        MatSnackBarModule,
        NgxsModule.forRoot([AuthState, FeatureConfigState, UserState]),
        PipesModule,
        ReactiveFormsModule],
    providers: [
        {
            provide: ActivatedRoute,
            useValue: { snapshot: { data: { formConfig: {} } } },
        },
        { provide: Router, useValue: { navigate: jest.fn().mockResolvedValue(true) } },
        provideHttpClient(withInterceptorsFromDi()),
    ]
}).compileComponents();

    fixture = TestBed.createComponent(UserProfileComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("hides the delete-account danger zone without app.account.delete", async () => {
    const store = TestBed.inject(Store);
    store.dispatch(new SetPermissions([], {}));

    fixture.detectChanges();
    await fixture.whenStable();

    expect(fixture.nativeElement.querySelector(".danger-zone")).toBeFalsy();
  });

  it("shows the delete-account danger zone with app.account.delete", async () => {
    const store = TestBed.inject(Store);
    store.dispatch(new SetPermissions([Permission.AppAccountDelete], {}));

    fixture.detectChanges();
    await fixture.whenStable();

    expect(fixture.nativeElement.querySelector(".danger-zone")).toBeTruthy();
  });

  it("gates the edit button on app.account.update", () => {
    const store = TestBed.inject(Store);

    store.dispatch(new SetPermissions([], {}));
    expect((component as any).canEdit()).toEqual(false);

    store.dispatch(new SetPermissions([Permission.AppAccountUpdate], {}));
    expect((component as any).canEdit()).toEqual(true);
  });

  it("should init form correctly", () => {
    const store = TestBed.inject(Store);
    store.reset({
      auth: {
        username: "cheetos",
        displayname: "burger",
        defaultAvatarColor: "#CD5C5C",
      },
    });

    component.ngOnInit();

    expect(component.form.value).toEqual({
      username: "cheetos",
      displayName: "burger",
      defaultAvatarColor: "#CD5C5C",
    });
  });

  it("should submit form and update state correctly", () => {
    const store = TestBed.inject(Store);
    const serviceSpy = jest.spyOn(TestBed.inject(UserService), "updateUserProfile");
    const authSpy = jest.spyOn(TestBed.inject(TokenRefreshService), "refreshToken");

    jest.spyOn(TestBed.inject(UserService), "getUserClaims").mockReturnValue(
      of({
        userId: "1",
        displayname: "store",
        username: "general",
      } as any)
    );

    serviceSpy.mockReturnValue(of(undefined) as any);
    authSpy.mockReturnValue(of(undefined as any));

    store.reset({
      users: { users: [{ id: 1, displayName: "cheetos", username: "burger" }] },
      auth: {
        userId: "1",
        username: "cheetos",
        displayname: "burger",
        defaultAvatarColor: "#CD5C5C",
      },
    });

    component.ngOnInit();
    component.form.patchValue({
      username: "general",
      displayName: "store",
    });

    component.submit();

    const updatedUser = store.selectSnapshot(AuthState.loggedInUser);
    const updatedUsers = store.selectSnapshot(UserState.users);

    expect(serviceSpy).toHaveBeenCalledWith({
      username: "general",
      displayName: "store",
      defaultAvatarColor: "#CD5C5C",
    } as any);
    expect(authSpy).toHaveBeenCalled();
  });

  it("should open delete account dialog when deleteAccount is called", () => {
    const matDialog = TestBed.inject(MatDialog);
    const dialogRefMock = {
      afterClosed: () => of(false),
      componentInstance: {},
    } as any;

    const dialogSpy = jest.spyOn(matDialog, "open").mockReturnValue(dialogRefMock);

    component.deleteAccount();

    expect(dialogSpy).toHaveBeenCalledWith(
      DeleteAccountDialogComponent,
      expect.any(Object),
    );
    // No connections loaded (a failed load looks the same): the password dialog.
    expect(dialogSpy.mock.calls[0][1]).not.toHaveProperty("data");
  });

  it("should delete account, dispatch logout, and navigate on confirmation", () => {
    const matDialog = TestBed.inject(MatDialog);
    const store = TestBed.inject(Store);
    const router = TestBed.inject(Router);
    const userService = TestBed.inject(UserService);

    const dialogRefMock = {
      afterClosed: () => of({ password: "userpassword" }),
      componentInstance: {},
    } as any;

    jest.spyOn(matDialog, "open").mockReturnValue(dialogRefMock);
    const deleteAccountSpy = jest.spyOn(userService, "deleteAccount").mockReturnValue(of(undefined) as any);
    const dispatchSpy = jest.spyOn(store, "dispatch").mockReturnValue(of(undefined));
    const navigateSpy = jest.spyOn(router, "navigate").mockResolvedValue(true);

    component.deleteAccount();

    expect(deleteAccountSpy).toHaveBeenCalledWith({ password: "userpassword" });
    expect(dispatchSpy).toHaveBeenCalledWith(expect.any(Logout));
    expect(navigateSpy).toHaveBeenCalledWith(["/"]);
  });

  it("should re-open dialog when deleteAccount returns 401", () => {
    const matDialog = TestBed.inject(MatDialog);
    const store = TestBed.inject(Store);
    const userService = TestBed.inject(UserService);

    const passwordDialogRef = {
      afterClosed: () => of({ password: "userpassword" }),
      componentInstance: {},
    } as any;
    const cancelledDialogRef = {
      afterClosed: () => of(false),
      componentInstance: {},
    } as any;

    const dialogSpy = jest.spyOn(matDialog, "open")
      .mockReturnValueOnce(passwordDialogRef)
      .mockReturnValueOnce(cancelledDialogRef);
    jest.spyOn(userService, "deleteAccount").mockReturnValue(
      throwError(() => ({ status: 401, error: { errorMsg: "Error deleting account." } }))
    );
    const dispatchSpy = jest.spyOn(store, "dispatch");

    component.deleteAccount();

    expect(dialogSpy).toHaveBeenCalledTimes(2);
    expect(dispatchSpy).not.toHaveBeenCalledWith(expect.any(Logout));
  });

  it("keeps the password dialog for an account whose connections were linked, not provisioned", () => {
    const matDialog = TestBed.inject(MatDialog);
    const userService = TestBed.inject(UserService);
    const store = TestBed.inject(Store);
    component.connections.set([connection({ provisionedUser: false })]);

    const dialogSpy = jest.spyOn(matDialog, "open").mockReturnValue({
      afterClosed: () => of({ password: "userpassword" }),
      componentInstance: {},
    } as any);
    const deleteAccountSpy = jest.spyOn(userService, "deleteAccount").mockReturnValue(of(undefined) as any);
    jest.spyOn(store, "dispatch").mockReturnValue(of(undefined));

    component.deleteAccount();

    expect(component.isPasswordless()).toBe(false);
    expect(dialogSpy.mock.calls[0][1]).not.toHaveProperty("data");
    expect(deleteAccountSpy).toHaveBeenCalledWith({ password: "userpassword" });
  });

  it("offers a passwordless account its providers and navigates to the chosen one", () => {
    const matDialog = TestBed.inject(MatDialog);
    const userService = TestBed.inject(UserService);
    component.connections.set([
      connection({ providerName: "my idp", providerDisplayName: "My IdP" }),
      connection({ providerName: "google", providerDisplayName: "Google" }),
    ]);

    const dialogSpy = jest.spyOn(matDialog, "open").mockReturnValue({
      afterClosed: () => of({ providerName: "my idp" }),
      componentInstance: {},
    } as any);
    const navigateSpy = jest
      .spyOn(component as any, "navigateToUrl")
      .mockImplementation(() => undefined);
    const deleteAccountSpy = jest.spyOn(userService, "deleteAccount");

    component.deleteAccount();

    expect(component.isPasswordless()).toBe(true);
    expect(dialogSpy).toHaveBeenCalledWith(
      DeleteAccountDialogComponent,
      expect.objectContaining({
        data: {
          providers: [
            { name: "my idp", displayName: "My IdP" },
            { name: "google", displayName: "Google" },
          ],
        },
      }),
    );
    expect(navigateSpy).toHaveBeenCalledWith(
      "/api/oidc/connections/my%20idp/delete-confirmation"
    );
    // The code comes back on the redirect; nothing is spent yet.
    expect(deleteAccountSpy).not.toHaveBeenCalled();
  });

  it("escapes a provider name in the delete-confirmation URL", () => {
    expect(buildOidcDeleteConfirmationUrl("my provider/../evil")).toBe(
      "/api/oidc/connections/my%20provider%2F..%2Fevil/delete-confirmation"
    );
  });

  it("navigates the whole page to the link URL when connecting a provider", () => {
    const navigateSpy = jest
      .spyOn(component as any, "navigateToUrl")
      .mockImplementation(() => undefined);

    component.connectProvider("google");

    expect(navigateSpy).toHaveBeenCalledWith("/api/oidc/link/google");
  });

  describe("returning from a delete confirmation", () => {
    const stripParams = {
      queryParams: { deleteConfirmation: null, deleteError: null },
      queryParamsHandling: "merge",
      replaceUrl: true,
    };

    function initWithQueryParams(queryParams: Record<string, string>): void {
      const route = TestBed.inject(ActivatedRoute) as any;
      route.snapshot.queryParams = queryParams;
      jest
        .spyOn(TestBed.inject(OidcService), "getOidcConnections")
        .mockReturnValue(of([]) as any);

      component.ngOnInit();
    }

    it("strips the code, spends it, signs out and leaves the page", () => {
      const store = TestBed.inject(Store);
      const router = TestBed.inject(Router);
      const route = TestBed.inject(ActivatedRoute);
      const deleteAccountSpy = jest
        .spyOn(TestBed.inject(UserService), "deleteAccount")
        .mockReturnValue(of(undefined) as any);
      const dispatchSpy = jest.spyOn(store, "dispatch").mockReturnValue(of(undefined));
      const successSpy = jest
        .spyOn(TestBed.inject(SnackbarService), "success")
        .mockImplementation(() => undefined as any);

      initWithQueryParams({ tab: "user-profile", deleteConfirmation: "the-code" });

      expect(router.navigate).toHaveBeenCalledWith([], { relativeTo: route, ...stripParams });
      expect(deleteAccountSpy).toHaveBeenCalledWith({ confirmationCode: "the-code" });
      // Stripped before spending, so a reload cannot resubmit a spent code.
      expect((router.navigate as jest.Mock).mock.invocationCallOrder[0]).toBeLessThan(
        deleteAccountSpy.mock.invocationCallOrder[0]
      );
      expect(dispatchSpy).toHaveBeenCalledWith(expect.any(Logout));
      expect(successSpy).toHaveBeenCalledWith("Your account has been successfully deleted");
      expect(router.navigate).toHaveBeenLastCalledWith(["/"]);
    });

    it("stays signed in when the code is refused", () => {
      const store = TestBed.inject(Store);
      const router = TestBed.inject(Router);
      jest
        .spyOn(TestBed.inject(UserService), "deleteAccount")
        .mockReturnValue(throwError(() => ({ status: 400 })));
      const dispatchSpy = jest.spyOn(store, "dispatch");

      initWithQueryParams({ deleteConfirmation: "spent-code" });

      expect(dispatchSpy).not.toHaveBeenCalledWith(expect.any(Logout));
      expect(router.navigate).not.toHaveBeenCalledWith(["/"]);
    });

    it.each([
      ["identity_mismatch", "You signed in with a different account than the one connected here."],
      ["stale_login", "Your provider didn't ask you to sign in again. Sign out of it and try again."],
      ["not_linked", "That provider is no longer connected to your account."],
      ["forbidden", "Your account can't be deleted this way."],
      ["provider_error", "Could not confirm the account deletion. Please try again."],
      ["constructor", "Could not confirm the account deletion. Please try again."],
    ])("strips deleteError=%s and reports it", (code, message) => {
      const router = TestBed.inject(Router);
      const route = TestBed.inject(ActivatedRoute);
      const deleteAccountSpy = jest.spyOn(TestBed.inject(UserService), "deleteAccount");
      const errorSpy = jest
        .spyOn(TestBed.inject(SnackbarService), "error")
        .mockImplementation(() => undefined as any);

      initWithQueryParams({ deleteError: code });

      expect(router.navigate).toHaveBeenCalledWith([], { relativeTo: route, ...stripParams });
      expect(errorSpy).toHaveBeenCalledWith(message);
      expect(deleteAccountSpy).not.toHaveBeenCalled();
    });

    it("leaves the URL alone without either parameter", () => {
      const router = TestBed.inject(Router);

      initWithQueryParams({ tab: "user-profile" });

      expect(router.navigate).not.toHaveBeenCalled();
    });
  });

  it("should not call deleteAccount API when dialog is cancelled", () => {
    const matDialog = TestBed.inject(MatDialog);
    const userService = TestBed.inject(UserService);

    const dialogRefMock = {
      afterClosed: () => of(false),
      componentInstance: {},
    } as any;

    jest.spyOn(matDialog, "open").mockReturnValue(dialogRefMock);
    const deleteAccountSpy = jest.spyOn(userService, "deleteAccount");

    component.deleteAccount();

    expect(deleteAccountSpy).not.toHaveBeenCalled();
  });
});
