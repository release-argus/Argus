/* eslint-disable @typescript-eslint/no-explicit-any */
import { ReactNode, useEffect, useState } from 'react';

/**
 * Delays the rendering of a component.
 *
 * @param delay - The delay to wait before rendering.
 * @returns A function to render the component after the delay.
 */
export const useDelayedRender = (delay: number): any => {
	const [delayed, setDelayed] = useState<boolean>(true);
	useEffect(() => {
		const timeout = setTimeout(() => setDelayed(false), delay);
		return () => clearTimeout(timeout);
	}, [delay]);
	return (fn: any, placeholder: ReactNode = null) =>
		delayed ? placeholder : fn();
};
