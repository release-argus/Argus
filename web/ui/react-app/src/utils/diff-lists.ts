type DiffListsOptions<A, B, K extends keyof A & keyof B = keyof A & keyof B> = {
	/* The first list to compare. */
	listA: A[];
	/* The second list to compare. */
	listB: B[];
	/* The specific key to use for comparison. */
	key?: K;
	/* The string used to join and separate elements when forming comparable strings. */
	separator?: string;
};

/**
 * Compares two lists to identify whether they are different based on the specified key or their string representations.
 *
 * @template T The type of the objects in the lists.
 * @template K The key of the object properties to compare, defaulting to all keys.
 * @param listA - The first list to compare.
 * @param listB - The second list to compare.
 * @param key  - The specific key to use for comparison. If not provided, the string representation of the entire object is used.
 * @param separator -  The string used to join and separate elements when forming comparable strings.
 * @returns true if the lists are different; otherwise, false.
 */
const diffLists = <A, B, K extends keyof A & keyof B = keyof A & keyof B>({
	listA,
	listB,
	key,
	separator = '-_-',
}: DiffListsOptions<A, B, K>): boolean => {
	if (listA.length !== listB.length) return true;

	// Function to extract the key from an object.
	const extract = (item: A | B) => (key ? String(item[key]) : String(item));

	const sortedA = listA
		.map(extract)
		.toSorted((a, b) => a.localeCompare(b))
		.join(separator);
	const sortedB = listB
		.map(extract)
		.toSorted((a, b) => a.localeCompare(b))
		.join(separator);

	return sortedA !== sortedB;
};

export default diffLists;
