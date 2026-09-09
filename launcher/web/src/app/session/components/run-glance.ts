import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';

import { LauncherStore } from '@app/server/launcher-store';
import { Badge } from '@app/ui/badge';
import { Panel } from '@app/ui/panel';

/**
 * The run at a glance: whether the bridge is on the multiworld, and what it has
 * counted. The numbers are the bridge's own, so a disagreement between this and
 * the tracker is the bridge's to explain.
 */
@Component({
  selector: 'app-run-glance',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge, Panel],
  template: `
    <app-panel heading="The run">
      <app-badge panelAside [tone]="connected() ? 'good' : 'neutral'">
        {{ connected() ? 'on the multiworld' : 'not connected' }}
      </app-badge>

      @if (error()) {
        <p class="error">{{ error() }}</p>
      }

      <dl>
        <div>
          <dt>Slot</dt>
          <dd>{{ health()?.slot || 'unknown' }}</dd>
        </div>
        <div>
          <dt>Seed</dt>
          <dd>{{ health()?.seed || 'unknown' }}</dd>
        </div>
        <div>
          <dt>Checks sent</dt>
          <dd>{{ health()?.checks ?? 0 }}</dd>
        </div>
        <div>
          <dt>Items held</dt>
          <dd>{{ health()?.items ?? 0 }}</dd>
        </div>
        <div>
          <dt>Death link</dt>
          <dd>{{ health()?.deathLink ? 'on' : 'off' }}</dd>
        </div>
        <div>
          <dt>Goal</dt>
          <dd>{{ health()?.goalSent ? 'sent' : 'not yet' }}</dd>
        </div>
      </dl>

      @if (health()?.lastError) {
        <p class="error">{{ health()?.lastError }}</p>
      }
    </app-panel>
  `,
  styles: `
    dl {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
      gap: var(--gap);
      margin: 0;
    }

    dt {
      font-size: var(--text-xs);
      text-transform: uppercase;
      letter-spacing: 0.04em;
      color: var(--text-faint);
    }

    dd {
      margin: 2px 0 0;
      font-family: var(--font-mono);
      font-size: var(--text-lg);
      word-break: break-all;
    }

    .error {
      margin: 0;
      color: var(--state-warn);
      font-size: var(--text-sm);
    }
  `,
})
export class RunGlance {
  private readonly store = inject(LauncherStore);

  readonly health = computed(() => this.store.health());
  readonly connected = computed(() => this.store.health()?.connected ?? false);
  readonly error = computed(() => this.store.sessionError());
}
