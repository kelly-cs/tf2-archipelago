import { Component, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { form } from '@angular/forms/signals';
import { beforeEach, describe, expect, it } from 'vitest';

import { ChoiceRow } from '@app/settings/components/choice-row';
import { NumberRow } from '@app/settings/components/number-row';
import { PasswordRow } from '@app/settings/components/password-row';
import { TextRow } from '@app/settings/components/text-row';
import { ToggleRow } from '@app/settings/components/toggle-row';
import { create } from '@bufbuild/protobuf';

import { Field, FieldSchema, Kind, Option, OptionSchema } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * One test per Kind, because the renderer switches on it: a Kind drawn as the
 * wrong control looks fine and loses the player's answer at Save. What each one
 * checks is the control it produced and the value it sends back.
 */

function field(over: Omit<Partial<Field>, '$typeName'>): Field {
  return create(FieldSchema, { id: 'a.row', label: 'A row', ...over });
}

function option(value: string, label: string): Option {
  return create(OptionSchema, { value, label });
}

@Component({
  imports: [TextRow, PasswordRow, NumberRow, ChoiceRow, ToggleRow],
  template: `
    @switch (row().kind) {
      @case (Kind.TEXT) {
        <app-text-row
          [field]="row()"
          [control]="page.value"
          [value]="held()"
          (typed)="said.set($event)"
        />
      }
      @case (Kind.PASSWORD) {
        <app-password-row [field]="row()" [control]="page.value" (typed)="said.set($event)" />
      }
      @case (Kind.NUMBER) {
        <app-number-row [field]="row()" [control]="page.value" (typed)="said.set($event)" />
      }
      @case (Kind.CHOICE) {
        <app-choice-row [field]="row()" [control]="page.value" (typed)="said.set($event)" />
      }
      @case (Kind.TOGGLE) {
        <app-toggle-row [field]="row()" [value]="held()" (typed)="said.set($event)" />
      }
    }
  `,
})
class Host {
  readonly Kind = Kind;
  readonly row = signal<Field>(field({}));
  readonly held = signal('');
  readonly said = signal('');
  readonly page = form(signal({ value: '' }));
}

describe('one control per Kind', () => {
  let host: Host;
  let element: HTMLElement;

  function draw(row: Field, held = ''): void {
    host.row.set(row);
    host.held.set(held);
    TestBed.tick();
  }

  beforeEach(() => {
    const fixture = TestBed.createComponent(Host);
    host = fixture.componentInstance;
    element = fixture.nativeElement as HTMLElement;
    fixture.detectChanges();
  });

  it('draws a text line, and a Browse row gets a button beside it', () => {
    draw(field({ kind: Kind.TEXT, placeholder: 'somewhere on this machine' }));
    const input = element.querySelector<HTMLInputElement>('input');
    expect(input?.type).toBe('text');
    expect(input?.placeholder).toBe('somewhere on this machine');
    expect(element.querySelector('app-button')).toBeNull();

    draw(field({ kind: Kind.TEXT, browse: true }));
    expect(element.querySelector('app-button')?.textContent).toContain('Browse');
  });

  // The launcher blanks a password before the model leaves it, so an empty box
  // means unchanged rather than empty. The placeholder is what says so.
  it('never shows a password back, and says the box means unchanged', () => {
    draw(field({ kind: Kind.PASSWORD }));
    const input = element.querySelector<HTMLInputElement>('input');
    expect(input?.type).toBe('password');
    expect(input?.value).toBe('');
    expect(input?.placeholder).toBe('unchanged');
  });

  it('draws a number with its bounds beside it', () => {
    draw(field({ kind: Kind.NUMBER, low: 1, high: 26 }));
    expect(element.querySelector<HTMLInputElement>('input')?.type).toBe('number');
    expect(element.querySelector('.bounds')?.textContent).toContain('1 to 26');
  });

  it('draws a choice with the words the player reads, not the saved value', () => {
    draw(
      field({
        kind: Kind.CHOICE,
        options: [
          option('progression', 'Required for progression'),
          option('useful', 'Nice to have'),
        ],
      }),
    );
    const options = element.querySelectorAll('option');
    expect(options).toHaveLength(2);
    expect(options[0].value).toBe('progression');
    expect(options[0].textContent).toBe('Required for progression');
  });

  it('sends true and false for a tick, and says a different word for each', () => {
    draw(field({ kind: Kind.TOGGLE, hint: 'in the pool', hintOff: 'left out' }), 'true');
    const tick = element.querySelector<HTMLInputElement>('input[type=checkbox]');
    expect(tick?.checked).toBe(true);
    expect(element.querySelector('.hint')?.textContent).toContain('in the pool');

    draw(field({ kind: Kind.TOGGLE, hint: 'in the pool', hintOff: 'left out' }), 'false');
    expect(element.querySelector('.hint')?.textContent).toContain('left out');

    tick?.click();
    expect(host.said()).toBe('true');
  });

  // Most rows read the same either way, and hintOff empty is what says so.
  it('uses the one word when a tick has only one', () => {
    draw(field({ kind: Kind.TOGGLE, hint: 'on' }), 'false');
    expect(element.querySelector('.hint')?.textContent).toContain('on');
  });
});
