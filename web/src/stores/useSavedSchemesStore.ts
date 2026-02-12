import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { ColorPair } from "@/types/api";

export interface SavedScheme {
  id: string;
  name: string;
  pairs: ColorPair[];
  createdAt: number;
}

interface SavedSchemesState {
  schemes: SavedScheme[];
  save: (name: string, pairs: ColorPair[]) => string;
  remove: (id: string) => void;
  rename: (id: string, newName: string) => void;
}

const MAX_SCHEMES = 20;

export const useSavedSchemesStore = create<SavedSchemesState>()(
  persist(
    (set) => ({
      schemes: [],

      save: (name, pairs) => {
        const id = crypto.randomUUID();
        set((state) => ({
          schemes: [
            { id, name, pairs, createdAt: Date.now() },
            ...state.schemes,
          ].slice(0, MAX_SCHEMES),
        }));
        return id;
      },

      remove: (id) =>
        set((state) => ({
          schemes: state.schemes.filter((s) => s.id !== id),
        })),

      rename: (id, newName) =>
        set((state) => ({
          schemes: state.schemes.map((s) =>
            s.id === id ? { ...s, name: newName } : s
          ),
        })),
    }),
    {
      name: "thesisviz-saved-schemes",
    }
  )
);
