import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed, } from "@angular/core/testing";
import { FormArray, FormControl, ReactiveFormsModule } from "@angular/forms";
import { MatDialog, MatDialogModule, MatDialogRef } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { ActivatedRoute } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { CarouselModule } from "ngx-bootstrap/carousel";
import { of } from "rxjs";
import { SharedUiModule } from "src/shared-ui/shared-ui.module";
import { LayoutState } from "src/store/layout.state";
import { ReceiptFileUploadCommand } from "../../interfaces";
import { ApiModule, ReceiptService, ReceiptStatus } from "../../open-api";
import { PipesModule } from "../../pipes";
import { SnackbarService } from "../../services";
import { AuthState, GroupState } from "../../store";
import { QUICK_SCAN_COMMENT_MAX_LENGTH, QuickScanDialogComponent } from "./quick-scan-dialog.component";
import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";

describe("QuickScanDialogComponent", () => {
  let component: QuickScanDialogComponent;
  let fixture: ComponentFixture<QuickScanDialogComponent>;
  let store: Store;

  beforeEach(() => {
    TestBed.configureTestingModule({
    declarations: [QuickScanDialogComponent],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    imports: [ApiModule,
        CarouselModule,
        MatDialogModule,
        MatSnackBarModule,
        NgxsModule.forRoot([AuthState, GroupState, LayoutState]),
        NoopAnimationsModule,
        PipesModule,
        ReactiveFormsModule,
        SharedUiModule],
    providers: [
        {
            provide: ActivatedRoute,
            useValue: {},
        },
        {
            provide: MatDialog,
            useValue: {}
        },
        {
            provide: MatDialogRef<QuickScanDialogComponent>,
            useValue: {
                close: () => { },
            },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
});
    fixture = TestBed.createComponent(QuickScanDialogComponent);
    store = TestBed.inject(Store);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should init form correctly", () => {
    component.ngOnInit();

    expect(component.form.value).toEqual({
      paidByUserIds: [],
      statuses: [],
      groupIds: [],
      categories: [],
      tags: [],
      comments: [],
    });
  });

  it("should push fileData into images when loaded", () => {
    const originalCreateObjectURL = URL.createObjectURL;
    URL.createObjectURL = jest.fn().mockReturnValue("awesome");
    const fileData = {} as ReceiptFileUploadCommand;
    component.fileLoaded(fileData);

    expect(component.images).toEqual([fileData]);
    URL.createObjectURL = originalCreateObjectURL;
  });

  it("should close the dialog", () => {
    const dialogSpy = jest.spyOn(TestBed.inject(MatDialogRef), "close");

    component.cancelButtonClicked();

    expect(dialogSpy).toHaveBeenCalled();
  });

  it("should show error if no image has been selected", () => {
    const snackbarSpy = jest.spyOn(TestBed.inject(SnackbarService), "error");

    component.submitButtonClicked();

    expect(snackbarSpy).toHaveBeenCalledWith(
      "Please select images to upload"
    );
  });

  it("should push new image when there are no user preferences", () => {
    component.fileLoaded({} as any);

    expect(component.form.value).toEqual({
      paidByUserIds: [""],
      statuses: [""],
      groupIds: [""],
      categories: [[]],
      tags: [[]],
      comments: [""],
    });
    expect(component.images).toEqual([{} as any]);
  });

  it("should push new image when there user preferences", () => {
    store.reset({
      auth: {
        userPreferences: {
          quickScanDefaultPaidById: 1,
          quickScanDefaultStatus: ReceiptStatus.Open,
          quickScanDefaultGroupId: 1,
        },
      },
      groups: { groups: [], selectedGroupId: "", selectedDashboardId: "" },
    });

    component.fileLoaded({} as any);

    expect(component.form.value).toEqual({
      paidByUserIds: [1],
      statuses: [ReceiptStatus.Open],
      groupIds: [1],
      categories: [[]],
      tags: [[]],
      comments: [""],
    });
    expect(component.images).toEqual([{} as any]);
  });

  it("seeds the user's only group when they have no default group preference", () => {
    // The All group ships alongside every real group and is never a quick-scan
    // target, so a single-group user still has two entries in state.
    store.reset({
      auth: {},
      groups: {
        groups: [
          { id: 1, name: "All Groups", isAllGroup: true },
          { id: 7, name: "My Receipts", isAllGroup: false },
        ],
        selectedGroupId: "",
        selectedDashboardId: "",
      },
    });

    component.fileLoaded({} as any);

    expect(component.groupIds.value).toEqual([7]);
  });

  it("prefers the quick-scan default group over the user's only group", () => {
    store.reset({
      auth: { userPreferences: { quickScanDefaultGroupId: 3 } },
      groups: {
        groups: [{ id: 7, name: "My Receipts", isAllGroup: false }],
        selectedGroupId: "",
        selectedDashboardId: "",
      },
    });

    component.fileLoaded({} as any);

    expect(component.groupIds.value).toEqual([3]);
  });

  it("leaves the group blank when the user belongs to more than one", () => {
    store.reset({
      auth: {},
      groups: {
        groups: [
          { id: 7, name: "My Receipts", isAllGroup: false },
          { id: 8, name: "Household", isAllGroup: false },
        ],
        selectedGroupId: "",
        selectedDashboardId: "",
      },
    });

    component.fileLoaded({} as any);

    expect(component.groupIds.value).toEqual([""]);
  });

  it("should drive field visibility and required-ness from the selected group's settings", () => {
    const group = {
      id: 2,
      groupReceiptSettings: {
        quickScanPaidByEnabled: false,
        quickScanPaidByRequired: false,
        quickScanStatusEnabled: true,
        quickScanStatusRequired: true,
        quickScanCategoriesEnabled: true,
        quickScanCategoriesRequired: true,
        quickScanTagsEnabled: false,
        quickScanTagsRequired: false,
      },
    };
    store.reset({
      auth: {},
      groups: { groups: [group], selectedGroupId: "", selectedDashboardId: "" },
    });

    component.fileLoaded({} as any);
    component.groupIds.at(0).setValue(2);

    expect(component.showPaidBy(0)).toBe(false);
    expect(component.showStatus(0)).toBe(true);
    expect(component.showCategories(0)).toBe(true);
    expect(component.showTags(0)).toBe(false);

    // Required category with no selection is invalid; status is required and empty.
    expect(component.categories.at(0).valid).toBe(false);
    expect(component.statuses.at(0).valid).toBe(false);
    // Hidden paid-by is not required.
    expect(component.paidByUserIds.at(0).valid).toBe(true);
  });

  it("should send category and tag ids per image on submit", () => {
    const originalCreateObjectURL = URL.createObjectURL;
    try {
      URL.createObjectURL = jest.fn().mockReturnValue("blob");

      const group = {
        id: 2,
        groupReceiptSettings: {
          quickScanPaidByEnabled: false,
          quickScanPaidByRequired: false,
          quickScanStatusEnabled: true,
          quickScanStatusRequired: false,
          quickScanCategoriesEnabled: true,
          quickScanCategoriesRequired: true,
          quickScanTagsEnabled: true,
          quickScanTagsRequired: false,
        },
      };
      store.reset({
        auth: {
          groupCategories: { 2: [{ id: 10, name: "c" }] },
          groupTags: { 2: [{ id: 20, name: "t" }] },
        },
        groups: { groups: [group], selectedGroupId: "", selectedDashboardId: "" },
      });

      const receiptService = TestBed.inject(ReceiptService);
      const serviceSpy = jest
        .spyOn(receiptService, "quickScanReceipt")
        .mockReturnValue(of({} as any));

      const fileData = { file: { name: "a" } } as any;
      component.fileLoaded(fileData);
      component.groupIds.at(0).setValue(2);
      // Categories/tags are per-image FormArrays; selections are pushed onto them
      // (as the autocomplete does), not set wholesale.
      (component.categories.at(0) as FormArray).push(new FormControl({ id: 10, name: "c" }));
      (component.tags.at(0) as FormArray).push(new FormControl({ id: 20, name: "t" }));

      component.submitButtonClicked();

      expect(serviceSpy).toHaveBeenCalledWith(
        [fileData.file],
        [2],
        [""],
        [""],
        ["10"],
        ["20"],
        [""]
      );
    } finally {
      URL.createObjectURL = originalCreateObjectURL;
    }
  });
  describe("comment field", () => {
    // The comment field needs the group config AND group.comments.create, so every case seeds both.
    const seedStore = (
      settings: Record<string, boolean>,
      groupPermissions: Record<number, string[]> = { 2: ["group.comments.create"] },
    ) => {
      store.reset({
        auth: { groupPermissions },
        groups: {
          groups: [{ id: 2, groupReceiptSettings: settings }],
          selectedGroupId: "",
          selectedDashboardId: "",
        },
      });
    };

    it("should show and require the comment per the group config", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(true);
      // Required and empty.
      expect(component.comments.at(0).valid).toBe(false);

      component.comments.at(0).setValue("A note");
      expect(component.comments.at(0).valid).toBe(true);
    });

    it("should show an optional comment without requiring it", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: false });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(true);
      expect(component.comments.at(0).valid).toBe(true);
    });

    it("should hide the comment by default", () => {
      seedStore({});

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(false);
    });

    it("should hide the comment when the group hides comments", () => {
      seedStore({ hideComments: true, quickScanCommentEnabled: true, quickScanCommentRequired: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(false);
      // Not required either - otherwise the hidden field would block every submit.
      expect(component.comments.at(0).valid).toBe(true);
    });

    it("should hide the comment without group.comments.create", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: true }, { 2: [] });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(false);
      expect(component.comments.at(0).valid).toBe(true);
    });

    it("should accept a comment at the maximum length and reject one over it", () => {
      seedStore({ quickScanCommentEnabled: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      component.comments.at(0).setValue("a".repeat(QUICK_SCAN_COMMENT_MAX_LENGTH));
      expect(component.comments.at(0).valid).toBe(true);

      component.comments.at(0).setValue("a".repeat(QUICK_SCAN_COMMENT_MAX_LENGTH + 1));
      expect(component.comments.at(0).valid).toBe(false);
      expect(component.comments.at(0).hasError("maxlength")).toBe(true);
    });

    // The backend counts runes, so the cap must count code points too - Validators.maxLength
    // counts UTF-16 code units and would reject 500 emoji the API accepts.
    it("should count the comment length in code points, not UTF-16 units", () => {
      seedStore({ quickScanCommentEnabled: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      // 500 supplementary characters = 1000 UTF-16 units but 500 code points.
      component.comments.at(0).setValue("😀".repeat(QUICK_SCAN_COMMENT_MAX_LENGTH));
      expect(component.comments.at(0).valid).toBe(true);

      component.comments.at(0).setValue("😀".repeat(QUICK_SCAN_COMMENT_MAX_LENGTH + 1));
      expect(component.comments.at(0).hasError("maxlength")).toBe(true);
    });

    // QuickScanCommand trims each comment on parse, so whitespace-only arrives empty and the
    // group's required check 400s - the form has to refuse it first.
    it("should reject a whitespace-only comment when one is required", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      component.comments.at(0).setValue("   \n\t ");
      expect(component.comments.at(0).valid).toBe(false);
      expect(component.comments.at(0).hasError("required")).toBe(true);

      component.comments.at(0).setValue("  A real note  ");
      expect(component.comments.at(0).valid).toBe(true);
    });

    it("should allow a whitespace-only comment when one is not required", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: false });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);

      component.comments.at(0).setValue("   ");
      expect(component.comments.at(0).valid).toBe(true);
    });

    // setRequired removes by reference, so a recompute that turns required off must actually
    // drop the trim-aware validator rather than leaving the field permanently required.
    it("should drop the required check when the group stops requiring a comment", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);
      component.comments.at(0).setValue("   ");
      expect(component.comments.at(0).valid).toBe(false);

      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: false });
      component.groupIds.at(0).setValue(2);
      expect(component.comments.at(0).valid).toBe(true);
    });

    // setRequired composes validators additively, so a show/require recompute must not drop the
    // length cap that was attached when the control was created.
    it("should keep the length cap after the group config is re-applied", () => {
      seedStore({ quickScanCommentEnabled: true, quickScanCommentRequired: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);
      component.groupIds.at(0).setValue(2);

      component.comments.at(0).setValue("a".repeat(QUICK_SCAN_COMMENT_MAX_LENGTH + 1));
      expect(component.comments.at(0).hasError("maxlength")).toBe(true);
    });

    it("should clear a hidden comment so nothing stale is submitted", () => {
      seedStore({ quickScanCommentEnabled: true });

      component.fileLoaded({} as any);
      component.groupIds.at(0).setValue(2);
      component.comments.at(0).setValue("typed before the group hid the field");

      seedStore({ quickScanCommentEnabled: false });
      component.groupIds.at(0).setValue(2);

      expect(component.showComment(0)).toBe(false);
      expect(component.comments.at(0).value).toBe("");
    });

    it("should send the comment per image on submit", () => {
      const originalCreateObjectURL = URL.createObjectURL;
      try {
        URL.createObjectURL = jest.fn().mockReturnValue("blob");
        seedStore({
          quickScanPaidByEnabled: false,
          quickScanStatusEnabled: false,
          quickScanCommentEnabled: true,
          quickScanCommentRequired: true,
        });

        const receiptService = TestBed.inject(ReceiptService);
        const serviceSpy = jest
          .spyOn(receiptService, "quickScanReceipt")
          .mockReturnValue(of({} as any));

        const fileData = { file: { name: "a" } } as any;
        component.fileLoaded(fileData);
        component.groupIds.at(0).setValue(2);
        component.comments.at(0).setValue("Client dinner");

        component.submitButtonClicked();

        expect(serviceSpy).toHaveBeenCalledWith(
          [fileData.file],
          [2],
          [""],
          [""],
          [""],
          [""],
          ["Client dinner"]
        );
      } finally {
        URL.createObjectURL = originalCreateObjectURL;
      }
    });
  });
});
