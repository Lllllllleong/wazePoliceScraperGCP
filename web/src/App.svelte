<script lang="ts">
  import { onMount } from 'svelte';
  import Explorer from './components/Explorer.svelte';
  import { decodeDataset, type Dataset } from './lib/data/dataset.ts';
  import { loadDatasetFile } from './lib/data/load.ts';
  import { DATASET_PATH } from './lib/data/schema.ts';

  type Status =
    | { kind: 'loading'; loaded: number; total: number }
    | { kind: 'decoding' }
    | { kind: 'error'; message: string }
    | { kind: 'ready'; ds: Dataset };

  let status = $state.raw<Status>({ kind: 'loading', loaded: 0, total: 0 });

  async function load() {
    status = { kind: 'loading', loaded: 0, total: 0 };
    try {
      const file = await loadDatasetFile(`${import.meta.env.BASE_URL}${DATASET_PATH}`, (loaded, total) => {
        status = { kind: 'loading', loaded, total };
      });
      status = { kind: 'decoding' };
      // Let "Preparing" paint before decoding blocks the main thread.
      await new Promise((resolve) => setTimeout(resolve, 0));
      status = { kind: 'ready', ds: decodeDataset(file) };
    } catch (e) {
      status = { kind: 'error', message: e instanceof Error ? e.message : String(e) };
    }
  }

  onMount(load);

  const mb = (bytes: number) => (bytes / 1048576).toFixed(1);
</script>

{#if status.kind === 'ready'}
  <Explorer ds={status.ds} />
{:else}
  <div class="splash">
    <h1>Police alerts on the Sydney–Canberra corridor</h1>
    {#if status.kind === 'error'}
      <p role="alert">{status.message}</p>
      <button type="button" class="button" onclick={load}>Try again</button>
    {:else}
      <div class="progress" role="status">
        <p>{status.kind === 'decoding' ? 'Preparing the data…' : 'Loading six months of alerts…'}</p>
        <div class="bar">
          <span
            style:width="{status.kind === 'decoding' ? 100 : status.total ? Math.min(100, (status.loaded / status.total) * 100) : 5}%"
          ></span>
        </div>
        {#if status.kind === 'loading' && status.total}
          <p class="size">{mb(Math.min(status.loaded, status.total))} of {mb(status.total)} MB</p>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .splash {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 20px;
    min-height: 100vh;
    padding: 24px;
    text-align: center;
  }

  h1 {
    font-size: 18px;
    font-weight: 600;
  }

  .progress {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    width: min(320px, 100%);
    color: var(--ink-2);
  }

  .bar {
    width: 100%;
    height: 4px;
    border-radius: 2px;
    background: var(--line);
    overflow: hidden;
  }

  .bar span {
    display: block;
    height: 100%;
    background: var(--accent);
    transition: width 150ms;
  }

  .size {
    font-size: 12px;
    color: var(--ink-3);
  }
</style>
