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
  format: "matplotlib",
  language: "en",
  colorScheme: "academic_blue",
  model: "gemini-2.5-flash",
  setFormat: (format) => set({ format }),
  setLanguage: (language) => set({ language }),
  setColorScheme: (colorScheme) => set({ colorScheme }),
  setModel: (model) => set({ model }),
}));
