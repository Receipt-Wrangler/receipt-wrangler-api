import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ReactiveFormsModule, Validators } from "@angular/forms";
import { MatDialogModule } from "@angular/material/dialog";
import { MatSnackBarModule } from "@angular/material/snack-bar";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { ActivatedRoute } from "@angular/router";
import { Store } from "@ngxs/store";
import { BehaviorSubject, of } from "rxjs";
import { FormMode } from "src/enums/form-mode.enum";
import { PipesModule } from "src/pipes/pipes.module";
import { SharedUiModule } from "src/shared-ui/shared-ui.module";
import { ApiModule, CustomFieldType, Permission, ReceiptImageService, ReceiptStatus } from "../../open-api";
import { SnackbarService } from "../../services";
import { QueueMode } from "../../services/receipt-queue.service";
import { StatefulMenuItem } from "../../standalone/components/filtered-stateful-menu/stateful-menu-item";
import { SetPermissions } from "../../store/auth.state.actions";
import { SetGroups, SetSelectedGroupId } from "../../store/group.state.actions";
import { StoreModule } from "../../store/store.module";
import { ReceiptFormComponent } from "./receipt-form.component";

describe("ReceiptFormComponent", () => {
  let component: ReceiptFormComponent;
  let fixture: ComponentFixture<ReceiptFormComponent>;
  let routeDataSubject: BehaviorSubject<any>;

  beforeEach(async () => {
    routeDataSubject = new BehaviorSubject<any>({});
    await TestBed.configureTestingModule({
      declarations: [ReceiptFormComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [ApiModule,
        PipesModule,
        MatDialogModule,
        MatSnackBarModule,
        StoreModule,
        NoopAnimationsModule,
        PipesModule,
        ReactiveFormsModule,
        SharedUiModule
      ],
      providers: [
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              data: {}, queryParams: {}
            },
            data: routeDataSubject,
            params: of({})
          },
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(ReceiptFormComponent);
    component = fixture.componentInstance;
    component.mode = FormMode.edit;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("falls back to the receipt's embedded custom field definitions when the catalog is empty", () => {
    const definition = { id: 7, name: "Project", type: 0 } as any;
    routeDataSubject.next({
      mode: FormMode.edit,
      customFields: [],
      receipt: {
        id: 1,
        name: "R",
        amount: "1.00",
        customFields: [{ customFieldId: 7, customField: definition }],
      } as any,
    });

    component.ngOnInit();

    expect(component.customFields).toEqual([definition]);
  });

  it("should init form correctly when there is no initial data", () => {
    jest.useFakeTimers();
    const mockedDate = new Date(2020, 0, 1);
    jest.setSystemTime(mockedDate);
    component.ngOnInit();

    expect(component.form.value).toEqual({
      name: "",
      amount: "",
      categories: [],
      tags: [],
      date: mockedDate,
      paidByUserId: "",
      // "" rather than the 0 this used to seed via Number(""): both read as
      // blank in the picker, but Validators.required treats 0 as PRESENT, so
      // the old sentinel left a group-less form valid.
      groupId: "",
      status: ReceiptStatus.Open,
      customFields: [],
      receiptItems: [],
      syncAmountWithItems: false,
    });
    jest.useRealTimers();
  });

  it("should patch magic fill values correctly", () => {
    // Mock timezone offset to be EST
    Date.prototype.getTimezoneOffset = () => 240;
    component.images.set([{ id: 1 } as any]);
    component.ngOnInit();
    component.mode = FormMode.edit;
    Object.defineProperty(component, 'carouselComponent', {
      value: () => ({
        currentlyShownImageIndex: 0,
      }),
      configurable: true,
    });
    component.categories = [
      { id: 1, name: "category" } as any,
      { id: 2, name: "category2" } as any,
    ];
    component.tags = [
      { id: 1, name: "tag" } as any,
      { id: 2, name: "tag2" } as any,
    ];

    const magicReceipt = {
      name: "magic",
      amount: "482.32",
      date: "2023-08-05T00:00:00.000Z",
      categories: [{ id: 1 } as any],
      tags: [
        {
          id: 2,
        },
      ],
    } as any;

    const receiptImageServiceSpy = jest.spyOn(
      TestBed.inject(ReceiptImageService),
      "magicFillReceipt"
    ).mockReturnValue(of(magicReceipt));

    const snackbarSpy = jest.spyOn(
      TestBed.inject(SnackbarService),
      "success"
    ).mockReturnValue(undefined);

    component.magicFill();

    expect(receiptImageServiceSpy).toHaveBeenCalledWith(1, undefined);

    const receiptValue = component.form.getRawValue();

    expect(receiptValue.name).toEqual(magicReceipt.name);
    expect(receiptValue.amount).toEqual(magicReceipt.amount);
    expect(receiptValue.date).toEqual(new Date("2023-08-05T04:00:00.000Z"));
    expect(receiptValue.categories).toEqual([component.categories[0]]);
    expect(receiptValue.tags).toEqual([component.tags[1]]);
    expect(snackbarSpy).toHaveBeenCalledWith(
      "Magic fill successfully filled name, amount, date, categories, tags from selected image!",
      { duration: 10000 }
    );
  });

  it("should not patch magic fill values if they are the defaults", () => {
    component.images.set([{ id: 1 } as any]);
    component.ngOnInit();
    component.mode = FormMode.edit;
    Object.defineProperty(component, 'carouselComponent', {
      value: () => ({
        currentlyShownImageIndex: 0,
      }),
      configurable: true,
    });

    const originalData = {
      name: "a different name",
      amount: "482.32",
      date: "2023-08-05T04:09:12.316Z",
    } as any;

    component.form.patchValue(originalData);

    const magicReceipt = {
      name: "magic",
      amount: "0",
      date: "0001-01-01T00:00:00Z",
    } as any;

    const receiptImageServiceSpy = jest.spyOn(
      TestBed.inject(ReceiptImageService),
      "magicFillReceipt"
    ).mockReturnValue(of(magicReceipt));

    component.magicFill();

    expect(receiptImageServiceSpy).toHaveBeenCalledWith(1, undefined,);

    const receiptValue = component.form.getRawValue();

    expect(receiptValue.name).toEqual(magicReceipt.name);
    expect(receiptValue.amount).toEqual(originalData.amount);
    expect(receiptValue.date).toEqual(originalData.date);
  });

  it("should not patch any values when they are all default values and pop error snackbar", () => {
    component.images.set([{ id: 1 } as any]);
    component.ngOnInit();
    component.mode = FormMode.edit;
    Object.defineProperty(component, 'carouselComponent', {
      value: () => ({
        currentlyShownImageIndex: 0,
      }),
      configurable: true,
    });


    const originalData = {
      name: "a different name",
      amount: "482.32",
      date: "2023-08-05T04:09:12.316Z",
    } as any;

    component.form.patchValue(originalData);

    const magicReceipt = {
      name: "",
      amount: "0",
      date: "0001-01-01T00:00:00Z",
    } as any;

    const receiptImageServiceSpy = jest.spyOn(
      TestBed.inject(ReceiptImageService),
      "magicFillReceipt"
    ).mockReturnValue(of(magicReceipt));

    const snackbarSpy = jest.spyOn(
      TestBed.inject(SnackbarService),
      "error"
    ).mockReturnValue(undefined);

    component.magicFill();

    expect(receiptImageServiceSpy).toHaveBeenCalledWith(1, undefined);

    const receiptValue = component.form.getRawValue();

    expect(receiptValue.name).toEqual(originalData.name);
    expect(receiptValue.amount).toEqual(originalData.amount);
    expect(receiptValue.date).toEqual(originalData.date);
    expect(snackbarSpy).toHaveBeenCalledWith(
      "Could not find any values to fill! Try reuploading a clearer image."
    );
  });

  it("should set queue data when there is no data", () => {
    component.ngOnInit();

    expect(component.queueIndex).toEqual(-1);
    expect(component.queueIds).toEqual([]);
    expect(component.queueMode).toEqual(undefined);
    expect(component.submitButtonText).toEqual("Save");
  });

  it("should set queue data when there is data", () => {
    TestBed.inject(ActivatedRoute).snapshot.queryParams = {
      ids: ["1", "2", "3"],
      queueMode: QueueMode.VIEW,
    };
    routeDataSubject.next({
      receipt: { id: 2 } as any,
    });

    expect(component.queueIndex).toEqual(1);
    expect(component.queueIds).toEqual(["1", "2", "3"]);
    expect(component.queueMode).toEqual(QueueMode.VIEW);
    expect(component.submitButtonText).toEqual("Save & Next");
  });

  it("rebuilds form state when route data emits a new receipt (duplicate navigation)", () => {
    routeDataSubject.next({
      receipt: { id: 1, name: "Original", amount: "10.00", customFields: [] } as any,
    });

    expect(component.originalReceipt?.id).toEqual(1);
    expect(component.form.get("name")?.value).toEqual("Original");
    expect(component.editLink).toEqual("/receipts/1/edit");

    routeDataSubject.next({
      receipt: { id: 2, name: "Duplicate", amount: "10.00", customFields: [] } as any,
    });

    expect(component.originalReceipt?.id).toEqual(2);
    expect(component.form.get("name")?.value).toEqual("Duplicate");
    expect(component.editLink).toEqual("/receipts/2/edit");
  });

  // Full-receipt magic-fill ingest: proves the form takes in every field the
  // backend can return (paid-by, status, items, shares, custom fields, comments),
  // not just the header fields, and drops nothing on the way into the form.
  describe("magicFill — full receipt ingest", () => {
    // A test below overrides the timezone offset for a deterministic date
    // assertion; capture the real one (before any test runs) and restore it so
    // the override can't leak into later tests.
    const originalGetTimezoneOffset = Date.prototype.getTimezoneOffset;
    afterEach(() => {
      Date.prototype.getTimezoneOffset = originalGetTimezoneOffset;
    });

    function stubCarousel(index = 0): void {
      Object.defineProperty(component, "carouselComponent", {
        value: () => ({ currentlyShownImageIndex: index }),
        configurable: true,
      });
    }

    function stubItemAndShareLists(): { setItems: jest.Mock; setUserItemMap: jest.Mock } {
      const setItems = jest.fn();
      const setUserItemMap = jest.fn();
      Object.defineProperty(component, "itemListComponent", {
        value: () => ({ setItems }),
        configurable: true,
      });
      Object.defineProperty(component, "shareListComponent", {
        value: () => ({ setUserItemMap }),
        configurable: true,
      });
      return { setItems, setUserItemMap };
    }

    function stubCommentsComponent(): { addMagicFilledComments: jest.Mock } {
      const addMagicFilledComments = jest.fn();
      Object.defineProperty(component, "receiptCommentsComponent", {
        value: () => ({ addMagicFilledComments }),
        configurable: true,
      });
      return { addMagicFilledComments };
    }

    function stubPaidByAutocomplete(): { syncSingleDisplay: jest.Mock } {
      const syncSingleDisplay = jest.fn();
      Object.defineProperty(component, "paidByAutocomplete", {
        value: () => ({ autocompleteComponent: () => ({ syncSingleDisplay }) }),
        configurable: true,
      });
      return { syncSingleDisplay };
    }

    function mockMagicFill(magicReceipt: any): jest.SpyInstance {
      return jest
        .spyOn(TestBed.inject(ReceiptImageService), "magicFillReceipt")
        .mockReturnValue(of(magicReceipt));
    }

    function spySuccess(): jest.SpyInstance {
      return jest.spyOn(TestBed.inject(SnackbarService), "success").mockReturnValue(undefined);
    }

    function spyError(): jest.SpyInstance {
      return jest.spyOn(TestBed.inject(SnackbarService), "error").mockReturnValue(undefined);
    }

    const customFieldDefs = [
      { id: 1, name: "Text Field", type: CustomFieldType.Text },
      { id: 2, name: "Date Field", type: CustomFieldType.Date },
      { id: 3, name: "Select Field", type: CustomFieldType.Select },
      { id: 4, name: "Currency Field", type: CustomFieldType.Currency },
      { id: 5, name: "Boolean Field", type: CustomFieldType.Boolean },
    ] as any[];

    it("ingests every field of a full magic-filled receipt", () => {
      // Fix the timezone offset so the date assertion is deterministic.
      Date.prototype.getTimezoneOffset = () => 0;
      component.images.set([{ id: 1 } as any]);
      routeDataSubject.next({ mode: FormMode.edit, customFields: customFieldDefs });
      component.mode = FormMode.edit;

      stubCarousel();
      const { setItems, setUserItemMap } = stubItemAndShareLists();
      const { addMagicFilledComments } = stubCommentsComponent();
      const { syncSingleDisplay } = stubPaidByAutocomplete();

      component.categories = [
        { id: 1, name: "category" } as any,
        { id: 2, name: "category2" } as any,
      ];
      component.tags = [
        { id: 1, name: "tag" } as any,
        { id: 2, name: "tag2" } as any,
      ];

      const magicReceipt = {
        name: "Full Receipt",
        amount: "100.00",
        date: "2023-08-05T00:00:00.000Z",
        paidByUserId: 7,
        status: ReceiptStatus.NeedsAttention,
        categories: [{ id: 1 }],
        tags: [{ id: 2 }],
        receiptItems: [
          {
            name: "Regular Item",
            amount: "40.00",
            status: "OPEN",
            categories: [{ id: 5, name: "item cat" }],
            tags: [{ id: 6, name: "item tag" }],
          },
          {
            name: "Shared Item",
            amount: "20.00",
            status: "OPEN",
            chargedToUserId: 7,
            linkedItems: [
              { name: "Linked", amount: "10.00", status: "OPEN", chargedToUserId: 7 },
            ],
          },
        ],
        customFields: [
          { customFieldId: 1, stringValue: "hello" },
          { customFieldId: 2, dateValue: "2023-08-05T00:00:00.000Z" },
          { customFieldId: 3, selectValue: 42 },
          { customFieldId: 4, currencyValue: "9.99" },
          { customFieldId: 5, booleanValue: true },
        ],
        comments: [{ comment: "auto comment", userId: 7 }],
      } as any;

      mockMagicFill(magicReceipt);
      const successSpy = spySuccess();

      component.magicFill();

      const value = component.form.getRawValue();

      // Scalars
      expect(value.name).toEqual("Full Receipt");
      expect(value.amount).toEqual("100.00");
      expect(value.date).toEqual(new Date("2023-08-05T00:00:00.000Z"));
      expect(value.paidByUserId).toEqual(7);
      expect(value.status).toEqual(ReceiptStatus.NeedsAttention);
      expect(syncSingleDisplay).toHaveBeenCalled();

      // Categories/tags matched from the pool by id
      expect(value.categories).toEqual([component.categories[0]]);
      expect(value.tags).toEqual([component.tags[1]]);

      // Items (regular + share + linked + per-item cats/tags + taxed)
      expect(value.receiptItems.length).toEqual(2);
      const regular = value.receiptItems[0];
      expect(regular.name).toEqual("Regular Item");
      expect(regular.amount).toEqual("40.00");
      expect(regular.status).toEqual("OPEN");
      expect(regular.chargedToUserId).toBeFalsy();
      expect(regular.categories).toEqual([{ id: 5, name: "item cat" }]);
      expect(regular.tags).toEqual([{ id: 6, name: "item tag" }]);
      expect(regular.linkedItems).toEqual([]);

      const shared = value.receiptItems[1];
      expect(shared.name).toEqual("Shared Item");
      expect(shared.chargedToUserId).toEqual(7);
      expect(shared.linkedItems.length).toEqual(1);
      expect(shared.linkedItems[0].name).toEqual("Linked");
      expect(shared.linkedItems[0].chargedToUserId).toEqual(7);

      expect(setItems).toHaveBeenCalled();
      expect(setUserItemMap).toHaveBeenCalled();

      // Custom fields land in the type-specific column
      expect(value.customFields.length).toEqual(5);
      const byId = (id: number) => value.customFields.find((c: any) => c.customFieldId === id);
      expect(byId(1).stringValue).toEqual("hello");
      expect(byId(2).dateValue).toEqual("2023-08-05T00:00:00.000Z");
      expect(byId(3).selectValue).toEqual(42);
      expect(byId(4).currencyValue).toEqual("9.99");
      expect(byId(5).booleanValue).toEqual(true);
      // Manage-fields menu entries flip to selected
      customFieldDefs.forEach((def) => {
        const menuItem = component.customFieldsStatefulMenuItems.find(
          (m) => m.value === def.id.toString()
        );
        expect(menuItem?.selected).toBe(true);
      });

      // Comments handed to the (mode-aware) comments child
      expect(addMagicFilledComments).toHaveBeenCalledWith(magicReceipt.comments);

      // Snackbar lists every filled field with reader-friendly labels
      expect(successSpy).toHaveBeenCalledWith(
        "Magic fill successfully filled name, amount, date, paid by, status, categories, tags, items, custom fields, comments from selected image!",
        { duration: 10000 }
      );
    });

    it("treats items with a chargedToUserId as shares and excludes them from the item total", () => {
      component.images.set([{ id: 1 } as any]);
      component.ngOnInit();
      component.mode = FormMode.edit;
      stubCarousel();
      stubItemAndShareLists();

      const magicReceipt = {
        name: "Items",
        amount: "100.00",
        date: "2023-08-05T00:00:00.000Z",
        receiptItems: [
          { name: "Regular", amount: "40.00", status: "OPEN" },
          { name: "Share", amount: "20.00", status: "OPEN", chargedToUserId: 9 },
        ],
      } as any;

      mockMagicFill(magicReceipt);
      spySuccess();

      component.magicFill();

      // Only the non-share item counts toward the receipt item total
      expect((component as any).calculateItemsTotal()).toEqual(40);
      // The share carries a required chargedToUserId validator; the regular item does not
      const shareControl = component.receiptItemsFormArray.at(1).get("chargedToUserId");
      const regularControl = component.receiptItemsFormArray.at(0).get("chargedToUserId");
      expect(shareControl?.hasValidator(Validators.required)).toBe(true);
      expect(regularControl?.hasValidator(Validators.required)).toBe(false);
    });

    it("fills nothing and shows an error when the response is all defaults/empty", () => {
      component.images.set([{ id: 1 } as any]);
      component.ngOnInit();
      component.mode = FormMode.edit;
      stubCarousel();

      const originalData = {
        name: "kept name",
        amount: "482.32",
        paidByUserId: 3,
        status: ReceiptStatus.Open,
      } as any;
      component.form.patchValue(originalData);

      const magicReceipt = {
        name: "",
        amount: "0",
        date: "0001-01-01T00:00:00Z",
        paidByUserId: 0,
        status: "",
        categories: [],
        tags: [],
        receiptItems: [],
        customFields: [],
        comments: [],
      } as any;

      mockMagicFill(magicReceipt);
      const errorSpy = spyError();

      component.magicFill();

      const value = component.form.getRawValue();
      expect(value.name).toEqual("kept name");
      expect(value.amount).toEqual("482.32");
      expect(value.paidByUserId).toEqual(3);
      expect(value.status).toEqual(ReceiptStatus.Open);
      expect(value.receiptItems).toEqual([]);
      expect(value.customFields).toEqual([]);
      expect(errorSpy).toHaveBeenCalledWith(
        "Could not find any values to fill! Try reuploading a clearer image."
      );
    });

    it("fills only the fields present and labels just those in the snackbar", () => {
      component.images.set([{ id: 1 } as any]);
      component.ngOnInit();
      component.mode = FormMode.edit;
      stubCarousel();
      stubItemAndShareLists();

      component.form.patchValue({ name: "untouched", amount: "5.00" });

      const magicReceipt = {
        receiptItems: [{ name: "Only Item", amount: "3.00", status: "OPEN" }],
      } as any;

      mockMagicFill(magicReceipt);
      const successSpy = spySuccess();

      component.magicFill();

      const value = component.form.getRawValue();
      expect(value.name).toEqual("untouched");
      expect(value.amount).toEqual("5.00");
      expect(value.receiptItems.length).toEqual(1);
      expect(successSpy).toHaveBeenCalledWith(
        "Magic fill successfully filled items from selected image!",
        { duration: 10000 }
      );
    });

    it("skips custom field values whose field is not in the catalog pool", () => {
      component.images.set([{ id: 1 } as any]);
      routeDataSubject.next({ mode: FormMode.edit, customFields: [customFieldDefs[0]] });
      component.mode = FormMode.edit;
      stubCarousel();

      const magicReceipt = {
        name: "CF",
        amount: "1.00",
        date: "2023-08-05T00:00:00.000Z",
        customFields: [{ customFieldId: 999, stringValue: "orphan" }],
      } as any;

      mockMagicFill(magicReceipt);
      const successSpy = spySuccess();

      component.magicFill();

      expect(component.customFieldsFormArray.length).toEqual(0);
      // "custom fields" is absent from the snackbar since nothing was ingested
      expect(successSpy).toHaveBeenCalledWith(
        "Magic fill successfully filled name, amount, date from selected image!",
        { duration: 10000 }
      );
    });

    it("does not duplicate a custom field the receipt already carries a value for", () => {
      component.images.set([{ id: 1 } as any]);
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: [customFieldDefs[0]],
        receipt: {
          id: 5,
          name: "R",
          amount: "1.00",
          customFields: [{ customFieldId: 1, stringValue: "existing" }],
        } as any,
      });
      component.mode = FormMode.edit;
      stubCarousel();

      const magicReceipt = {
        name: "CF",
        amount: "1.00",
        date: "2023-08-05T00:00:00.000Z",
        customFields: [{ customFieldId: 1, stringValue: "new value" }],
      } as any;

      mockMagicFill(magicReceipt);
      spySuccess();

      component.magicFill();

      expect(component.customFieldsFormArray.length).toEqual(1);
      expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("existing");
    });

    it("appends magic-filled items and categories onto existing ones", () => {
      component.images.set([{ id: 1 } as any]);
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: [],
        receipt: {
          id: 5,
          name: "R",
          amount: "100.00",
          categories: [{ id: 1, name: "existing cat" }],
          receiptItems: [{ name: "Existing Item", amount: "10.00", status: "OPEN" }],
          customFields: [],
        } as any,
      });
      component.mode = FormMode.edit;
      stubCarousel();
      stubItemAndShareLists();
      component.categories = [
        { id: 1, name: "existing cat" } as any,
        { id: 2, name: "new cat" } as any,
      ];

      const magicReceipt = {
        categories: [{ id: 2 }],
        receiptItems: [{ name: "New Item", amount: "15.00", status: "OPEN" }],
      } as any;

      mockMagicFill(magicReceipt);
      spySuccess();

      component.magicFill();

      const value = component.form.getRawValue();
      expect(value.categories).toEqual([
        { id: 1, name: "existing cat" },
        { id: 2, name: "new cat" },
      ]);
      expect(value.receiptItems.map((i: any) => i.name)).toEqual([
        "Existing Item",
        "New Item",
      ]);
    });

    it("does not duplicate a category the receipt already carries", () => {
      component.images.set([{ id: 1 } as any]);
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: [],
        receipt: {
          id: 5,
          name: "R",
          amount: "100.00",
          categories: [{ id: 1, name: "existing cat" }],
          customFields: [],
        } as any,
      });
      component.mode = FormMode.edit;
      stubCarousel();
      component.categories = [{ id: 1, name: "existing cat" } as any];

      const magicReceipt = { categories: [{ id: 1 }] } as any;

      mockMagicFill(magicReceipt);
      const errorSpy = spyError();

      component.magicFill();

      // The already-present category is not appended a second time, and since
      // nothing new was added the fill reports empty (error toast).
      expect(component.form.getRawValue().categories).toEqual([
        { id: 1, name: "existing cat" },
      ]);
      expect(errorSpy).toHaveBeenCalledWith(
        "Could not find any values to fill! Try reuploading a clearer image."
      );
    });

    it("round-trips negative refund amounts on the receipt and items", () => {
      component.images.set([{ id: 1 } as any]);
      component.ngOnInit();
      component.mode = FormMode.edit;
      stubCarousel();
      stubItemAndShareLists();

      const magicReceipt = {
        name: "Refund",
        amount: "-50.00",
        date: "2023-08-05T00:00:00.000Z",
        receiptItems: [{ name: "Refunded", amount: "-50.00", status: "OPEN" }],
      } as any;

      mockMagicFill(magicReceipt);
      spySuccess();

      component.magicFill();

      const value = component.form.getRawValue();
      expect(value.amount).toEqual("-50.00");
      expect(value.receiptItems[0].amount).toEqual("-50.00");
    });

    it("skips paid-by when the backend returns the unset sentinel (0)", () => {
      component.images.set([{ id: 1 } as any]);
      component.ngOnInit();
      component.mode = FormMode.edit;
      stubCarousel();
      const { syncSingleDisplay } = stubPaidByAutocomplete();
      component.form.patchValue({ paidByUserId: 4 });

      const magicReceipt = {
        name: "No Payer",
        amount: "1.00",
        date: "2023-08-05T00:00:00.000Z",
        paidByUserId: 0,
      } as any;

      mockMagicFill(magicReceipt);
      const successSpy = spySuccess();

      component.magicFill();

      // Existing paid-by preserved, display not refreshed, not listed in the snackbar
      expect(component.form.getRawValue().paidByUserId).toEqual(4);
      expect(syncSingleDisplay).not.toHaveBeenCalled();
      expect(successSpy).toHaveBeenCalledWith(
        "Magic fill successfully filled name, amount, date from selected image!",
        { duration: 10000 }
      );
    });
  });

  // The group's configured default custom fields, applied to the receipt form as
  // a "smart swap" when the group is chosen or changed.
  describe("group default custom fields", () => {
    let store: Store;

    const catalog = [
      { id: 1, name: "Cost Centre", type: CustomFieldType.Text },
      { id: 2, name: "PO Number", type: CustomFieldType.Text },
      { id: 3, name: "Reimbursable", type: CustomFieldType.Boolean },
    ] as any[];

    const groupWithDefaults = (id: number, defaultCustomFieldIds: number[]): any => ({
      id,
      name: `Group ${id}`,
      isAllGroup: false,
      groupMembers: [],
      groupReceiptSettings: { defaultCustomFieldIds },
    });

    const attachedIds = (): number[] =>
      component.customFieldsFormArray.value.map((value: any) => value.customFieldId);

    const menuItemFor = (customFieldId: number): StatefulMenuItem =>
      component.customFieldsStatefulMenuItems.find(
        (item) => item.value === customFieldId.toString()
      ) as StatefulMenuItem;

    const openAddForm = (selectedGroupId: string, customFields: any[] = catalog): void => {
      store.dispatch(new SetSelectedGroupId(selectedGroupId));
      routeDataSubject.next({ mode: FormMode.add, customFields });
    };

    beforeEach(() => {
      store = TestBed.inject(Store);
      store.dispatch(new SetPermissions([Permission.AppCustomFieldsRead], {}));
      store.dispatch(
        new SetGroups([
          groupWithDefaults(1, [1]),
          groupWithDefaults(2, [2]),
          groupWithDefaults(3, []),
          groupWithDefaults(4, [3]),
        ])
      );
    });

    it("seeds the selected group's defaults on an add form", () => {
      openAddForm("1");

      expect(attachedIds()).toEqual([1]);
      expect(menuItemFor(1).selected).toBe(true);
    });

    it("swaps an untouched default out for the new group's on a group change", () => {
      openAddForm("1");

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([2]);
      expect(menuItemFor(1).selected).toBe(false);
      expect(menuItemFor(2).selected).toBe(true);
    });

    it("keeps a default the user typed into when the group changes", () => {
      openAddForm("1");
      component.customFieldsFormArray.at(0).get("stringValue")!.setValue("R&D");

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([1, 2]);
      expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("R&D");
      expect(menuItemFor(1).selected).toBe(true);
    });

    it("keeps a manually added custom field when the group changes", () => {
      // Group 3 configures no defaults, so field 2 can only be there because the
      // user added it.
      openAddForm("3");
      component.customFieldChanged(menuItemFor(2));
      expect(attachedIds()).toEqual([2]);

      component.form.get("groupId")!.setValue(1);

      expect(attachedIds()).toEqual([2, 1]);
    });

    it("keeps a default the user toggled off and back on across a later group change", () => {
      openAddForm("1");

      component.customFieldChanged(menuItemFor(1));
      expect(attachedIds()).toEqual([]);
      component.customFieldChanged(menuItemFor(1));
      expect(attachedIds()).toEqual([1]);

      component.form.get("groupId")!.setValue(2);

      // Empty, but user-owned since they re-added it by hand - never swapped out.
      expect(attachedIds()).toEqual([1, 2]);
    });

    it("treats a boolean default left false as empty and swaps it out", () => {
      openAddForm("4");

      expect(attachedIds()).toEqual([3]);
      expect(component.customFieldsFormArray.at(0).value.booleanValue).toBe(false);

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([2]);
    });

    it("applies nothing to an edit-mode receipt on load", () => {
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: catalog,
        receipt: { id: 9, name: "R", amount: "1.00", groupId: 1, customFields: [] } as any,
      });

      expect(attachedIds()).toEqual([]);
    });

    it("applies the new group's defaults on an active group change in edit mode", () => {
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: catalog,
        receipt: { id: 9, name: "R", amount: "1.00", groupId: 1, customFields: [] } as any,
      });

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([2]);
    });

    it("never applies defaults in view mode", () => {
      routeDataSubject.next({
        mode: FormMode.view,
        customFields: catalog,
        receipt: { id: 9, name: "R", amount: "1.00", groupId: 1, customFields: [] } as any,
      });

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([]);
    });

    it("never applies defaults without app.custom-fields.read", () => {
      store.dispatch(new SetPermissions([], {}));

      openAddForm("1");
      expect(attachedIds()).toEqual([]);

      component.form.get("groupId")!.setValue(2);
      expect(attachedIds()).toEqual([]);
    });

    it("skips a default id that is missing from the loaded catalog", () => {
      // Group 1 defaults to field 1, which is not in this (restricted) catalog.
      openAddForm("2", [catalog[1]]);
      expect(attachedIds()).toEqual([2]);

      component.form.get("groupId")!.setValue(1);

      expect(attachedIds()).toEqual([]);
    });

    it("is a no-op when the group is cleared", () => {
      openAddForm("1");

      component.form.get("groupId")!.setValue("");

      expect(attachedIds()).toEqual([1]);
      expect(menuItemFor(1).selected).toBe(true);
    });

    it("resets the auto-applied set when route data re-emits", () => {
      openAddForm("1");
      expect(attachedIds()).toEqual([1]);

      // A fresh navigation rebuilds the form: field 1 is now the SAVED receipt's
      // own value, so a stale auto-applied set would wrongly swap it out below.
      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: catalog,
        receipt: {
          id: 9,
          name: "R",
          amount: "1.00",
          groupId: 1,
          customFields: [{ customFieldId: 1 }],
        } as any,
      });
      expect(attachedIds()).toEqual([1]);

      component.form.get("groupId")!.setValue(2);

      expect(attachedIds()).toEqual([1, 2]);
    });

    // Magic fill and group defaults interact: the default has already put an EMPTY
    // control on the form, so a plain "skip what's present" would drop the magic
    // value for exactly the fields a group pre-adds.
    describe("magic fill interaction", () => {
      const magicFillWith = (customFields: any[]): void => {
        // magicFill() reads the carousel viewChild for the current image index and,
        // in add mode, the file at that index off filesToUpload.
        Object.defineProperty(component, "carouselComponent", {
          value: () => ({ currentlyShownImageIndex: 0 }),
          configurable: true,
        });
        component.filesToUpload.set([{ file: new Blob() } as any]);
        jest
          .spyOn(TestBed.inject(ReceiptImageService), "magicFillReceipt")
          .mockReturnValue(of({ customFields } as any));
        jest
          .spyOn(TestBed.inject(SnackbarService), "success")
          .mockReturnValue(undefined);
        component.magicFill();
      };

      it("fills a blank auto-applied default instead of discarding the value", () => {
        openAddForm("1");
        expect(attachedIds()).toEqual([1]);

        magicFillWith([{ customFieldId: 1, stringValue: "R&D" }]);

        expect(attachedIds()).toEqual([1]);
        expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("R&D");
        expect(menuItemFor(1).selected).toBe(true);
      });

      it("does not overwrite a default the user already typed into", () => {
        openAddForm("1");
        component.customFieldsFormArray.at(0).get("stringValue")!.setValue("mine");

        magicFillWith([{ customFieldId: 1, stringValue: "R&D" }]);

        expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("mine");
      });

      // Filling it makes it the user's data, so the swap must stop managing it.
      it("keeps a magic-filled default when the group then changes", () => {
        openAddForm("1");
        magicFillWith([{ customFieldId: 1, stringValue: "R&D" }]);

        component.form.get("groupId")!.setValue(2);

        expect(attachedIds()).toEqual([1, 2]);
        expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("R&D");
      });

      it("still appends a magic value for a field not already on the form", () => {
        openAddForm("3");
        expect(attachedIds()).toEqual([]);

        magicFillWith([{ customFieldId: 2, stringValue: "PO-1" }]);

        expect(attachedIds()).toEqual([2]);
        expect(component.customFieldsFormArray.at(0).value.stringValue).toEqual("PO-1");
      });
    });
  });
  describe("group seeding on the add form", () => {
    let store: Store;

    const group = (id: number, isAllGroup = false): any => ({
      id,
      name: `Group ${id}`,
      isAllGroup,
      groupMembers: [],
    });

    const openAddForm = (selectedGroupId: string): void => {
      store.dispatch(new SetSelectedGroupId(selectedGroupId));
      routeDataSubject.next({ mode: FormMode.add, customFields: [] });
    };

    beforeEach(() => {
      store = TestBed.inject(Store);
    });

    it("seeds the user's only group when the All group is the active one", () => {
      // The All group sorts first server-side, so it is what a fresh
      // single-group user lands on - and it is never a receipt target.
      store.dispatch(new SetGroups([group(1, true), group(7)]));

      openAddForm("1");

      expect(component.form.value.groupId).toEqual(7);
    });

    it("leaves the group blank when the user belongs to more than one", () => {
      store.dispatch(new SetGroups([group(1, true), group(7), group(8)]));

      openAddForm("1");

      expect(component.form.value.groupId).toEqual("");
      // Not 0: Validators.required treats 0 as present, so a 0 seed would let a
      // group-less receipt submit.
      expect(component.form.get("groupId")!.valid).toBe(false);
    });

    it("seeds the user's only group when the active selection is stale", () => {
      // selectedGroupId is persisted, so it outlives a group the user has left.
      // Number("404") is truthy, which used to skip the fallback and seed an id
      // the picker does not offer.
      store.dispatch(new SetGroups([group(1, true), group(7)]));

      openAddForm("404");

      expect(component.form.value.groupId).toEqual(7);
    });

    it("still leaves it blank on a stale selection with several groups", () => {
      store.dispatch(new SetGroups([group(1, true), group(7), group(8)]));

      openAddForm("404");

      expect(component.form.value.groupId).toEqual("");
      expect(component.form.get("groupId")!.valid).toBe(false);
    });

    it("gates add-mode permissions on the seeded group, not the browsed one", () => {
      // The All group carries its own membership, which can be missing create
      // while the user's real group has it. The gate must follow the group the
      // receipt is actually going into.
      store.dispatch(new SetGroups([group(1, true), group(7)]));
      store.dispatch(
        new SetPermissions([], { 7: [Permission.GroupReceiptsCreate] })
      );

      openAddForm("1");

      expect(component.form.value.groupId).toEqual(7);
      expect(component.canEditReceipt()).toBe(true);
    });

    it("keeps gating on the browsed group when there is no single add target", () => {
      // Multi-group user on the All group: no add target, so the gate must stay
      // on the All group exactly as before - resolving the blank seed instead
      // would deny and render the whole form read-only.
      store.dispatch(new SetGroups([group(1, true), group(7), group(8)]));
      store.dispatch(
        new SetPermissions([], { 1: [Permission.GroupReceiptsCreate] })
      );

      openAddForm("1");

      expect(component.form.value.groupId).toEqual("");
      expect(component.canEditReceipt()).toBe(true);
    });

    it("keeps the actively selected group over the sole-group fallback", () => {
      store.dispatch(new SetGroups([group(1, true), group(7)]));

      openAddForm("7");

      expect(component.form.value.groupId).toEqual(7);
    });

    it("keeps an existing receipt's own group in edit mode", () => {
      store.dispatch(new SetGroups([group(1, true), group(7)]));
      store.dispatch(new SetSelectedGroupId("1"));

      routeDataSubject.next({
        mode: FormMode.edit,
        customFields: [],
        receipt: { id: 1, name: "R", amount: "1.00", groupId: 9 } as any,
      });

      expect(component.form.value.groupId).toEqual(9);
    });
  });
});
