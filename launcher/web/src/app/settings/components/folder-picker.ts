import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import { rxResource } from '@angular/core/rxjs-interop';

import { LauncherCommands } from '@app/server/launcher-commands';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';

/**
 * A folder picker without a native dialog.
 *
 * The window had one from walk and a browser has none, so the launcher reads a
 * folder and this draws it: one behaviour on Windows and on Linux, and one that
 * can be tested. The typed path in the row beside it stays editable, because
 * typing a path is still the fastest way to reach one you already know.
 */
@Component({
  selector: 'app-folder-picker',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, EmptyState],
  template: `
    <div class="picker" role="dialog" aria-label="Choose a folder">
      <header>
        <app-button
          size="small"
          [disabled]="!folder.value()?.parent"
          (press)="goTo(folder.value()?.parent ?? '')"
        >
          Up
        </app-button>
        <span class="here">{{ folder.value()?.path ?? 'reading' }}</span>
      </header>

      @if (folder.isLoading()) {
        <app-empty-state>Reading the folder.</app-empty-state>
      } @else if ((folder.value()?.folders ?? []).length === 0) {
        <app-empty-state>Nothing inside this one.</app-empty-state>
      } @else {
        <ul>
          @for (name of folder.value()?.folders ?? []; track name) {
            <li>
              <button type="button" (click)="goTo(join(name))">{{ name }}</button>
            </li>
          }
        </ul>
      }

      <footer>
        <app-button tone="ghost" size="small" (press)="dismissed.emit()">Cancel</app-button>
        <app-button tone="primary" size="small" (press)="chosen.emit(folder.value()?.path ?? '')">
          Use this folder
        </app-button>
      </footer>
    </div>
  `,
  styleUrl: './folder-picker.scss',
})
export class FolderPicker {
  private readonly commands = inject(LauncherCommands);

  /** Where to start. Empty lands where the player's own files are. */
  readonly start = input('');

  readonly chosen = output<string>();
  readonly dismissed = output<void>();

  private readonly at = signal<string | undefined>(undefined);
  private readonly asked = computed(() => this.at() ?? this.start());

  readonly folder = rxResource({
    params: () => this.asked(),
    stream: ({ params }) => this.commands.listFolder(params),
  });

  goTo(path: string): void {
    if (path !== '') {
      this.at.set(path);
    }
  }

  join(name: string): string {
    const here = this.folder.value()?.path ?? '';
    return here.endsWith('/') || here.endsWith('\\') ? here + name : `${here}/${name}`;
  }
}
