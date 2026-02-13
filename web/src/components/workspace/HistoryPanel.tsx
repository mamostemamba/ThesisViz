"use client";

import { useState } from "react";
import { useGenerations } from "@/lib/queries";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { GenerationCard } from "./GenerationCard";
import { getGenerationDetail } from "@/lib/api";
import { Loader2 } from "lucide-react";
import type { Generation } from "@/types/api";

interface HistoryPanelProps {
  projectId: string;
  onLoadResult?: () => void;
}

export function HistoryPanel({ projectId, onLoadResult }: HistoryPanelProps) {
  const { data, isLoading } = useGenerations(projectId);
  const setCode = useGenerateStore((s) => s.setCode);
  const setImageUrl = useGenerateStore((s) => s.setImageUrl);
  const setResult = useGenerateStore((s) => s.setResult);
  const setIsCancelled = useGenerateStore((s) => s.setIsCancelled);
  const setFormat = useSettingsStore((s) => s.setFormat);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  const generations = data?.items ?? [];

  const handleSelect = async (gen: Generation) => {
    setSelectedId(gen.id);
    setLoadingDetail(true);
    setIsCancelled(false);

    try {
      const detail = await getGenerationDetail(gen.id);
      if (detail.code) {
        setCode(detail.code);
      }
      if (detail.image_url) {
        setImageUrl(detail.image_url);
      }
      setFormat(detail.format as "tikz" | "matplotlib" | "mermaid");
      setResult({
        generationId: detail.id,
        code: detail.code || "",
        format: detail.format,
        explanation: detail.explanation || "",
        imageUrl: detail.image_url || "",
        reviewPassed: true,
        reviewRounds: 0,
        reviewCritique: "",
        reviewIssues: [],
        fullTex: "",
      });
      onLoadResult?.();
    } catch {
      // silently fail
    } finally {
      setLoadingDetail(false);
    }
  };

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
        保存项目后可查看生成历史
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (generations.length === 0) {
    return (
      <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
        暂无生成记录。使用智能分析或专家工具箱来创建图表。
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">
          生成历史（{generations.length}）
        </h3>
        {loadingDetail && (
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        )}
      </div>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {generations.map((gen) => (
          <GenerationCard
            key={gen.id}
            generation={gen}
            isSelected={selectedId === gen.id}
            onClick={() => handleSelect(gen)}
          />
        ))}
      </div>
    </div>
  );
}
