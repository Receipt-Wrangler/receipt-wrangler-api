import { Pipe, PipeTransform } from "@angular/core";
import {
  dateOperationOptions,
  FILTER_OPERATION_DISPLAY_VALUES,
  listOperationOptions,
  numberOperationOptions,
  textOperationOptions,
  usersOperationOptions
} from "src/constants";
import { FilterOperation } from "src/open-api";

@Pipe({
    name: "operations",
    standalone: false
})
export class OperationsPipe implements PipeTransform {
  public transform(type: string, display: boolean): string[] {
    let operationOptions: FilterOperation[] = [];

    switch (type) {
      case "date":
        return dateOperationOptions.map((option) => this.getDisplayValue(option, display));

      case "text":
        return textOperationOptions.map((option) => this.getDisplayValue(option, display));

      case "number":
        return numberOperationOptions.map((option) => this.getDisplayValue(option, display));

      case "list":
        return listOperationOptions.map((option) => this.getDisplayValue(option, display));

      case "users":
        return usersOperationOptions.map((option) => this.getDisplayValue(option, display));

      default:
        return [];
    }
  }

  private getDisplayValue(option: FilterOperation | null, display: boolean): string {
    if (display) {
      return FILTER_OPERATION_DISPLAY_VALUES?.[option ?? ""] ?? "";
    } else {
      return option ?? "";
    }
  }
}
