import type { OptionType } from '@/components/ui/react-select/custom-components';

/**
 * Converts a `string` array to an `OptionType` array,
 * or returns the input unchanged when not a `string[]`.
 *
 * @param input - The input value to check and convert.
 * @param sort - Whether to sort the options alphabetically.
 */
export const convertStringArrayToOptionTypeArray = (
	input: string[] | OptionType[],
	sort?: boolean,
): OptionType[] => {
	// Check whether already converted to an 'Options' list.
	if (Array.isArray(input) && input.every((item) => typeof item === 'string')) {
		// Convert to a list of `Option` objects.
		if (sort) {
			return input
				.toSorted((a, b) => a.localeCompare(b))
				.map((opt) => createOption(opt));
		}
		return input.map((opt) => createOption(opt));
	}

	// Already a list of `Option` objects, return it.
	if (sort) return input.toSorted((a, b) => a.label.localeCompare(b.label));
	return input;
};

// Create an OptionType object from a string.
export const createOption = (inputValue: string): OptionType => ({
	label: inputValue,
	value: inputValue,
});
