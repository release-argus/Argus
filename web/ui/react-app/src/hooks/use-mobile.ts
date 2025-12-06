import * as React from 'react';

/* Screen width at which the UI switches to mobile layout. */
const MOBILE_BREAKPOINT = 768;

/**
 * @returns Whether the screen is smaller than the mobile breakpoint.
 */
export const useIsMobile = () => {
	const [isMobile, setIsMobile] = React.useState<boolean | undefined>(
		undefined,
	);

	React.useEffect(() => {
		const mql = globalThis.matchMedia(
			`(max-width: ${MOBILE_BREAKPOINT - 1}px)`,
		);
		const onChange = () => {
			setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
		};
		mql.addEventListener('change', onChange);
		setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
		return () => {
			mql.removeEventListener('change', onChange);
		};
	}, []);

	return !!isMobile;
};
