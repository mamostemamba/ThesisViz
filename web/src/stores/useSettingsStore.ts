import { create } from "zustand";

interface SettingsState {
  format: "tikz" | "matplotlib" | "mermaid";
  language: "en" | "zh";
  colorScheme: string;
  setFormat: (format: SettingsState["format"]) => void;
  setLanguage: (language: SettingsState["language"]) => void;
  setColorScheme: (scheme: string) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  format: "matplotlib",
  language: "en",
  colorScheme: "academic_blue",
  setFormat: (format) => set({ format }),
  setLanguage: (language) => set({ language }),
  setColorScheme: (colorScheme) => set({ colorScheme }),
}));
