const SUBTYPE_LABELS: Record<string, string> = {
  '': 'Unspecified',
  POLICE_WITH_MOBILE_CAMERA: 'Mobile camera',
  POLICE_HIDING: 'Hidden',
  POLICE_VISIBLE: 'Visible',
};

export function subtypeLabel(raw: string): string {
  if (raw in SUBTYPE_LABELS) return SUBTYPE_LABELS[raw];
  const words = raw.replace(/^POLICE_/, '').replace(/_/g, ' ').toLowerCase();
  return words.charAt(0).toUpperCase() + words.slice(1);
}

/** Waze road type codes. */
const ROAD_TYPE_LABELS: Record<number, string> = {
  0: 'Unknown',
  1: 'Street',
  2: 'Primary street',
  3: 'Freeway',
  4: 'Ramp',
  5: 'Walking trail',
  6: 'Major highway',
  7: 'Minor highway',
  8: 'Dirt road',
  9: 'Walkway',
  10: 'Boardwalk',
  15: 'Ferry',
  16: 'Stairway',
  17: 'Private road',
  18: 'Railway',
  19: 'Runway',
  20: 'Parking lot',
  22: 'Narrow street',
};

export function roadTypeLabel(code: number): string {
  return ROAD_TYPE_LABELS[code] ?? `Road type ${code}`;
}

export function townLabel(name: string): string {
  return name || 'No town recorded';
}

export const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
export const WEEKDAYS_LONG = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];
