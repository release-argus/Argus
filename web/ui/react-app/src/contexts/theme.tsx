import { createContext, useContext } from "react";

export type themeName = "theme-light" | "theme-dark";
export type themeSetting = themeName | "auto";

export interface ThemeCtx {
  theme: themeName;
  themePreference: themeSetting;
  setTheme: (t: themeSetting) => void;
}

// defaults, will be overridden in App.tsx
export const ThemeContext = createContext<ThemeCtx>({
  theme: "theme-dark",
  themePreference: "auto",
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  setTheme: (s: themeSetting) => {},
});

export const useTheme = () => {
  return useContext(ThemeContext);
};
