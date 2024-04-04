/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";

/**
 * Returns the component after a delay
 *
 * @param delay - The delay to wait before rendering
 * @returns A function to render the component after the delay
 */
export const useDelayedRender = (delay: number): any => {
  const [delayed, setDelayed] = useState<boolean>(true);
  useEffect(() => {
    const timeout = setTimeout(() => setDelayed(false), delay);
    return () => clearTimeout(timeout);
  }, [delay]);
  return (fn: any) => !delayed && fn();
};
