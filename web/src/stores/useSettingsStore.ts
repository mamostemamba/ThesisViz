import { create } from "zustand";

interface SettingsState {
  format: "tikz" | "matplotlib" | "mermaid";
  language: "en" | "zh";
  colorScheme: string;
  model: string;
  setFormat: (format: SettingsState["format"]) => void;
  setLanguage: (language: SettingsState["language"]) => void;
  setColorScheme: (scheme: string) => void;
  setModel: (model: string) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  format: "tikz",
  language: "zh",
  colorScheme: "drawio",
  model: "gemini-3-pro-preview",
  setFormat: (format) => set({ format }),
  setLanguage: (language) => set({ language }),
  setColorScheme: (colorScheme) => set({ colorScheme }),
  setModel: (model) => set({ model }),
}));
