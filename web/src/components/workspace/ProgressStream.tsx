"use client";

import { useState, useEffect, useMemo, useRef } from "react";
import type { WSMessage } from "@/lib/ws";
import { DiffViewer } from "./DiffViewer";
import { CheckCircle2, Circle, Loader2, AlertCircle, XCircle, Code2, Image as ImageIcon, GitCompareArrows, ZoomIn, MessageSquareText } from "lucide-react";
import { ImageLightbox } from "./ImageLightbox";

interface ProgressStreamProps {
  messages: WSMessage[];
  phase: string;
}

const phaseOrder = ["generating", "compiling", "reviewing", "rerolling", "fixing", "explaining", "done"];

const phaseLabels: Record<string, string> = {
  generating: "代码生成",
  compiling: "编译渲染",
  reviewing: "视觉审查",
  rerolling: "重新生成",
  fixing: "润色修复",
  explaining: "代码说明",
  done: "完成",
};

// ── Phase timing helpers ──

interface PhaseTiming {
  startTs: number;
  endTs?: number;
}

/** Build a map of phase → { startTs, endTs } from timestamped messages. */
function buildPhaseTimings(messages: WSMessage[]): Map<string, PhaseTiming> {
  const map = new Map<string, PhaseTiming>();
  for (const m of messages) {
    const ts = m._ts;
    if (!ts) continue;
    const existing = map.get(m.phase);
    if (!existing) {
      map.set(m.phase, { startTs: ts, endTs: undefined });
    }
  }
  // Set endTs for each phase = startTs of the next phase that appeared
  const ordered = phaseOrder.filter((p) => map.has(p));
  for (let i = 0; i < ordered.length - 1; i++) {
    const cur = map.get(ordered[i])!;
    const next = map.get(ordered[i + 1])!;
    cur.endTs = next.startTs;
  }
  return map;
}

function formatElapsed(ms: number): string {
  const secs = Math.floor(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = secs % 60;
  return `${mins}m${remSecs.toString().padStart(2, "0")}s`;
}

/** Displays elapsed time for a phase. Ticks every second if active. */
function PhaseTimer({ timing, isActive }: { timing: PhaseTiming | undefined; isActive: boolean }) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!isActive) return;
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, [isActive]);

  if (!timing) return null;

  const elapsed = (isActive ? now : (timing.endTs ?? now)) - timing.startTs;
  if (elapsed < 0) return null;

  return (
    <span className="text-xs text-muted-foreground tabular-nums">
      {formatElapsed(elapsed)}
    </span>
  );
}

// ── Phase icon ──

function PhaseIcon({ phase, currentPhase, appeared }: { phase: string; currentPhase: string; appeared: boolean }) {
  if (!appeared) {
    return <Circle className="h-4 w-4 text-muted-foreground" />;
  }
  const currentIdx = phaseOrder.indexOf(currentPhase);
  const phaseIdx = phaseOrder.indexOf(phase);

  if (currentPhase === phase && phase !== "done") {
    return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
  }
  if (phaseIdx < currentIdx || currentPhase === "done") {
    return <CheckCircle2 className="h-4 w-4 text-green-500" />;
  }
  return <Circle className="h-4 w-4 text-muted-foreground" />;
}

// ── Snapshot collectors ──

/** Extract all distinct image snapshots from progress messages, in order. */
export function collectImageSnapshots(messages: WSMessage[]) {
  const snapshots: { round: number; imageUrl: string; label: string }[] = [];
  const seen = new Set<string>();

  for (const m of messages) {
    const url = m.data.image_url;
    if (!url || seen.has(url)) continue;
    seen.add(url);

    const round = m.data.round || 0;
    let label: string;
    if (m.phase === "compiling") {
      label = "初次编译";
    } else if (m.phase === "rerolling") {
      label = `第 ${round} 次重画`;
    } else if (m.phase === "fixing") {
      label = `第 ${round} 轮润色`;
    } else if (m.phase === "reviewing") {
      label = "审查结果";
    } else {
      label = "渲染预览";
    }
    snapshots.push({ round, imageUrl: url, label });
  }

  return snapshots;
}

/** Extract ordered code snapshots for diff display. */
function collectCodeSnapshots(messages: WSMessage[]) {
  const snapshots: { code: string; label: string }[] = [];

  for (const m of messages) {
    const code = m.data.code;
    if (!code) continue;
    if (snapshots.length > 0 && snapshots[snapshots.length - 1].code === code) continue;

    if (m.phase === "generating") {
      snapshots.push({ code, label: "初始生成" });
    } else if (m.phase === "rerolling") {
      const round = m.data.round || snapshots.length;
      snapshots.push({ code, label: `第 ${round} 次重画` });
    } else if (m.phase === "fixing") {
      const round = m.data.round || snapshots.length;
      snapshots.push({ code, label: `第 ${round} 轮润色` });
    } else if (m.phase === "reviewing") {
      snapshots.push({ code, label: "审查后" });
    } else {
      snapshots.push({ code, label: m.phase });
    }
  }

  return snapshots;
}

// ── Small components ──

