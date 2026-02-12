"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Generation } from "@/types/api";

interface GenerationCardProps {
  generation: Generation;
  isSelected: boolean;
  onClick: () => void;
}

export function GenerationCard({
  generation,
  isSelected,
  onClick,
}: GenerationCardProps) {
  const formatLabel = {
    tikz: "TikZ",
    matplotlib: "Matplotlib",
    mermaid: "Mermaid",
  }[generation.format] || generation.format;

  const statusColor = {
    success: "text-green-600",
    failed: "text-red-600",
    processing: "text-yellow-600",
    queued: "text-muted-foreground",
  }[generation.status] || "text-muted-foreground";

  const promptPreview =
    generation.prompt.length > 80
      ? generation.prompt.slice(0, 80) + "..."
      : generation.prompt;

  return (
    <Card
      className={`cursor-pointer transition-colors ${
        isSelected
          ? "border-primary bg-primary/5"
          : "hover:border-muted-foreground/30"
      }`}
      onClick={onClick}
    >
      <CardContent className="p-3 space-y-2">
        <div className="flex items-center justify-between">
          <Badge variant="secondary" className="text-xs">
            {formatLabel}
          </Badge>
          <span className={`text-xs ${statusColor}`}>
            {generation.status}
          </span>
        </div>
        <p className="text-xs text-muted-foreground line-clamp-2">
          {promptPreview}
        </p>
        <p className="text-xs text-muted-foreground/60">
          {new Date(generation.created_at).toLocaleString()}
        </p>
      </CardContent>
    </Card>
  );
}
