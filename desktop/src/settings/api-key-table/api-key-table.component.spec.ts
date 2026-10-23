import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { HttpTestingController, provideHttpClientTesting } from "@angular/common/http/testing";
import { provideZonelessChangeDetection } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatDialogModule } from "@angular/material/dialog";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { provideRouter } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { ApiKeyScope, ApiKeyView, Permission } from "../../open-api";
import { PipesModule } from "../../pipes/pipes.module";
import { SharedUiModule } from "../../shared-ui/shared-ui.module";
import { AuthState } from "../../store/auth.state";
import { SetAuthState, SetPermissions } from "../../store/auth.state.actions";
import { ApiKeyTableState } from "../../store/api-key-table.state";
import { TableModule } from "../../table/table.module";

import { ApiKeyTableComponent } from "./api-key-table.component";

describe("ApiKeyTableComponent", () => {
  let component: ApiKeyTableComponent;
  let fixture: ComponentFixture<ApiKeyTableComponent>;
  let store: Store;

  const ownKey = { id: 1, userId: 7 } as ApiKeyView;
  const otherKey = { id: 2, userId: 99 } as ApiKeyView;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ApiKeyTableComponent],
      imports: [
        SharedUiModule,
        PipesModule,
        NgxsModule.forRoot([AuthState, ApiKeyTableState]),
        TableModule,
        MatDialogModule,
        NoopAnimationsModule,
      ],
      providers: [
        provideZonelessChangeDetection(),
        provideRouter([]),
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    store = TestBed.inject(Store);
    store.dispatch(
      new SetAuthState({ userId: 7 } as any)
    );

    fixture = TestBed.createComponent(ApiKeyTableComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  describe("canViewAll", () => {
    it("is true when the user holds read-any", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysReadAny], {}));
      component.ngOnInit();
      expect(component.canViewAll()).toBe(true);
    });

    it("is false without read-any", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysRead], {}));
      component.ngOnInit();
      expect(component.canViewAll()).toBe(false);
    });
  });

  describe("canCreate", () => {
    it("is true when the user holds create", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysCreate], {}));
      component.ngOnInit();
      expect(component.canCreate()).toBe(true);
    });

    it("is false without create", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysRead], {}));
      component.ngOnInit();
      expect(component.canCreate()).toBe(false);
    });
  });

  describe("per-row gating", () => {
    beforeEach(() => {
      component.ngOnInit();
    });

    it("isOwner reflects the current user", () => {
      expect(component.isOwner(ownKey)).toBe(true);
      expect(component.isOwner(otherKey)).toBe(false);
    });

    it("canEdit requires ownership AND update", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysUpdate], {}));
      expect(component.canEdit(ownKey)).toBe(true);
      expect(component.canEdit(otherKey)).toBe(false);
    });

    it("canEdit is false without the update permission", () => {
      store.dispatch(new SetPermissions([], {}));
      expect(component.canEdit(ownKey)).toBe(false);
    });

    it("canDelete allows owners with the delete permission only", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysDelete], {}));
      expect(component.canDelete(ownKey)).toBe(true);
      expect(component.canDelete(otherKey)).toBe(false);
    });

    it("canDelete allows any key with delete-any", () => {
      store.dispatch(new SetPermissions([Permission.AppApiKeysDeleteAny], {}));
      expect(component.canDelete(ownKey)).toBe(true);
      expect(component.canDelete(otherKey)).toBe(true);
    });

    it("canDelete is false without any delete permission", () => {
      store.dispatch(new SetPermissions([], {}));
      expect(component.canDelete(ownKey)).toBe(false);
      expect(component.canDelete(otherKey)).toBe(false);
    });
  });

  describe("scope column", () => {
    beforeEach(() => {
      fixture.detectChanges(false);
    });

    it("is displayed after the description and is not sortable", () => {
      const displayed = component.displayedColumns;
      expect(displayed.indexOf("scope")).toBe(displayed.indexOf("description") + 1);

      const column = component.columns.find((c) => c.matColumnDef === "scope");
      expect(column?.columnHeader).toBe("Scope");
      expect(column?.sortable).toBe(false);
    });

    it("renders each scope's label, and an unknown value as-is", async () => {
      const httpMock = TestBed.inject(HttpTestingController);
      const data = [
        { id: "a", name: "read key", scope: ApiKeyScope.R },
        { id: "b", name: "write key", scope: ApiKeyScope.W },
        { id: "c", name: "read write key", scope: ApiKeyScope.Rw },
        { id: "d", name: "odd key", scope: "x" },
      ];
      httpMock.match(() => true).forEach((req) => req.flush({ totalCount: data.length, data }));
      await fixture.whenStable();
      fixture.detectChanges(false);

      const cells = Array.from(
        (fixture.nativeElement as HTMLElement).querySelectorAll('[data-testid="api-key-scope"]')
      ).map((cell) => cell.textContent?.trim());
      expect(cells).toEqual(["Read", "Write", "Read/Write", "x"]);
    });
  });
});
