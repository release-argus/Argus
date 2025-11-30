/**
 * The first non-zero length argument.
 *
 * @param args - The arguments to check for the first non-zero length.
 * @returns The first non-zero length argument from the list of arguments.
 */
const firstNonEmpty = <T extends unknown[] | undefined>(
	...args: T[]
): NonNullable<T> => {
	for (const arg of args) {
		if (arg?.length) return arg;
	}

	return [] as unknown as NonNullable<T>;
};

export default firstNonEmpty;
