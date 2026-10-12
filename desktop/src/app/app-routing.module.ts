import { NgModule } from "@angular/core";
import { RouterModule, Routes } from "@angular/router";
import { appPermissionGuard } from "src/guards/app-permission.guard";
import { AuthGuard } from "src/guards/auth.guard";
import { SidebarComponent } from "src/layout/sidebar/sidebar.component";
import { Permission } from "../open-api";

// set up dashboard
const routes: Routes = [
  {
    path: "",
    component: SidebarComponent,
    children: [
      {
        path: "auth",
        loadChildren: () =>
          import("../auth").then(
            (m) => m.AuthModule
          ),
        canActivate: [AuthGuard],
      },
      {
        path: "dashboard",
        loadChildren: () =>
          import("../dashboard/dashboard.module").then(
            (m) => m.DashboardModule
          ),
        canActivate: [AuthGuard],
      },
      {
        path: "categories",
        loadChildren: () =>
          import("../categories/categories.module").then(
            (m) => m.CategoriesModule
          ),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [Permission.AppCategoriesRead],
        },
      },
      {
        path: "tags",
        loadChildren: () =>
          import("../tags/tags.module").then((m) => m.TagsModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [Permission.AppTagsRead],
        },
      },
      {
        path: "custom-fields",
        loadChildren: () =>
          import("../custom-fields/custom-fields.module").then((m) => m.CustomFieldsModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [Permission.AppCustomFieldsRead],
        },
      },
      {
        path: "groups",
        loadChildren: () =>
          import("../group/group.module").then((m) => m.GroupsModule),
        canActivate: [AuthGuard],
      },
      {
        path: "receipts",
        loadChildren: () =>
          import("../receipts/receipts.module").then((m) => m.ReceiptsModule),
        canActivate: [AuthGuard],
      },
      {
        path: "settings",
        loadChildren: () =>
          import("../settings/settings.module").then((m) => m.SettingsModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [
            Permission.AppAccountRead,
            Permission.AppUserPreferencesRead,
            Permission.AppApiKeysRead,
          ],
        },
      },
      {
        path: "system-settings",
        loadChildren: () =>
          import("../system-settings/system-settings.module").then((m) => m.SystemSettingsModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [
            Permission.AppSystemSettingsRead,
            Permission.AppPromptsRead,
            Permission.AppReceiptProcessingSettingsRead,
            Permission.AppSystemEmailsRead,
            Permission.AppOidcProvidersRead,
            Permission.AppSystemTasksRead,
          ],
        },
      },
      {
        path: "users",
        loadChildren: () =>
          import("../user/user.module").then((m) => m.UserModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [Permission.AppUsersRead],
        },
      },
      {
        path: "roles",
        loadChildren: () =>
          import("../roles/roles.module").then((m) => m.RolesModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          appPermissions: [Permission.AppRolesRead],
        },
      },
      {
        path: "reports",
        loadChildren: () =>
          import("../reports/reports.module").then((m) => m.ReportsModule),
        canActivate: [AuthGuard, appPermissionGuard],
        data: {
          // read OR readAll: an admin with only the "*All" bypass still reaches reports.
          appPermissions: [Permission.AppReportsRead, Permission.AppReportsReadAll],
        },
      },
      {
        path: "",
        redirectTo: "/auth/login",
        pathMatch: "full",
      },
    ],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
