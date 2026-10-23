import { Component, OnInit } from "@angular/core";
import { FormBuilder } from "@angular/forms";
import { ActivatedRoute, Router } from "@angular/router";
import { Store } from "@ngxs/store";
import { switchMap, take, tap } from "rxjs";
import { BaseFormComponent } from "src/form/base-form/base-form.component";
import { FormMode } from "../../enums/form-mode.enum";
import { Group, GroupsService, Permission } from "../../open-api";
import { SnackbarService } from "../../services";
import { AuthState, UpdateGroup } from "../../store";

@Component({
  selector: "app-group-settings",
  templateUrl: "./group-settings.component.html",
  styleUrls: ["./group-settings.component.scss"],
  standalone: false
})
export class GroupSettingsComponent
  extends BaseFormComponent
  implements OnInit {

  public editSettingsUrl: string = "";

  public group!: Group;

  public canEditEmailSettings: boolean = false;

  constructor(
    private activatedRoute: ActivatedRoute,
    private formBuilder: FormBuilder,
    private groupsService: GroupsService,
    private snackbarService: SnackbarService,
    private store: Store,
    private router: Router
  ) {
    super();
  }

  public ngOnInit(): void {
    this.canEditEmailSettings = this.store.selectSnapshot(AuthState.hasAppPermission(Permission.AppGroupsUpdateSettings));
    this.setFormConfigFromRoute(this.activatedRoute);
    if (!this.canEditEmailSettings) {
      this.formConfig.mode = FormMode.view;
    }
    this.initForm();
    this.group = this.activatedRoute.snapshot.data["group"];
    this.editSettingsUrl = `/groups/${this.group.id}/settings/edit`;
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      systemEmailId: "",
      emailIntegrationEnabled: false,
      emailBodyProcessingEnabled: false,
      emailRequireSignedSender: true,
      subjectLineRegexes: this.formBuilder.array([]),
      emailWhiteList: this.formBuilder.array([]),
      emailDefaultReceiptStatus: "",
      emailDefaultReceiptPaidById: "",
      promptId: "",
      fallbackPromptId: "",
    });
  }

  public submit(): void {
    if (this.form.valid) {
      this.groupsService
        .updateGroupSettings(this.group.id, this.form.value)
        .pipe(
          take(1),
          switchMap((updatedGroupSettings) => {
            this.group.groupSettings = updatedGroupSettings;
            return this.store.dispatch(new UpdateGroup(this.group));
          }),
          tap(() => {
            this.snackbarService.success("Group settings updated");
            this.router.navigate([`/groups/${this.group.id}/settings/view`], {
              queryParams: { tab: "settings" }
            });
          })
        )
        .subscribe();
    }
  }
}
