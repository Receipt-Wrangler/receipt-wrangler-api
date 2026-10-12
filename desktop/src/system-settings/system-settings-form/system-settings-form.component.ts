import { AfterViewInit, Component, ElementRef, OnInit, viewChild } from "@angular/core";
import { FormArray, FormBuilder, FormGroup, ValidatorFn, Validators } from "@angular/forms";
import { ActivatedRoute, Router } from "@angular/router";
import { UntilDestroy, untilDestroyed } from "@ngneat/until-destroy";
import { Store } from "@ngxs/store";
import { startWith, switchMap, take, tap } from "rxjs";
import { fadeInOut } from "../../animations/index";
import { AutocomleteComponent } from "../../autocomplete/autocomlete/autocomlete.component";
import { BaseFormComponent } from "../../form";
import { FormOption } from "../../interfaces/form-option.interface";
import {
  CurrencySeparator,
  CurrencySymbolPosition,
  FeatureConfigService,
  QueueName,
  ReceiptProcessingSettings,
  SystemSettings,
  SystemSettingsService
} from "../../open-api";
import { InputReadonlyPipe } from "../../pipes/input-readonly.pipe";
import { SnackbarService } from "../../services";
import { SetFeatureConfig } from "../../store";
import { SetCurrencyData, SetCurrencyDisplay, SetTimeZone } from "../../store/system-settings.state.actions";
import { DurationUnit, maxForUnit, minForUnit, splitHours, toHours } from "../../utils";
import { DEFAULT_APP_TIME_ZONE, timeZoneOptions } from "../../utils/app-time-zone";
import { absoluteUrlValidator, durationValueValidator } from "../../validators";

interface QueueData extends FormOption {
  description: string;
}

@UntilDestroy()
@Component({
  selector: "app-system-settings-form",
  templateUrl: "./system-settings-form.component.html",
  styleUrl: "./system-settings-form.component.scss",
  providers: [InputReadonlyPipe],
  animations: [fadeInOut],
  standalone: false
})
export class SystemSettingsFormComponent extends BaseFormComponent implements OnInit, AfterViewInit {
  public readonly fallbackReceiptProcessingSettings = viewChild.required<AutocomleteComponent>("fallbackReceiptProcessingSettings");

  public readonly alert = viewChild.required("alert", { read: ElementRef });

  public originalSystemSettings!: SystemSettings;

  public allReceiptProcessingSettings: ReceiptProcessingSettings[] = [];

  public filteredReceiptProcessingSettings: ReceiptProcessingSettings[] = [];

  public showRestartTaskServerAlert = false;

  public readonly queueData: QueueData[] = [
    {
      value: QueueName.EmailPolling,
      displayValue: "Email Polling",
      description: "Polls system emails for receipts to process"
    },
    {
      value: QueueName.EmailReceiptProcessing,
      displayValue: "Email Receipt Processing",
      description: "Processing of captured emails"
    },
    {
      value: QueueName.EmailReceiptImageCleanup,
      displayValue: "Email Receipt Image Cleanup (retired)",
      description: "No longer used. Temporary file cleanup moved to System Clean Up."
    },
    {
      value: QueueName.QuickScan,
      displayValue: "Quick Scan",
      description: "Processes quick scan receipts"
    },
    {
      value: QueueName.SystemCleanUp,
      displayValue: "System Clean Up",
      description: "Expires refresh tokens and reclaims temporary files"
    }
  ];

  // The backend caps a refresh-token lifetime at 720 hours (30 days) and treats
  // 0 as "unset, use the default". Mirrored here so the form rejects what the
  // server would reject.
  public static readonly MAX_TOKEN_LIFETIME_HOURS = 720;

  // Temp-file retention has its own, much wider bounds — do not reuse the token
  // cap above. 24 hours is a real floor on the server, so it is enforced here
  // too or the entry fails with a bare 400 and no message on the field.
  public static readonly MIN_TEMP_FILE_RETENTION_HOURS = 24;

