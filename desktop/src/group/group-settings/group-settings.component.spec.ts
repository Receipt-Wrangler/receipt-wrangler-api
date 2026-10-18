import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { ActivatedRoute } from "@angular/router";
import { ApiModule, Group } from "../../open-api";
import { StoreModule } from "../../store/store.module";
import { GroupSettingsComponent } from "./group-settings.component";

describe("GroupSettingsComponent", () => {
  let component: GroupSettingsComponent;
  let fixture: ComponentFixture<GroupSettingsComponent>;
  const group = { id: 1 } as Group;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [GroupSettingsComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [ApiModule,
        StoreModule,
        MatSnackBarModule
      ],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              data: {
                group: group,
                formConfig: {},
              },
            },
          },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ]
    });
    fixture = TestBed.createComponent(GroupSettingsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });


  it("should set edit settings url correctly", () => {
    expect(component.editSettingsUrl).toEqual(`/groups/${group.id}/settings/edit`);
  });

  it("should init form with empty values", () => {
    component.ngOnInit();

    expect(component.form.value).toEqual({
      systemEmailId: "",
      emailIntegrationEnabled: false,
      emailBodyProcessingEnabled: false,
      emailRequireSignedSender: true,
      subjectLineRegexes: [],
      emailWhiteList: [],
      emailDefaultReceiptStatus: "",
      emailDefaultReceiptPaidById: "",
      promptId: "",
      fallbackPromptId: ""
    });
  });
});
