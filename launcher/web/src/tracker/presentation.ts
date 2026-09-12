import { defaultHost } from './source-address';
import { Objective, TrackerSource } from './types';

export function objectiveLabel(objective: Objective): string {
  if (objective.kind === 'wave_cleared') return String(objective.wave);
  if (objective.kind === 'tank_destroyed') return 'Tank';
  if (objective.kind === 'giant_killed') return 'Giant';
  if (objective.kind === 'mission_cleared') return 'Done';
  return '';
}

export function hideBrokenImage(event: Event): void {
  (event.target as HTMLImageElement).hidden = true;
}

export function initialLocation(search: string): { demo: boolean; input?: string } {
  const params = new URLSearchParams(search);
  if (params.has('demo')) return { demo: true };
  const kind = params.has('room') ? 'room' : params.has('tracker') ? 'tracker' : undefined;
  if (kind === undefined) return { demo: false };
  const id = params.get(kind);
  if (id === null) return { demo: false };
  return { demo: false, input: `${params.get('host') ?? defaultHost}/${kind}/${id}` };
}

export function rememberSource(source: TrackerSource): void {
  const params = new URLSearchParams();
  params.set(source.kind, source.id);
  if (source.host !== defaultHost) params.set('host', source.host);
  history.replaceState(null, '', `${location.pathname}?${params}`);
}