  public static readonly MAX_TEMP_FILE_RETENTION_HOURS = 8760;

  public static readonly DEFAULT_TEMP_FILE_RETENTION_HOURS = 720;

  public readonly durationUnits: FormOption[] = [
    {
      displayValue: "Hours",
      value: "HOURS"
    },
    {
      displayValue: "Days",
      value: "DAYS"
    }
  ];

  public readonly symbolPositions: FormOption[] = [
    {
      displayValue: "Start",
      value: CurrencySymbolPosition.Start
    },
    {
      displayValue: "End",
      value: CurrencySymbolPosition.End,
    }
  ];

  public readonly decimalSeparators: FormOption[] = [
    {
      displayValue: ", (Comma)",
      value: CurrencySeparator.Comma
    },
    {
      displayValue: ". (Dot)",
      value: CurrencySeparator.Period
    }
  ];

  // Plain IANA names (~420): the autocomplete filters and renders bare strings
  // itself, so there is no value/display split to model. Built once — the list
  // is static for the browser's lifetime.
  public readonly timeZones: string[] = timeZoneOptions();

  // The autocomplete's default displayWith renders "", which would blank the
  // field once a zone is picked.
  public readonly timeZoneDisplayWith = (zone: string | null): string => zone ?? "";

  constructor(
    private activatedRoute: ActivatedRoute,
    private featureConfigService: FeatureConfigService,
    private formBuilder: FormBuilder,
    private router: Router,
    private snackbarService: SnackbarService,
    private store: Store,
    private systemSettingsService: SystemSettingsService,
    private inputReadonlyPipe: InputReadonlyPipe,
  ) {
    super();
  }

  public ngOnInit(): void {
    this.setFormConfigFromRoute(this.activatedRoute);
    this.allReceiptProcessingSettings = this.activatedRoute.snapshot.data?.["allReceiptProcessingSettings"];
    this.originalSystemSettings = this.activatedRoute.snapshot.data?.["systemSettings"];
    this.showRestartTaskServerAlert = this.activatedRoute.snapshot?.queryParams?.["restartTaskServer"] === "true";
    this.initForm();
  }

  public ngAfterViewInit(): void {
    setTimeout(() => {

      if (this.showRestartTaskServerAlert) {
        this.alert().nativeElement.scrollIntoView({ behavior: "smooth" });
      }

    }, 0);
  }

