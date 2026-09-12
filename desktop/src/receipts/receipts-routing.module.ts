import { NgModule } from "@angular/core";
import { RouterModule, Routes } from "@angular/router";
import { FormMode } from "src/enums/form-mode.enum";
import { groupPermissionGuard } from "src/guards/group-permission.guard";
import { GroupGuard } from "src/guards/group.guard";
import { receiptGuardGuard } from "src/guards/receipt-guard.guard";
import { Permission } from "../open-api";
import { customFieldResolverFn } from "../resolvers/custom-field.resolver";
import { receiptResolverFn } from "../resolvers/receipt.resolver";
import { ReceiptFormComponent } from "./receipt-form/receipt-form.component";
import { ReceiptsTableComponent } from "./receipts-table/receipts-table.component";

const routes: Routes = [
  {
    path: "group/:groupId",
    component: ReceiptsTableComponent,
    canActivate: [GroupGuard],
    data: {
      groupGuardBasePath: `/receipts/group`,
    },
  },
  {
    path: "add",
    component: ReceiptFormComponent,
    resolve: {
      customFields: customFieldResolverFn,
    },
    data: {
      mode: FormMode.add,
      groupPermission: Permission.GroupReceiptsCreate,
      // Gate on the group the receipt would be created in, which is what the
      // form seeds - not merely the group being browsed. Without this a member
      // of exactly one group is bounced from /receipts/add whenever they are on
      // the synthetic "All" group, which is where login lands them.
      useAddTargetGroupId: true,
    },
    canActivate: [groupPermissionGuard],
  },
  {
    path: ":id/view",
    component: ReceiptFormComponent,
    resolve: {
      receipt: receiptResolverFn,
      customFields: customFieldResolverFn,
    },
    data: {
      mode: FormMode.view,
      permission: Permission.GroupReceiptsRead,
    },
    canActivate: [receiptGuardGuard],
  },
  {
    path: ":id/edit",
    component: ReceiptFormComponent,
    resolve: {
      receipt: receiptResolverFn,
      customFields: customFieldResolverFn,
    },
    data: {
      mode: FormMode.edit,
      permission: Permission.GroupReceiptsUpdate,
    },
    canActivate: [receiptGuardGuard],
  },
  {
    path: "",
    redirectTo: "",
    pathMatch: "full",
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class ReceiptsRoutingModule {
}
