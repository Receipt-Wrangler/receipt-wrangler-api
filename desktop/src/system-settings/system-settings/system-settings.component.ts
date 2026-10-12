import { Component, OnInit } from "@angular/core";
import { Store } from "@ngxs/store";
import { Permission } from "../../open-api";
import { TabConfig } from "../../shared-ui/tabs/tab-config.interface";
import { AuthState } from "../../store";

interface SystemSettingsTab extends TabConfig {
  readPermission: Permission;
}

@Component({
    selector: "app-system-settings",
    templateUrl: "./system-settings.component.html",
    styleUrl: "./system-settings.component.scss",
    standalone: false
})
export class SystemSettingsComponent implements OnInit {
  public tabs: TabConfig[] = [];

  constructor(private store: Store) {
  }

  public ngOnInit(): void {
    this.initTabs();
  }

  private initTabs(): void {
    const allTabs: SystemSettingsTab[] = [
      {
        label: "System Settings",
        routerLink: "settings/view",
        name: "settings",
        readPermission: Permission.AppSystemSettingsRead,
      },
      {
        label: "Receipt Processing Settings",
        routerLink: "receipt-processing-settings",
        name: "receipt-processing-settings",
        readPermission: Permission.AppReceiptProcessingSettingsRead,
      },
      {
        label: "Prompts",
        routerLink: "prompts",
        name: "prompts",
        readPermission: Permission.AppPromptsRead,
      },
      {
        label: "System Emails",
        routerLink: "system-emails",
        name: "system-emails",
        readPermission: Permission.AppSystemEmailsRead,
      },
      {
        label: "OIDC Providers",
        routerLink: "oidc-providers",
        name: "oidc-providers",
        readPermission: Permission.AppOidcProvidersRead,
      },
      {
        label: "System Tasks",
        routerLink: "system-tasks",
        name: "system-tasks",
        readPermission: Permission.AppSystemTasksRead,
      },
    ];

    this.tabs = allTabs
      .filter((tab) =>
        this.store.selectSnapshot(AuthState.hasAppPermission(tab.readPermission))
      )
      .map(({ readPermission, ...tab }) => tab);
  }
}
