import { create } from "zustand";
import type { components } from "@/api/types.gen";

type Session = components["schemas"]["Session"];

type State = {
  sessions: Session[];
  current: Session | null;
  setSessions: (s: Session[]) => void;
  setCurrent: (s: Session | null) => void;
};

export const useSessionStore = create<State>((set) => ({
  sessions: [],
  current: null,
  setSessions: (sessions) => set({ sessions }),
  setCurrent: (current) => set({ current }),
}));