  private initForm(): void {
    const refreshTokenLifetime = splitHours(this.originalSystemSettings?.refreshTokenValidForHours);
    const mcpRefreshTokenLifetime = splitHours(this.originalSystemSettings?.mcpRefreshTokenValidForHours);
    // splitHours falls back to 24 hours for an unset value, which would render
    // this setting as "1 Days". Its own default is 720, so substitute that
    // first — the server reads 0 the same way.
    const tempFileRetention = splitHours(
      this.originalSystemSettings?.tempFileRetentionHours || SystemSettingsFormComponent.DEFAULT_TEMP_FILE_RETENTION_HOURS
    );

    this.form = this.formBuilder.group({
      enableLocalSignUp: [this.originalSystemSettings?.enableLocalSignUp],
      debugOcr: [this.originalSystemSettings?.debugOcr],
      emailPollingInterval: [this.originalSystemSettings?.emailPollingInterval, [Validators.required, Validators.min(0)]],
      currencyDisplay: [this.originalSystemSettings?.currencyDisplay],
      currencyThousandthsSeparator: [this.originalSystemSettings.currencyThousandthsSeparator, [Validators.required]],
      currencyDecimalSeparator: [this.originalSystemSettings.currencyDecimalSeparator, [Validators.required]],
      currencySymbolPosition: [this.originalSystemSettings.currencySymbolPosition, [Validators.required]],
      currencyHideDecimalPlaces: [this.originalSystemSettings.currencyHideDecimalPlaces],
      timeZone: [this.originalSystemSettings?.timeZone || DEFAULT_APP_TIME_ZONE, [Validators.required]],
      receiptProcessingSettingsId: [this.originalSystemSettings?.receiptProcessingSettingsId],
      fallbackReceiptProcessingSettingsId: [this.originalSystemSettings?.fallbackReceiptProcessingSettingsId],
      pdfDpi: [this.originalSystemSettings?.pdfDpi, [Validators.min(72), Validators.max(1200)]],
      taskConcurrency: [this.originalSystemSettings?.taskConcurrency, [Validators.min(0), Validators.required]],
      taskQueueConfigurations: this.formBuilder.array(this.buildAsynqQueueConfigurations()),
      mcpEnabled: [this.originalSystemSettings?.mcpEnabled],
      mcpPublicUrl: [this.originalSystemSettings?.mcpPublicUrl],
      // Always sent, never conditionally required: the backend accepts a blank
      // value, and the cross-setting rule ("required once a provider is
      // enabled") is enforced where providers are saved so an admin cannot
      // register one that has no redirect URI to publish.
      serverPublicUrl: [this.originalSystemSettings?.serverPublicUrl, [absoluteUrlValidator()]],
      showLoginQr: [this.originalSystemSettings?.showLoginQr],
      mobileServerUrl: [this.originalSystemSettings?.mobileServerUrl],
      // Form-only controls. The API stores hours; the unit is presentation and
      // is folded back into `<name>Hours` in submit().
      refreshTokenValidForValue: [refreshTokenLifetime.value],
      refreshTokenValidForUnit: [refreshTokenLifetime.unit],
      mcpRefreshTokenValidForValue: [mcpRefreshTokenLifetime.value],
      mcpRefreshTokenValidForUnit: [mcpRefreshTokenLifetime.unit],
      tempFileRetentionValue: [tempFileRetention.value],
      tempFileRetentionUnit: [tempFileRetention.unit],
    });

    if (this.inputReadonlyPipe.transform(this.formConfig.mode)) {
      this.form.get("debugOcr")?.disable();
      this.form.get("enableLocalSignUp")?.disable();
      this.form.get("currencyThousandthsSeparator")?.disable();
      this.form.get("currencyDecimalSeparator")?.disable();
      this.form.get("currencySymbolPosition")?.disable();
      this.form.get("currencyHideDecimalPlaces")?.disable();
      this.form.get("timeZone")?.disable();
      this.form.get("mcpEnabled")?.disable();
      this.form.get("mcpPublicUrl")?.disable();
      this.form.get("serverPublicUrl")?.disable();
      this.form.get("showLoginQr")?.disable();
      this.form.get("mobileServerUrl")?.disable();
      this.form.get("refreshTokenValidForValue")?.disable();
      this.form.get("refreshTokenValidForUnit")?.disable();
      this.form.get("mcpRefreshTokenValidForValue")?.disable();
      this.form.get("mcpRefreshTokenValidForUnit")?.disable();
      this.form.get("tempFileRetentionValue")?.disable();
      this.form.get("tempFileRetentionUnit")?.disable();
    }

    this.listenForReceiptProcessingSettingsChanges();
    this.listenForHideDecimalPlacesChanges();
    this.listenForMcpEnabledChanges();
    this.listenForShowLoginQrChanges();
    this.listenForDurationUnitChanges(
      "refreshTokenValidForValue",
      "refreshTokenValidForUnit",
      SystemSettingsFormComponent.MAX_TOKEN_LIFETIME_HOURS
    );
    this.listenForDurationUnitChanges(
      "mcpRefreshTokenValidForValue",
      "mcpRefreshTokenValidForUnit",
      SystemSettingsFormComponent.MAX_TOKEN_LIFETIME_HOURS
    );
    this.listenForDurationUnitChanges(
      "tempFileRetentionValue",
      "tempFileRetentionUnit",
      SystemSettingsFormComponent.MAX_TEMP_FILE_RETENTION_HOURS,
      SystemSettingsFormComponent.MIN_TEMP_FILE_RETENTION_HOURS
    );
  }

