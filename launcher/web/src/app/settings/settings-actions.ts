import { Injectable, inject } from '@angular/core';
import { Observable, map, of, reduce } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { SettingsStore } from '@app/settings/settings-store';
import { orRefusal } from '@app/transport/refusal';
import { FileTarget } from '@gen/tf2ap/launcher/v1/files_pb';

/**
 * Where a settings button goes.
 *
 * Most of them are the launcher's own work and go through DispatchAction. Four
 * are not: showing a file or a folder is something a browser cannot do, so
 * those go to FilesService and the desktop opens them. Checking Tailscale
 * Funnel is the launcher's too, but its answer is a page to visit.
 *
 * The map is by the id form declares, which is the same id the window and the
 * terminal dispatched. A button that is neither shown here nor handled in
 * Dispatch fails loudly on the Go side rather than doing nothing here.
 */
const shows: Record<string, FileTarget> = {
  'run.open_settings_file': FileTarget.SETTINGS_FILE,
  'run.open_player_file': FileTarget.PLAYER_FILE,
  'run.open_folder': FileTarget.INSTALL_ROOT,
  'run.generate': FileTarget.GENERATED_SEED,
};

@Injectable({ providedIn: 'root' })
export class SettingsActions {
  private readonly commands = inject(LauncherCommands);
  private readonly store = inject(SettingsStore);

  /**
   * press answers with what to tell the player, or empty when the launcher will
   * say it on the stream. A refusal is an answer too, not an error: the buttons
   * share one subscription, and one refused press used to end it for all of
   * them.
   */
  press(id: string): Observable<string> {
    return this.route(id).pipe(orRefusal((refusal) => of(refusal)));
  }

  private route(id: string): Observable<string> {
    const target = shows[id];
    if (target !== undefined) {
      return this.commands.showFile(target).pipe(map((answer) => `opened ${answer.path}`));
    }
    if (id === 'net.check_funnel') {
      return this.commands
        .approveFunnel()
        .pipe(map((answer) => answer.approvalUrl || answer.message));
    }
    if (id === 'server.debug_bundle') {
      return this.download();
    }
    return this.store.dispatch(id);
  }

  /**
   * The bundle is a download because the player has to attach it somewhere,
   * usually to a bug report. It arrives as a stream of chunks: the first
   * message names the file and every one after it carries bytes, so this
   * collects them and hands the browser one blob to save.
   */
  private download(): Observable<string> {
    return this.commands.debugBundle().pipe(
      reduce(
        (collected: Bundle, message) => ({
          name: message.filename || collected.name,
          chunks:
            message.chunk.length > 0 ? [...collected.chunks, message.chunk] : collected.chunks,
        }),
        { name: 'tf2ap-debug.zip', chunks: [] },
      ),
      map((bundle) => save(bundle)),
    );
  }
}

interface Bundle {
  readonly name: string;
  readonly chunks: readonly Uint8Array[];
}

function save(bundle: Bundle): string {
  // Each chunk is a view on its own buffer, so the copies are what Blob takes.
  const parts = bundle.chunks.map((chunk) => chunk.slice().buffer);
  const blob = new Blob(parts, { type: 'application/zip' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = bundle.name;
  link.click();
  URL.revokeObjectURL(url);
  return `saved ${bundle.name}`;
}
