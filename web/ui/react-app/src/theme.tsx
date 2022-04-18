import { Button, ButtonGroup } from "react-bootstrap";
import { FC, useEffect } from "react";
import { faAdjust, faMoon, faSun } from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useTheme } from "./contexts/theme";

export const themeLocalStorageKey = "user-prefers-color-scheme";

export const Theme: FC = () => {
  const { theme } = useTheme();

  useEffect(() => {
    document.body.classList.toggle("theme-dark", theme === "theme-dark");
    document.body.classList.toggle("theme-light", theme === "theme-light");
  }, [theme]);

  return null;
};

export const ThemeToggle: FC = () => {
  const { themePreference, setTheme } = useTheme();

  return (
    <ButtonGroup size="sm">
      <Button
        key="light"
        variant="secondary"
        title="Use light theme"
        active={themePreference === "theme-light"}
        onClick={() => setTheme("theme-light")}
      >
        <FontAwesomeIcon
          icon={faSun}
          className={
            themePreference === "theme-light" ? "text-white" : "text-dark"
          }
        />
      </Button>
      <Button
        key="dark"
        variant="secondary"
        title="Use dark theme"
        active={themePreference === "theme-dark"}
        onClick={() => setTheme("theme-dark")}
      >
        <FontAwesomeIcon
          icon={faMoon}
          className={
            themePreference === "theme-dark" ? "text-white" : "text-dark"
          }
        />
      </Button>
      <Button
        key="auto"
        variant="secondary"
        title="Use browser-preferred theme"
        active={themePreference === "auto"}
        onClick={() => setTheme("auto")}
      >
        <FontAwesomeIcon
          icon={faAdjust}
          className={themePreference === "auto" ? "text-white" : "text-dark"}
        />
      </Button>
    </ButtonGroup>
  );
};
