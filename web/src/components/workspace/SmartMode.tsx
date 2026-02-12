"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useSettingsStore } from "@/stores/useSettingsStore";
import { useGenerateStore } from "@/stores/useGenerateStore";
import { useAnalyze, useGenerateCreate } from "@/lib/queries";
import { connectGeneration, type WSMessage } from "@/lib/ws";
import { ProgressStream, collectImageSnapshots } from "./ProgressStream";
import { ResultPanel } from "./ResultPanel";
import { generateCreate } from "@/lib/api";
import { Search, Sparkles, Loader2 } from "lucide-react";
import type { Recommendation } from "@/types/api";

interface SmartModeProps {
  projectId?: string;
}

export function SmartMode({ projectId }: SmartModeProps) {
  const [text, setText] = useState("");
  const [thesisTitle, setThesisTitle] = useState("");
  const [thesisAbstract, setThesisAbstract] = useState("");
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [selectedRec, setSelectedRec] = useState<Recommendation | null>(null);
  const [editingPrompt, setEditingPrompt] = useState<string | null>(null);
  const [parentCode, setParentCode] = useState<string | undefined>(undefined);

  const language = useSettingsStore((s) => s.language);
  const format = useSettingsStore((s) => s.format);
  const colorScheme = useSettingsStore((s) => s.colorScheme);
  const model = useSettingsStore((s) => s.model);

  const taskId = useGenerateStore((s) => s.taskId);
  const phase = useGenerateStore((s) => s.phase);
  const progress = useGenerateStore((s) => s.progress);
  const isGenerating = useGenerateStore((s) => s.isGenerating);
  const result = useGenerateStore((s) => s.result);
  const generateError = useGenerateStore((s) => s.generateError);
  const setTaskId = useGenerateStore((s) => s.setTaskId);
  const setPhase = useGenerateStore((s) => s.setPhase);
  const pushProgress = useGenerateStore((s) => s.pushProgress);
  const setIsGenerating = useGenerateStore((s) => s.setIsGenerating);
  const setResult = useGenerateStore((s) => s.setResult);
  const setGenerateError = useGenerateStore((s) => s.setGenerateError);
  const resetGeneration = useGenerateStore((s) => s.resetGeneration);

  const analyzeMutation = useAnalyze();
  const wsCleanupRef = useRef<(() => void) | null>(null);
  const progressRef = useRef<HTMLDivElement>(null);
  const resultRef = useRef<HTMLDivElement>(null);

  // Reset global generation store & cleanup WS when project changes (component remounts via key)
  useEffect(() => {
    resetGeneration();
    return () => {
      wsCleanupRef.current?.();
    };
  }, [resetGeneration]);

  // Auto-scroll to progress when generation starts or new messages arrive
  useEffect(() => {
    if (isGenerating && progress.length > 0) {
      progressRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }, [isGenerating, progress.length]);

  // Auto-scroll to result when done
  useEffect(() => {
    if (result) {
      resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }, [result]);

  const handleAnalyze = useCallback(async () => {
    if (!text.trim()) return;
    setRecommendations([]);
    setSelectedRec(null);
    setEditingPrompt(null);
    resetGeneration();

    const res = await analyzeMutation.mutateAsync({
      text,
      language,
      thesis_title: thesisTitle || undefined,
      thesis_abstract: thesisAbstract || undefined,
      model,
    });
    setRecommendations(res.recommendations || []);
  }, [text, language, thesisTitle, thesisAbstract, model, analyzeMutation, resetGeneration]);

  const startGeneration = useCallback(
    async (prompt: string, fmt?: string, keepResult?: boolean) => {
      if (keepResult) {
        // Refine: keep old result visible, only clear progress/error
        useGenerateStore.setState({
          taskId: null,
          phase: "",
          progress: [],
          generateError: null,
        });
      } else {
        resetGeneration();
        setParentCode(undefined);
      }
      setIsGenerating(true);

      try {
        const res = await generateCreate({
          project_id: projectId || undefined,
          format: fmt || format,
          prompt,
          language,
          color_scheme: colorScheme,
          thesis_title: thesisTitle || undefined,
          thesis_abstract: thesisAbstract || undefined,
          model,
        });

        setTaskId(res.task_id);

        // Connect WebSocket
        wsCleanupRef.current?.();
        wsCleanupRef.current = connectGeneration(
          res.task_id,
          (msg: WSMessage) => {
            pushProgress(msg);
            setPhase(msg.phase);

            if (msg.type === "result" && msg.phase === "done") {
              setResult({
                generationId: msg.data.generation_id || "",
                code: msg.data.code || "",
                format: msg.data.format || format,
                explanation: msg.data.explanation || "",
                imageUrl: msg.data.image_url || "",
                reviewPassed: msg.data.review_passed || false,
                reviewRounds: msg.data.review_rounds || 0,
                reviewCritique: msg.data.critique || "",
                reviewIssues: msg.data.issues || [],
              });
              setIsGenerating(false);
            }

            if (msg.type === "error") {
              setGenerateError(msg.data.message || "Generation failed");
              setIsGenerating(false);
            }
          },
          () => {
            // on close, if still generating mark as done
            if (useGenerateStore.getState().isGenerating) {
              setIsGenerating(false);
            }
          }
        );
      } catch (err) {
        setGenerateError(
          err instanceof Error ? err.message : "Failed to start generation"
        );
        setIsGenerating(false);
      }
    },
    [
      projectId,
      format,
      language,
      colorScheme,
      thesisTitle,
      thesisAbstract,
      model,
      resetGeneration,
      setTaskId,
      setPhase,
      pushProgress,
      setIsGenerating,
      setResult,
      setGenerateError,
    ]
  );

  const handleRefine = useCallback(
    (modification: string) => {
      if (!result) return;
      setParentCode(result.code);
      startGeneration(
        `Modify this existing code:\n\n${result.code}\n\nModification: ${modification}`,
        result.format,
        true // keep old result visible during refine
      );
    },
    [result, startGeneration]
  );

  return (
    <div className="space-y-6">
      {/* Step 1: Input */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">描述你的论文</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <input
            type="text"
            placeholder="论文标题（可选）"
            value={thesisTitle}
            onChange={(e) => setThesisTitle(e.target.value)}
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
          />
          <Textarea
            placeholder="论文摘要（可选）"
            className="min-h-[80px]"
            value={thesisAbstract}
            onChange={(e) => setThesisAbstract(e.target.value)}
          />
          <Textarea
            placeholder="粘贴需要配图的段落内容..."
            className="min-h-[150px]"
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
          <Button
            onClick={handleAnalyze}
            disabled={analyzeMutation.isPending || !text.trim()}
            className="w-full"
          >
            {analyzeMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Search className="mr-2 h-4 w-4" />
            )}
            {analyzeMutation.isPending ? "分析中..." : "分析"}
          </Button>
        </CardContent>
      </Card>

      {/* Step 2: Recommendations */}
      {recommendations.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-sm font-semibold">推荐图表</h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {recommendations.map((rec, i) => (
              <Card
                key={i}
                className={`cursor-pointer transition-colors ${
                  selectedRec === rec
                    ? "border-primary bg-primary/5"
                    : "hover:border-muted-foreground/30"
                }`}
                onClick={() => {
                  setSelectedRec(rec);
                  setEditingPrompt(rec.drawing_prompt);
                }}
              >
                <CardContent className="p-3 space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">{rec.title}</span>
                    <Badge variant="secondary" className="text-xs">
                      P{rec.priority}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {rec.description}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Step 3: Prompt editing */}
      {editingPrompt !== null && selectedRec && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">画图提示词（可编辑）</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Textarea
              value={editingPrompt}
              onChange={(e) => setEditingPrompt(e.target.value)}
              className="min-h-[120px] text-sm"
            />
            <Button
              onClick={() =>
                startGeneration(
                  editingPrompt,
                  selectedRec.format || format
                )
              }
              disabled={!editingPrompt.trim() || isGenerating}
              className="w-full"
            >
              {isGenerating ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Sparkles className="mr-2 h-4 w-4" />
              )}
              {isGenerating ? "生成中..." : "确认生成"}
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Progress — keep visible after completion so review results remain */}
      {progress.length > 0 && (
        <div ref={progressRef}>
          <ProgressStream messages={progress} phase={phase} />
        </div>
      )}

      {/* Error */}
      {generateError && !isGenerating && (
        <div className="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300">
          {generateError}
        </div>
      )}

      {/* Result */}
      {result && (
        <div ref={resultRef}>
        <ResultPanel
          code={result.code}
          format={result.format}
          explanation={result.explanation}
          imageUrl={result.imageUrl}
          reviewPassed={result.reviewPassed}
          reviewRounds={result.reviewRounds}
          reviewCritique={result.reviewCritique}
          reviewIssues={result.reviewIssues}
          onRefine={handleRefine}
          isRefining={isGenerating}
          parentCode={parentCode}
          imageSnapshots={collectImageSnapshots(progress)}
        />
        </div>
      )}
    </div>
  );
}
