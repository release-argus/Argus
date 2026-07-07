import type { OptionReadonly } from '@/components/ui/react-select/custom-components';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';

/**
 * The option from the provided options array matching the value (case-insensitive).
 *
 * @param options - The options to search.
 * @param value - The value to search for.
 */
export const normaliseForSelect = <T extends readonly OptionReadonly[]>(
	options: T,
	value?: string | null,
): { value: T[number]['value']; label: T[number]['label'] } | undefined => {
	if (value == null) return undefined;

	const valueLower = value.toLowerCase();
	return options.find((option) => option.value.toLowerCase() === valueLower);
};

/**
 * Prepends a `'<label> (default)'` option (selecting it clears back to the
 * schema default) to `options`, when `defaultValue` matches one of them.
 *
 * @param options - The options to search for the default value in.
 * @param defaultValue - The schema default value, if any.
 * @param tailOptions - The options to append after the default entry - defaults
 *   to `options`. Pass a filtered list when `options` already contains its own
 *   blank/no-op entry that would otherwise collide with the synthetic default one.
 */
export const withDefaultOption = <T extends readonly OptionReadonly[]>(
	options: T,
	defaultValue?: string | null,
	tailOptions: readonly OptionReadonly[] = options,
): readonly OptionReadonly[] => {
	const defaultOption = normaliseForSelect(options, defaultValue);
	if (!defaultOption) return options;

	return [
		{ label: `${defaultOption.label} (default)`, value: nullString },
		...tailOptions,
	];
};
