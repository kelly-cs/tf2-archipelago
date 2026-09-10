import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';

import { BotLineup } from '@app/bots/components/bot-lineup';
import { SettingsStore } from '@app/settings/settings-store';
import { Chip } from '@app/ui/chip';
import { Panel } from '@app/ui/panel';

/**
 * The Bot Switcher: who holds RED's seats, and which classes the mod may draw
 * for the ones nobody named.
 *
 * Both halves are rows form already declares, read straight off the model:
 * which classes exist, what each is called and whether it is allowed all come
 * from there, so a class the game adds later appears here without this file
 * knowing about it. A class allowed here is the same answer the settings screen
 * holds, so the two cannot disagree.
 */
@Component({
  selector: 'app-bots-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [BotLineup, Chip, Panel],
  template: `
    <app-panel heading="Bot Switcher">
      <app-bot-lineup />
    </app-panel>

    <app-panel heading="Classes the mod may draw">
      <p class="lead">For seats left to the mod. Click to allow or forbid.</p>

      @if (chips().length === 0) {
        <p class="lead">Open the lineup to change these.</p>
      } @else {
        <div role="group" aria-label="Classes the mod may draw" class="chips">
          @for (mercenary of chips(); track mercenary.id) {
            <app-chip
              tone="allow"
              [pressed]="mercenary.allowed"
              [hint]="
                mercenary.allowed
                  ? mercenary.name + ' may be drawn'
                  : mercenary.name + ' is forbidden'
              "
              (press)="flip(mercenary.id)"
            >
              {{ mercenary.name }}
            </app-chip>
          }
        </div>
      }
    </app-panel>
  `,
  styleUrl: './bots-page.scss',
})
export class BotsPage {
  private readonly settings = inject(SettingsStore);

  /** Only the allowed rows. bots.class.<key>.loadout is the class's default
      loadout and lives on the settings screen, not on a chip. */
  readonly chips = computed(() =>
    this.settings
      .fieldsMatching('bots.class.')
      .filter((field) => field.id.endsWith('.allowed'))
      .map((field) => ({
        id: field.id,
        name: field.label,
        allowed: this.settings.value(field.id) === 'true',
      })),
  );

  flip(id: string): void {
    const allowed = this.settings.value(id) === 'true';
    // Never all nine forbidden: a lineup the mod cannot draw from leaves the
    // seats empty, and an empty seat is a wave short of six defenders.
    if (allowed && this.chips().filter((one) => one.allowed).length === 1) {
      return;
    }
    this.settings.change(id, allowed ? 'false' : 'true');
  }
}
