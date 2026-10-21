<script lang="ts" module>
  export interface Chip {
    key: string;
    label: string;
    value: string;
    remove: () => void;
  }
</script>

<script lang="ts">
  import { fmtInt, fmtPercent } from '../lib/format.ts';

  interface Props {
    selected: number;
    total: number;
    chips: Chip[];
    onreset: () => void;
  }

  let { selected, total, chips, onreset }: Props = $props();
</script>

<div class="summary">
  <p class="count" aria-live="polite">
    <strong>{fmtInt(selected)}</strong>
    {#if chips.length}
      <span>of {fmtInt(total)} alerts ({fmtPercent(selected / total)})</span>
    {:else}
      <span>alerts</span>
    {/if}
  </p>

  {#if chips.length}
    <ul class="chips" aria-label="Active filters">
      {#each chips as chip (chip.key)}
        <li>
          <span class="k">{chip.label}</span>
          <span class="v">{chip.value}</span>
          <button type="button" aria-label="Remove {chip.label.toLowerCase()} filter" onclick={chip.remove}>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M2 2l6 6M8 2l-6 6" /></svg>
          </button>
        </li>
      {/each}
    </ul>
    <button type="button" class="button reset" onclick={onreset}>Reset all</button>
  {:else}
    <p class="hint">
      Every chart is a filter: drag across the timeline, or click a bar or a heatmap cell. Shift-click to choose several. Click a
      point on the map for details.
    </p>
  {/if}
</div>

<style>
  .summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 16px;
  }

  .count strong {
    font-size: 24px;
    font-weight: 600;
    line-height: 1.2;
  }

  .count span {
    color: var(--ink-3);
  }

  .chips {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .chips li {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 32px;
    padding: 0 2px 0 12px;
    border-radius: 16px;
    background: var(--accent-soft);
    color: var(--accent-ink);
    font-size: 13px;
  }

  .k {
    opacity: 0.8;
  }

  .v {
    font-weight: 500;
  }

  .chips button {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border: 0;
    border-radius: 14px;
    background: none;
    color: inherit;
  }

  .chips button:hover {
    background: color-mix(in srgb, var(--accent) 18%, transparent);
  }

  .reset {
    margin-left: auto;
  }

  .hint {
    flex: 1 1 320px;
    font-size: 13px;
    color: var(--ink-3);
  }
</style>
