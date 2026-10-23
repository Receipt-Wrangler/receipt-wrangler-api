import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { FormArray, FormControl, FormGroup, Validators } from "@angular/forms";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { ActivatedRoute } from "@angular/router";
import { PipesModule } from "src/pipes/pipes.module";
import { FormMode } from "../../enums/form-mode.enum";
import { FormConfig } from "../../interfaces";
import { Group, ReceiptStatus } from "../../open-api";

import { GroupSettingsEmailComponent } from "./group-settings-email.component";

describe("GroupSettingsEmailComponent", () => {
  let component: GroupSettingsEmailComponent;
  let fixture: ComponentFixture<GroupSettingsEmailComponent>;
  const formConfig = {
    mode: FormMode.edit,
  } as FormConfig;


  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [GroupSettingsEmailComponent],
      imports: [MatSnackBarModule, PipesModule, PipesModule],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              data: {
                formConfig: formConfig,
                group: { id: 1 },
              },
            },
          },
        },
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    });
    fixture = TestBed.createComponent(GroupSettingsEmailComponent);

    component = fixture.componentInstance;
    component.form = new FormGroup({
      systemEmailId: new FormControl(""),
      emailWhiteList: new FormArray([]),
      subjectLineRegexes: new FormArray([]),
      emailDefaultReceiptStatus: new FormControl(""),
      emailDefaultReceiptPaidById: new FormControl(""),
      emailIntegrationEnabled: new FormControl(false),
      emailBodyProcessingEnabled: new FormControl(false),
      emailRequireSignedSender: new FormControl(true),
    });
    component.formCommand.subscribe((command) =>
      component.handleFormCommand(command)
    );
    component.formConfig = formConfig;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should init form with empty values", () => {
    component.ngOnInit();

    expect(component.form.value).toEqual({
      systemEmailId: undefined,
      emailIntegrationEnabled: undefined,
      emailBodyProcessingEnabled: undefined,
      emailRequireSignedSender: true,
      subjectLineRegexes: [],
      emailWhiteList: [],
      emailDefaultReceiptStatus: undefined,
      emailDefaultReceiptPaidById: undefined,
    });

    const form = component.form;
    expect(
      form.get("systemEmailId")?.hasValidator(Validators.required)
    ).toBe(false);

    expect(
      form.get("emailDefaultReceiptStatus")?.hasValidator(Validators.required)
    ).toBe(false);
    expect(
      form.get("emailDefaultReceiptPaidById")?.hasValidator(Validators.required)
    ).toBe(false);
  });

  it("should init form with initial data", () => {
    const activatedRoute = TestBed.inject(ActivatedRoute);
    const group: Group = {
      id: 1,
      groupSettings: {
        id: 1,
        groupId: 1,
        emailIntegrationEnabled: true,
        systemEmailId: "test@test.com",
        subjectLineRegexes: [
          {
            regex: "test",
          },
        ],
        emailWhiteList: [
          {
            email: "oneHundred@test.com",
          },
          {
            email: "twoHundred@test.com",
          },
        ],
        emailDefaultReceiptStatus: ReceiptStatus.Open,
        emailDefaultReceiptPaidById: 1,
      },
    } as any;
    activatedRoute.snapshot.data = {
      ...activatedRoute.snapshot.data,
      group: group,
    };

    component.ngOnInit();

    expect(component.form.value).toEqual({
      systemEmailId: "test@test.com",
      emailIntegrationEnabled: true,
      emailBodyProcessingEnabled: undefined,
      emailRequireSignedSender: true,
      subjectLineRegexes: [{ regex: "test" }],
      emailWhiteList: [
        {
          email: "oneHundred@test.com",
        },
        {
          email: "twoHundred@test.com",
        },
      ],
      emailDefaultReceiptStatus: "OPEN",
      emailDefaultReceiptPaidById: 1,
    });

    const form = component.form;
    expect(
      form.get("systemEmailId")?.hasValidator(Validators.required)
    ).toBe(true);

    expect(
      form.get("emailDefaultReceiptStatus")?.hasValidator(Validators.required)
    ).toBe(true);
    expect(
      form.get("emailDefaultReceiptPaidById")?.hasValidator(Validators.required)
    ).toBe(true);
  });

  it("should seed the signed sender check from the group settings", () => {
    const activatedRoute = TestBed.inject(ActivatedRoute);
    activatedRoute.snapshot.data = {
      ...activatedRoute.snapshot.data,
      group: { id: 1, groupSettings: { id: 1, groupId: 1, emailRequireSignedSender: false } } as any,
    };

    component.ngOnInit();

    expect(component.form.get("emailRequireSignedSender")?.value).toBe(false);
  });

  it("should keep the signed sender check on when the settings omit it", () => {
    const activatedRoute = TestBed.inject(ActivatedRoute);
    component.form.get("emailRequireSignedSender")?.setValue(false);
    activatedRoute.snapshot.data = {
      ...activatedRoute.snapshot.data,
      group: { id: 1, groupSettings: { id: 1, groupId: 1 } } as any,
    };

    component.ngOnInit();

    // The form always submits this control, so an unset value must not read as
    // "off" and silently disable the check on save.
    expect(component.form.get("emailRequireSignedSender")?.value).toBe(true);
  });

  it("should render the signed sender checkbox bound to its control", () => {
    component.ngOnInit();
    fixture.detectChanges();

    const checkbox = fixture.nativeElement.querySelector('app-checkbox[label="Require Signed Sender (DKIM)"]');
    expect(checkbox).toBeTruthy();
  });
});