function ImageWithLoading({ src, alt, className }: { src: string; alt: string; className?: string }) {
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState(false);

  return (
    <div className="relative min-h-[200px]">
      {!loaded && !error && (
        <div className="absolute inset-0 flex items-center justify-center rounded border border-dashed bg-muted/50">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            图片加载中...
          </div>
        </div>
      )}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center rounded border border-dashed bg-muted/50">
          <span className="text-sm text-muted-foreground">图片加载失败</span>
        </div>
      )}
      <img
        src={src}
        alt={alt}
        className={className}
        style={{ display: loaded ? "block" : "none" }}
        onLoad={() => setLoaded(true)}
        onError={() => setError(true)}
      />
    </div>
  );
}

function ScoreBadge({ score }: { score: number }) {
  let color: string;
  if (score >= 9) {
    color = "bg-green-100 text-green-700 border-green-300 dark:bg-green-950 dark:text-green-300 dark:border-green-700";
  } else if (score >= 4) {
    color = "bg-yellow-100 text-yellow-700 border-yellow-300 dark:bg-yellow-950 dark:text-yellow-300 dark:border-yellow-700";
  } else {
    color = "bg-red-100 text-red-700 border-red-300 dark:bg-red-950 dark:text-red-300 dark:border-red-700";
  }
  return (
    <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-semibold ${color}`}>
      {score}/10
    </span>
  );
}

// ── Per-phase helpers ──

function getPhaseStatus(messages: WSMessage[], targetPhase: string): string | undefined {
  return messages
    .filter((m) => m.phase === targetPhase && m.data.message)
    .at(-1)?.data.message;
}

function getPhaseScore(messages: WSMessage[], targetPhase: string): number | undefined {
  return messages
    .filter((m) => m.phase === targetPhase && m.data.score && m.data.score > 0)
    .at(-1)?.data.score;
}

// ── Main component ──

export function ProgressStream({ messages, phase }: ProgressStreamProps) {
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  const latestIssues = messages
    .filter((m) => m.data.issues && m.data.issues.length > 0)
    .at(-1)?.data.issues;

  const latestCritique = messages
    .filter((m) => m.data.critique)
    .at(-1)?.data.critique;

  const latestScore = messages
    .filter((m) => m.data.score && m.data.score > 0)
    .at(-1)?.data.score;

  const appearedPhases = new Set(messages.map((m) => m.phase));
  appearedPhases.add(phase);

  const didReroll = messages.some((m) => m.phase === "rerolling");
  const didFix = messages.some((m) => m.phase === "fixing");

  const errorMsg = messages.find((m) => m.type === "error")?.data.message;

  const imageSnapshots = collectImageSnapshots(messages);
  const codeSnapshots = collectCodeSnapshots(messages);
  const latestImage = imageSnapshots.length > 0 ? imageSnapshots[imageSnapshots.length - 1] : null;

  const phaseTimings = useMemo(() => buildPhaseTimings(messages), [messages]);

  const visiblePhases = phaseOrder.filter((p) => {
    if (p === "generating" || p === "compiling" || p === "reviewing" || p === "explaining" || p === "done") return true;
    if (p === "rerolling") return didReroll;
    if (p === "fixing") return didFix;
    return false;
  });

  return (
    <div className="rounded-lg border bg-card p-4 space-y-4">
      <h3 className="text-sm font-semibold">生成进度</h3>

      <div className="grid gap-4 lg:grid-cols-[1fr_1fr]">
        {/* Left: phase timeline + issues */}
        <div className="space-y-3">
          <div className="space-y-2">
            {visiblePhases.map((p) => {
              const phaseScore = getPhaseScore(messages, p);
              const statusText = getPhaseStatus(messages, p);
              const isActive = phase === p && p !== "done";
              const timing = phaseTimings.get(p);

              return (
                <div key={p} className="space-y-0.5">
                  <div className="flex items-center gap-2 text-sm">
                    <PhaseIcon phase={p} currentPhase={phase} appeared={appearedPhases.has(p)} />
                    <span className={isActive ? "font-medium text-foreground" : "text-muted-foreground"}>
                      {phaseLabels[p]}
                    </span>
                    {(p === "reviewing" || p === "rerolling" || p === "fixing") && phaseScore != null && phaseScore > 0 && (
                      <ScoreBadge score={phaseScore} />
                    )}
                    <PhaseTimer timing={timing} isActive={isActive} />
                  </div>
                  {statusText && isActive && (
                    <p className="ml-6 text-xs text-muted-foreground">{statusText}</p>
                  )}
                </div>
              );
            })}
          </div>

          {errorMsg && (
            <div className="rounded border border-red-200 bg-red-50 p-2 text-xs dark:border-red-800 dark:bg-red-950">
              <div className="flex items-center gap-1 font-medium text-red-700 dark:text-red-300">
                <XCircle className="h-3 w-3" />
                错误：{errorMsg}
              </div>
            </div>
          )}

          {latestIssues && latestIssues.length > 0 && (
            <div className="rounded border border-yellow-200 bg-yellow-50 p-2 text-xs dark:border-yellow-800 dark:bg-yellow-950">
              <div className="flex items-center gap-1 font-medium text-yellow-700 dark:text-yellow-300 mb-1">
                <AlertCircle className="h-3 w-3" />
                审查意见
              </div>
              <ul className="list-disc list-inside space-y-0.5 text-yellow-600 dark:text-yellow-400">
                {latestIssues.map((issue, i) => (
                  <li key={i}>{issue}</li>
                ))}
              </ul>
            </div>
          )}

          {latestCritique && (
            <div className="rounded border border-blue-200 bg-blue-50 p-2 text-xs dark:border-blue-800 dark:bg-blue-950">
              <div className="flex items-center gap-1.5 font-medium text-blue-700 dark:text-blue-300 mb-1">
                <MessageSquareText className="h-3 w-3" />
                AI 点评
                {latestScore != null && latestScore > 0 && <ScoreBadge score={latestScore} />}
              </div>
              <p className="text-blue-600 dark:text-blue-400">{latestCritique}</p>
            </div>
          )}
        </div>

        {/* Right: current image preview */}
        {latestImage ? (
          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground flex items-center gap-1">
              <ImageIcon className="h-3.5 w-3.5" />
              当前渲染 — {latestImage.label}
            </p>
            <div className="relative group">
              <ImageWithLoading
                src={latestImage.imageUrl}
                alt={latestImage.label}
                className="max-h-[350px] rounded border object-contain w-full"
              />
              <button
                onClick={() => setLightboxSrc(latestImage.imageUrl)}
                className="absolute top-2 right-2 rounded-md bg-black/50 p-1.5 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/70"
                title="放大查看"
              >
                <ZoomIn className="h-4 w-4" />
              </button>
            </div>
          </div>
        ) : phase !== "done" && phase !== "generating" && (
          <div className="flex items-center justify-center min-h-[200px] rounded border border-dashed bg-muted/30">
            <div className="flex flex-col items-center gap-2 text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin" />
              <span className="text-sm">图片渲染中，请稍候...</span>
            </div>
          </div>
        )}
      </div>

      {/* Collapsible sections */}
      <div className="space-y-2">
        {imageSnapshots.length > 1 && (
          <details className="rounded border bg-muted/30 text-xs">
            <summary className="cursor-pointer px-2 py-1.5 font-medium flex items-center gap-1.5 text-muted-foreground hover:text-foreground">
              <ImageIcon className="h-3.5 w-3.5" />
              所有版本渲染 ({imageSnapshots.length} 版)
            </summary>
            <div className="border-t p-2 grid gap-3 sm:grid-cols-2">
              {imageSnapshots.map((snap, i) => (
                <div key={snap.imageUrl} className="space-y-1">
                  <p className="text-xs text-muted-foreground">
                    {i + 1}. {snap.label}
                  </p>
                  <div className="relative group">
                    <img
                      src={snap.imageUrl}
                      alt={snap.label}
                      className="max-h-[200px] rounded border object-contain w-full"
                    />
                    <button
                      onClick={() => setLightboxSrc(snap.imageUrl)}
                      className="absolute top-1 right-1 rounded bg-black/50 p-1 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/70"
                      title="放大查看"
                    >
                      <ZoomIn className="h-3 w-3" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </details>
        )}

        {codeSnapshots.length > 0 && (
          <details className="rounded border bg-muted/30 text-xs">
            <summary className="cursor-pointer px-2 py-1.5 font-medium flex items-center gap-1.5 text-muted-foreground hover:text-foreground">
              <Code2 className="h-3.5 w-3.5" />
              生成的代码 ({codeSnapshots.length} 版)
            </summary>
            <div className="border-t">
              <pre className="px-2 py-1.5 overflow-x-auto max-h-[200px] overflow-y-auto">
                <code>{codeSnapshots[codeSnapshots.length - 1].code}</code>
              </pre>
            </div>
          </details>
        )}

        {codeSnapshots.length > 1 && (
          <details className="rounded border bg-muted/30 text-xs">
            <summary className="cursor-pointer px-2 py-1.5 font-medium flex items-center gap-1.5 text-muted-foreground hover:text-foreground">
              <GitCompareArrows className="h-3.5 w-3.5" />
              代码修改对比 ({codeSnapshots.length - 1} 次修改)
            </summary>
            <div className="border-t p-2 space-y-3">
              {codeSnapshots.slice(1).map((snap, i) => (
                <div key={i} className="space-y-1">
                  <p className="text-xs font-medium text-muted-foreground">
                    {codeSnapshots[i].label} → {snap.label}
                  </p>
                  <DiffViewer
                    oldCode={codeSnapshots[i].code}
                    newCode={snap.code}
                    oldLabel={codeSnapshots[i].label}
                    newLabel={snap.label}
                  />
                </div>
              ))}
            </div>
          </details>
        )}
      </div>

      {lightboxSrc && (
        <ImageLightbox
          src={lightboxSrc}
          alt="放大查看"
          onClose={() => setLightboxSrc(null)}
        />
      )}
    </div>
  );
}
