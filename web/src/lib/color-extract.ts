import type { ColorPair } from "@/types/api";

/** Load image file into pixel data via offscreen canvas (max 256×256). */
function loadImageData(file: File): Promise<ImageData> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      const MAX = 256;
      const scale = Math.min(1, MAX / Math.max(img.width, img.height));
      const w = Math.round(img.width * scale);
      const h = Math.round(img.height * scale);

      const canvas = document.createElement("canvas");
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext("2d")!;
      ctx.drawImage(img, 0, 0, w, h);
      resolve(ctx.getImageData(0, 0, w, h));
      URL.revokeObjectURL(img.src);
    };
    img.onerror = () => {
      URL.revokeObjectURL(img.src);
      reject(new Error("Failed to load image"));
    };
    img.src = URL.createObjectURL(file);
  });
}

export type RGB = [number, number, number];

/** Convert a hex color string (#RRGGBB) to an RGB tuple. */
export function hexToRgb(hex: string): RGB {
  const h = hex.replace("#", "");
  return [
    parseInt(h.substring(0, 2), 16),
    parseInt(h.substring(2, 4), 16),
    parseInt(h.substring(4, 6), 16),
  ];
}

/** Sample up to `maxSamples` non-transparent, non-white pixels. */
function samplePixels(data: ImageData, maxSamples: number): RGB[] {
  const { data: d, width, height } = data;
  const total = width * height;
  const pixels: RGB[] = [];

  // Collect all valid pixel indices first
  const valid: number[] = [];
  for (let i = 0; i < total; i++) {
    const off = i * 4;
    const a = d[off + 3];
    if (a < 128) continue; // skip transparent
    const r = d[off], g = d[off + 1], b = d[off + 2];
    if (r > 240 && g > 240 && b > 240) continue; // skip near-white
    if (r < 15 && g < 15 && b < 15) continue; // skip near-black
    valid.push(i);
  }

  if (valid.length === 0) return pixels;

  // Reservoir sampling if too many
  if (valid.length <= maxSamples) {
    for (const i of valid) {
      const off = i * 4;
      pixels.push([d[off], d[off + 1], d[off + 2]]);
    }
  } else {
    const step = valid.length / maxSamples;
    for (let j = 0; j < maxSamples; j++) {
      const i = valid[Math.floor(j * step)];
      const off = i * 4;
      pixels.push([d[off], d[off + 1], d[off + 2]]);
    }
  }

  return pixels;
}

/** Euclidean distance squared between two RGB values. */
function distSq(a: RGB, b: RGB): number {
  const dr = a[0] - b[0], dg = a[1] - b[1], db = a[2] - b[2];
  return dr * dr + dg * dg + db * db;
}

interface Cluster {
  centroid: RGB;
  count: number;
}

/** K-Means++ initialization + Lloyd's iteration. */
function kMeans(pixels: RGB[], k: number, maxIter: number): Cluster[] {
  if (pixels.length === 0) return [];
  k = Math.min(k, pixels.length);

  // K-Means++ init
  const centroids: RGB[] = [pixels[Math.floor(Math.random() * pixels.length)]];
  const dists = new Float64Array(pixels.length).fill(Infinity);

  for (let c = 1; c < k; c++) {
    let totalDist = 0;
    for (let i = 0; i < pixels.length; i++) {
      const d = distSq(pixels[i], centroids[c - 1]);
      if (d < dists[i]) dists[i] = d;
      totalDist += dists[i];
    }
    let target = Math.random() * totalDist;
    for (let i = 0; i < pixels.length; i++) {
      target -= dists[i];
      if (target <= 0) {
        centroids.push(pixels[i]);
        break;
      }
    }
    if (centroids.length === c) centroids.push(pixels[Math.floor(Math.random() * pixels.length)]);
  }

  // Lloyd's iteration
  const assignments = new Int32Array(pixels.length);
  for (let iter = 0; iter < maxIter; iter++) {
    // Assign
    let changed = false;
    for (let i = 0; i < pixels.length; i++) {
      let bestC = 0, bestD = Infinity;
      for (let c = 0; c < k; c++) {
        const d = distSq(pixels[i], centroids[c]);
        if (d < bestD) { bestD = d; bestC = c; }
      }
      if (assignments[i] !== bestC) { assignments[i] = bestC; changed = true; }
    }
    if (!changed) break;

    // Update centroids
    const sums = Array.from({ length: k }, () => [0, 0, 0]);
    const counts = new Int32Array(k);
    for (let i = 0; i < pixels.length; i++) {
      const c = assignments[i];
      sums[c][0] += pixels[i][0];
      sums[c][1] += pixels[i][1];
      sums[c][2] += pixels[i][2];
      counts[c]++;
    }
    for (let c = 0; c < k; c++) {
      if (counts[c] > 0) {
        centroids[c] = [
          Math.round(sums[c][0] / counts[c]),
          Math.round(sums[c][1] / counts[c]),
          Math.round(sums[c][2] / counts[c]),
        ];
      }
    }
  }

  // Build clusters sorted by count descending
  const counts = new Int32Array(k);
  for (let i = 0; i < pixels.length; i++) counts[assignments[i]]++;

  const clusters: Cluster[] = centroids.map((centroid, i) => ({
    centroid,
    count: counts[i],
  }));
  clusters.sort((a, b) => b.count - a.count);
  return clusters;
}

function toHex(r: number, g: number, b: number): string {
  const clamp = (v: number) => Math.max(0, Math.min(255, Math.round(v)));
  return "#" + [r, g, b].map(clamp).map((v) => v.toString(16).padStart(2, "0")).join("");
}

/** Derive a fill (lightened) / line (darkened) pair from a centroid color. */
export function deriveColorPair(c: RGB): ColorPair {
  // Fill: mix 60% with white to lighten
  const fill = toHex(
    c[0] + (255 - c[0]) * 0.6,
    c[1] + (255 - c[1]) * 0.6,
    c[2] + (255 - c[2]) * 0.6,
  );
  // Line: darken by 0.7
  const line = toHex(c[0] * 0.7, c[1] * 0.7, c[2] * 0.7);
  return { fill, line };
}

/**
 * Extract dominant colors from an image file using client-side K-Means.
 * Returns ColorPair[] (fill/line pairs) sorted by dominance.
 */
export async function extractColorsFromImage(
  file: File,
  k = 6,
): Promise<ColorPair[]> {
  const imageData = await loadImageData(file);
  const pixels = samplePixels(imageData, 10000);
  if (pixels.length < k) {
    throw new Error("Image has too few usable colors");
  }
  const clusters = kMeans(pixels, k, 20);
  return clusters.map((cl) => deriveColorPair(cl.centroid));
}
