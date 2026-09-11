import {
  ChangeDetectionStrategy,
  Component,
  computed,
  DestroyRef,
  inject,
  input,
  OnInit,
  signal,
} from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { FormArray, FormBuilder, FormControl, FormGroup } from "@angular/forms";
import { MatDialog } from "@angular/material/dialog";
import { Store } from "@ngxs/store";
import { merge } from "rxjs";
import { DEFAULT_DIALOG_CONFIG } from "src/constants/dialog.constant";
import { GroupState } from "src/store";
import { BadgeTone, CUSTOM_FIELD_BADGE } from "src/shared-ui/badge/badge.component";
import { ReportColumn, ReportDetail, ReportPeriod } from "../../open-api";
import {
  REPORT_DOCUMENT_VARIABLES,
  REPORT_PERIOD_PRESETS,
  ReportField,
  toFieldOptions,
} from "../models/report-catalog.constants";
import { CHIP_COLORS, groupInitials } from "../models/report-chip.util";
import { isDimensionColumnDisabled, ReportColumnValue } from "../models/report-command.mapper";
import {
  buildColumnGroup,
  buildGroupByGroup,
  readGroupByFields,
} from "../models/report-form.factory";
import { formatPeriodRange, resolvePeriodRange } from "../models/report-period.util";
import { ReportCatalogService } from "../services/report-catalog.service";
import {
  AddGroupDialogComponent,
  AddGroupDialogData,
} from "../dialogs/add-group-dialog/add-group-dialog.component";
import {
  ColumnPickerDialogComponent,
  ColumnPickerDialogData,
} from "../dialogs/column-picker-dialog/column-picker-dialog.component";

interface ScopeChip {
  index: number;
  name: string;
  initials: string;
  color: string;
}

interface GroupByLevel {
  index: number;
  label: string;
  isCustom: boolean;
  isFirst: boolean;
  isLast: boolean;
}

interface ColumnRow {
  index: number;
  id: string;
  label: string;
  description: string;
  kindLabel: string;
  kindIcon: string;
  kindClass: string;
  kindTone: BadgeTone;
  isCustom: boolean;
  isFirst: boolean;
  isLast: boolean;
  disabled: boolean;
  disabledReason: string;
}

// The id the grouping level's synthetic column carries into the picker. The
// picker only reads it as "this is an edit, not a new column" (and hands it back
// untouched), so it is a constant rather than a real row id.
const GROUPING_LEVEL_COLUMN_ID = "grouping-level";

// cssClass colours the row's icon chip, which sits on a known background and so
// keeps its translucent fill; tone colours the badge, which is the shared
// component and carries its own opaque one.
const KIND_META: Record<
  ReportColumn.KindEnum,
  { label: string; icon: string; cssClass: string; tone: BadgeTone }
> = {
  dimension: { label: "Dim", icon: "sell", cssClass: "kind-dimension", tone: "slate" },
  aggregate: { label: "Agg", icon: "functions", cssClass: "kind-aggregate", tone: "blue" },
  formula: { label: "Formula", icon: "calculate", cssClass: "kind-formula", tone: "green" },
};

/**
 * The report builder's left configuration panel: report details + scope, period,
 * filters, grouping, detail mode, columns, totals, and the document. It binds
 * inline controls (name, period, detail, totals, document) straight to the shared
 * form via the formGet pipe, and manages the structural lists (scope, grouping,
 * columns) with mutation methods; a revision signal re-renders those lists after
 * dialog-driven changes under zoneless change detection.
 */
