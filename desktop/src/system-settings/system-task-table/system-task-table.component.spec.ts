import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MatChipsModule } from "@angular/material/chips";
import { MatIconModule } from "@angular/material/icon";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { ActivatedRoute } from "@angular/router";
import { NgxsModule, Store } from "@ngxs/store";
import { ButtonComponent } from "../../button/index";
import { FilterOperation } from "../../open-api";
import { SharedUiModule } from "../../shared-ui/shared-ui.module";
import { UserState } from "../../store";
import { buildDefaultSystemTaskFilter, SystemTaskTableState } from "../../store/system-task-table.state";
import {
  SetPage,
  SetSystemTaskFilter,
  SetSystemTaskFilterField
} from "../../store/system-task-table.state.actions";
import { TableModule } from "../../table/table.module";

import { SystemTaskTableComponent } from "./system-task-table.component";

describe("SystemTaskTableComponent", () => {
  let component: SystemTaskTableComponent;
  let fixture: ComponentFixture<SystemTaskTableComponent>;
  let store: Store;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [SystemTaskTableComponent, ButtonComponent],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
      imports: [SharedUiModule,
        MatChipsModule,
        MatIconModule,
        NgxsModule.forRoot([SystemTaskTableState, UserState]),
        TableModule,
        NoopAnimationsModule],
      providers: [{
        provide: ActivatedRoute,
        useValue: {
          snapshot: {
            data: {
              prompts: [],
              allReceiptProcessingSettings: []
            }
          }
        }
      }, provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
    })
      .compileComponents();

    store = TestBed.inject(Store);

    fixture = TestBed.createComponent(SystemTaskTableComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("starts with no chips and an empty badge", () => {
    expect(component.filterChips()).toEqual([]);
    expect(component.numFiltersApplied()).toBe(0);
  });

  it("builds one chip per active condition off the stored filter", () => {
    store.dispatch(new SetSystemTaskFilter({
      ...buildDefaultSystemTaskFilter(),
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
    } as any));

    expect(component.filterChips().map((chip) => chip.key)).toEqual(["type"]);
    expect(component.numFiltersApplied()).toBe(1);
  });

  // Clearing a chip from page 7 must not leave the user on a page the narrowed
  // result set no longer has.
  it("clears a chip and resets to page one", () => {
    store.dispatch(new SetPage(7));
    store.dispatch(new SetSystemTaskFilter({
      ...buildDefaultSystemTaskFilter(),
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
    } as any));

    const dispatchSpy = jest.spyOn(store, "dispatch");
    jest.spyOn(component, "refresh").mockImplementation(() => {});

    component.filterChipCleared("type");

    expect(dispatchSpy).toHaveBeenCalledWith(new SetSystemTaskFilterField("type", null));
    expect(dispatchSpy).toHaveBeenCalledWith(new SetPage(1));
    expect(store.selectSnapshot(SystemTaskTableState.page)).toBe(1);
    expect(component.filterChips()).toEqual([]);
  });
});
