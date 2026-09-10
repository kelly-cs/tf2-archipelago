import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, concatMap, tap } from 'rxjs';

import { LauncherStore } from '@app/server/launcher-store';
import { SettingsStore } from '@app/settings/settings-store';
import { Button } from '@app/ui/button';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/** One seat, as the editor draws it: who plays it and what they carry. */
interface Seat {
  readonly number: number;
  readonly label: string;
  readonly playedBy: Field;
  readonly carries: Field | undefined;
}

/**
 * The lineup: a class and a loadout for each of RED's seats.
 *
 * Every one of these is a row form already declares, so this picks them out of
 * the model by id rather than inventing a second way to say who plays seat two.
 * A change goes to the launcher immediately and reaches a running server on the
 * bot's next respawn, with no restart: saveplan calls that a Team change and
 * botlive turns it into console commands.
 */
@Component({
  selector: 'app-bot-lineup',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button],
  templateUrl: './bot-lineup.html',
  styleUrl: './bot-lineup.scss',
})
export class BotLineup {
  private readonly settings = inject(SettingsStore);
  private readonly launcher = inject(LauncherStore);

  readonly open = computed(() => this.settings.open());
  readonly notRunning = computed(() => !this.launcher.running());
  readonly feedback = signal('');

  readonly teamSize = computed(() => this.settings.field('bots.team_size'));
  readonly savedTeams = computed(() => this.settings.field('bots.team_preset'));
  readonly teamName = computed(() => this.settings.field('bots.team_name'));

  readonly seats = computed<Seat[]>(() => {
    const classes = this.settings.fieldsMatching('bots.seat.');
    const seats: Seat[] = [];
    for (const played of classes.filter((one) => one.id.endsWith('.class'))) {
      const number = seats.length + 1;
      seats.push({
        number,
        label: `Seat ${number}`,
        playedBy: played,
        carries: classes.find((one) => one.id === played.id.replace('.class', '.loadout')),
      });
    }
    return seats;
  });

  /** What Save says, which depends on whether the typed name is already one of
      the saved lineups. */
  readonly saveLabel = computed(() => {
    const typed = this.value(this.teamName()).trim().toLowerCase();
    const known = (this.savedTeams()?.options ?? []).some(
      (option) => option.label.toLowerCase() === typed && typed !== '',
    );
    return known ? 'Update' : 'Save as new';
  });

  readonly openSettings = new Subject<void>();
  readonly fired = new Subject<string>();
  readonly answered = new Subject<{ id: string; value: string; said: string }>();

  constructor() {
    this.answered
      .pipe(
        tap((change) => {
          this.settings.change(change.id, change.value);
          this.feedback.set(change.said);
        }),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.openSettings
      .pipe(
        concatMap(() => this.settings.openSettings('Bots')),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.fired
      .pipe(
        concatMap((id) => this.settings.dispatch(id)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  setName(value: string): void {
    const field = this.teamName();
    if (field !== undefined) {
      this.settings.change(field.id, value);
    }
  }

  loadTeam(value: string): void {
    const field = this.savedTeams();
    if (field === undefined) {
      return;
    }
    const chosen = field.options.find((option) => option.value === value);
    this.answered.next({
      id: field.id,
      value,
      said: `Loaded ${chosen?.label ?? value}. Bots switch on their next respawn.`,
    });
  }

  saveTeam(): void {
    this.fired.next('bots.save_team');
    this.feedback.set('Saved the current seats.');
  }

  removeTeam(): void {
    this.fired.next('bots.remove_team');
    this.feedback.set('Removed that lineup.');
  }

  value(field: Field | undefined): string {
    return field === undefined ? '' : this.settings.value(field.id);
  }

  setSeatClass(seat: Seat, value: string): void {
    const chosen = seat.playedBy.options.find((option) => option.value === value);
    this.answered.next({
      id: seat.playedBy.id,
      value,
      said: `${seat.label} → ${chosen?.label ?? value}. Applies on the bot's next respawn.`,
    });
  }

  setSeatLoadout(seat: Seat, value: string): void {
    if (seat.carries === undefined) {
      return;
    }
    const chosen = seat.carries.options.find((option) => option.value === value);
    this.answered.next({
      id: seat.carries.id,
      value,
      said: `${seat.label} carries ${chosen?.label ?? value}.`,
    });
  }

  setTeamSize(value: string): void {
    const field = this.teamSize();
    if (field !== undefined) {
      this.answered.next({ id: field.id, value, said: `RED now fills to ${value}.` });
    }
  }
}
