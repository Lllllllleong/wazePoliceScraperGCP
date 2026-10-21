<script lang="ts" module>
  export interface FacetItem {
    id: number;
    label: string;
    /** Alerts under every filter except this facet's. */
    count: number;
    /** Alerts in the whole dataset. */
    total: number;
    /** A stand-in like "No town recorded": listed last and muted. */
    placeholder?: boolean;
  }
</script>

<script lang="ts">
  import { fmtInt } from '../lib/format.ts';

  interface Props {
    title: string;
    items: FacetItem[];
    selected: ReadonlySet<number>;
    onpick: (id: number, additive: boolean) => void;
    /** Rows shown before "Show all". */
    limit?: number;
    /** Long lists: sort by current count, hide empty rows, offer a search box. */
    long?: boolean;
    /** Plural noun for the "Show all" button. */
    noun: string;
  }

  let { title, items, selected, onpick, limit = 6, long = false, noun }: Props = $props();

  let expanded = $state(false);
  let query = $state('');

  const chartTotal = $derived(items.reduce((a, it) => a + it.count, 0));
  const datasetTotal = $derived(items.reduce((a, it) => a + it.total, 0));
  /** Scales an item's dataset total to what this chart would show if its mix were typical. */
  const ratio = $derived(datasetTotal ? chartTotal / datasetTotal : 0);

  /** Ticks only mean something once another chart is filtering this one. */
  const comparing = $derived(items.some((it) => it.count !== it.total));

  const ordered = $derived(
    long
      ? items
          .filter((it) => it.count > 0 || selected.has(it.id))
          .sort((a, b) => Number(!!a.placeholder) - Number(!!b.placeholder) || b.count - a.count || b.total - a.total)
      : items,
  );

  const visible = $derived.by(() => {
    if (expanded) {
      const q = query.trim().toLowerCase();
      return (q ? ordered.filter((it) => it.label.toLowerCase().includes(q)) : ordered).slice(0, 300);
    }
    const top = ordered.slice(0, limit);
    // Keep selected rows in view even when they aren't in the top few.
    const extra = ordered.filter((it) => selected.has(it.id) && !top.includes(it));
    return [...top, ...extra];
  });

  const scale = $derived(Math.max(1, ...visible.map((it) => Math.max(it.count, comparing ? it.total * ratio : 0))));
  const titleId = $derived(`facet-${title.toLowerCase().replace(/\W+/g, '-')}`);

  const additive = (e: MouseEvent) => e.shiftKey || e.metaKey || e.ctrlKey;
</script>

<section class="card" aria-labelledby={titleId}>
  <h2 class="card-title" id={titleId}>{title}</h2>

  {#if expanded && long}
    <input class="search" type="search" placeholder="Filter {noun}" aria-label="Filter {noun}" bind:value={query} />
  {/if}

  <ul class:scroll={expanded}>
    {#each visible as it (it.id)}
      {@const on = selected.has(it.id)}
      {@const expected = Math.round(it.total * ratio)}
      <li>
        <button
          type="button"
          class:on
          class:dim={selected.size > 0 && !on}
          aria-pressed={on}
          aria-label={comparing ? `${it.label}: ${fmtInt(it.count)} alerts, ${fmtInt(expected)} if typical` : `${it.label}: ${fmtInt(it.count)} alerts`}
          onclick={(e) => onpick(it.id, additive(e))}
        >
          <span class="line">
            <span class="name" class:placeholder={it.placeholder}>{it.label}</span>
            <span class="count">{fmtInt(it.count)}</span>
          </span>
          <span class="track">
            <span class="fill" style:width="{(it.count / scale) * 100}%"></span>
            {#if comparing && ratio > 0}
              <span class="tick" style:left="{Math.min(100, ((it.total * ratio) / scale) * 100)}%" title="{fmtInt(expected)} if typical"></span>
            {/if}
          </span>
        </button>
      </li>
    {:else}
      <li class="empty">Nothing matches</li>
    {/each}
  </ul>

  {#if ordered.length > limit}
    <button type="button" class="text-button" onclick={() => ((expanded = !expanded), (query = ''))}>
      {expanded ? 'Show fewer' : `Show all ${fmtInt(ordered.length)} ${noun}`}
    </button>
  {/if}
</section>

<style>
  section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .search {
    height: 34px;
    padding: 0 10px;
    border: 1px solid var(--line-strong);
    border-radius: 6px;
    background: var(--surface);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  ul.scroll {
    max-height: 360px;
    overflow-y: auto;
    margin-right: -8px;
    padding-right: 8px;
  }

  li button {
    display: flex;
    flex-direction: column;
    gap: 5px;
    min-height: 40px;
    padding: 6px 8px;
    margin: 0 -8px;
    width: calc(100% + 16px);
    border: 0;
    border-radius: 6px;
    background: none;
    text-align: left;
  }

  li button:hover {
    background: var(--surface-2);
  }

  .line {
    display: flex;
    gap: 8px;
    font-size: 13px;
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .count {
    color: var(--ink-2);
  }

  .name.placeholder {
    color: var(--ink-3);
    font-style: italic;
  }

  .on .name,
  .on .count {
    font-weight: 600;
    color: var(--ink);
  }

  .track {
    position: relative;
    display: block;
    height: 6px;
    border-radius: 3px;
    background: var(--surface-2);
  }

  .fill {
    position: absolute;
    inset: 0 auto 0 0;
    border-radius: 3px;
    background: var(--accent);
  }

  .dim .fill {
    background: var(--ghost);
  }

  .tick {
    position: absolute;
    top: -3px;
    bottom: -3px;
    width: 2px;
    margin-left: -1px;
    background: var(--tick);
  }

  .empty {
    padding: 8px 0;
    font-size: 13px;
    color: var(--ink-3);
  }

  .text-button {
    align-self: flex-start;
  }
</style>
