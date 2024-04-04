import { Button, ButtonGroup } from "react-bootstrap";
import { FC, useEffect } from "react";
import { faAdjust, faMoon, faSun } from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import useLocalStorage from "hooks/local-storage";
import { useMedia } from "hooks/media";

export type themeName = "light" | "dark";
export type themeSetting = themeName | "auto";
export const themeLocalStorageKey = "user-prefers-color-scheme";

/**
 * @returns A button group that allows the user to toggle between light, dark, and auto themes
 */
export const ThemeToggle: FC = () => {
  const [activeTheme, setActiveTheme] = useLocalStorage<themeSetting>(
    themeLocalStorageKey,
    "auto"
  );
  const browserHasThemes = useMedia("(prefers-color-scheme)");
  const browserWantsDarkTheme = useMedia("(prefers-color-scheme: dark)");

  useEffect(() => {
    let theme: themeName;
    if (activeTheme !== "auto") {
      theme = activeTheme;
    } else {
      theme = browserHasThemes && browserWantsDarkTheme ? "dark" : "light";
    }
    document.documentElement.setAttribute("data-bs-theme", theme);
  }, [activeTheme]);

  return (
    <ButtonGroup size="sm">
      <Button
        key="light"
        variant="secondary"
        title="Use light theme"
        active={activeTheme === "light"}
        onClick={() => setActiveTheme("light")}
      >
        <FontAwesomeIcon icon={faSun} />
      </Button>
      <Button
        key="dark"
        variant="secondary"
        title="Use dark theme"
        active={activeTheme === "dark"}
        onClick={() => setActiveTheme("dark")}
      >
        <FontAwesomeIcon icon={faMoon} />
      </Button>
      <Button
        key="auto"
        variant="secondary"
        title="Use browser-preferred theme"
        active={activeTheme === "auto"}
        onClick={() => setActiveTheme("auto")}
      >
        <FontAwesomeIcon icon={faAdjust} />
      </Button>
    </ButtonGroup>
  );
};