  // The bounds depend on the selected unit (720 hours === 30 days), so the value
  // control's validators are re-applied whenever the unit flips. Like
  // urlValidators() below, setValidators replaces the whole list, so required
  // has to be re-supplied each time rather than declared in initForm.
  //
  // The bounds are parameters rather than the token cap this once hardcoded:
  // temp-file retention runs to a year and has a real floor of its own.
  private listenForDurationUnitChanges(
    valueControlName: string,
    unitControlName: string,
    maxHours: number,
    minHours: number = 1
  ): void {
    const valueControl = this.form.get(valueControlName);

    this.form.get(unitControlName)?.valueChanges
      .pipe(
        startWith(this.form.get(unitControlName)?.value),
        untilDestroyed(this),
        tap((unit: DurationUnit) => {
          valueControl?.setValidators([
            Validators.required,
            durationValueValidator(maxForUnit(maxHours, unit), minForUnit(minHours, unit)),
          ]);
          valueControl?.updateValueAndValidity({ emitEvent: false });
        })
      )
      .subscribe();
  }

  // TODO: finish implementing UI for taskQueueConfigurations
  private buildAsynqQueueConfigurations(): FormGroup[] {
    return (this.originalSystemSettings?.taskQueueConfigurations ?? []).map(config => {
      return this.formBuilder.group({
        name: [config.name],
        priority: [config.priority],
      });
    });
  }

  private listenForReceiptProcessingSettingsChanges(): void {
    this.form.get("receiptProcessingSettingsId")?.valueChanges
      .pipe(
        startWith(this.form.get("receiptProcessingSettingsId")?.value),
        untilDestroyed(this),
        tap((value: number) => {
          this.filteredReceiptProcessingSettings = this.allReceiptProcessingSettings.filter((rps) => rps.id !== value);

          if (!value) {
            this.fallbackReceiptProcessingSettings()?.clearFilter();
          }
        })
      )
      .subscribe();
  }

  private listenForHideDecimalPlacesChanges(): void {
    this.form.get("currencyHideDecimalPlaces")?.valueChanges
      .pipe(
        startWith(this.form.get("currencyHideDecimalPlaces")?.value),
        untilDestroyed(this),
        tap((hide: boolean) => {
          if (hide) {
            this.form.get("currencyDecimalSeparator")?.disable();
          } else {
            this.form.get("currencyDecimalSeparator")?.enable();
          }
        })
      )
      .subscribe();
  }

  // A public URL is required to enable the MCP server, mirroring the backend
  // validation. The control's validators are toggled with the enable flag; the
  // format check stays on either way because the backend validates a non-empty
  // URL regardless of the toggle.
  private listenForMcpEnabledChanges(): void {
    const mcpPublicUrl = this.form.get("mcpPublicUrl");

    this.form.get("mcpEnabled")?.valueChanges
      .pipe(
        startWith(this.form.get("mcpEnabled")?.value),
        untilDestroyed(this),
        tap((enabled: boolean) => {
          mcpPublicUrl?.setValidators(this.urlValidators(enabled));
          mcpPublicUrl?.updateValueAndValidity({ emitEvent: false });
        })
      )
      .subscribe();
  }

  // A mobile server URL is required to show the login QR code, mirroring the
  // backend validation. The control's validators track the toggle.
  private listenForShowLoginQrChanges(): void {
    const mobileServerUrl = this.form.get("mobileServerUrl");

    this.form.get("showLoginQr")?.valueChanges
      .pipe(
        startWith(this.form.get("showLoginQr")?.value),
        untilDestroyed(this),
        tap((enabled: boolean) => {
          mobileServerUrl?.setValidators(this.urlValidators(enabled));
          mobileServerUrl?.updateValueAndValidity({ emitEvent: false });
        })
      )
      .subscribe();
  }

