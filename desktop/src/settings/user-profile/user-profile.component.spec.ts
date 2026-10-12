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
import { ApiModule, Permission, UserService } from "../../open-api";
import { TokenRefreshService } from "../../services";
import { AuthState, FeatureConfigState, Logout, UserState } from "../../store";
import { SetPermissions } from "../../store/auth.state.actions";
import { UserProfileComponent } from "./user-profile.component";
import { DeleteAccountDialogComponent } from "../delete-account-dialog/delete-account-dialog.component";

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
  });

  it("should delete account, dispatch logout, and navigate on confirmation", () => {
    const matDialog = TestBed.inject(MatDialog);
    const store = TestBed.inject(Store);
    const router = TestBed.inject(Router);
    const userService = TestBed.inject(UserService);

    const dialogRefMock = {
      afterClosed: () => of("userpassword"),
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
      afterClosed: () => of("userpassword"),
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
