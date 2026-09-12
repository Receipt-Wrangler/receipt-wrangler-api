import { HttpTestingController, provideHttpClientTesting } from "@angular/common/http/testing";
import { CUSTOM_ELEMENTS_SCHEMA } from "@angular/core";
import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NgxsModule } from "@ngxs/store";
import { FilterOperation, SystemTaskPagedRequestFilter } from "../../open-api";
import { TABLE_SERVICE_INJECTION_TOKEN } from "../../services/injection-tokens/table-service";
import { SystemEmailTaskTableService } from "../../services/system-email-task-table.service";
import { SystemEmailTaskTableState } from "../../store/system-email-task-table.state";

import { TaskTableComponent } from "./task-table.component";
import { provideHttpClient, withInterceptorsFromDi } from "@angular/common/http";

describe("TaskTableComponent", () => {
  let component: TaskTableComponent;
  let fixture: ComponentFixture<TaskTableComponent>;
  let httpTesting: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
    declarations: [TaskTableComponent],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    imports: [NgxsModule.forRoot([SystemEmailTaskTableState])],
    providers: [
        {
            provide: TABLE_SERVICE_INJECTION_TOKEN,
            useClass: SystemEmailTaskTableService
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
})
      .compileComponents();

    httpTesting = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(TaskTableComponent);
    component = fixture.componentInstance;
  });

  afterEach(() => {
    httpTesting.verify();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("sends the filter the provider reports with the paged request", () => {
    const filter = {
      type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] },
    } as unknown as SystemTaskPagedRequestFilter;
    fixture.componentRef.setInput("filterProvider", () => filter);

    component.getTableData();

    const request = httpTesting.expectOne("/api/systemTask/getPagedSystemTasks");
    expect(request.request.body.filter).toEqual(filter);
    request.flush({ data: [], totalCount: 0 });
  });

  // The page dispatches a filter change and refreshes in the same synchronous
  // turn, before change detection can push a new input value in. Reading the
  // filter at request time is what makes a cleared chip actually clear.
  it("reads the filter at request time, not when the provider was bound", () => {
    let filter: any = { type: { operation: FilterOperation.Contains, value: ["QUICK_SCAN"] } };
    fixture.componentRef.setInput("filterProvider", () => filter);

    component.getTableData();
    httpTesting.expectOne("/api/systemTask/getPagedSystemTasks").flush({ data: [], totalCount: 0 });

    filter = { type: { operation: null, value: [] } };
    component.getTableData();

    const request = httpTesting.expectOne("/api/systemTask/getPagedSystemTasks");
    expect(request.request.body.filter).toEqual(filter);
    request.flush({ data: [], totalCount: 0 });
  });

  // The two embedded task tables never bind it, so the key must be absent
  // rather than sent as null -- the API then sees a zero-value filter and adds
  // no predicates.
  it("omits the filter key when no provider is bound", () => {
    component.getTableData();

    const request = httpTesting.expectOne("/api/systemTask/getPagedSystemTasks");
    expect(request.request.body.filter).toBeUndefined();
    request.flush({ data: [], totalCount: 0 });
  });

  // Two chip clears in quick succession put two fetches in flight; the last
  // request must win, not the last response.
  it("supersedes an in-flight refresh rather than letting it repaint the table", () => {
    component.getTableData();
    component.getTableData();

    const requests = httpTesting.match("/api/systemTask/getPagedSystemTasks");
    expect(requests.length).toBe(2);

    // The first was cancelled by switchMap, so only the second can land.
    expect(requests[0].cancelled).toBe(true);
    requests[1].flush({ data: [], totalCount: 3 });

    expect(component.totalCount()).toBe(3);
  });
});
