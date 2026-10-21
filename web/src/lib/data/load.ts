import { DATASET_VERSION, type DatasetFile } from './schema.ts';

/**
 * Fetches and parses the dataset file, reporting download progress.
 *
 * The file is stored gzipped. A server may also send it with
 * `Content-Encoding: gzip`, in which case the browser has already unpacked it,
 * so check the bytes rather than trusting the name.
 */
export async function loadDatasetFile(
  url: string,
  onProgress?: (loadedBytes: number, totalBytes: number) => void,
): Promise<DatasetFile> {
  const res = await fetch(url);
  if (!res.ok || !res.body) throw new Error(`The data file could not be loaded (HTTP ${res.status}).`);

  const total = Number(res.headers.get('content-length')) || 0;
  const reader = res.body.getReader();
  const chunks: Uint8Array[] = [];
  let loaded = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    chunks.push(value);
    loaded += value.length;
    onProgress?.(loaded, total);
  }

  const bytes = new Uint8Array(new ArrayBuffer(loaded));
  let offset = 0;
  for (const chunk of chunks) {
    bytes.set(chunk, offset);
    offset += chunk.length;
  }

  const text = isGzip(bytes) ? await gunzip(bytes) : new TextDecoder().decode(bytes);
  const file = JSON.parse(text) as DatasetFile;
  if (file.version !== DATASET_VERSION) {
    throw new Error(`Unexpected data format version ${file.version}; this page expects ${DATASET_VERSION}.`);
  }
  return file;
}

function isGzip(bytes: Uint8Array): boolean {
  return bytes.length > 2 && bytes[0] === 0x1f && bytes[1] === 0x8b;
}

async function gunzip(bytes: Uint8Array<ArrayBuffer>): Promise<string> {
  const stream = new Blob([bytes]).stream().pipeThrough(new DecompressionStream('gzip'));
  return new Response(stream).text();
}
