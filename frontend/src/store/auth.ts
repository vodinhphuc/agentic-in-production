import { create } from "zustand";

type AuthState = {
  loggedIn: boolean;
  username?: string;
  setLoggedIn: (username: string) => void;
  setLoggedOut: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  loggedIn: false,
  setLoggedIn: (username) => set({ loggedIn: true, username }),
  setLoggedOut: () => set({ loggedIn: false, username: undefined }),
}));
