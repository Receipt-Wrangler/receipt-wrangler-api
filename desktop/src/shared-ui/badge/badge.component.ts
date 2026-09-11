import { Component, input } from "@angular/core";

/** The colour a badge is drawn in. See `badge.component.scss` for the pairings. */
export type BadgeTone = "purple" | "slate" | "blue" | "green";

/** The badge text marking a custom field wherever one is listed. */
export const CUSTOM_FIELD_BADGE = "Custom";

/**
 * A small uppercase badge for marking an item in a list — a custom field, a
 * column's kind, anything that needs a word of classification beside its label.
 *
 * Presentational and input-driven. Purple is the tone that marks a custom field,
 * so it is the default: that is what the badge is for nearly everywhere it
 * appears, and hard-coding it at each call site is how the two implementations
 * this replaced ended up different colours.
 *
 * @example
 * <app-badge [text]="customFieldBadge"></app-badge>
 * <app-badge text="Agg" tone="blue"></app-badge>
 */
@Component({
  selector: "app-badge",
  templateUrl: "./badge.component.html",
  styleUrl: "./badge.component.scss",
  standalone: true,
})
export class BadgeComponent {
  public readonly text = input.required<string>();

  public readonly tone = input<BadgeTone>("purple");
}
