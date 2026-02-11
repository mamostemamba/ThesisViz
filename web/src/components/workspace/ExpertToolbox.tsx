"use client";

import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { CodeEditor } from "./CodeEditor";
import { ImagePreview } from "./ImagePreview";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useRender } from "@/lib/queries";
import { exportTeX } from "@/lib/api";
import { Play, Copy, Check } from "lucide-react";
import { useState } from "react";

export function ExpertToolbox() {
  const code = useGenerateStore((s) => s.code);
  const imageUrl = useGenerateStore((s) => s.imageUrl);
  const setImageUrl = useGenerateStore((s) => s.setImageUrl);
  const setIsRendering = useGenerateStore((s) => s.setIsRendering);
  const setRenderError = useGenerateStore((s) => s.setRenderError);
  const isRendering = useGenerateStore((s) => s.isRendering);

  const format = useSettingsStore((s) => s.format);
  const language = useSettingsStore((s) => s.language);
  const colorScheme = useSettingsStore((s) => s.colorScheme);

  const renderMutation = useRender();
  const [copied, setCopied] = useState(false);

  // Track if we've rendered at least once (to enable auto-render on scheme change)
  const hasRendered = useRef(false);
  const prevScheme = useRef(colorScheme);

  const handleRender = async () => {
    if (!code.trim()) return;

    setIsRendering(true);
    setRenderError(null);
    setImageUrl(null);

    try {
      const result = await renderMutation.mutateAsync({
        code,
        format,
        language,
        color_scheme: colorScheme,
      });

      if (result.status === "ok" && result.image_url) {
        setImageUrl(result.image_url);
        hasRendered.current = true;
      } else {
        setRenderError(result.error || "Unknown render error");
      }
    } catch (err) {
      setRenderError(
        err instanceof Error ? err.message : "Failed to render"
      );
    } finally {
      setIsRendering(false);
    }
  };

  // Auto-render when color scheme changes (only if we've already rendered once)
  useEffect(() => {
    if (prevScheme.current !== colorScheme) {
      prevScheme.current = colorScheme;
      if (hasRendered.current && code.trim() && !isRendering) {
        handleRender();
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [colorScheme]);

  const handleCopyForOverleaf = async () => {
    if (!code.trim() || format !== "tikz") return;

    try {
      const result = await exportTeX({
        code,
        language,
        color_scheme: colorScheme,
      });
      await navigator.clipboard.writeText(result.tex);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback: copy raw code
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="flex h-[calc(100vh-10rem)] gap-4">
      {/* Left: Code editor */}
      <div className="flex w-1/2 flex-col gap-3">
        <div className="min-h-0 flex-1">
          <CodeEditor />
        </div>
        <div className="flex gap-2">
          <Button
            className="flex-1"
            onClick={handleRender}
            disabled={isRendering || !code.trim()}
          >
            <Play className="mr-2 h-4 w-4" />
            {isRendering ? "Rendering..." : "Render"}
          </Button>
          {format === "tikz" && (
            <Button
              variant="outline"
              onClick={handleCopyForOverleaf}
              disabled={!code.trim()}
            >
              {copied ? (
                <Check className="mr-2 h-4 w-4" />
              ) : (
                <Copy className="mr-2 h-4 w-4" />
              )}
              {copied ? "Copied!" : "Copy for Overleaf"}
            </Button>
          )}
        </div>
      </div>

      {/* Right: Preview */}
      <div className="flex w-1/2 flex-col gap-2">
        <label className="text-sm font-medium">Preview</label>
        <div className="min-h-0 flex-1 overflow-auto">
          <ImagePreview />
        </div>
        {imageUrl && (
          <a
            href={imageUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-center text-xs text-muted-foreground hover:underline"
          >
            Open full image in new tab
          </a>
        )}
      </div>
    </div>
  );
}
