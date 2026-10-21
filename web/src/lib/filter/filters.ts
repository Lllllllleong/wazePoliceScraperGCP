export interface Filters {
  /** Inclusive range of day indexes. */
  days: readonly [number, number] | null;
  /** Hour-of-week cells, weekday * 24 + hour. */
  cells: ReadonlySet<number>;
  subtypes: ReadonlySet<number>;
  roadTypes: ReadonlySet<number>;
  roads: ReadonlySet<number>;
  towns: ReadonlySet<number>;
  /** Count only alerts inside the current map view. */
  inView: boolean;
}

export type SetDimension = 'cells' | 'subtypes' | 'roadTypes' | 'roads' | 'towns';

export const NO_FILTERS: Filters = {
  days: null,
  cells: new Set(),
  subtypes: new Set(),
  roadTypes: new Set(),
  roads: new Set(),
  towns: new Set(),
  inView: false,
};

export function isFiltered(f: Filters): boolean {
  return (
    f.days !== null ||
    f.cells.size > 0 ||
    f.subtypes.size > 0 ||
    f.roadTypes.size > 0 ||
    f.roads.size > 0 ||
    f.towns.size > 0 ||
    f.inView
  );
}

/**
 * A click on a chart. A plain click selects just `values`, or clears the
 * selection if that was already all of it. An additive (shift) click adds
 * `values`, or removes them if they were all selected already.
 */
export function pick(current: ReadonlySet<number>, values: readonly number[], additive: boolean): ReadonlySet<number> {
  const allSelected = values.every((v) => current.has(v));
  if (additive) {
    const next = new Set(current);
    for (const v of values) {
      if (allSelected) next.delete(v);
      else next.add(v);
    }
    return next;
  }
  if (allSelected && current.size === values.length) return new Set();
  return new Set(values);
}
