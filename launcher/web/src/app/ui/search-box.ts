import { ChangeDetectionStrategy, Component, input, model } from '@angular/core';
import { FormsModule } from '@angular/forms';

/**
 * A filter over a list. Two-way through model(), because what is typed here is
 * never saved anywhere: it narrows what is on screen and leaving the page
 * throws it away.
 */
@Component({
  selector: 'app-search-box',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  template: `
    <label>
      <span class="sr-only">{{ label() }}</span>
      <input
        type="search"
        [attr.placeholder]="label()"
        [ngModel]="text()"
        (ngModelChange)="text.set($event)"
      />
    </label>
  `,
  styles: `
    @use 'parts';

    :host {
      display: block;
    }

    label {
      display: block;
    }

    input {
      width: 100%;
      background: var(--surface-void);
      border: 1px solid var(--line);
      border-radius: var(--radius);
      color: var(--text);
      font-family: var(--font-body);
      font-size: var(--text-md);
      padding: 7px 10px;
      transition: border-color var(--quick) var(--ease-out);
    }

    input:hover {
      border-color: var(--line-strong);
    }

    .sr-only {
      @include parts.only-for-screen-readers;
    }
  `,
})
export class SearchBox {
  readonly label = input('Search');
  readonly text = model('');
}
