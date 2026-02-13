import { create } from "zustand";
import type { WSMessage } from "@/lib/ws";

interface GenerateState {
  // Editor state (existing)
  code: string;
  imageUrl: string | null;
  isRendering: boolean;
  renderError: string | null;
  setCode: (code: string) => void;
  setImageUrl: (url: string | null) => void;
  setIsRendering: (v: boolean) => void;
  setRenderError: (err: string | null) => void;

  // AI analysis state (smart mode "分析" step)
  isAnalyzing: boolean;
  setIsAnalyzing: (v: boolean) => void;

  // AI generation state
  taskId: string | null;
  phase: string;
  progress: WSMessage[];
  isGenerating: boolean;
  isCancelled: boolean;
  result: {
    generationId: string;
    code: string;
    format: string;
    explanation: string;
    imageUrl: string;
    reviewPassed: boolean;
    reviewRounds: number;
    reviewCritique: string;
    reviewIssues: string[];
    fullTex: string;
  } | null;
  explanation: string;
  generateError: string | null;

  setTaskId: (id: string | null) => void;
  setPhase: (phase: string) => void;
  pushProgress: (msg: WSMessage) => void;
  setIsGenerating: (v: boolean) => void;
  setIsCancelled: (v: boolean) => void;
  setResult: (result: GenerateState["result"]) => void;
  setExplanation: (text: string) => void;
  setGenerateError: (err: string | null) => void;
  resetGeneration: () => void;
}

export const useGenerateStore = create<GenerateState>((set) => ({
  // Editor state
  code: "",
  imageUrl: null,
  isRendering: false,
  renderError: null,
  setCode: (code) => set({ code }),
  setImageUrl: (imageUrl) => set({ imageUrl }),
  setIsRendering: (isRendering) => set({ isRendering }),
  setRenderError: (renderError) => set({ renderError }),

  // AI analysis state
  isAnalyzing: false,
  setIsAnalyzing: (isAnalyzing) => set({ isAnalyzing }),

  // AI generation state
  taskId: null,
  phase: "",
  progress: [],
  isGenerating: false,
  isCancelled: false,
  result: null,
  explanation: "",
  generateError: null,

  setTaskId: (taskId) => set({ taskId }),
  setPhase: (phase) => set({ phase }),
  pushProgress: (msg) =>
    set((state) => ({ progress: [...state.progress, msg] })),
  setIsGenerating: (isGenerating) => set({ isGenerating }),
  setIsCancelled: (isCancelled) => set({ isCancelled }),
  setResult: (result) => set({ result }),
  setExplanation: (explanation) => set({ explanation }),
  setGenerateError: (generateError) => set({ generateError }),
  resetGeneration: () =>
    set({
      taskId: null,
      phase: "",
      progress: [],
      isGenerating: false,
      isCancelled: false,
      result: null,
      explanation: "",
      generateError: null,
    }),
}));
