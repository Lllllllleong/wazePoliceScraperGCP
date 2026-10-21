# Police alerts explorer

The public site for this project: an explorer for the 241,717 police alerts collected between 26 Sep 2025 and 16 Mar 2026. It replaces `dataAnalysis/`.

There's no backend. The whole dataset is one static 3 MB file; the page loads it once and computes every filter in the browser.

## Develop

```bash
npm install
npm run dev     # http://localhost:5173
npm run check   # svelte-check and TypeScript
npm test        # Vitest
npm run build   # → dist/
```

## How it works

- `scripts/build-data.ts` turns the daily archives into `public/data/alerts.v1.json.gz` (format in `src/lib/data/schema.ts`): one array per field, deduplicated, with how much of each day the scraper was running. The raw Waze payloads, which include reporter usernames, are left out.
- `src/lib/data/dataset.ts` decodes the file into typed arrays and works out local hour, weekday and road groups (`roads.ts` merges names like "M31 - Hume Mwy" and "M31 - Hume Hwy").
- `src/lib/filter/engine.ts` applies the filters in one pass over every alert. Each chart counts with every filter except its own, so it keeps showing the alternatives to its selection.
- `src/lib/map/pointsLayer.ts` draws all alerts as a MapLibre custom WebGL layer. Filtering rewrites one byte per alert, so the map keeps up with dragging.
- Filters live in the URL fragment (`src/lib/filter/url.ts`), so any view can be shared.

Basemap tiles come from [OpenFreeMap](https://openfreemap.org/), which needs no key.

## Rebuilding the data

The data is final and committed, so this is only needed if the format changes (bump `DATASET_VERSION` when it does). It needs Node 24+ and read access to the archive bucket:

```bash
gcloud storage cp "gs://wazepolicescrapergcp-archive/*.jsonl" /tmp/archives
npm run build:data -- /tmp/archives
```

## Deploy

Firebase Hosting, in the same project as the old site, so going live updates wazepolicescrapergcp.web.app and both custom domains at once. Run these from this folder (from `dataAnalysis/` they'd deploy the old site). Every deploy builds first.

Check a preview channel, then promote exactly that build:

```bash
firebase hosting:channel:deploy preview --expires 7d
firebase hosting:clone wazepolicescrapergcp:preview wazepolicescrapergcp:live
```

To undo a release, roll back from Hosting › Release history in the Firebase console.
