<script lang="ts">
  import type { Dataset } from '../lib/data/dataset.ts';
  import { fmtDate, fmtInt, fmtPercent } from '../lib/format.ts';

  interface Props {
    ds: Dataset;
    open: boolean;
    onclose: () => void;
  }

  let { ds, open, onclose }: Props = $props();

  let dialog: HTMLDialogElement;

  const first = $derived(ds.days[0].date);
  const last = $derived(ds.days[ds.days.length - 1].date);
  const incomplete = $derived(ds.days.filter((d) => d.incomplete));

  $effect(() => {
    if (open && !dialog.open) dialog.showModal();
    else if (!open && dialog.open) dialog.close();
  });
</script>

<dialog bind:this={dialog} {onclose} aria-labelledby="about-title">
  <div class="head">
    <h2 id="about-title">About the data</h2>
    <button type="button" class="close" aria-label="Close" onclick={() => dialog.close()}>
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 3l8 8M11 3l-8 8" /></svg>
    </button>
  </div>

  <div class="body">
    <p>
      These are the {fmtInt(ds.count)} police alerts that Waze users reported along the Sydney–Canberra corridor between
      {fmtDate(first)} and {fmtDate(last)}. A scraper polled Waze every minute and kept each alert until it disappeared. Collection
      stopped when Waze began refusing requests in March 2026, so the dataset will not grow.
    </p>

    <h3>What’s counted</h3>
    <ul>
      <li>Each alert is counted once, on the day and hour it was reported (Sydney time).</li>
      <li><strong>Active for</strong> is the time from the report until the alert was last seen.</li>
      <li><strong>Thumbs up</strong> are other drivers confirming the alert; <strong>reliability</strong> and <strong>confidence</strong> are Waze’s own 0–10 scores.</li>
      <li>Most alerts carry no detail. “Hidden”, “visible” and “mobile camera” are what the reporter chose when they did.</li>
      <li>Roads are grouped by route number and name, so “M31 - Hume Mwy” and “M31 - Hume Hwy” count as one road.</li>
    </ul>

    <h3>Coverage</h3>
    <p>
      Only the areas left unshaded on the map were scraped. Gaps in the record are days the scraper wasn’t running for part of
      the day; they’re hatched on the timeline.
    </p>
    {#if incomplete.length}
      <ul class="days">
        {#each incomplete as d (d.index)}
          <li><span>{fmtDate(d.date)}</span><span>{fmtPercent(d.coverage)} observed</span></li>
        {/each}
      </ul>
    {/if}

    <h3>Please note</h3>
    <p>
      These are reports from the public, not police records. They can be wrong, late or duplicated. This site is a personal
      project, unaffiliated with Waze, Google or any police force, and isn’t meant for planning trips or avoiding enforcement.
    </p>

    <p class="credits">
      Map data © OpenStreetMap contributors; map tiles by OpenFreeMap.
      <a href="https://github.com/Lllllllleong/wazePoliceScraperGCP" target="_blank" rel="noreferrer">Source code on GitHub</a>.
    </p>
  </div>
</dialog>

<style>
  dialog {
    width: min(640px, calc(100vw - 32px));
    max-height: calc(100vh - 64px);
    padding: 0;
    border: 1px solid var(--line);
    border-radius: 12px;
    background: var(--surface);
    color: var(--ink);
  }

  dialog::backdrop {
    background: rgba(10, 12, 16, 0.45);
  }

  .head {
    position: sticky;
    top: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 14px 24px;
    border-bottom: 1px solid var(--line);
    background: var(--surface);
  }

  h2 {
    font-size: 17px;
    font-weight: 600;
  }

  .close {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--ink-2);
  }

  .close:hover {
    background: var(--surface-2);
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px 24px 24px;
    line-height: 1.55;
  }

  h3 {
    margin-top: 8px;
    font-size: 14px;
    font-weight: 600;
  }

  ul {
    margin: 0;
    padding-left: 20px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .days {
    list-style: none;
    padding: 0;
    gap: 0;
    font-size: 13px;
  }

  .days li {
    display: flex;
    justify-content: space-between;
    padding: 6px 0;
    border-bottom: 1px solid var(--line);
  }

  .days li span:last-child {
    color: var(--ink-3);
  }

  .credits {
    margin-top: 8px;
    font-size: 12px;
    color: var(--ink-3);
  }
</style>
