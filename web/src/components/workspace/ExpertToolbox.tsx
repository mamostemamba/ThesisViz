"use client";

import { Button } from "@/components/ui/button";
import { CodeEditor } from "./CodeEditor";
import { ImagePreview } from "./ImagePreview";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useRender } from "@/lib/queries";
import { Play } from "lucide-react";

export function ExpertToolbox() {
  const code = useGenerateStore((s) => s.code);
  const setImageUrl = useGenerateStore((s) => s.setImageUrl);
  const setIsRendering = useGenerateStore((s) => s.setIsRendering);
  const setRenderError = useGenerateStore((s) => s.setRenderError);
  const isRendering = useGenerateStore((s) => s.isRendering);

  const format = useSettingsStore((s) => s.format);
  const language = useSettingsStore((s) => s.language);
  const colorScheme = useSettingsStore((s) => s.colorScheme);

  const renderMutation = useRender();

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

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div className="space-y-4">
        <CodeEditor />
        <Button
          className="w-full"
          onClick={handleRender}
          disabled={isRendering || !code.trim()}
        >
          <Play className="mr-2 h-4 w-4" />
          {isRendering ? "Rendering..." : "Render"}
        </Button>
      </div>
      <div>
        <label className="mb-2 block text-sm font-medium">Preview</label>
        <ImagePreview />
      </div>
    </div>
  );
}
