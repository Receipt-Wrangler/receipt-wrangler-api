import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA, provideZonelessChangeDetection } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { FormArray, FormControl, FormGroup, ReactiveFormsModule } from "@angular/forms";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { ActivatedRoute, Router } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { of, throwError } from "rxjs";
import { AutocompleteModule } from "../../autocomplete/autocomplete.module";
import { CheckboxModule } from "../../checkbox/checkbox.module";
import { InputModule } from "../../input/index";
import { FormMode } from "../../enums/form-mode.enum";
import { CurrencySeparator, CurrencySymbolPosition, FeatureConfigService, QueueName, SystemSettingsService } from "../../open-api";
import { PipesModule } from "../../pipes";
import { CustomCurrencyPipe } from "../../pipes/custom-currency.pipe";
import { SnackbarService } from "../../services";
import { SharedUiModule } from "../../shared-ui/shared-ui.module";
import { AuthState } from "../../store/auth.state";
import { SystemSettingsState } from "../../store/system-settings.state";
import { SetTimeZone } from "../../store/system-settings.state.actions";
import { TaskQueueFormControlPipe } from "../pipes/task-queue-form-control.pipe";

import { SystemSettingsFormComponent } from "./system-settings-form.component";
import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";

describe("SystemSettingsFormComponent", () => {
  let component: SystemSettingsFormComponent;
  let fixture: ComponentFixture<SystemSettingsFormComponent>;
  let store: Store;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [SystemSettingsFormComponent, CustomCurrencyPipe, TaskQueueFormControlPipe],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    imports: [AutocompleteModule,
        CheckboxModule,
        InputModule,
        NgxsModule.forRoot([SystemSettingsState, AuthState]),
        PipesModule,
        ReactiveFormsModule,
        SharedUiModule,
        NoopAnimationsModule],
    providers: [
        provideZonelessChangeDetection(),
        CustomCurrencyPipe,
        {
            provide: ActivatedRoute,
            useValue: {
                snapshot: {
                    data: {
                        allReceiptProcessingSettings: [],
                        systemSettings: {
                            taskQueueConfigurations: [
                                { name: "email_polling", priority: 1 },
                                { name: "email_receipt_processing", priority: 1 },
                                { name: "email_receipt_image_cleanup", priority: 1 },
                                { name: "quick_scan", priority: 1 }
                            ]
                        },
                        formConfig: {}
                    }
                }
            }
        },
        { provide: Router, useValue: { navigate: jest.fn().mockResolvedValue(true) } },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting()
    ]
})
      .compileComponents();

    store = TestBed.inject(Store);
    fixture = TestBed.createComponent(SystemSettingsFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  describe("layout contracts", () => {
    // Both classes below are load-bearing and silently droppable: nothing fails
    // to compile without them, the page just goes back to overlapping itself.
    it("marks both duration rows so the flex-basis rule can reach them", () => {
      const rows: NodeListOf<HTMLElement> =
        fixture.nativeElement.querySelectorAll(".duration-row");

      // Session and MCP Server. `.duration-row > *` in the component's SCSS is
      // what bases each column at 0; without the class they fall back to
      // flex-basis: auto and the hinted column swallows the row.
      expect(rows.length).toEqual(3);
      rows.forEach((row) => expect(row.classList).toContain("d-flex"));
    });

    it("pairs every host margin with d-block", () => {
      const hosts: NodeListOf<HTMLElement> =
        fixture.nativeElement.querySelectorAll("app-checkbox.mb-3, app-input.mb-3");

      // app-checkbox and app-input have no :host { display: block }, so a bare
      // mb-3 on the host is an inline element's vertical margin — it does
      // nothing at all until d-block comes with it.
      expect(hosts.length).toBeGreaterThan(0);
      hosts.forEach((host) => expect(host.classList).toContain("d-block"));
    });
  });

  it("init form with no data", () => {
    component.ngOnInit();

    expect(component.form.value).toEqual({
      enableLocalSignUp: null,
      debugOcr: null,
      currencyDisplay: null,
      emailPollingInterval: null,
      receiptProcessingSettingsId: null,
      fallbackReceiptProcessingSettingsId: null,
      currencyThousandthsSeparator: null,
      currencyDecimalSeparator: null,
      currencySymbolPosition: null,
      currencyHideDecimalPlaces: null,
      // Unset on the wire (an older server): seeds the server's own default.
      timeZone: "UTC",
      pdfDpi: null,
      taskConcurrency: null,
      taskQueueConfigurations: [
        { name: "email_polling", priority: 1 },
        { name: "email_receipt_processing", priority: 1 },
        { name: "email_receipt_image_cleanup", priority: 1 },
        { name: "quick_scan", priority: 1 }
      ],
      mcpEnabled: null,
      mcpPublicUrl: null,
      serverPublicUrl: null,
      showLoginQr: null,
      mobileServerUrl: null,
      // With no stored value, splitHours falls back to the 24h default and
      // renders it as the friendlier "1 Days".
      refreshTokenValidForValue: 1,
      refreshTokenValidForUnit: "DAYS",
      mcpRefreshTokenValidForValue: 1,
      mcpRefreshTokenValidForUnit: "DAYS",
      // Unset on the wire, so it renders its own 720h default rather than
      // splitHours' generic 24h fallback.
      tempFileRetentionValue: 30,
      tempFileRetentionUnit: "DAYS",
    });
  });

  it("init form with data", () => {
    const activatedRoute = TestBed.inject(ActivatedRoute);
    activatedRoute.snapshot.data["systemSettings"] = {
      enableLocalSignUp: true,
      debugOcr: true,
      currencyDisplay: "USD",
      emailPollingInterval: 5,
      receiptProcessingSettingsId: 1,
      fallbackReceiptProcessingSettingsId: 2,
      currencyThousandthsSeparator: CurrencySeparator.Comma,
      currencyDecimalSeparator: CurrencySeparator.Period,
      currencySymbolPosition: CurrencySymbolPosition.Start,
      currencyHideDecimalPlaces: true,
      timeZone: "America/New_York",
      pdfDpi: 300,
      taskConcurrency: 12,
      taskQueueConfigurations: [{
        name: QueueName.QuickScan,
        priority: 1,
      }],
      mcpEnabled: true,
      mcpPublicUrl: "https://receipts.example.com",
      serverPublicUrl: "https://receipts.example.com",
      showLoginQr: true,
      mobileServerUrl: "https://receipts.example.com/api",
      refreshTokenValidForHours: 720,
      mcpRefreshTokenValidForHours: 6,
    };

    component.ngOnInit();

    expect(component.form.getRawValue()).toEqual({
      enableLocalSignUp: true,
      debugOcr: true,
      currencyDisplay: "USD",
      emailPollingInterval: 5,
      receiptProcessingSettingsId: 1,
      fallbackReceiptProcessingSettingsId: 2,
      currencyThousandthsSeparator: CurrencySeparator.Comma,
      currencyDecimalSeparator: CurrencySeparator.Period,
      currencySymbolPosition: CurrencySymbolPosition.Start,
      currencyHideDecimalPlaces: true,
      timeZone: "America/New_York",
      pdfDpi: 300,
      taskConcurrency: 12,
      taskQueueConfigurations: [{
        name: QueueName.QuickScan,
        priority: 1,
      }],
      mcpEnabled: true,
      mcpPublicUrl: "https://receipts.example.com",
      serverPublicUrl: "https://receipts.example.com",
      showLoginQr: true,
      mobileServerUrl: "https://receipts.example.com/api",
      // 720 hours divides evenly into days, so it renders as 30 Days; 6 does
      // not, so it stays in hours.
      refreshTokenValidForValue: 30,
      refreshTokenValidForUnit: "DAYS",
      mcpRefreshTokenValidForValue: 6,
      mcpRefreshTokenValidForUnit: "HOURS",
      tempFileRetentionValue: 30,
      tempFileRetentionUnit: "DAYS",
    });
  });

  it("requires a mobile server url only when the login QR is enabled", () => {
    component.ngOnInit();

    const showLoginQr = component.form.get("showLoginQr")!;
    const mobileServerUrl = component.form.get("mobileServerUrl")!;

    // Off by default: an empty url is valid.
    showLoginQr.setValue(false);
    mobileServerUrl.setValue("");
    expect(mobileServerUrl.valid).toBe(true);

    // Enabled: an empty url is now required (invalid).
    showLoginQr.setValue(true);
    mobileServerUrl.setValue("");
    expect(mobileServerUrl.hasError("required")).toBe(true);

    // Enabled with a url is valid again.
    mobileServerUrl.setValue("https://receipts.example.com/api");
    expect(mobileServerUrl.valid).toBe(true);

    // Disabling clears the requirement.
    showLoginQr.setValue(false);
    mobileServerUrl.setValue("");
    expect(mobileServerUrl.valid).toBe(true);
  });

  it("rejects a mobile server url that is not an absolute http(s) url", () => {
    component.ngOnInit();

    const showLoginQr = component.form.get("showLoginQr")!;
    const mobileServerUrl = component.form.get("mobileServerUrl")!;
    showLoginQr.setValue(true);

    // Whitespace satisfies Validators.required but the backend trims first, so
    // it would be rejected server-side as missing.
    mobileServerUrl.setValue("   ");
    expect(mobileServerUrl.hasError("url")).toBe(true);

    mobileServerUrl.setValue("receipts.example.com/api");
    expect(mobileServerUrl.hasError("url")).toBe(true);

    mobileServerUrl.setValue("ftp://receipts.example.com");
    expect(mobileServerUrl.hasError("url")).toBe(true);

    // Credentials would be published verbatim by the public login QR.
    mobileServerUrl.setValue("https://user:token@receipts.example.com/api");
    expect(mobileServerUrl.hasError("url")).toBe(true);

    // http stays valid -- LAN / bare-IP self-hosting is supported.
    mobileServerUrl.setValue("http://192.168.1.50:8081/api");
    expect(mobileServerUrl.valid).toBe(true);

    mobileServerUrl.setValue("https://receipts.example.com/api");
    expect(mobileServerUrl.valid).toBe(true);

    // The format check applies with the toggle off too, mirroring the backend,
    // which validates any non-empty url regardless of the toggle.
    showLoginQr.setValue(false);
    mobileServerUrl.setValue("ftp://receipts.example.com");
    expect(mobileServerUrl.hasError("url")).toBe(true);
  });

  // `new URL()` is more lenient than Go's `url.Parse`: it normalizes these three
  // authority-less forms into a valid URL, while `url.Parse` leaves Host empty
  // and the backend rejects them. Without a literal `http(s)://` prefix check
  // the form would accept a value the server 400s on -- exactly what this
  // validator exists to prevent.
  it("rejects url forms that the backend rejects", () => {
    component.ngOnInit();

    const showLoginQr = component.form.get("showLoginQr")!;
    const mobileServerUrl = component.form.get("mobileServerUrl")!;
    showLoginQr.setValue(true);

    for (const value of [
      "https:receipts.example.com/api",
      "https:/receipts.example.com/api",
      "https:\\\\receipts.example.com/api",
    ]) {
      mobileServerUrl.setValue(value);
      expect(mobileServerUrl.hasError("url")).toBe(true);
    }

    // But the scheme is case-insensitive server-side (`url.Parse` lowercases
    // it), so an uppercase scheme must stay VALID -- a case-sensitive prefix
    // check would introduce the very mismatch above in the other direction.
    mobileServerUrl.setValue("HTTPS://receipts.example.com/api");
    expect(mobileServerUrl.valid).toBe(true);
  });

  it("rejects an mcp public url that is not an absolute http(s) url", () => {
    component.ngOnInit();

    const mcpEnabled = component.form.get("mcpEnabled")!;
    const mcpPublicUrl = component.form.get("mcpPublicUrl")!;
    mcpEnabled.setValue(true);

    mcpPublicUrl.setValue("   ");
    expect(mcpPublicUrl.hasError("url")).toBe(true);

    mcpPublicUrl.setValue("receipts.example.com");
    expect(mcpPublicUrl.hasError("url")).toBe(true);

    mcpPublicUrl.setValue("https://user:token@receipts.example.com");
    expect(mcpPublicUrl.hasError("url")).toBe(true);

    // The dev default must keep validating.
    mcpPublicUrl.setValue("http://localhost:8081");
    expect(mcpPublicUrl.valid).toBe(true);

    mcpPublicUrl.setValue("https://receipts.example.com");
    expect(mcpPublicUrl.valid).toBe(true);
  });

  it("caps the session lifetime at 720 hours or 30 days depending on the unit", () => {
    component.ngOnInit();

    const value = component.form.get("refreshTokenValidForValue")!;
    const unit = component.form.get("refreshTokenValidForUnit")!;

    unit.setValue("HOURS");
    value.setValue(720);
    expect(value.valid).toBe(true);

    value.setValue(721);
    expect(value.valid).toBe(false);

    // Flipping the unit has to retarget the max, or 30 days would be rejected
    // as if it were 30 hours over the limit.
    unit.setValue("DAYS");
    value.setValue(30);
    expect(value.valid).toBe(true);

    value.setValue(31);
    expect(value.valid).toBe(false);
  });

  it("caps temp file retention at 8760 hours or 365 days depending on the unit", () => {
    component.ngOnInit();

    const value = component.form.get("tempFileRetentionValue")!;
    const unit = component.form.get("tempFileRetentionUnit")!;

    unit.setValue("HOURS");
    value.setValue(8760);
    expect(value.valid).toBe(true);

    value.setValue(8761);
    expect(value.valid).toBe(false);

    unit.setValue("DAYS");
    value.setValue(365);
    expect(value.valid).toBe(true);

    value.setValue(366);
    expect(value.valid).toBe(false);
  });

  // The server floors this at 24 hours. Without a client-side floor the entry
  // passes validation and comes back as a bare 400 with nothing on the field.
  it("enforces the 24 hour retention floor in whichever unit is selected", () => {
    component.ngOnInit();

    const value = component.form.get("tempFileRetentionValue")!;
    const unit = component.form.get("tempFileRetentionUnit")!;

    unit.setValue("HOURS");
    value.setValue(23);
    expect(value.valid).toBe(false);
    expect(value.errors?.["duration"]).toBe("Must be at least 24.");

    value.setValue(24);
    expect(value.valid).toBe(true);

    unit.setValue("DAYS");
    value.setValue(1);
    expect(value.valid).toBe(true);
  });

  it("bounds the retention and the session lifetime independently", () => {
    component.ngOnInit();

    component.form.get("tempFileRetentionUnit")!.setValue("HOURS");
    component.form.get("tempFileRetentionValue")!.setValue(5000);
    expect(component.form.get("tempFileRetentionValue")!.valid).toBe(true);

    // The session lifetime keeps its own, much lower cap.
    component.form.get("refreshTokenValidForUnit")!.setValue("HOURS");
    component.form.get("refreshTokenValidForValue")!.setValue(5000);
    expect(component.form.get("refreshTokenValidForValue")!.valid).toBe(false);
  });

  it("requires a positive whole-number session lifetime", () => {
    component.ngOnInit();

    const value = component.form.get("refreshTokenValidForValue")!;

    value.setValue(0);
    expect(value.valid).toBe(false);

    value.setValue(null);
    expect(value.valid).toBe(false);

    // The API stores this as a Go int, so a fractional value would fail
    // json.Unmarshal instead of coming back as a field error.
    value.setValue(1.5);
    expect(value.valid).toBe(false);

    value.setValue(1);
    expect(value.valid).toBe(true);
  });

  it("surfaces a readable message when the lifetime is out of range", () => {
    component.ngOnInit();

    const value = component.form.get("refreshTokenValidForValue")!;
    component.form.get("refreshTokenValidForUnit")!.setValue("DAYS");

    value.setValue(31);

    // The message rides on the error value so BaseInputComponent renders it —
    // Validators.max has no message mapping and would show a blank error.
    expect(value.errors?.["duration"]).toBe("Must be at most 30.");
  });

  it("caps the mcp connector lifetime independently of the session lifetime", () => {
    component.ngOnInit();

    component.form.get("mcpRefreshTokenValidForUnit")!.setValue("DAYS");
    const mcpValue = component.form.get("mcpRefreshTokenValidForValue")!;

    mcpValue.setValue(31);
    expect(mcpValue.valid).toBe(false);

    mcpValue.setValue(30);
    expect(mcpValue.valid).toBe(true);

    // The app-session control keeps its own unit and bound.
    expect(component.form.get("refreshTokenValidForUnit")!.value).toBe("DAYS");
  });

  it("should submit form", () => {
    const systemSettingsService = TestBed.inject(SystemSettingsService);
    const snackbarService = TestBed.inject(SnackbarService);
    const router = TestBed.inject(Router);

    const updateSystemSettingsSpy = jest.spyOn(systemSettingsService, "updateSystemSettings").mockReturnValue(of(null as any));
    const snackbarServiceSpy = jest.spyOn(snackbarService, "success");
    const routerSpy = jest.spyOn(router, "navigate");

    component.originalSystemSettings.taskQueueConfigurations = [{
      name: QueueName.QuickScan,
      priority: "1",
    } as any];

    component.form.patchValue({
      enableLocalSignUp: true,
      debugOcr: true,
      currencyDisplay: "USD",
      emailPollingInterval: "5",
      receiptProcessingSettingsId: 1,
      fallbackReceiptProcessingSettingsId: 2,
      currencyThousandthsSeparator: CurrencySeparator.Comma,
      currencyDecimalSeparator: CurrencySeparator.Period,
      currencySymbolPosition: CurrencySymbolPosition.Start,
      currencyHideDecimalPlaces: false,
      timeZone: "Europe/Berlin",
      pdfDpi: "300",
      taskConcurrency: "12",
      mcpEnabled: true,
      mcpPublicUrl: "https://receipts.example.com",
      serverPublicUrl: "https://receipts.example.com",
      showLoginQr: true,
      mobileServerUrl: "https://receipts.example.com/api",
      refreshTokenValidForValue: "14",
      refreshTokenValidForUnit: "DAYS",
      mcpRefreshTokenValidForValue: "12",
      mcpRefreshTokenValidForUnit: "HOURS",
    });

    // Update the quick_scan queue priority specifically
    const queueArray = component.form.get("taskQueueConfigurations") as FormArray;
    const quickScanIndex = queueArray.controls.findIndex(control => 
      control.get('name')?.value === 'quick_scan'
    );
    if (quickScanIndex >= 0) {
      queueArray.at(quickScanIndex).get('priority')?.setValue("1");
    }

    component.submit();

    expect(updateSystemSettingsSpy).toHaveBeenCalledWith({
      enableLocalSignUp: true,
      debugOcr: true,
      currencyDisplay: "USD",
      emailPollingInterval: 5,
      receiptProcessingSettingsId: 1,
      fallbackReceiptProcessingSettingsId: 2,
      currencyThousandthsSeparator: CurrencySeparator.Comma,
      currencyDecimalSeparator: CurrencySeparator.Period,
      currencySymbolPosition: CurrencySymbolPosition.Start,
      currencyHideDecimalPlaces: false,
      timeZone: "Europe/Berlin",
      pdfDpi: 300,
      taskConcurrency: 12,
      taskQueueConfigurations: [
        { name: 'email_polling', priority: 1 },
        { name: 'email_receipt_processing', priority: 1 },
        { name: 'email_receipt_image_cleanup', priority: 1 },
        { name: 'quick_scan', priority: 1 }
      ],
      mcpEnabled: true,
      mcpPublicUrl: "https://receipts.example.com",
      serverPublicUrl: "https://receipts.example.com",
      showLoginQr: true,
      mobileServerUrl: "https://receipts.example.com/api",
      // The value/unit pairs are folded into hours and the presentation-only
      // controls are stripped — toEqual would fail if any of them leaked.
      refreshTokenValidForHours: 336,
      mcpRefreshTokenValidForHours: 12,
      tempFileRetentionHours: 720,
    });

    expect(snackbarServiceSpy).toHaveBeenCalled();
    expect(routerSpy).toHaveBeenCalled();
  });

  describe("time zone", () => {
    function setRouteSettings(settings: object): void {
      const activatedRoute = TestBed.inject(ActivatedRoute);
      activatedRoute.snapshot.data["systemSettings"] = {
        taskQueueConfigurations: [],
        ...settings,
      };
    }

    it("seeds the stored zone", () => {
      setRouteSettings({ timeZone: "Asia/Tokyo" });
      component.ngOnInit();

      expect(component.form.get("timeZone")!.value).toEqual("Asia/Tokyo");
    });

    it("is required", () => {
      component.ngOnInit();
      const timeZone = component.form.get("timeZone")!;

      timeZone.setValue(null);
      expect(timeZone.hasError("required")).toBe(true);

      timeZone.setValue("America/New_York");
      expect(timeZone.valid).toBe(true);
    });

    it("offers UTC first, then the browser's IANA zones", () => {
      expect(component.timeZones[0]).toEqual("UTC");
      expect(component.timeZones.filter((zone) => zone === "UTC").length).toEqual(1);
    });

    it("shows the picked zone's name rather than the autocomplete's blank default", () => {
      expect(component.timeZoneDisplayWith("Europe/Berlin")).toEqual("Europe/Berlin");
      expect(component.timeZoneDisplayWith(null)).toEqual("");
    });

    it("renders the picker with its test id", () => {
      const picker: HTMLElement | null = fixture.nativeElement.querySelector(
        '[data-testid="system-settings-time-zone"]'
      );

      expect(picker).not.toBeNull();
      expect(picker!.tagName.toLowerCase()).toEqual("app-autocomlete");
    });

    it("is disabled in view mode but still read by getRawValue", () => {
      setRouteSettings({ timeZone: "America/New_York" });
      const activatedRoute = TestBed.inject(ActivatedRoute);
      activatedRoute.snapshot.data["formConfig"] = { mode: FormMode.view };
      component.ngOnInit();

      expect(component.form.get("timeZone")!.disabled).toBe(true);
      expect(component.form.getRawValue().timeZone).toEqual("America/New_York");
    });

    it("dispatches the saved zone once the save succeeds", () => {
      const systemSettingsService = TestBed.inject(SystemSettingsService);
      const featureConfigService = TestBed.inject(FeatureConfigService);
      jest.spyOn(systemSettingsService, "updateSystemSettings").mockReturnValue(of(null as any));
      jest.spyOn(featureConfigService, "getFeatureConfig").mockReturnValue(of({} as any));
      const dispatchSpy = jest.spyOn(store, "dispatch");

      setRouteSettings({ timeZone: "UTC" });
      component.ngOnInit();
      component.form.get("timeZone")!.setValue("America/New_York");
      component.submit();

      expect(systemSettingsService.updateSystemSettings).toHaveBeenCalledWith(
        expect.objectContaining({ timeZone: "America/New_York" })
      );
      expect(dispatchSpy).toHaveBeenCalledWith(new SetTimeZone("America/New_York"));
      expect(store.selectSnapshot(SystemSettingsState.timeZone)).toEqual("America/New_York");
    });

    it("does not dispatch a zone the server rejected", () => {
      const systemSettingsService = TestBed.inject(SystemSettingsService);
      jest
        .spyOn(systemSettingsService, "updateSystemSettings")
        .mockReturnValue(throwError(() => new Error("400")) as any);
      const dispatchSpy = jest.spyOn(store, "dispatch");

      component.ngOnInit();
      component.form.get("timeZone")!.setValue("America/New_York");
      component.submit();

      expect(dispatchSpy).not.toHaveBeenCalledWith(new SetTimeZone("America/New_York"));
    });
  });
});
