import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * The mark: Archipelago's five circles with Mann Co.'s wrench across them.
 * Inline rather than an asset so it takes the page's colours and never flashes
 * in after the rest of the header.
 */
@Component({
  selector: 'app-logo',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <svg
      viewBox="0 0 200 200"
      [attr.width]="size()"
      [attr.height]="size()"
      role="img"
      aria-label="Mann vs Archipelago"
    >
      <circle cx="100" cy="46" r="36" fill="#c9788a" />
      <circle cx="47" cy="78" r="36" fill="#ebe08a" />
      <circle cx="153" cy="78" r="36" fill="#7cc47a" />
      <circle cx="47" cy="138" r="36" fill="#7a7fc0" />
      <circle cx="153" cy="138" r="36" fill="#cb96c4" />
      <g transform="translate(100 160) rotate(-8)">
        <circle r="38" fill="#b5561f" />
        <rect x="-42" y="-4.5" width="84" height="9" fill="#241b10" />
        <rect x="-4.5" y="-42" width="9" height="84" fill="#241b10" />
        <circle r="13" fill="#241b10" />
      </g>
    </svg>
  `,
  styles: `
    :host {
      display: inline-flex;
      flex: none;
    }
  `,
})
export class Logo {
  readonly size = input(46);
}
