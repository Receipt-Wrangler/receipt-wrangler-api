import { Component, Input, OnInit } from "@angular/core";
import { FormBuilder, FormGroup, Validators } from "@angular/forms";
import { MatDialogRef } from "@angular/material/dialog";
import { take, tap } from "rxjs";
import { API_KEY_SCOPE_OPTIONS } from "../../constants";
import { ApiKeyResult, ApiKeyScope, ApiKeyService, ApiKeyView, UpsertApiKeyCommand } from "../../open-api";
import { SnackbarService } from "../../services";

@Component({
  selector: "app-api-key-form-dialog",
  templateUrl: "./api-key-form-dialog.component.html",
  styleUrls: ["./api-key-form-dialog.component.scss"],
  standalone: false
})
export class ApiKeyFormDialogComponent implements OnInit {
  @Input() public headerText: string = "";
  @Input() public apiKey?: ApiKeyView;

  public form!: FormGroup;
  public apiKeyResult?: ApiKeyResult;
  public apiKeyScopes = API_KEY_SCOPE_OPTIONS;

  constructor(
    private dialogRef: MatDialogRef<ApiKeyFormDialogComponent>,
    private formBuilder: FormBuilder,
    private apiKeyService: ApiKeyService,
    private snackbarService: SnackbarService
  ) {}

  public ngOnInit(): void {
    this.initForm();
  }

  private initForm(): void {
    this.form = this.formBuilder.group({
      name: [this.apiKey?.name ?? "", Validators.required],
      description: [this.apiKey?.description ?? ""],
      scope: [this.apiKey?.scope ?? ApiKeyScope.R, Validators.required]
    });
  }

  public submitButtonClicked(): void {
    if (this.form.valid) {
      const command: UpsertApiKeyCommand = this.form.value;

      if (this.apiKey && this.apiKey.id) {
        this.apiKeyService.updateApiKey(this.apiKey.id, command)
          .pipe(
            take(1),
            tap(() => {
              this.snackbarService.success("API key updated successfully");
              this.dialogRef.close(true);
            })
          )
          .subscribe();
      } else {
        this.apiKeyService.createApiKey(command)
          .pipe(
            take(1),
            tap((result: ApiKeyResult) => {
              this.apiKeyResult = result;
              this.snackbarService.success("API key created successfully");
            })
          )
          .subscribe();
      }
    } else {
      this.snackbarService.error("Please fill in all required fields");
    }
  }

  public cancelButtonClicked(): void {
    this.dialogRef.close();
  }

  public copyToClipboard(): void {
    if (this.apiKeyResult?.key) {
      navigator.clipboard.writeText(this.apiKeyResult.key).then(() => {
        this.snackbarService.success("API key copied to clipboard");
      }).catch(() => {
        this.snackbarService.error("Failed to copy API key to clipboard");
      });
    }
  }

  public closeDialog(): void {
    this.dialogRef.close(this.apiKey ? true : this.apiKeyResult);
  }
}
