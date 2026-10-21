<script lang="ts">
  import { onMount } from 'svelte';
  import { extentOf, type Dataset } from '../lib/data/dataset.ts';
  import { describeCells, listLabel } from '../lib/describe.ts';
  import { crossfilter } from '../lib/filter/engine.ts';
  import { isFiltered, NO_FILTERS, pick, type Filters, type SetDimension } from '../lib/filter/filters.ts';
  import { decodeFilters, encodeFilters, type Limits } from '../lib/filter/url.ts';
  import { fmtDateRange, fmtInt } from '../lib/format.ts';
  import type { Bbox } from '../lib/geo.ts';
  import { roadTypeLabel, subtypeLabel, townLabel } from '../lib/labels.ts';
  import AboutDialog from './AboutDialog.svelte';
  import AlertDetail from './AlertDetail.svelte';
  import Facet, { type FacetItem } from './Facet.svelte';
  import FilterSummary, { type Chip } from './FilterSummary.svelte';
  import Heatmap from './Heatmap.svelte';
  import MapView from './MapView.svelte';
  import SearchBox, { type SearchHit } from './SearchBox.svelte';
  import Timeline from './Timeline.svelte';

  let { ds }: { ds: Dataset } = $props();

  const limits = $derived<Limits>({
    days: ds.days.length,
    subtypes: ds.subtypes.length,
    roadTypes: ds.roadTypeCount,
    roads: ds.roads.length,
    towns: ds.towns.length,
  });

  // Read the URL before anything renders, so the first render already shows a shared view.
  let filters = $state.raw<Filters>(filtersFromUrl());
  let view = $state.raw<Bbox | null>(null);
  let picked = $state<number | null>(null);
  let focus = $state.raw<{ bounds: Bbox; seq: number } | null>(null);
  let aboutOpen = $state(false);

  // Only recompute on map moves when the map view is actually a filter.
  const activeView = $derived(filters.inView ? view : null);
  const result = $derived(crossfilter(ds, filters, activeView));
  const filtered = $derived(isFiltered(filters));

  function filtersFromUrl(): Filters {
    return decodeFilters(location.hash, limits);
  }

  onMount(() => {
    const onHash = () => (filters = filtersFromUrl());
    addEventListener('hashchange', onHash);
    return () => removeEventListener('hashchange', onHash);
  });

  // Keep the URL in step so any view can be shared. replaceState, because
  // dragging the timeline would otherwise fill the back button with history.
  $effect(() => {
    const hash = encodeFilters(filters);
    if (hash !== location.hash.replace(/^#/, '')) {
      history.replaceState(history.state, '', hash ? `#${hash}` : location.pathname + location.search);
    }
  });

  function update(patch: Partial<Filters>) {
    filters = { ...filters, ...patch };
  }

  function pickIn(dim: SetDimension, values: number[], additive: boolean) {
    update({ [dim]: pick(filters[dim], values, additive) });
  }

  function flyTo(bounds: Bbox) {
    focus = { bounds, seq: (focus?.seq ?? 0) + 1 };
  }

  function onsearch(hit: SearchHit) {
    update(hit.kind === 'road' ? { roads: new Set([hit.id]) } : { towns: new Set([hit.id]) });
    const bounds = extentOf(ds, hit.kind, hit.id);
    if (bounds) flyTo(bounds);
  }

  function zoomToSpot(i: number) {
    const d = 0.006;
    flyTo([ds.lon[i] - d, ds.lat[i] - d, ds.lon[i] + d, ds.lat[i] + d]);
  }

  const chips = $derived.by(() => {
    const f = filters;
    const out: Chip[] = [];
    const names = (values: ReadonlySet<number>, label: (id: number) => string) => listLabel([...values].map(label));
    if (f.days) {
      out.push({ key: 'days', label: 'Dates', value: fmtDateRange(ds.days[f.days[0]].date, ds.days[f.days[1]].date), remove: () => update({ days: null }) });
    }
    if (f.cells.size) out.push({ key: 'cells', label: 'When', value: describeCells(f.cells), remove: () => update({ cells: new Set() }) });
    if (f.subtypes.size) {
      out.push({ key: 'subtypes', label: 'Type', value: names(f.subtypes, (i) => subtypeLabel(ds.subtypes[i])), remove: () => update({ subtypes: new Set() }) });
    }
    if (f.roadTypes.size) {
      out.push({ key: 'roadTypes', label: 'Road type', value: names(f.roadTypes, roadTypeLabel), remove: () => update({ roadTypes: new Set() }) });
    }
    if (f.roads.size) out.push({ key: 'roads', label: 'Road', value: names(f.roads, (i) => ds.roads[i].name), remove: () => update({ roads: new Set() }) });
    if (f.towns.size) {
      out.push({ key: 'towns', label: 'Town', value: names(f.towns, (i) => townLabel(ds.towns[i])), remove: () => update({ towns: new Set() }) });
    }
    if (f.inView) out.push({ key: 'view', label: 'Area', value: 'Map view', remove: () => update({ inView: false }) });
    return out;
  });

  const byTotal = (a: FacetItem, b: FacetItem) => b.total - a.total;

  const subtypeItems = $derived(
    ds.subtypes.map((s, id) => ({ id, label: subtypeLabel(s), count: result.subtypes[id], total: ds.totals.subtypes[id] })).sort(byTotal),
  );
  const roadTypeItems = $derived(
    Array.from({ length: ds.roadTypeCount }, (_, id) => ({ id, label: roadTypeLabel(id), count: result.roadTypes[id], total: ds.totals.roadTypes[id] }))
      .filter((it) => it.total > 0)
      .sort(byTotal),
  );
  const roadItems = $derived(
    ds.roads.map((r, id) => ({ id, label: r.name, count: result.roads[id], total: ds.totals.roads[id], placeholder: !r.variants[0] })),
  );
  const townItems = $derived(
    ds.towns.map((t, id) => ({ id, label: townLabel(t), count: result.towns[id], total: ds.totals.towns[id], placeholder: !t })),
  );
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === 'Escape' && picked !== null && !aboutOpen) picked = null;
  }}
