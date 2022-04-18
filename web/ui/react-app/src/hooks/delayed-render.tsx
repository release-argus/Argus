/* eslint-disable @typescript-eslint/no-explicit-any */
import { useEffect, useState } from "react";

export const useDelayedRender = (delay: number): any => {
  const [delayed, setDelayed] = useState<boolean>(true);
  useEffect(() => {
    const timeout = setTimeout(() => setDelayed(false), delay);
    return () => clearTimeout(timeout);
  }, [delay]);
  return (fn: any) => !delayed && fn();
};
