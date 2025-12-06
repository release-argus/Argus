type EnsureValueParams<T> = {
	path: string;
	getValues: (path: string) => T | undefined;
	setValue: (path: string, value: T) => void;
	defaultValue?: T | null;
	fallback: T;
};

/**
 * Ensures a value is set for a given path by checking against existing values
 * and applying a default or fallback as needed.
 *
 * @template T
 * @param path - The path to the value in the form.
 * @param getValues - The `getValues` function from `useFormContext`.
 * @param setValue - The `setValue` function from `useFormContext`.
 * @param defaultValue - The default value for the given path, used when the current value is empty.
 * @param fallback - A fallback value to use if no default value provided, and the current value is empty.
 */
export const ensureValue = <T>({
	path,
	getValues,
	setValue,
	defaultValue,
	fallback,
}: EnsureValueParams<T | ''>) => {
	const current = getValues(path);
	const newValue = defaultValue ? '' : fallback;

	// If we have no value/were using the default value, set it.
	if (!current) setValue(path, newValue);
};
