"use client";

import { useMemo } from "react";
import { createTwoFilesPatch } from "diff";

interface DiffViewerProps {
  oldCode: string;
  newCode: string;
  oldLabel?: string;
  newLabel?: string;
}

export function DiffViewer({
  oldCode,
  newCode,
  oldLabel = "修改前",
  newLabel = "修改后",
}: DiffViewerProps) {
  const lines = useMemo(() => {
    const patch = createTwoFilesPatch(oldLabel, newLabel, oldCode, newCode, "", "", {
      context: 3,
    });
    // Skip the first two header lines (--- / +++)
    const allLines = patch.split("\n");
    return allLines.slice(2);
  }, [oldCode, newCode, oldLabel, newLabel]);

  if (oldCode === newCode) {
    return (
      <div className="flex items-center justify-center rounded border border-dashed py-8 text-sm text-muted-foreground">
        无变更
      </div>
    );
  }

  return (
    <pre className="max-h-[400px] overflow-auto rounded border bg-muted p-3 text-xs font-mono leading-relaxed">
      {lines.map((line, i) => {
        let className = "";
        if (line.startsWith("+")) {
          className = "bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-300";
        } else if (line.startsWith("-")) {
          className = "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300";
        } else if (line.startsWith("@@")) {
          className = "text-blue-600 dark:text-blue-400";
        }
        return (
          <div key={i} className={className}>
            {line}
          </div>
        );
      })}
    </pre>
  );
}
