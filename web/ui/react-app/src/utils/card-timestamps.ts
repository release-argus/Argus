import {
	CARD_TIMESTAMPS_STORAGE_KEY,
	type CardTimestampType,
	toolbarTimestampOptions,
} from '@/constants/toolbar';

/**
 * Reads which timestamps to show at the bottom of the service card.
 * An empty string means every timestamp was switched off, which is distinct
 * from the key being absent (never configured).
 *
 * @param defaults - The timestamps to show when this browser has no preference.
 * @returns The enabled timestamps.
 */
export const loadCardTimestamps = (
	defaults: readonly CardTimestampType[],
): CardTimestampType[] => {
	const stored = localStorage.getItem(CARD_TIMESTAMPS_STORAGE_KEY);
	if (stored === null) return [...defaults];

	const enabled = new Set(stored.split(',').filter(Boolean));
	return toolbarTimestampOptions
		.map(({ key }) => key)
		.filter((key) => enabled.has(key));
};

/**
 * Persists the timestamps to show at the bottom of the service card.
 *
 * @param timestamps - The enabled timestamps.
 */
export const persistCardTimestamps = (timestamps: CardTimestampType[]) => {
	localStorage.setItem(CARD_TIMESTAMPS_STORAGE_KEY, timestamps.join(','));
};

/**
 * Drops this browser's timestamp choice, so the saved default applies again.
 */
export const clearCardTimestamps = () => {
	localStorage.removeItem(CARD_TIMESTAMPS_STORAGE_KEY);
};
