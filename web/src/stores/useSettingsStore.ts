import { create } from "zustand";
import type { CustomColors } from "@/types/api";

interface SettingsState {
  format: "tikz" | "matplotlib" | "mermaid";
  language: "en" | "zh";
  colorScheme: string;
  customColors: CustomColors | null;
  model: string;
  setFormat: (format: SettingsState["format"]) => void;
  setLanguage: (language: SettingsState["language"]) => void;
  setColorScheme: (scheme: string) => void;
  setCustomColors: (colors: CustomColors | null) => void;
  setModel: (model: string) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  format: "tikz",
  language: "zh",
  colorScheme: "drawio",
  customColors: null,
  model: "gemini-3-pro-preview",
  setFormat: (format) => set({ format }),
  setLanguage: (language) => set({ language }),
  setColorScheme: (colorScheme) => set({ colorScheme }),
  setCustomColors: (customColors) => set({ customColors }),
  setModel: (model) => set({ model }),
}));
