import { Component, DestroyRef, OnInit, ViewEncapsulation, inject, viewChild } from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { FormArray, FormBuilder, FormControl, FormGroup, Validators } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { Store } from "@ngxs/store";
import { take, tap } from "rxjs";
import { ReceiptFileUploadCommand } from "../../interfaces";
import { setRequired } from "../../form";
import { Category, GroupReceiptSettings, Permission, ReceiptService, ReceiptStatus, Tag } from "../../open-api";
import { SnackbarService } from "../../services";
import { AuthState, GroupState } from "../../store";
import { codePointMaxLengthValidator, trimmedRequiredValidator } from "../../validators";
import { UploadImageComponent } from "../upload-image/upload-image.component";
import { QuickScanFieldConfig, resolveQuickScanFieldConfig } from "./quick-scan-field-config";

// Mirrors the backend's models.MaxCommentLength (the Comment column is varchar(500), which
// MySQL/Postgres measure in characters) and mobile's FormBuilderTextField maxLength, so an
// over-length comment is caught here instead of coming back as a 400 after submit. Counted in
// code points, like the backend's rune count -- Validators.maxLength counts UTF-16 code units,
// which would reject 500 emoji the API accepts.
export const QUICK_SCAN_COMMENT_MAX_LENGTH = 500;

// Built once so setRequired can remove them by reference on a later recompute.
const commentMaxLengthValidator = codePointMaxLengthValidator(QUICK_SCAN_COMMENT_MAX_LENGTH);
// QuickScanCommand trims each comment on parse, so a whitespace-only one arrives empty and fails
// the group's required check -- Validators.required would have called it valid and eaten a 400.
const commentRequiredValidator = trimmedRequiredValidator();

@Component({
    selector: "app-quick-scan-dialog",
    templateUrl: "./quick-scan-dialog.component.html",
    styleUrls: ["./quick-scan-dialog.component.scss"],
    encapsulation: ViewEncapsulation.None,
    standalone: false
})
export class QuickScanDialogComponent implements OnInit {
  public readonly uploadImageComponent = viewChild.required(UploadImageComponent);

  public form: FormGroup = new FormGroup({});

  public images: ReceiptFileUploadCommand[] = [];

  public currentlySelectedIndex: number = 0;

  // base-input has no built-in message for the maxlength error, so an unmapped one would render an
  // empty mat-error. Held as a field rather than an inline object literal so the binding is stable.
  public readonly commentErrorMessages: { [key: string]: string } = {
    maxlength: `Comment must be ${QUICK_SCAN_COMMENT_MAX_LENGTH} characters or fewer.`,
  };

  private readonly commentPermissionByGroup = new Map<number, boolean>();

  private readonly destroyRef = inject(DestroyRef);

  constructor(
    private dialogRef: MatDialogRef<QuickScanDialogComponent>,
    private formBuilder: FormBuilder,
    private receiptService: ReceiptService,
    private snackbarService: SnackbarService,
    private store: Store
  ) {}

  public get paidByUserIds(): FormArray {
    return this.form.get("paidByUserIds") as FormArray;
  }

  public get statuses(): FormArray {
    return this.form.get("statuses") as FormArray;
  }

  public get groupIds(): FormArray {
    return this.form.get("groupIds") as FormArray;
  }

  public get categories(): FormArray {
    return this.form.get("categories") as FormArray;
  }

  public get tags(): FormArray {
    return this.form.get("tags") as FormArray;
  }

  public get comments(): FormArray {
    return this.form.get("comments") as FormArray;
  }