  // setValidators replaces the whole list, so the format check has to be
  // re-supplied on every toggle rather than declared once in initForm.
  private urlValidators(required: boolean): ValidatorFn[] {
    return required
      ? [Validators.required, absoluteUrlValidator()]
      : [absoluteUrlValidator()];
  }

  public displayWith(id: number): string {
    return this.allReceiptProcessingSettings.find((rps) => rps.id === id)?.name ?? "";
  }

  private doesTaskServerRequiresRestart(): boolean {
    let requiresRestart = this.originalSystemSettings.taskConcurrency !== this.form.get("taskConcurrency")?.value;
    for (let i = 0; i < this.originalSystemSettings.taskQueueConfigurations.length; i++) {
      const originalConfig = this.originalSystemSettings.taskQueueConfigurations[i];
      const formConfig = (this.form.get("taskQueueConfigurations") as FormArray).controls.find((control) => control.get("name")?.value === originalConfig.name);

      if (originalConfig.priority !== formConfig?.get("priority")?.value) {
        requiresRestart = true;
        break;
      }
    }

    return requiresRestart;
  }

  public submit(): void {
    const formValue = this.form.getRawValue();
    // Fold the value/unit pairs back into the hours the API stores, then drop
    // the presentation-only controls so they never reach the wire.
    formValue["refreshTokenValidForHours"] = toHours(
      formValue["refreshTokenValidForValue"],
      formValue["refreshTokenValidForUnit"]
    );
    formValue["mcpRefreshTokenValidForHours"] = toHours(
      formValue["mcpRefreshTokenValidForValue"],
      formValue["mcpRefreshTokenValidForUnit"]
    );
    formValue["tempFileRetentionHours"] = toHours(
      formValue["tempFileRetentionValue"],
      formValue["tempFileRetentionUnit"]
    );
    delete formValue["refreshTokenValidForValue"];
    delete formValue["refreshTokenValidForUnit"];
    delete formValue["mcpRefreshTokenValidForValue"];
    delete formValue["mcpRefreshTokenValidForUnit"];
    delete formValue["tempFileRetentionValue"];
    delete formValue["tempFileRetentionUnit"];
    formValue["emailPollingInterval"] = Number.parseInt(formValue["emailPollingInterval"]);
    formValue["taskConcurrency"] = Number.parseInt(formValue["taskConcurrency"]);
    formValue["pdfDpi"] = Number.parseInt(formValue["pdfDpi"]);
    (formValue["taskQueueConfigurations"] as Array<any>).forEach(config => {
      config.priority = Number.parseInt(config.priority);
    });
    const restartTaskServer = this.doesTaskServerRequiresRestart();

    this.systemSettingsService.updateSystemSettings(formValue)
      .pipe(
        take(1),
        tap(() => {
          this.snackbarService.success("System settings updated successfully");
          this.router.navigate(["/system-settings/settings/view"], {
            queryParams: {
              restartTaskServer: restartTaskServer
            }
          });
        }),
        switchMap(() => this.featureConfigService.getFeatureConfig()),
        tap((featureConfig) => this.store.dispatch(new SetFeatureConfig(featureConfig))),
        switchMap(() => this.store.dispatch(new SetCurrencyDisplay(formValue["currencyDisplay"]?.toString()))),
        switchMap(() => this.store.dispatch(
          new SetCurrencyData(formValue["currencySymbolPosition"],
            formValue["currencyDecimalSeparator"],
            formValue["currencyThousandthsSeparator"],
            formValue["currencyHideDecimalPlaces"]
          ))),
        // Every appDate pipe and "today" reads the zone from the store, so the
        // app repaints in the new zone without a reload.
        switchMap(() => this.store.dispatch(new SetTimeZone(formValue["timeZone"]))),
      )
      .subscribe();
  }

  public restartTaskServer(): void {
    this.systemSettingsService.restartTaskServer().pipe(
      take(1),
      tap(() => {
          this.snackbarService.success("Task server restarted successfully");
          this.showRestartTaskServerAlert = false;
        },
      )).subscribe();
  }
}
