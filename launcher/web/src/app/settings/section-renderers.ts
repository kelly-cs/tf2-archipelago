import { InjectionToken, Type } from '@angular/core';

/**
 * A section of the settings a domain draws itself, instead of as the rows.
 *
 * The rows are the truth and every section could be drawn as them. Some are
 * better as something else: six seats are a grid, nine classes are a table, a
 * loadout is a builder. The domain that knows how registers here, keyed by the
 * page slug and the section slug, and the settings page never imports it: the
 * component arrives through this token, loaded when the section is opened.
 */
export interface SectionRenderer {
  /** `<page slug>/<section slug>`, both as slugOf makes them. */
  readonly key: string;
  readonly load: () => Promise<Type<object>>;
}

export const SECTION_RENDERERS = new InjectionToken<readonly SectionRenderer[]>(
  'settings section renderers',
  { providedIn: 'root', factory: () => [] },
);
