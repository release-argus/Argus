/**
 * Whether the value is empty, null, or undefined.
 *
 * @param value - The value to check.
 * @returns true when value is empty, null, or undefined. false otherwise.
 */
const isEmptyOrNull = (value: unknown): boolean => {
	return (value ?? '') === '';
};

export default isEmptyOrNull;
