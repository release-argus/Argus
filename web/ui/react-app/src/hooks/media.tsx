import { useEffect, useState } from 'react';

/**
 * Whether the CSS media query matches.
 *
 * @param query - The CSS media query to match.
 * @returns Whether the CSS media query matches.
 */
export const useMedia = (query: string): boolean => {
	const mediaQuery = window.matchMedia(query);
	const [matches, setMatches] = useState(mediaQuery.matches);

	useEffect(() => {
		const handler = () => setMatches(mediaQuery.matches);
		mediaQuery.addEventListener('change', handler);
		return () => mediaQuery.removeEventListener('change', handler);
	}, [mediaQuery]);

	return matches;
};
