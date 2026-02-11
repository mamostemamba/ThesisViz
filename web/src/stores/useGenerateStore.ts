import { create } from "zustand";

interface GenerateState {
  code: string;
  imageUrl: string | null;
  isRendering: boolean;
  renderError: string | null;
  setCode: (code: string) => void;
  setImageUrl: (url: string | null) => void;
  setIsRendering: (v: boolean) => void;
  setRenderError: (err: string | null) => void;
}

export const useGenerateStore = create<GenerateState>((set) => ({
  code: "",
  imageUrl: null,
  isRendering: false,
  renderError: null,
  setCode: (code) => set({ code }),
  setImageUrl: (imageUrl) => set({ imageUrl }),
  setIsRendering: (isRendering) => set({ isRendering }),
  setRenderError: (renderError) => set({ renderError }),
}));