/>

<header class="masthead">
  <div class="title">
    <h1>Police alerts on the Sydney–Canberra corridor</h1>
    <p>Waze reports · {fmtDateRange(ds.days[0].date, ds.days[ds.days.length - 1].date)} · {fmtInt(ds.count)} alerts</p>
  </div>
  <div class="tools">
    <SearchBox {ds} onselect={onsearch} />
    <button type="button" class="text-button" onclick={() => (aboutOpen = true)}>About the data</button>
  </div>
</header>

<div class="summary-bar">
  <FilterSummary selected={result.selected} total={ds.count} {chips} onreset={() => (filters = NO_FILTERS)} />
</div>

<main>
  <Timeline days={ds.days} counts={result.perDay} range={filters.days} onchange={(days) => update({ days })} />

  <div class="row">
    <section class="card map-card" aria-labelledby="map-title">
      <div class="map-head">
        <div>
          <h2 class="card-title" id="map-title">Where</h2>
          <p class="card-sub">Click a point for details</p>
        </div>
        <label class="toggle">
          <input type="checkbox" checked={filters.inView} onchange={(e) => update({ inView: e.currentTarget.checked })} />
          Count only alerts in view
        </label>
      </div>
      <div class="map-area" class:with-detail={picked !== null}>
        <div class="map-frame">
          <MapView {ds} mapState={result.mapState} {picked} {focus} onview={(b) => (view = b)} onpick={(i) => (picked = i)} />
          <div class="map-legend">
            {#if filtered}
              <span class="legend"><span class="swatch" style:background="var(--point-selected)"></span>Matching</span>
              <span class="legend"><span class="swatch" style:background="var(--point-context)"></span>Other alerts</span>
            {/if}
            <span class="legend"><span class="swatch shaded"></span>Not scraped</span>
          </div>
        </div>
        {#if picked !== null}
          <AlertDetail {ds} index={picked} mapState={result.mapState} {filtered} onclose={() => (picked = null)} onzoom={() => zoomToSpot(picked!)} />
        {/if}
      </div>
    </section>

    <div class="side">
      <Heatmap counts={result.cells} selected={filters.cells} onpick={(cells, add) => pickIn('cells', cells, add)} />
      <Facet title="Alert type" noun="types" items={subtypeItems} selected={filters.subtypes} onpick={(id, add) => pickIn('subtypes', [id], add)} />
    </div>
  </div>

  {#if filtered}
    <p class="tick-legend"><span class="tick-mark"></span>Ticks show each bar’s count if the selection had the same mix as all alerts</p>
  {/if}
  <div class="facets">
    <Facet title="Road type" noun="road types" items={roadTypeItems} selected={filters.roadTypes} onpick={(id, add) => pickIn('roadTypes', [id], add)} />
    <Facet title="Road" noun="roads" long items={roadItems} selected={filters.roads} onpick={(id, add) => pickIn('roads', [id], add)} />
    <Facet title="Town" noun="towns" long items={townItems} selected={filters.towns} onpick={(id, add) => pickIn('towns', [id], add)} />
  </div>
</main>

<footer>
  <p>
    Reports from Waze users, collected by a personal project. Not affiliated with Waze, Google or any police force.
    <a href="https://github.com/Lllllllleong/wazePoliceScraperGCP" target="_blank" rel="noreferrer">Source code</a>
  </p>
</footer>

<AboutDialog {ds} open={aboutOpen} onclose={() => (aboutOpen = false)} />

<style>
  .masthead {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 24px;
    padding: 12px 24px;
    background: var(--surface);
    border-bottom: 1px solid var(--line);
  }

  .title {
    margin-right: auto;
  }

  h1 {
    font-size: 16px;
    font-weight: 600;
  }

  .title p {
    margin-top: 2px;
    font-size: 12px;
    color: var(--ink-3);
  }

  .tools {
    display: flex;
    align-items: center;
    gap: 16px;
    flex: 0 1 460px;
  }

  .tools > :global(:first-child) {
    flex: 1;
  }

  .summary-bar {
    position: sticky;
    top: 0;
    z-index: 10;
    padding: 14px 24px;
    background: color-mix(in srgb, var(--bg) 92%, transparent);
    backdrop-filter: blur(6px);
    border-bottom: 1px solid var(--line);
  }

  main {
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 1600px;
    margin: 0 auto;
    padding: 16px 24px 32px;
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }

  .map-card {
    flex: 3 1 560px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .side {
    flex: 2 1 380px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-width: 0;
  }

  .map-head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px 16px;
  }

  .toggle {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 32px;
    font-size: 13px;
    color: var(--ink-2);
    cursor: pointer;
  }

  .toggle input {
    width: 16px;
    height: 16px;
    margin: 0;
    accent-color: var(--accent);
  }

  .map-area {
    position: relative;
  }

  .map-frame {
    position: relative;
    height: 540px;
    border-radius: 6px;
    overflow: hidden;
    background: var(--surface-2);
  }

  .map-area :global(.detail) {
    position: absolute;
    top: 12px;
    right: 12px;
    bottom: 12px;
    width: 320px;
    z-index: 2;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }

  .map-legend {
    position: absolute;
    left: 10px;
    top: 10px;
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    padding: 6px 10px;
    border-radius: 6px;
    background: color-mix(in srgb, var(--surface) 90%, transparent);
    pointer-events: none;
  }

  .swatch.shaded {
    background: var(--map-mask);
    box-shadow: inset 0 0 0 1px var(--line-strong);
  }

  .tick-legend {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    font-size: 12px;
    color: var(--ink-2);
  }

  .tick-mark {
    width: 2px;
    height: 14px;
    background: var(--tick);
  }

  .facets {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
    align-items: start;
  }

  footer {
    padding: 8px 24px 32px;
    text-align: center;
    font-size: 12px;
    color: var(--ink-3);
  }

  @media (max-width: 720px) {
    .masthead,
    .summary-bar {
      padding-left: 16px;
      padding-right: 16px;
    }

    main {
      padding: 12px 16px 24px;
    }

    .tools {
      flex-basis: 100%;
    }

    .summary-bar {
      position: static;
    }

    .map-frame {
      height: 380px;
    }

    .map-area :global(.detail) {
      position: static;
      width: auto;
      margin-top: 12px;
      box-shadow: none;
    }
  }
</style>
