import { useEffect, useState } from "react";

/**
 * useMedia is a hook to determine whether a CSS media query finds any matches
 *
 * @param query - The CSS media query
 * @returns Whether the CSS media query finds any matches
 */
export const useMedia = (query: string): boolean => {
  const mediaQuery = window.matchMedia(query);
  const [matches, setMatches] = useState(mediaQuery.matches);

  useEffect(() => {
    const handler = () => setMatches(mediaQuery.matches);
    mediaQuery.addEventListener("change", handler);
    return () => mediaQuery.removeEventListener("change", handler);
  }, [mediaQuery]);

  return matches;
};
