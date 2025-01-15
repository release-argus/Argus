/**
 * Whether the value is empty, null, or undefined.
 *
 * @param value - The value to check.
 * @returns true when value is empty, null, or undefined. false otherwise.
 */
const isEmptyOrNull = (value: unknown): value is null | undefined | '' => {
	return (value ?? '') === '';
};

export default isEmptyOrNull;
