import {
	CARD_TIMESTAMPS_STORAGE_KEY,
	type CardTimestampType,
	DEFAULT_CARD_TIMESTAMPS,
	toolbarTimestampOptions,
} from '@/constants/toolbar';

/**
 * Reads which timestamps to show at the bottom of the service card.
 * An empty string means every timestamp was switched off, which is distinct
 * from the key being absent (never configured).
 *
 * @returns The enabled timestamps.
 */
export const loadCardTimestamps = (): CardTimestampType[] => {
	const stored = localStorage.getItem(CARD_TIMESTAMPS_STORAGE_KEY);
	if (stored === null) return [...DEFAULT_CARD_TIMESTAMPS];

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
