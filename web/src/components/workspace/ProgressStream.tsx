"use client";

import type { WSMessage } from "@/lib/ws";
import { CheckCircle2, Circle, Loader2, AlertCircle, XCircle } from "lucide-react";

interface ProgressStreamProps {
  messages: WSMessage[];
  phase: string;
}

const phaseOrder = ["generating", "compiling", "reviewing", "fixing", "explaining", "done"];

const phaseLabels: Record<string, string> = {
  generating: "代码生成",
  compiling: "编译渲染",
  reviewing: "视觉审查",
  fixing: "修复问题",
  explaining: "代码说明",
  done: "完成",
};

function PhaseIcon({ phase, currentPhase }: { phase: string; currentPhase: string }) {
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

export function ProgressStream({ messages, phase }: ProgressStreamProps) {
  const latestIssues = messages
    .filter((m) => m.data.issues && m.data.issues.length > 0)
    .at(-1)?.data.issues;

  const latestPreviewUrl = messages
    .filter((m) => m.data.image_url)
    .at(-1)?.data.image_url;

  const errorMsg = messages.find((m) => m.type === "error")?.data.message;

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <h3 className="text-sm font-semibold">生成进度</h3>

      <div className="space-y-2">
        {phaseOrder
          .filter((p) => p !== "fixing") // fixing is shown inline with reviewing
          .map((p) => (
            <div key={p} className="flex items-center gap-2 text-sm">
              <PhaseIcon phase={p} currentPhase={phase} />
              <span
                className={
                  phase === p && p !== "done"
                    ? "font-medium text-foreground"
                    : "text-muted-foreground"
                }
              >
                {phaseLabels[p]}
              </span>
            </div>
          ))}
      </div>

      {latestIssues && latestIssues.length > 0 && (
        <div className="rounded border border-yellow-200 bg-yellow-50 p-2 text-xs dark:border-yellow-800 dark:bg-yellow-950">
          <div className="flex items-center gap-1 font-medium text-yellow-700 dark:text-yellow-300 mb-1">
            <AlertCircle className="h-3 w-3" />
            发现问题
          </div>
          <ul className="list-disc list-inside space-y-0.5 text-yellow-600 dark:text-yellow-400">
            {latestIssues.map((issue, i) => (
              <li key={i}>{issue}</li>
            ))}
          </ul>
        </div>
      )}

      {latestPreviewUrl && (
        <div className="mt-2">
          <p className="text-xs text-muted-foreground mb-1">当前预览：</p>
          <img
            src={latestPreviewUrl}
            alt="Preview"
            className="max-h-[150px] rounded border object-contain"
          />
        </div>
      )}

      {errorMsg && (
        <div className="rounded border border-red-200 bg-red-50 p-2 text-xs dark:border-red-800 dark:bg-red-950">
          <div className="flex items-center gap-1 font-medium text-red-700 dark:text-red-300">
            <XCircle className="h-3 w-3" />
            错误：{errorMsg}
          </div>
        </div>
      )}
    </div>
  );
}
