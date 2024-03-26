/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";

/**
 * useDelayedRender is a hook to delay rendering of a component
 *
 * @param delay - The delay in milliseconds
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
