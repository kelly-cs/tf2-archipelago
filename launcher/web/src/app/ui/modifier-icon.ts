import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/** Compact HUD-inspired silhouettes for mission modifier labels. */
@Component({
  selector: 'app-modifier-icon',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <svg viewBox="0 0 32 32" aria-hidden="true">
      @switch (modifier()) {
        @case ('low_gravity') {
          <path d="M16 27V7M9 14l7-7 7 7M10 26h12" />
        }
        @case ('high_gravity') {
          <path d="M16 5v20M9 18l7 7 7-7M9 6h14" />
        }
        @case ('blast_plating') {
          <path d="M16 3l10 4v8c0 7-4 11-10 14C10 26 6 22 6 15V7z" />
          <path d="M13 9l3 4 4-3-2 5 5 2-6 1-1 6-2-5-5 2 4-5-4-3 5 1z" />
        }
        @case ('thermal_shielding') {
          <path
            d="M17 3c2 6-3 7 1 11 2-3 5-3 6-7 5 8 3 21-8 22C5 29 4 18 9 12c0 5 3 5 4 7 3-5-2-9 4-16z"
          />
        }
        @case ('ballistic_plating') {
          <path d="M5 12h15l7 4-7 4H5zM8 12V8h9v4M8 20v4h9v-4" />
          <path d="M19 12v8" />
        }
        @case ('fragile_mercenaries') {
          <path d="M16 28S4 21 4 12C4 5 13 3 16 9c3-6 12-4 12 3 0 9-12 16-12 16z" />
          <path d="M18 8l-5 8 5 1-4 10" />
        }
        @case ('overclocked_servos') {
          <path d="M3 9h9l7 7-7 7H3l7-7zM15 9h5l8 7-8 7h-5l8-7z" />
        }
        @case ('loaded_dice') {
          <rect x="5" y="5" width="22" height="22" rx="4" />
          <circle cx="11" cy="11" r="1.6" />
          <circle cx="21" cy="11" r="1.6" />
          <circle cx="16" cy="16" r="1.6" />
          <circle cx="11" cy="21" r="1.6" />
          <circle cx="21" cy="21" r="1.6" />
        }
        @case ('weaponized_tanks') {
          <path d="M4 19h23l-3 7H8zM8 19l2-8h13l3 8M14 11V7h8v4M22 8l6-2" />
          <circle cx="11" cy="23" r="2" />
          <circle cx="21" cy="23" r="2" />
        }
        @case ('miniature_menace') {
          <rect x="8" y="9" width="16" height="14" rx="2" />
          <path d="M16 9V5M12 27v-4M20 27v-4M4 13h4M24 13h4" />
          <circle cx="13" cy="15" r="1.5" />
          <circle cx="19" cy="15" r="1.5" />
          <path d="M12 19h8" />
        }
        @case ('bot_surge') {
          <path d="M3 25l7-8 5 5 7-12 7 15M4 8h8M8 4v8" />
        }
        @case ('faulty_calibration') {
          <circle cx="16" cy="16" r="8" />
          <path d="M16 3v6M16 23v6M3 16h6M23 16h6M13 15l7 4M20 19l4-1" />
        }
        @case ('loose_footing') {
          <path d="M7 6c4 5 7 7 13 9l5 2c2 1 2 5-1 6l-8 2c-5 1-9-1-10-5L4 14" />
          <path d="M4 25h7M16 29h10M2 19h5" />
        }
        @case ('mental') {
          <path d="M8 20V11l4-5h8l4 5v9l-4 5h-8zM8 14H5m3 5H3m21-5h3m-3 5h5" />
          <path d="M11 13h10v5H11zM14 22h4" />
          <circle cx="14" cy="15.5" r="1" />
          <circle cx="18" cy="15.5" r="1" />
        }
        @case ('make_it_count') {
          <path d="M12 27V12l4-7 4 7v15zM12 20h8M10 29h12" />
          <path d="M9 9 6 6m20 0-3 3M8 16H4m20 0h4M9 23l-3 3m20 0-3-3" />
        }
        @default {
          <path d="M16 3l13 23H3zM16 11v7M16 23v1" />
        }
      }
    </svg>
  `,
  styles: `
    :host {
      display: grid;
      width: 24px;
      height: 24px;
      flex: 0 0 24px;
      place-items: center;
    }

    svg {
      width: 100%;
      height: 100%;
      overflow: visible;
      fill: none;
      stroke: currentcolor;
      stroke-linecap: round;
      stroke-linejoin: round;
      stroke-width: 2;
    }

    circle {
      fill: currentcolor;
      stroke: none;
    }
  `,
})
export class ModifierIcon {
  readonly modifier = input.required<string>();
}