@Component({
  selector: "app-report-config-panel",
  templateUrl: "./report-config-panel.component.html",
  styleUrls: ["./report-config-panel.component.scss"],
  standalone: false,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ReportConfigPanelComponent implements OnInit {
  public readonly form = input.required<FormGroup>();

  private readonly catalog = inject(ReportCatalogService);
  private readonly dialog = inject(MatDialog);
  private readonly formBuilder = inject(FormBuilder);
  private readonly store = inject(Store);
  private readonly destroyRef = inject(DestroyRef);

  private readonly allGroups = this.store.selectSignal(GroupState.groupsWithoutAll);
  private readonly revision = signal(0);

  public readonly dimensions = this.catalog.dimensions;
  public readonly measures = this.catalog.measures;

  public readonly periodOptions = REPORT_PERIOD_PRESETS.map((preset) => ({
    value: preset.id,
    displayValue: preset.label,
  }));
  public readonly documentVariables = REPORT_DOCUMENT_VARIABLES;
  public readonly customFieldBadge = CUSTOM_FIELD_BADGE;
  public readonly ReportPeriodPreset = ReportPeriod.PresetEnum;
  public readonly ReportDetailMode = ReportDetail.ModeEnum;

  // The "Add grouping level…" picker is a stateless dispatcher: app-select is
  // FormControl-driven (no change output), so a scratch control feeds each pick
  // into addGroupBy and is reset to the blank option afterward.
  public readonly addGroupControl = new FormControl<string | null>(null);

  public readonly dimensionOptions = computed(() => toFieldOptions(this.dimensions()));

  public readonly addableDimensionOptions = computed(() => toFieldOptions(this.addableDimensions()));

  public readonly scopeChips = computed<ScopeChip[]>(() => {
    this.revision();
    const groups = this.allGroups();
    return this.scopeArray.controls.map((control, index) => {
      const id = control.value as string;
      const group = groups.find((candidate) => candidate.id?.toString() === id);
      const name = group?.name ?? id;
      return { index, name, initials: groupInitials(name), color: CHIP_COLORS[index % CHIP_COLORS.length] };
    });
  });

  public readonly groupByLevels = computed<GroupByLevel[]>(() => {
    this.revision();
    const controls = this.groupByArray.controls;
    return controls.map((_, index) => ({
      index,
      // A level renders as a leading column in the report, so it shows the
      // heading that column will carry: the user's override when set.
      label: this.headingForLevel(index),
      isCustom: this.isCustomField(this.fieldForLevel(index)),
      isFirst: index === 0,
      isLast: index === controls.length - 1,
    }));
  });

  public readonly addableDimensions = computed<ReportField[]>(() => {
    this.revision();
    const used = new Set(readGroupByFields(this.groupByArray));
    return this.dimensions().filter((field) => !used.has(field.key));
  });

  public readonly columnRows = computed<ColumnRow[]>(() => {
    this.revision();
    const mode = this.detailMode;
    const detailBy = this.form().get("detail.by")!.value as string;
    const groupBy = readGroupByFields(this.groupByArray);
    const controls = this.columnsArray.controls;
    return controls.map((control, index) => {
      const value = control.value as ReportColumnValue;
      const meta = KIND_META[value.kind];
      const disabled = isDimensionColumnDisabled(value, mode, detailBy, groupBy);
      return {
        index,
        id: value.id,
        label: value.label,
        description: this.describeColumn(value),
        kindLabel: meta.label,
        kindIcon: meta.icon,
        kindClass: meta.cssClass,
        kindTone: meta.tone,
        isCustom: this.columnReadsACustomField(value),
        isFirst: index === 0,
        isLast: index === controls.length - 1,
        disabled,
        disabledReason: disabled
          ? `"${value.label}" is hidden — you're summarizing by ${this.labelForField(detailBy)}, so a ` +
            `column can only show ${this.labelForField(detailBy)} or something you're grouping by. ` +
            `To show ${value.label}, group by it or summarize by it.`
          : "",
      };
    });
  });

  /**
   * "aggregate by" is bound straight to detail.by and grouping mutates a FormArray,
   * so re-tick the revision signal when either changes to recompute which columns
   * are disabled (a dimension column is only shown when it matches one of them).
   */
  public ngOnInit(): void {
    merge(this.form().get("detail")!.valueChanges, this.groupByArray.valueChanges)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => this.bump());

    this.addGroupControl.valueChanges
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((key) => {
        if (!key) {
          return;
        }
        this.addGroupBy(key);
        this.addGroupControl.setValue(null, { emitEvent: false });
      });
  }

  public get scopeArray(): FormArray {
    return this.form().get("scope") as FormArray;
  }

  public get groupByArray(): FormArray {
    return this.form().get("groupBy") as FormArray;
  }

  public get columnsArray(): FormArray {
    return this.form().get("columns") as FormArray;
  }

  public get filterGroup(): FormGroup {
    return this.form().get("filter") as FormGroup;
  }

  public get periodPreset(): ReportPeriod.PresetEnum {
    return this.form().get("period.preset")!.value;
  }

  public get detailMode(): ReportDetail.ModeEnum {
    return this.form().get("detail.mode")!.value;
  }

  /** The resolved date window shown under the period picker (display only). */
  public periodLabel(): string {
    const start = this.form().get("period.startDate")!.value as Date | null;
    const end = this.form().get("period.endDate")!.value as Date | null;
    if (this.periodPreset === ReportPeriod.PresetEnum.Custom && (!start || !end)) {
      return "a custom range";
    }
    return formatPeriodRange(resolvePeriodRange(this.periodPreset, start, end));
  }

  // ---- scope -------------------------------------------------------------

  public openAddGroups(): void {
    const data: AddGroupDialogData = {
      selectedGroupIds: this.scopeArray.controls.map((control) => control.value as string),
    };
    this.dialog
      .open(AddGroupDialogComponent, { ...DEFAULT_DIALOG_CONFIG, data })
      .afterClosed()
      .subscribe((result?: string[]) => {
        if (!result) {
          return;
        }
        this.scopeArray.clear();
        result.forEach((id) => this.scopeArray.push(this.formBuilder.control(id)));
        this.bump();
      });
  }

  public removeScope(index: number): void {
    this.scopeArray.removeAt(index);
    this.bump();
  }

  // ---- grouping ----------------------------------------------------------

  public addGroupBy(key: string): void {
    if (!key) {
      return;
    }
    this.groupByArray.push(buildGroupByGroup(this.formBuilder, key));
    this.bump();
  }

  /**
   * Renames the column this grouping level renders as, through the same picker a
   * regular dimension column uses — with its field locked, since the field is
   * chosen by the grouping level itself. The override is dropped when the entered
   * label is the field's own catalog label, so retyping the default resets it.
   */
  public openGroupingLabelPicker(index: number): void {
    const level = this.groupByArray.at(index);
    const field = this.fieldForLevel(index);
    const data: ColumnPickerDialogData = {
      dimensions: this.dimensions(),
      measures: this.measures(),
      existingColumns: this.columnsArray.controls.map(
        (control) => control.value as ReportColumnValue
      ),
      // A grouping level *is* a dimension column, so it is handed to the picker as
      // one: an id makes the picker treat it as an edit (seeding the label and
      // opening straight on the dimension step) rather than a new column.
      column: {
        id: GROUPING_LEVEL_COLUMN_ID,
        kind: ReportColumn.KindEnum.Dimension,
        name: field,
        label: this.headingForLevel(index),
        field,
      },
      lockField: true,
    };
    this.dialog
      .open(ColumnPickerDialogComponent, { ...DEFAULT_DIALOG_CONFIG, data })
      .afterClosed()
      .subscribe((result?: ReportColumnValue) => {
        if (!result) {
          return;
        }
        const label = result.label.trim();
        level.get("label")!.setValue(label === this.labelForField(field) ? "" : label);
        this.bump();
      });
  }

  public moveGroupBy(index: number, delta: number): void {
    this.moveInArray(this.groupByArray, index, delta);
  }

  public removeGroupBy(index: number): void {
    this.groupByArray.removeAt(index);
    this.bump();
  }

  // ---- detail mode -------------------------------------------------------

  public setDetailMode(mode: ReportDetail.ModeEnum): void {
    this.form().get("detail.mode")!.setValue(mode);
  }

  // ---- columns -----------------------------------------------------------

  public openColumnPicker(index?: number): void {
    const existing =
      index === undefined ? undefined : (this.columnsArray.at(index).value as ReportColumnValue);
    const data: ColumnPickerDialogData = {
      dimensions: this.dimensions(),
      measures: this.measures(),
      existingColumns: this.columnsArray.controls
        .map((control) => control.value as ReportColumnValue)
        .filter((column) => column.id !== existing?.id),
      column: existing,
    };
    this.dialog
      .open(ColumnPickerDialogComponent, { ...DEFAULT_DIALOG_CONFIG, data })
      .afterClosed()
      .subscribe((result?: ReportColumnValue) => {
        if (!result) {
          return;
        }
        if (index === undefined) {
          this.columnsArray.push(buildColumnGroup(this.formBuilder, result));
        } else {
          this.columnsArray.at(index).patchValue(result);
        }
        this.bump();
      });
  }

  public moveColumn(index: number, delta: number): void {
    this.moveInArray(this.columnsArray, index, delta);
  }

  public removeColumn(index: number): void {
    this.columnsArray.removeAt(index);
    this.bump();
  }

  // ---- document ----------------------------------------------------------

  public insertVariable(token: string): void {
    const control = this.form().get("document.intro")!;
    const current = (control.value as string) ?? "";
    const separator = current && !current.endsWith(" ") ? " " : "";
    control.setValue(current + separator + token);
  }

  // ---- helpers -----------------------------------------------------------

  private labelForField(key: string): string {
    return this.dimensions().find((field) => field.key === key)?.label ?? key;
  }

  /** The engine field key the grouping level at [index] nests by. */
  private fieldForLevel(index: number): string {
    return this.groupByArray.at(index).get("field")!.value as string;
  }

  /**
   * The heading the grouping level's column carries: the user's override when
   * set, otherwise the field catalog's label. Mirrors the backend's
   * buildDimensions, which resolves the same way when rendering the report.
   */
  private headingForLevel(index: number): string {
    const override = ((this.groupByArray.at(index).get("label")!.value as string) ?? "").trim();
    return override || this.labelForField(this.fieldForLevel(index));
  }

  /**
   * Whether a field key names a custom field. Resolved through the catalog rather
   * than by matching the key's shape, so a user who cannot read the custom-field
   * pool (the catalog swallows that 403) sees no badge instead of a badge on a
   * field the builder cannot even name.
   */
  private isCustomField(key: string): boolean {
    const field =
      this.dimensions().find((candidate) => candidate.key === key) ??
      this.measures().find((candidate) => candidate.key === key);
    return field?.isCustom === true;
  }

  /** A column is custom when the field it reads is — a dimension's or an aggregate's measure. */
  private columnReadsACustomField(column: ReportColumnValue): boolean {
    if (column.kind === ReportColumn.KindEnum.Dimension) {
      return this.isCustomField(column.field ?? "");
    }
    if (column.kind === ReportColumn.KindEnum.Aggregate) {
      return this.isCustomField(column.measure ?? "");
    }
    return false;
  }

  private describeColumn(column: ReportColumnValue): string {
    if (column.kind === ReportColumn.KindEnum.Dimension) {
      return this.labelForField(column.field ?? "");
    }
    if (column.kind === ReportColumn.KindEnum.Aggregate) {
      if (column.aggFunc === ReportColumn.AggFuncEnum.Count) {
        return "COUNT()";
      }
      const measure = this.measures().find((field) => field.key === column.measure)?.label ?? column.measure;
      return `${column.aggFunc}(${measure})`;
    }
    return column.expr ?? "";
  }

  private moveInArray(array: FormArray, index: number, delta: number): void {
    const target = index + delta;
    if (target < 0 || target >= array.length) {
      return;
    }
    const control = array.at(index);
    array.removeAt(index);
    array.insert(target, control);
    this.bump();
  }

  private bump(): void {
    this.revision.update((value) => value + 1);
  }
}
