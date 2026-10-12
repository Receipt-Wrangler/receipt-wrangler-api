import { NgModule } from "@angular/core";
import { RouterModule, Routes } from "@angular/router";
import { FormMode } from "../enums/form-mode.enum";
import { appPermissionGuard } from "../guards/app-permission.guard";
import { systemSettingsLandingGuard } from "../guards/system-settings-landing.guard";
import { FormConfig } from "../interfaces";
import { Permission } from "../open-api";
import { PromptFormComponent } from "../prompt/prompt-form/prompt-form.component";
import { PromptTableComponent } from "../prompt/prompt-table/prompt-table.component";
import { promptResolver } from "../prompt/prompt.resolver";
import { promptsResolver } from "../prompt/prompts.resolver";
import {
  ReceiptProcessingSettingsFormComponent
} from "../receipt-processing-settings/receipt-processing-settings-form/receipt-processing-settings-form.component";
import {
  ReceiptProcessingSettingsTableComponent
} from "../receipt-processing-settings/receipt-processing-settings-table/receipt-processing-settings-table.component";
import { receiptProcessingSettingsResolver } from "../receipt-processing-settings/receipt-processing-settings.resolver";
import { OidcProviderTableComponent } from "./oidc-provider-table/oidc-provider-table.component";
import { allReceiptProcessingSettingsResolver } from "./resolvers/receipt-processing-settings.resolver";
import { systemEmailResolver } from "./resolvers/system-email.resolver";
import { systemSettingsResolver } from "./resolvers/system-settings.resolver";
import { SystemEmailFormComponent } from "./system-email-form/system-email-form.component";
import { allGroupsResolver } from "./system-email-table/all-groups.resolver";
import { SystemEmailTableComponent } from "./system-email-table/system-email-table.component";
import { SystemSettingsFormComponent } from "./system-settings-form/system-settings-form.component";
import { SystemSettingsComponent } from "./system-settings/system-settings.component";
import { SystemTaskTableComponent } from "./system-task-table/system-task-table.component";

const routes: Routes = [
  {
    path: "",
    canActivate: [systemSettingsLandingGuard],
    children: [],
    pathMatch: "full",
  },
  {
    path: "",
    component: SystemSettingsComponent,
    children: [
      {
        path: "oidc-providers",
        component: OidcProviderTableComponent,
        canActivate: [appPermissionGuard],
        data: {
          appPermissions: [Permission.AppOidcProvidersRead],
        },
      },
      {
        path: "system-emails",
        component: SystemEmailTableComponent,
        canActivate: [appPermissionGuard],
        data: {
          appPermissions: [Permission.AppSystemEmailsRead],
        },
        resolve: {
          allGroups: allGroupsResolver,
        }
      },
      {
        path: "prompts",
        component: PromptTableComponent,
        canActivate: [appPermissionGuard],
        data: {
          appPermissions: [Permission.AppPromptsRead],
        },
        resolve: {
          allGroups: allGroupsResolver,
          allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
        }
      },
      {
        path: "receipt-processing-settings",
        component: ReceiptProcessingSettingsTableComponent,
        canActivate: [appPermissionGuard],
        data: {
          appPermissions: [Permission.AppReceiptProcessingSettingsRead],
        },
        resolve: {
          systemSettings: systemSettingsResolver,
        }
      },
      {
        path: "system-tasks",
        component: SystemTaskTableComponent,
        canActivate: [appPermissionGuard],
        data: {
          appPermissions: [Permission.AppSystemTasksRead],
        },
        resolve: {
          prompts: promptsResolver,
          allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
        }
      },
      {
        path: "settings/view",
        component: SystemSettingsFormComponent,
        canActivate: [appPermissionGuard],
        data: {
          formConfig: {
            mode: FormMode.view,
            headerText: "View System Settings",
          } as FormConfig,
          appPermissions: [Permission.AppSystemSettingsRead],
        },
        resolve: {
          allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
          systemSettings: systemSettingsResolver,
        }
      },
      {
        path: "settings/edit",
        component: SystemSettingsFormComponent,
        canActivate: [appPermissionGuard],
        data: {
          formConfig: {
            mode: FormMode.edit,
            headerText: "Edit System Settings",
          } as FormConfig,
          appPermissions: [Permission.AppSystemSettingsUpdate],
        },
        resolve: {
          allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
          systemSettings: systemSettingsResolver,
        }
      },
    ]
  },
  {
    path: "prompts/create",
    component: PromptFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.add,
        headerText: "Create Prompt",
      } as FormConfig,
      appPermissions: [Permission.AppPromptsCreate],
    },
  },
  {
    path: "prompts/:id/view",
    component: PromptFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.view,
        headerText: "View Prompt",
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppPromptsRead],
    },
    resolve: {
      prompt: promptResolver
    }
  },
  {
    path: "prompts/:id/edit",
    component: PromptFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.edit,
        headerText: "Edit Prompt",
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppPromptsUpdate],
    },
    resolve: {
      prompt: promptResolver
    }
  },
  {
    path: "system-emails/create",
    component: SystemEmailFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.add,
        headerText: "Create System Email",
      } as FormConfig,
      appPermissions: [Permission.AppSystemEmailsCreate],
    }
  },
  {
    path: "system-emails/:id/view",
    component: SystemEmailFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.view,
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppSystemEmailsRead],
    },
    resolve: {
      systemEmail: systemEmailResolver,
      prompts: promptsResolver,
      allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
    }
  },
  {
    path: "system-emails/:id/edit",
    component: SystemEmailFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.edit,
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppSystemEmailsUpdate],
    },
    resolve: {
      systemEmail: systemEmailResolver,
      prompts: promptsResolver,
      allReceiptProcessingSettings: allReceiptProcessingSettingsResolver,
    }
  },
  {
    path: "receipt-processing-settings/create",
    component: ReceiptProcessingSettingsFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.add,
        headerText: "Create Receipt Processing Settings",
      } as FormConfig,
      appPermissions: [Permission.AppReceiptProcessingSettingsCreate],
    },
    resolve: {
      prompts: promptsResolver,
    }
  },
  {
    path: "receipt-processing-settings/:id/view",
    component: ReceiptProcessingSettingsFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.view,
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppReceiptProcessingSettingsRead],
    },
    resolve: {
      prompts: promptsResolver,
      receiptProcessingSettings: receiptProcessingSettingsResolver,
    }
  },
  {
    path: "receipt-processing-settings/:id/edit",
    component: ReceiptProcessingSettingsFormComponent,
    canActivate: [appPermissionGuard],
    data: {
      formConfig: {
        mode: FormMode.edit,
      } as FormConfig,
      setHeaderText: true,
      appPermissions: [Permission.AppReceiptProcessingSettingsUpdate],
    },
    resolve: {
      prompts: promptsResolver,
      receiptProcessingSettings: receiptProcessingSettingsResolver,
    }
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class SystemSettingsRoutingModule {
}
