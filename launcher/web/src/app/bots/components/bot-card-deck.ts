import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { CdkDrag, CdkDragDrop, CdkDragHandle, CdkDropList } from '@angular/cdk/drag-drop';

import { BotTradingCard } from '@cards/bot-card';
import { BotCard, BotForm, botCards, cardById } from '@cards/catalog';
import { SettingsStore } from '@app/settings/settings-store';
import { LauncherStore } from '@app/server/launcher-store';

@Component({
  selector: 'app-bot-card-deck',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [BotTradingCard, CdkDrag, CdkDragHandle, CdkDropList],
  templateUrl: './bot-card-deck.html',
  styleUrl: './bot-card-deck.scss',
})
export class BotCardDeck {
  private readonly settings = inject(SettingsStore);
  readonly attached = inject(LauncherStore).managedExternally;
  readonly cards = botCards;
  readonly feedback = signal('');
  readonly selected = computed(() =>
    this.settings
      .value('bots.priority')
      .split(',')
      .filter(Boolean)
      .map(cardById)
      .filter((card): card is BotCard => card !== undefined),
  );
  readonly available = computed(() => {
    const selected = new Set(this.selected().map((card) => card.id));
    return botCards.filter((card) => !selected.has(card.id));
  });

  form(card: BotCard): BotForm {
    const value = this.settings.value(`bots.card.${card.id}.form`);
    return value === 'human' || value === 'giant' ? value : 'robot';
  }

  setForm(card: BotCard, form: string): void {
    if (form !== 'human' && form !== 'robot' && form !== 'giant' && form !== 'reroll') return;
    this.settings.change(`bots.card.${card.id}.form`, form);
    this.feedback.set(`${card.name} is ${this.form(card)}. Press Apply to save.`);
  }

  useDemoSquad(): void {
    this.settings.change('bots.demo_lineup', 'demo');
    this.feedback.set(
      'Six-card lineup drafted with a random two-of-each form mix. Press Apply to save it.',
    );
  }

  select(card: BotCard): void {
    const seat = this.recruitSeat();
    if (seat === undefined) {
      this.feedback.set('All six collectible seats are filled. Remove a card first.');
      return;
    }
    const replaced = this.settings.value(seat) !== '';
    this.settings.change(seat.replace('.class', '.card'), card.id);
    this.feedback.set(
      replaced
        ? `${card.name} replaced the manual bot in Seat ${this.seatNumber(seat)}. Press Apply to save.`
        : `${card.name} added to the team.`,
    );
  }

  recruitLabel(card: BotCard): string {
    const seat = this.recruitSeat();
    if (seat === undefined) return `Recruit ${card.name}`;
    return this.settings.value(seat) === ''
      ? `Recruit ${card.name}`
      : `Replace manual Seat ${this.seatNumber(seat)} with ${card.name}`;
  }

  remove(card: BotCard): void {
    const seat = this.cardSeat(card);
    if (seat !== undefined) {
      this.settings.change(seat, '');
      this.feedback.set(`${card.name} moved back to the collection.`);
    }
  }

  move(card: BotCard, direction: -1 | 1): void {
    const index = this.selected().findIndex((entry) => entry.id === card.id);
    this.reorder(index, index + direction);
  }

  drop(event: CdkDragDrop<BotCard[]>): void {
    this.reorder(event.previousIndex, event.currentIndex);
  }

  private recruitSeat(): string | undefined {
    const seats = this.settings
      .fieldsMatching('bots.seat.')
      .filter((field) => field.id.endsWith('.class'));
    const empty = seats.find((field) => this.settings.value(field.id) === '');
    if (empty) return empty.id;
    // A full manually configured roster should not block the first card.
    // Replace the lowest-priority manual seat, preserving selected cards.
    return seats
      .slice()
      .reverse()
      .find((field) => this.settings.value(field.id.replace('.class', '.card')) === '')?.id;
  }

  private seatNumber(id: string): number {
    return Number(id.split('.')[2]) + 1;
  }

  private cardSeat(card: BotCard): string | undefined {
    const fields = this.settings.fieldsMatching('bots.seat.');
    return fields.find(
      (field) => field.id.endsWith('.card') && this.settings.value(field.id) === card.id,
    )?.id;
  }

  private reorder(from: number, to: number): void {
    const ids = this.selected().map((card) => card.id);
    if (from < 0 || to < 0 || from >= ids.length || to >= ids.length || from === to) return;
    const [moved] = ids.splice(from, 1);
    ids.splice(to, 0, moved);
    this.settings.change('bots.priority', ids.join(','));
    this.feedback.set('Priority changed. Lower cards yield to arriving players first.');
  }
}
