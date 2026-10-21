/**
 * Groups Waze street names into roads.
 *
 * Waze records some roads under several names: the M31 appears as both
 * "M31 - Hume Mwy" and "M31 - Hume Hwy", and some names carry sign text
 * ("B69 - Narellan Rd to Campbelltown / University / Hospital"). Names with
 * the same route number and base name are merged. The route number alone is
 * not enough, since B73 runs along a dozen differently named streets.
 */

export interface RoadGroup {
  /** Display name, from the most common variant. */
  name: string;
  /** Every Waze street name in the group, most common first. */
  variants: string[];
}

const ROUTE_NAME = /^([A-Z]{1,2}\d+[A-Z]?) - (.+)$/;
const SIGN_TEXT = / to .*$/;
const ROAD_KIND = / (Mwy|Hwy|Fwy|Pwy)$/;

/** The grouping key for a street name. */
export function roadKey(street: string): string {
  const m = ROUTE_NAME.exec(street);
  if (!m) return street;
  const name = m[2].replace(SIGN_TEXT, '').replace(ROAD_KIND, '');
  return `${m[1]} - ${name}`;
}

export function displayRoadName(street: string): string {
  if (!street) return 'No road name';
  return street.replace(/^([A-Z]{1,2}\d+[A-Z]?) - /, '$1 – ');
}

/**
 * Groups street names, given how many alerts each has. Groups are ordered by
 * alert count, so group 0 is the busiest road.
 */
export function groupRoads(streets: readonly string[], counts: ArrayLike<number>) {
  const byKey = new Map<string, { members: number[]; total: number }>();
  streets.forEach((street, i) => {
    const key = roadKey(street);
    let group = byKey.get(key);
    if (!group) byKey.set(key, (group = { members: [], total: 0 }));
    group.members.push(i);
    group.total += counts[i];
  });

  const ordered = [...byKey.values()].sort((a, b) => b.total - a.total);
  const streetToRoad = new Uint16Array(streets.length);
  const roads: RoadGroup[] = ordered.map((group, id) => {
    const members = group.members.sort((a, b) => counts[b] - counts[a]);
    for (const m of members) streetToRoad[m] = id;
    const variants = members.map((m) => streets[m]);
    return { name: displayRoadName(variants[0]), variants };
  });
  return { roads, streetToRoad };
}
