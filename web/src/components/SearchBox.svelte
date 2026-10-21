<script lang="ts" module>
  export interface SearchHit {
    kind: 'road' | 'town';
    id: number;
  }
</script>

<script lang="ts">
  import type { Dataset } from '../lib/data/dataset.ts';
  import { fmtInt } from '../lib/format.ts';

  interface Props {
    ds: Dataset;
    onselect: (hit: SearchHit) => void;
  }

  let { ds, onselect }: Props = $props();

  interface Entry extends SearchHit {
    label: string;
    /** Lower-cased names to match against, including every Waze spelling of a road. */
    text: string;
    total: number;
  }

  let query = $state('');
  let open = $state(false);
  let active = $state(0);

  const entries = $derived.by(() => {
    const list: Entry[] = [];
    ds.roads.forEach((r, id) => {
      if (r.variants[0]) list.push({ kind: 'road', id, label: r.name, text: r.variants.join(' | ').toLowerCase(), total: ds.totals.roads[id] });
    });
    ds.towns.forEach((t, id) => {
      if (t) list.push({ kind: 'town', id, label: t, text: t.toLowerCase(), total: ds.totals.towns[id] });
    });
    return list;
  });

  const results = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (q.length < 2) return [];
    const starts = (e: Entry) => (e.label.toLowerCase().startsWith(q) || e.text.includes(` - ${q}`) ? 0 : 1);
    return entries
      .filter((e) => e.text.includes(q))
      .sort((a, b) => starts(a) - starts(b) || b.total - a.total)
      .slice(0, 8);
  });

  const showList = $derived(open && results.length > 0);

  function choose(e: Entry) {
    onselect({ kind: e.kind, id: e.id });
    query = '';
    open = false;
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      open = true;
      active = Math.min(results.length - 1, active + 1);
    } else if (e.key === 'ArrowUp') {
      active = Math.max(0, active - 1);
    } else if (e.key === 'Enter' && showList) {
      choose(results[active]);
    } else if (e.key === 'Escape') {
      open = false;
      return;
    } else {
      return;
    }
    e.preventDefault();
  }
</script>

<div class="search">
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
    <circle cx="7" cy="7" r="4.5" /><path d="M10.5 10.5L14 14" />
  </svg>
  <input
    type="search"
    role="combobox"
    placeholder="Search a road or town"
    aria-label="Search a road or town"
    aria-autocomplete="list"
    aria-controls="search-results"
    aria-expanded={showList}
    aria-activedescendant={showList ? `search-result-${active}` : undefined}
    autocomplete="off"
    spellcheck="false"
    bind:value={query}
    oninput={() => ((open = true), (active = 0))}
    onfocus={() => (open = true)}
    onblur={() => (open = false)}
    {onkeydown}
  />
  {#if showList}
    <ul id="search-results" role="listbox" aria-label="Roads and towns">
      {#each results as r, i (`${r.kind}-${r.id}`)}
        <li
          id="search-result-{i}"
          role="option"
          tabindex="-1"
          aria-selected={i === active}
          onpointerdown={(e) => (e.preventDefault(), choose(r))}
          onpointerenter={() => (active = i)}
        >
          <span class="label">{r.label}</span>
          <span class="meta">{r.kind === 'road' ? 'Road' : 'Town'} · {fmtInt(r.total)}</span>
        </li>
      {/each}
    </ul>
  {:else if open && query.trim().length >= 2}
    <p class="none" role="status">No roads or towns match “{query.trim()}”</p>
  {/if}
</div>

<style>
  .search {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 40px;
    padding: 0 12px;
    border: 1px solid var(--line-strong);
    border-radius: 6px;
    background: var(--surface);
    color: var(--ink-3);
  }

  .search:focus-within {
    border-color: var(--accent);
  }

  input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: none;
    background: transparent;
    color: var(--ink);
  }

  ul,
  .none {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    z-index: 20;
    margin: 0;
    padding: 4px;
    border: 1px solid var(--line-strong);
    border-radius: 6px;
    background: var(--surface);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }

  ul {
    list-style: none;
  }

  li {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 10px;
    border-radius: 4px;
    cursor: pointer;
    color: var(--ink);
  }

  li[aria-selected='true'] {
    background: var(--accent-soft);
  }

  .label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    flex: none;
    font-size: 12px;
    color: var(--ink-3);
  }

  .none {
    padding: 10px 12px;
    font-size: 13px;
  }
</style>
