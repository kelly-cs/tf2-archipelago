import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

/**
 * A mission's tier, in its own colour.
 *
 * Four colours a player learns once and then reads everywhere a mission is
 * named. The word is always there, so the colour is what makes the tier findable
 * at a glance rather than what carries the meaning.
 */
@Component({
  selector: 'app-tier',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `@if (tier()) {
    <span [class]="kind()">{{ tier() }}</span>
  }`,
  styles: `
    :host {
      display: inline-flex;
    }

    span {
      font-size: var(--text-xs);
      font-weight: 600;
      padding: 2px 9px;
      border-radius: var(--radius-pill);
      border: 1px solid currentcolor;
      white-space: nowrap;
    }

    .normal {
      color: var(--tier-normal);
      background: var(--tier-normal-bg);
    }

    .intermediate {
      color: var(--tier-intermediate);
      background: var(--tier-intermediate-bg);
    }

    .advanced {
      color: var(--tier-advanced);
      background: var(--tier-advanced-bg);
    }

    .expert {
      color: var(--tier-expert);
      background: var(--tier-expert-bg);
    }

    /* A tier gamedata does not name yet still gets a pill, in no colour: an
       uncoloured word reads as one nobody has decided about. */
    .unknown {
      color: var(--text-faint);
    }
  `,
})
export class Tier {
  readonly tier = input('');
  readonly kind = computed(() => {
    const name = this.tier().toLowerCase();
    return ['normal', 'intermediate', 'advanced', 'expert'].includes(name) ? name : 'unknown';
  });
}
