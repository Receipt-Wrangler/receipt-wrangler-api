import { Component, Input, OnInit, input, viewChildren } from "@angular/core";
import { FormArray, FormBuilder, FormControl, FormGroup, Validators, } from "@angular/forms";
import { ActivatedRoute } from "@angular/router";
import { startWith, tap } from "rxjs";
import { FormMode } from "src/enums/form-mode.enum";
import { BaseFormComponent, FormCommand } from "src/form";
import { FormListComponent } from "src/shared-ui/form-list/form-list.component";
import { FormConfig } from "../../interfaces";
import { Group, GroupSettingsWhiteListEmail, SubjectLineRegex, SystemEmail } from "../../open-api";

@Component({
    selector: "app-group-settings-email",
    templateUrl: "./group-settings-email.component.html",
    styleUrls: ["./group-settings-email.component.scss"],
    standalone: false
})
export class GroupSettingsEmailComponent
  extends BaseFormComponent
  implements OnInit {
  @Input() public override form: FormGroup = new FormGroup({});

  public readonly canEdit = input<boolean>(false);

  @Input() public override formConfig!: FormConfig;

  public readonly formListComponents = viewChildren(FormListComponent);

  public group!: Group;

  public systemEmails: SystemEmail[] = [];

  constructor(
    private activatedRoute: ActivatedRoute,
    private formBuilder: FormBuilder
  ) {
    super();
  }

  public get subjectLineRegexes(): FormArray {
    return this.form.get("subjectLineRegexes") as FormArray;
  }

  public get emailWhiteList(): FormArray {
    return this.form.get("emailWhiteList") as FormArray;
  }

  public ngOnInit(): void {
    this.group = this.activatedRoute.snapshot.data["group"];
    this.systemEmails = this.activatedRoute.snapshot.data["systemEmails"];
    if (!this.canEdit() && this.group.groupSettings?.systemEmail?.id) {
      this.systemEmails = [this.group.groupSettings.systemEmail];
    }
    this.initForm();
  }

  private initForm(): void {
    this.setInitialValues();
    this.addValidators();
    this.listenForEnableEmailIntegrationChanges();
    if (this.formConfig.mode === FormMode.view) {
      this.form.get("emailIntegrationEnabled")?.disable();
    }
  }

  private listenForEnableEmailIntegrationChanges(): void {
    const control = this.form.get("emailIntegrationEnabled");

    control?.valueChanges
      .pipe(
        startWith(control.value),
        tap((enabled) => {
          if (enabled) {
            this.emitFormCommand({
              path: "systemEmailId",
              command: "addValidators",
              payload: [Validators.required],
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptStatus",
              command: "addValidators",
              payload: [Validators.required],
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptPaidById",
              command: "addValidators",
              payload: [Validators.required],
            });
            this.emitFormCommand({
              path: "emailBodyProcessingEnabled",
              command: "enable",
            });
          } else {
            this.emitFormCommand({
              path: "systemEmailId",
              command: "removeValidators",
              payload: [Validators.required],
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptStatus",
              command: "removeValidators",
              payload: [Validators.required],
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptPaidById",
              command: "removeValidators",
              payload: [Validators.required],
            });

            const errors = control.errors;

            if (errors?.["required"]) {
              delete errors["required"];
            }

            this.emitFormCommand({
              path: "systemEmailId",
              command: "setErrors",
              payload: {
                required: null,
              },
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptStatus",
              command: "setErrors",
              payload: {
                required: null,
              },
            });
            this.emitFormCommand({
              path: "emailDefaultReceiptPaidById",
              command: "setErrors",
              payload: {
                required: null,
              },
            });
            this.emitFormCommand({
              path: "emailBodyProcessingEnabled",
              command: "patchValue",
              payload: false,
            });
            this.emitFormCommand({
              path: "emailBodyProcessingEnabled",
              command: "disable",
            });
          }
          this.emitFormCommand({
            path: "systemEmailId",
            command: "updateValueAndValidity",
          });
          this.emitFormCommand({
            path: "emailDefaultReceiptStatus",
            command: "updateValueAndValidity",
          });
          this.emitFormCommand({
            path: "emailDefaultReceiptPaidById",
            command: "updateValueAndValidity",
          });
        })
      )
      .subscribe();
  }

  private addValidators(): void {
  }

  private setInitialValues(): void {
    this.emitFormCommand({
      path: "emailIntegrationEnabled",
      command: "patchValue",
      payload: this.group?.groupSettings?.emailIntegrationEnabled,
    });

    this.emitFormCommand({
      path: "emailBodyProcessingEnabled",
      command: "patchValue",
      payload: this.group?.groupSettings?.emailBodyProcessingEnabled,
    });

    // Default to on when the settings omit it: the form always submits this
    // control, so an unset value would otherwise switch the check off.
    this.emitFormCommand({
      path: "emailRequireSignedSender",
      command: "patchValue",
      payload: this.group?.groupSettings?.emailRequireSignedSender ?? true,
    });

    const formCommand: FormCommand = {
      path: "systemEmailId",
      command: "patchValue",
      payload: this.group?.groupSettings?.systemEmailId,
    };
    this.emitFormCommand(formCommand);

    this.emitFormCommand({
      path: "emailDefaultReceiptStatus",
      command: "patchValue",
      payload: this.group?.groupSettings?.emailDefaultReceiptStatus,
    });

    this.emitFormCommand({
      path: "emailDefaultReceiptPaidById",
      command: "patchValue",
      payload: this.group?.groupSettings?.emailDefaultReceiptPaidById,
    });

    const groupSettingsEmails = (
      this.group?.groupSettings?.emailWhiteList || []
    ).map((email) => this.buildGroupSettingsEmail(email));

    groupSettingsEmails.forEach((groupSettingsEmail) => {
      this.emitFormCommand({
        path: "emailWhiteList",
        command: "push",
        payload: groupSettingsEmail,
      });
    });

    const subjectLineRegexes = (
      this.group?.groupSettings?.subjectLineRegexes || []
    ).map((regex) => this.buildSubjectLineRegexes(regex));

    subjectLineRegexes.forEach((regex) => {
      this.emitFormCommand({
        path: "subjectLineRegexes",
        command: "push",
        payload: regex,
      });
    });
  }

  private buildGroupSettingsEmail(
    groupSettingEmail?: GroupSettingsWhiteListEmail
  ): FormGroup {
    return this.formBuilder.group({
      email: new FormControl(groupSettingEmail?.email, [
        Validators.email,
        Validators.required,
      ]),
    });
  }

  private buildSubjectLineRegexes(regex?: SubjectLineRegex): FormGroup {
    return this.formBuilder.group({
      regex: new FormControl(regex?.regex ?? "", [Validators.required]),
    });
  }

  public addSubjectLineRegex(): void {
    this.emitFormCommand({
      path: "subjectLineRegexes",
      command: "push",
      payload: this.buildSubjectLineRegexes(),
    });
  }

  public itemDoneButtonClicked(index: number): void {
    if (this.form.valid) {
      this.formListComponents().at(index)?.resetEditingIndex();
    }
  }

  public subjectLineItemCancelButtonClicked(): void {
    const lastIndex = this.subjectLineRegexes.length - 1;
    const formCommand: FormCommand = {
      path: "subjectLineRegexes",
      command: "removeAt",
      payload: lastIndex,
    };
    this.emitFormCommand(formCommand);
  }

  public subjectLineItemDeleteButtonClicked(index: number): void {
    const formCommand: FormCommand = {
      path: "subjectLineRegexes",
      command: "removeAt",
      payload: index,
    };
    this.emitFormCommand(formCommand);
  }

  public addEmailWhiteList(): void {
    this.emitFormCommand({
      path: "emailWhiteList",
      command: "push",
      payload: this.buildGroupSettingsEmail(),
    });
  }

  public emailWhiteListItemCancelButtonClicked(): void {
    const lastIndex = this.emailWhiteList.length - 1;
    const formCommand: FormCommand = {
      path: "emailWhiteList",
      command: "removeAt",
      payload: lastIndex,
    };
    this.emitFormCommand(formCommand);
  }

  public emailWhiteListItemDeleteButtonClicked(index: number): void {
    const formCommand: FormCommand = {
      path: "emailWhiteList",
      command: "removeAt",
      payload: index,
    };
    this.emitFormCommand(formCommand);
  }

  public systemEmailDisplayWith(id: number): string {
    return this.systemEmails.find((email) => email.id === id)?.username ?? "";
  }
}