  public ngOnInit(): void {
    this.initForm();
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      paidByUserIds: this.formBuilder.array<number>([]),
      statuses: this.formBuilder.array<ReceiptStatus>([]),
      groupIds: this.formBuilder.array<number>([]),
      categories: this.formBuilder.array<Category[]>([]),
      tags: this.formBuilder.array<Tag[]>([]),
      comments: this.formBuilder.array<string>([]),
    });

    // Re-resolve each image's field config whenever anything changes (a group selection can flip
    // which fields are shown/required for that image).
    this.form.valueChanges
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => this.configureImages());
  }

  public fileLoaded(fileData: ReceiptFileUploadCommand): void {
    if (fileData.file && !fileData.encodedImage) {
      fileData.url = URL.createObjectURL(fileData.file);
    }
    this.images.push(fileData);
    const userPreferences = this.store.selectSnapshot(AuthState.userPreferences);

    // Group is always required; paid-by/status validators are applied per-image by configureImages.
    this.paidByUserIds.push(new FormControl(userPreferences?.quickScanDefaultPaidById ?? ""));
    this.statuses.push(new FormControl(userPreferences?.quickScanDefaultStatus ?? ""));
    // The user's quick-scan default group wins; failing that, a member of exactly
    // one group has no choice to make, so seed it. configureImages() below then
    // applies that group's show/require config exactly as a manual pick would.
    this.groupIds.push(new FormControl(
      userPreferences?.quickScanDefaultGroupId
        ?? this.store.selectSnapshot(GroupState.soleGroupId)
        ?? "",
      Validators.required
    ));
    // Categories/tags are multi-selects backed by app-category/tag-autocomplete,
    // which push selections onto a FormArray (see the receipt form). A plain
    // FormControl has no push(), so selecting one would throw — use a FormArray.
    this.categories.push(this.formBuilder.array([]));
    this.tags.push(this.formBuilder.array([]));
    // The comment is a scalar text field, so a plain FormControl - unlike categories/tags above.
    // maxLength is applied here rather than in configureImages because setRequired composes
    // validators additively, so it survives every show/require recompute and is never re-added.
    this.comments.push(new FormControl("", commentMaxLengthValidator));

    this.configureImages();
  }

  // Returns the receipt settings for the group currently selected on the given image, if any.
  public settingsForIndex(index: number): GroupReceiptSettings | undefined {
    const groupId = this.groupIds.at(index)?.value;
    if (!groupId) {
      return undefined;
    }

    return this.store.selectSnapshot(GroupState.getGroupById(groupId.toString()))?.groupReceiptSettings;
  }

  // Whether the user has picked a group for this image. This -- not "did settings resolve" -- is
  // what gates the field set: a group id the store does not know (a stale quickScanDefaultGroupId,
  // or AppData not carrying that group) is still a choice, and keeps the backend defaults.
  private hasGroupAt(index: number): boolean {
    return !!this.groupIds.at(index)?.value;
  }

  // Resolved per call rather than cached per index: removeImage() shifts every later image down, so
  // an index-keyed cache would hand an image another one's config. The template reads the show*
  // getters on every change-detection pass, which is what keeps them correct as the group changes.
  //
  // The comment field additionally requires group.comments.create: without it the field is hidden,
  // never required, and a comment sent anyway is dropped server-side. hideComments hides the whole
  // group's comments, so it hides this too. Mirrors the backend's IsQuickScanCommentShown.
  private configForIndex(index: number): QuickScanFieldConfig {
    return resolveQuickScanFieldConfig(this.settingsForIndex(index), {
      hasGroup: this.hasGroupAt(index),
      canCreateComments: this.canCommentForIndex(index),
    });
  }

  public showPaidBy(index: number): boolean {
    return this.configForIndex(index).showPaidBy;
  }

  public showStatus(index: number): boolean {
    return this.configForIndex(index).showStatus;
  }

  public showCategories(index: number): boolean {
    return this.configForIndex(index).showCategories;
  }

  public showTags(index: number): boolean {
    return this.configForIndex(index).showTags;
  }

  public showComment(index: number): boolean {
    return this.configForIndex(index).showComment;
  }

  // Cached per group: AuthState.hasGroupPermission allocates a new selector on each call, and
  // showComment is read from the template on every change-detection pass.
  private canCommentForIndex(index: number): boolean {
    const groupId = Number(this.groupIds.at(index)?.value);
    if (!groupId) {
      return false;
    }

    if (!this.commentPermissionByGroup.has(groupId)) {
      this.commentPermissionByGroup.set(
        groupId,
        this.store.selectSnapshot(AuthState.hasGroupPermission(groupId, Permission.GroupCommentsCreate))
      );
    }

    return this.commentPermissionByGroup.get(groupId) ?? false;
  }

  public categoriesForIndex(index: number): Category[] {
    const groupId = this.groupIds.at(index)?.value;
    return groupId ? this.store.selectSnapshot(AuthState.groupCategories(Number(groupId))) : [];
  }

  public tagsForIndex(index: number): Tag[] {
    const groupId = this.groupIds.at(index)?.value;
    return groupId ? this.store.selectSnapshot(AuthState.groupTags(Number(groupId))) : [];
  }

  // Applies the per-image required validators and clears hidden fields so the server backfills their
  // group-configured defaults (a hidden paid-by/status must be sent empty, not with a stale value).
  private configureImages(): void {
    for (let i = 0; i < this.images.length; i++) {
      const config = this.configForIndex(i);

      // Required is recomputed unconditionally: with no group nothing but Group is required, and a
      // required flag left over from a group the user just cleared has to come off.
      setRequired(this.form.get("paidByUserIds." + i), config.requirePaidBy);
      setRequired(this.form.get("statuses." + i), config.requireStatus);
      setRequired(this.form.get("categories." + i), config.requireCategories);
      setRequired(this.form.get("tags." + i), config.requireTags);
      setRequired(this.form.get("comments." + i), config.requireComment, commentRequiredValidator);

      // Clearing is gated on a group being SELECTED. "Hidden because this group's config hides it"
      // must clear, so the server backfills its configured default instead of receiving a stale
      // value. "Hidden because no group is picked yet" must NOT: the values sitting there are the
      // caller's quickScanDefault* prefills, and every FormArray.push in fileLoaded() emits
      // valueChanges, so this runs before groupIds is even pushed -- an ungated clear would wipe
      // those prefills on the very first pass and never put them back.
      if (!this.hasGroupAt(i)) {
        continue;
      }

      if (!config.showPaidBy) {
        this.form.get("paidByUserIds." + i)?.setValue("", { emitEvent: false });
      }
      if (!config.showStatus) {
        this.form.get("statuses." + i)?.setValue("", { emitEvent: false });
      }
      if (!config.showCategories) {
        (this.form.get("categories." + i) as FormArray | null)?.clear({ emitEvent: false });
      }
      if (!config.showTags) {
        (this.form.get("tags." + i) as FormArray | null)?.clear({ emitEvent: false });
      }
      if (!config.showComment) {
        this.form.get("comments." + i)?.setValue("", { emitEvent: false });
      }
    }
  }

  public openImageUploadComponent(): void {
    this.uploadImageComponent().clickInput();
  }

  public removeImage(index: number): void {
    this.paidByUserIds.removeAt(index);
    this.statuses.removeAt(index);
    this.groupIds.removeAt(index);
    this.categories.removeAt(index);
    this.tags.removeAt(index);
    this.comments.removeAt(index);
    this.images.splice(index, 1);
  }

  public submitButtonClicked(): void {
    if (this.form.valid && this.images.length > 0) {
      this.receiptService
        .quickScanReceipt(
          this.images.map((i) => i.file),
          this.groupIds.value,
          this.paidByUserIds.value,
          this.statuses.value,
          this.joinIds(this.categories),
          this.joinIds(this.tags),
          // Already one string per image - no id joining needed, unlike categories/tags.
          this.comments.value
        )
        .pipe(
          take(1),
          tap(() => {
            const imageWord = this.images.length === 1 ? "image" : "images";
            this.snackbarService.success(`Successfully queued ${imageWord} for processing`);
            this.dialogRef.close();
          }),
        )
        .subscribe();
    }
    if (this.images.length === 0) {
      this.snackbarService.error("Please select images to upload");
    }
    if (this.form.invalid) {
      this.snackbarService.error("Please fill in all required fields. Some images are missing required fields.");
    }
  }

  // Serializes each image's selected category/tag objects into a comma-joined id string (one entry
  // per image), matching the multipart shape the API expects.
  private joinIds(array: FormArray): string[] {
    return array.controls.map((control) =>
      ((control.value ?? []) as { id: number }[]).map((entity) => entity.id).join(",")
    );
  }

  public cancelButtonClicked(): void {
    this.dialogRef.close();
  }

  public navigateImages(delta: number): void {
    const newValue = this.currentlySelectedIndex + delta;
    if (newValue === -1) {
      this.currentlySelectedIndex = this.images.length - 1;
      return;
    }

    if (newValue > this.images.length - 1) {
      this.currentlySelectedIndex = 0;
      return;
    }

    this.currentlySelectedIndex = newValue;
  }
}
