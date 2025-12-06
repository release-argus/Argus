import { type ReactNode, useEffect, useState } from 'react';

/* Function to render a component. */
type RenderFunction = () => ReactNode;
type DelayedRenderer = (
	renderFn: RenderFunction,
	/* Fallback content to render whilst component delayed. */
	fallback?: ReactNode,
) => ReactNode;

/**
 * Delays the rendering of a component.
 *
 * @param delay - The delay to wait before rendering.
 * @returns A function to render the component after the delay.
 */
export const useDelayedRender = (delay: number): DelayedRenderer => {
	const [isDelayed, setIsDelayed] = useState<boolean>(true);

	useEffect(() => {
		const timeout = setTimeout(() => {
			setIsDelayed(false);
		}, delay);

		return () => {
			clearTimeout(timeout);
		};
	}, [delay]);

	return (renderFn: RenderFunction, fallback: ReactNode = null): ReactNode => {
		return isDelayed ? fallback : renderFn();
	};
};
