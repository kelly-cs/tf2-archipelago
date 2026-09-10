import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * A raised area with a heading. Every screen is made of these, so the border,
 * the radius and the space inside are decided once.
 */
@Component({
  selector: 'app-panel',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <section [attr.aria-label]="heading() || null">
      @if (heading()) {
        <header>
          <h2>{{ heading() }}</h2>
          <ng-content select="[panelAside]" />
        </header>
      }
      <ng-content />
    </section>
  `,
  styles: `
    :host {
      display: block;
    }

    section {
      display: flex;
      flex-direction: column;
      gap: var(--gap);
      height: 100%;
      background: var(--surface-panel);
      border: 1px solid var(--line-soft);
      border-radius: var(--radius-panel);
      padding: var(--gap-lg) 20px;
      box-sizing: border-box;
    }

    header {
      display: flex;
      justify-content: space-between;
      align-items: baseline;
      gap: var(--gap-sm);
    }

    h2 {
      font-size: var(--title);
    }
  `,
})
export class Panel {
  readonly heading = input('');
}
