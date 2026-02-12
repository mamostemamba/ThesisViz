"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CodeEditor } from "./CodeEditor";
import { ImagePreview } from "./ImagePreview";
import { ProgressStream } from "./ProgressStream";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useRender } from "@/lib/queries";
import { exportTeX, generateCreate } from "@/lib/api";
import { connectGeneration, type WSMessage } from "@/lib/ws";
import { Play, Copy, Check, Sparkles, Loader2, Download, ExternalLink } from "lucide-react";

interface ExpertToolboxProps {
  projectId?: string;
}

export function ExpertToolbox({ projectId }: ExpertToolboxProps) {
  const code = useGenerateStore((s) => s.code);
  const imageUrl = useGenerateStore((s) => s.imageUrl);
  const setCode = useGenerateStore((s) => s.setCode);
  const setImageUrl = useGenerateStore((s) => s.setImageUrl);
  const setIsRendering = useGenerateStore((s) => s.setIsRendering);
  const setRenderError = useGenerateStore((s) => s.setRenderError);
  const isRendering = useGenerateStore((s) => s.isRendering);

  const format = useSettingsStore((s) => s.format);
  const language = useSettingsStore((s) => s.language);
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const model = useSettingsStore((s) => s.model);

  const renderMutation = useRender();
  const [copied, setCopied] = useState(false);

  // AI Generate state
  const [aiPrompt, setAiPrompt] = useState("");
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiProgress, setAiProgress] = useState<WSMessage[]>([]);
  const [aiPhase, setAiPhase] = useState("");
  const [aiError, setAiError] = useState<string | null>(null);
  const wsCleanupRef = useRef<(() => void) | null>(null);

  // Track if we've rendered at least once (to enable auto-render on scheme change)
  const hasRendered = useRef(false);
  const prevScheme = useRef(colorScheme);

  // Cleanup WS on unmount
  useEffect(() => {
    return () => {
      wsCleanupRef.current?.();
    };
  }, []);

  const handleRender = useCallback(async () => {
    if (!code.trim()) return;

    // Mermaid renders client-side — just trigger a re-render via state
    if (format === "mermaid") {
      hasRendered.current = true;
      // Force ImagePreview to re-read code by toggling rendering state
      setRenderError(null);
      setImageUrl(null);
      return;
    }

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
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [code, format, language, colorScheme]);

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

  const handleAIGenerate = useCallback(async () => {
    if (!aiPrompt.trim()) return;
    setAiGenerating(true);
    setAiProgress([]);
    setAiPhase("");
    setAiError(null);

    try {
      const res = await generateCreate({
        project_id: projectId || undefined,
        format,
        prompt: aiPrompt,
        language,
        color_scheme: colorScheme,
        model,
      });

      wsCleanupRef.current?.();
      wsCleanupRef.current = connectGeneration(
        res.task_id,
        (msg: WSMessage) => {
          setAiProgress((prev) => [...prev, msg]);
          setAiPhase(msg.phase);

          if (msg.type === "result" && msg.phase === "done") {
            // Fill code into editor
            if (msg.data.code) {
              setCode(msg.data.code);
            }
            if (msg.data.image_url) {
              setImageUrl(msg.data.image_url);
              hasRendered.current = true;
            }
            setAiGenerating(false);
          }

          if (msg.type === "error") {
            setAiError(msg.data.message || "Generation failed");
            setAiGenerating(false);
          }
        },
        () => {
          setAiGenerating(false);
        }
      );
    } catch (err) {
      setAiError(
        err instanceof Error ? err.message : "Failed to start AI generation"
      );
      setAiGenerating(false);
    }
  }, [aiPrompt, projectId, format, language, colorScheme, model, setCode, setImageUrl]);

  return (
    <div className="flex h-[calc(100vh-10rem)] flex-col gap-3">
      {/* AI Generate bar */}
      <div className="flex gap-2">
        <Input
          placeholder="描述你想生成的图表，例如「三层神经网络架构图」"
          value={aiPrompt}
          onChange={(e) => setAiPrompt(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey && !aiGenerating) {
              handleAIGenerate();
            }
          }}
          className="flex-1"
        />
        <Button
          onClick={handleAIGenerate}
          disabled={aiGenerating || !aiPrompt.trim()}
          variant="secondary"
        >
          {aiGenerating ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Sparkles className="mr-2 h-4 w-4" />
          )}
          {aiGenerating ? "生成中..." : "AI 生成"}
        </Button>
      </div>

      {/* AI Progress */}
      {aiGenerating && aiProgress.length > 0 && (
        <ProgressStream messages={aiProgress} phase={aiPhase} />
      )}

      {/* AI Error */}
      {aiError && !aiGenerating && (
        <div className="rounded border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {aiError}
        </div>
      )}

      {/* Editor + Preview */}
      <div className="flex min-h-0 flex-1 gap-4">
        {/* Left: Code editor */}
        <div className="flex w-1/2 flex-col gap-3">
          <div className="min-h-0 flex-1">
            <CodeEditor />
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              className="flex-1"
              onClick={handleRender}
              disabled={isRendering || !code.trim()}
            >
              <Play className="mr-2 h-4 w-4" />
              {isRendering ? "渲染中..." : "渲染"}
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
                {copied ? "已复制！" : "复制到 Overleaf"}
              </Button>
            )}
            {format === "mermaid" && code.trim() && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={async () => {
                    const svgEl = document.querySelector(".mermaid-preview svg");
                    if (!svgEl) return;
                    const { toPng } = await import("html-to-image");
                    const dataUrl = await toPng(svgEl as unknown as HTMLElement, { backgroundColor: "#ffffff" });
                    const link = document.createElement("a");
                    link.download = "figure.png";
                    link.href = dataUrl;
                    link.click();
                  }}
                >
                  <Download className="mr-1 h-4 w-4" />
                  PNG
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const svgEl = document.querySelector(".mermaid-preview svg");
                    if (!svgEl) return;
                    const svgStr = new XMLSerializer().serializeToString(svgEl);
                    const blob = new Blob([svgStr], { type: "image/svg+xml" });
                    const url = URL.createObjectURL(blob);
                    const link = document.createElement("a");
                    link.download = "figure.svg";
                    link.href = url;
                    link.click();
                    URL.revokeObjectURL(url);
                  }}
                >
                  <Download className="mr-1 h-4 w-4" />
                  SVG
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const state = JSON.stringify({ code, mermaid: { theme: "default" } });
                    const encoded = btoa(unescape(encodeURIComponent(state)));
                    window.open(`https://mermaid.live/edit#base64:${encoded}`, "_blank");
                  }}
                >
                  <ExternalLink className="mr-1 h-4 w-4" />
                  Mermaid Live
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Right: Preview */}
        <div className="flex w-1/2 flex-col gap-2">
          <label className="text-sm font-medium">预览</label>
          <div className="mermaid-preview min-h-0 flex-1 overflow-auto">
            <ImagePreview />
          </div>
          {imageUrl && format !== "mermaid" && (
            <a
              href={imageUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-center text-xs text-muted-foreground hover:underline"
            >
              在新标签页打开完整图片
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
