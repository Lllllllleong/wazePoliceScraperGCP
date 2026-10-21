<script lang="ts">
  import type { FeatureCollection } from 'geojson';
  import { MapLibreMap, NavigationControl, setWorkerUrl, type GeoJSONSource, type MapMouseEvent } from 'maplibre-gl';
  // MapLibre finds its worker relative to its own module URL, which doesn't
  // survive bundling, so hand it the URL of the copy Vite serves.
  import workerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?url';
  import { onMount } from 'svelte';
  import type { Dataset } from '../lib/data/dataset.ts';
  import { COVERAGE_BOUNDS, mercatorX, mercatorY, outsideCoverage, rectsToGeoJSON, type Bbox } from '../lib/geo.ts';
  import { AlertPointsLayer, type PointColors, type Rgba } from '../lib/map/pointsLayer.ts';

  interface Props {
    ds: Dataset;
    mapState: Uint8Array;
    picked: number | null;
    /** Fly here when `seq` changes. */
    focus: { bounds: Bbox; seq: number } | null;
    onview: (bounds: Bbox) => void;
    onpick: (index: number | null) => void;
  }

  let { ds, mapState, picked, focus, onview, onpick }: Props = $props();

  const STYLES = {
    light: 'https://tiles.openfreemap.org/styles/positron',
    dark: 'https://tiles.openfreemap.org/styles/dark',
  };
  /** Click and hover tolerance, in CSS pixels. */
  const HIT_RADIUS = 10;

  let container: HTMLDivElement;
  let map: MapLibreMap | undefined;
  let layer: AlertPointsLayer | undefined;
  let failed = $state(false);

  setWorkerUrl(workerUrl);

  onMount(() => {
    const dark = matchMedia('(prefers-color-scheme: dark)');
    layer = new AlertPointsLayer('alert-points', ds.lon, ds.lat, readColors());
    layer.setState(mapState);

    const [w, s, e, n] = COVERAGE_BOUNDS;
    let m: MapLibreMap;
    try {
      m = new MapLibreMap({
        container,
        style: dark.matches ? STYLES.dark : STYLES.light,
        bounds: [w, s, e, n],
        fitBoundsOptions: { padding: 24 },
        minZoom: 5,
        maxZoom: 16,
        maxBounds: [
          [w - 4, s - 3],
          [e + 4, n + 3],
        ],
        dragRotate: false,
        pitchWithRotate: false,
        touchPitch: false,
        attributionControl: { compact: true },
      });
    } catch {
      // No WebGL: the charts still work without the map.
      failed = true;
      return;
    }
    map = m;
    m.touchZoomRotate.disableRotation();
    m.addControl(new NavigationControl({ showCompass: false }), 'top-right');
    m.on('style.load', addLayers);
    m.on('load', reportView);
    m.on('moveend', reportView);
    m.on('click', (ev) => onpick(hit(ev)));
    // Hit-testing scans every alert, so do it at most once a frame.
    let pending: MapMouseEvent | null = null;
    m.on('mousemove', (ev) => {
      if (!pending) {
        requestAnimationFrame(() => {
          if (pending) m.getCanvas().style.cursor = hit(pending) === null ? '' : 'pointer';
          pending = null;
        });
      }
      pending = ev;
    });

    const onScheme = () => {
      layer?.setColors(readColors());
      m.setStyle(dark.matches ? STYLES.dark : STYLES.light);
    };
    dark.addEventListener('change', onScheme);
    return () => {
      dark.removeEventListener('change', onScheme);
      m.remove();
    };
  });

  $effect(() => {
    layer?.setState(mapState);
  });

  $effect(() => {
    const data = pickedFeature(picked);
    (map?.getSource('picked') as GeoJSONSource | undefined)?.setData(data);
  });

  $effect(() => {
    if (!focus || !map) return;
    const [w, s, e, n] = focus.bounds;
    map.fitBounds([w, s, e, n], { padding: 48, maxZoom: 13, duration: 800 });
  });

  /**
   * Our layers go above every road and area in the basemap but under the
   * labels that come after them. Some styles put label layers before the
   * roads, so "below the first label" would hide alerts under the roads.
   * Re-added whenever the style changes.
   */
  function addLayers() {
    if (!map || !layer) return;
    const layers = map.getStyle().layers;
    const lastShape = layers.findLastIndex((l) => l.type !== 'symbol');
    const firstLabel = layers.slice(lastShape + 1).find((l) => l.type === 'symbol')?.id;
    const [w, s, e, n] = COVERAGE_BOUNDS;
    const outer: Bbox = [w - 10, s - 10, e + 10, n + 10];
    const css = getComputedStyle(document.documentElement);

    map.addSource('outside', { type: 'geojson', data: rectsToGeoJSON(outsideCoverage(outer)) });
    map.addLayer(
      {
        id: 'outside',
        type: 'fill',
        source: 'outside',
        // No antialiasing, or the edges between rectangles show as hairlines.
        paint: { 'fill-color': css.getPropertyValue('--map-mask').trim(), 'fill-antialias': false },
      },
      firstLabel,
    );
    map.addLayer(layer, firstLabel);
    map.addSource('picked', { type: 'geojson', data: pickedFeature(picked) });
    map.addLayer({
      id: 'picked',
      type: 'circle',
      source: 'picked',
      paint: {
        'circle-radius': 9,
        'circle-color': 'rgba(0, 0, 0, 0)',
        'circle-stroke-width': 2.5,
        'circle-stroke-color': css.getPropertyValue('--tick').trim(),
      },
    });
  }

  function reportView() {
    if (!map) return;
    const b = map.getBounds();
    onview([b.getWest(), b.getSouth(), b.getEast(), b.getNorth()]);
  }

  function hit(ev: MapMouseEvent): number | null {
    if (!map || !layer) return null;
    const worldSize = 512 * 2 ** map.getZoom();
    const i = layer.nearest(mercatorX(ev.lngLat.lng), mercatorY(ev.lngLat.lat), HIT_RADIUS / worldSize);
    return i < 0 ? null : i;
  }

  function pickedFeature(i: number | null): FeatureCollection {
    return {
      type: 'FeatureCollection',
      features:
        i === null ? [] : [{ type: 'Feature', properties: {}, geometry: { type: 'Point', coordinates: [ds.lon[i], ds.lat[i]] } }],
    };
  }

  function readColors(): PointColors {
    const css = getComputedStyle(document.documentElement);
    return { selected: hexToRgba(css.getPropertyValue('--point-selected')), context: hexToRgba(css.getPropertyValue('--point-context')) };
  }

  function hexToRgba(hex: string): Rgba {
    const m = /#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})/i.exec(hex);
    if (!m) return [0.5, 0.5, 0.5, 1];
    return [parseInt(m[1], 16) / 255, parseInt(m[2], 16) / 255, parseInt(m[3], 16) / 255, 1];
  }
</script>

<div class="map" bind:this={container} role="region" aria-label="Map of alerts. Click a point for details."></div>
{#if failed}
  <p class="failed">The map needs WebGL, which isn't available in this browser. The charts still work.</p>
{/if}

<style>
  .map {
    position: absolute;
    inset: 0;
  }

  .failed {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    padding: 24px;
    text-align: center;
    color: var(--ink-3);
  }

  .map :global(.maplibregl-ctrl-group) {
    border-radius: 6px;
    box-shadow: none;
    border: 1px solid var(--line-strong);
  }

  .map :global(.maplibregl-ctrl-group button) {
    width: 36px;
    height: 36px;
  }
</style>
