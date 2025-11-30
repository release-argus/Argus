import type { OptionReadonly } from '@/components/ui/react-select/custom-components';

/**
 * The option from the provided options array matching the value (case-insensitive).
 *
 * @param options - The options to search.
 * @param value - The value to search for.
 */
export const normaliseForSelect = (
	options: readonly OptionReadonly[],
	value?: string | null,
): { value: string; label: string } | undefined => {
	if (value == null) return undefined;

	const valueLower = value.toLowerCase();
	return options.find((option) => option.value.toLowerCase() === valueLower);
};
