"use client";

import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { MermaidRenderer } from "./MermaidRenderer";
import { ImageIcon } from "lucide-react";

export function ImagePreview() {
  const imageUrl = useGenerateStore((s) => s.imageUrl);
  const code = useGenerateStore((s) => s.code);
  const isRendering = useGenerateStore((s) => s.isRendering);
  const renderError = useGenerateStore((s) => s.renderError);
  const format = useSettingsStore((s) => s.format);
  const colorScheme = useSettingsStore((s) => s.colorScheme);

  if (isRendering) {
    return (
      <div className="flex h-full min-h-[400px] items-center justify-center rounded-md border border-dashed">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-current border-t-transparent" />
          <span className="text-sm">Rendering...</span>
        </div>
      </div>
    );
  }

  if (renderError) {
    return (
      <div className="flex h-full min-h-[400px] items-center justify-center rounded-md border border-destructive/50 bg-destructive/5 p-4">
        <div className="max-w-full space-y-2 text-center">
          <p className="text-sm font-medium text-destructive">Render Error</p>
          <p className="whitespace-pre-wrap break-words text-xs text-muted-foreground">
            {renderError}
          </p>
        </div>
      </div>
    );
  }

  // Mermaid: render client-side
  if (format === "mermaid" && code.trim()) {
    return <MermaidRenderer code={code} colorScheme={colorScheme} />;
  }

  if (imageUrl) {
    return (
      <div className="flex min-h-[400px] items-center justify-center rounded-md border bg-white p-2">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={imageUrl}
          alt="Rendered figure"
          className="max-h-[600px] max-w-full object-contain"
        />
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-[400px] items-center justify-center rounded-md border border-dashed">
      <div className="flex flex-col items-center gap-2 text-muted-foreground">
        <ImageIcon className="h-10 w-10" />
        <span className="text-sm">Rendered image will appear here</span>
      </div>
    </div>
  );
}
