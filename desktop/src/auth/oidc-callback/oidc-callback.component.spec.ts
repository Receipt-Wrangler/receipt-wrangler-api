import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { ActivatedRoute, Router } from "@angular/router";
import { RouterTestingModule } from "@angular/router/testing";
import { NgxsModule } from "@ngxs/store";
import { of, throwError } from "rxjs";
import { ApiModule, UserService } from "../../open-api";
import { AuthState } from "../../store/auth.state";
import { FeatureConfigState } from "../../store/feature-config.state";
import { GroupState } from "../../store/group.state";
import { UserState } from "../../store/user.state";
import { OidcCallbackComponent } from "./oidc-callback.component";

describe("OidcCallbackComponent", () => {
  let fixture: ComponentFixture<OidcCallbackComponent>;
  let router: Router;
  let userService: UserService;

  const appData: any = {
    claims: { userId: 1, username: "u", displayName: "U" },
    groups: [],
    users: [],
    featureConfig: { enableLocalSignUp: false, aiPoweredReceipts: false },
    categories: [],
    tags: [],
    appPermissions: [],
    groupPermissions: {},
    userPreferences: {},
    icons: [],
    about: {},
    currencyDisplay: "$",
  };

  async function setup(queryParams: Record<string, string> = {}): Promise<void> {
    await TestBed.configureTestingModule({
      declarations: [OidcCallbackComponent],
      imports: [
        ApiModule,
        NgxsModule.forRoot([AuthState, FeatureConfigState, GroupState, UserState]),
        NoopAnimationsModule,
        RouterTestingModule,
      ],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParams } },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    router = TestBed.inject(Router);
    userService = TestBed.inject(UserService);
    jest.spyOn(router, "navigate").mockResolvedValue(true);

    fixture = TestBed.createComponent(OidcCallbackComponent);
  }

  afterEach(() => {
    TestBed.resetTestingModule();
    jest.restoreAllMocks();
  });

  // This is the whole reason the route exists: AppInitService skips loading app
  // data when the store is empty, so a first-ever OIDC login would otherwise
  // boot signed-out while holding perfectly valid cookies.
  it("loads app data and navigates to the dashboard", async () => {
    await setup();
    const getAppData = jest
      .spyOn(userService, "getAppData")
      .mockReturnValue(of(appData) as any);

    fixture.detectChanges();
    await fixture.whenStable();

    expect(getAppData).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalled();
  });

  it("routes back to login when app data cannot be loaded", async () => {
    await setup();
    jest
      .spyOn(userService, "getAppData")
      .mockReturnValue(throwError(() => new Error("nope")) as any);

    fixture.detectChanges();
    await fixture.whenStable();

    expect(router.navigate).toHaveBeenCalledWith(["/auth/login"], {
      queryParams: { oidcError: "invalid_state" },
    });
  });

  it("forwards an error code to the login page without calling the API", async () => {
    await setup({ oidcError: "no_account" });
    const getAppData = jest.spyOn(userService, "getAppData");

    fixture.detectChanges();
    await fixture.whenStable();

    expect(getAppData).not.toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(["/auth/login"], {
      queryParams: { oidcError: "no_account" },
    });
  });
});
