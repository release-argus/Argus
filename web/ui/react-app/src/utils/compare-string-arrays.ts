/**
 * Whether two arrays are equal.
 *
 * @param arr1 - The first array.
 * @param arr2 - The second array.
 * @returns True if the arrays are equal, false otherwise.
 */
const compareStringArrays = (arr1: string[], arr2: string[]): boolean => {
	// Check whether the lengths differ.
	if (arr1.length !== arr2.length) {
		return false;
	}

	// Sort arrays so that order doesn't affect comparison.
	const sortedArrayOne = arr1.slice().sort((a, b) => a.localeCompare(b));
	const sortedArrayTwo = arr2.slice().sort((a, b) => a.localeCompare(b));

	// Compare each element.
	for (let i = 0; i < sortedArrayOne.length; i++) {
		if (sortedArrayOne[i] !== sortedArrayTwo[i]) {
			return false;
		}
	}

	// All elements equal.
	return true;
};

export default compareStringArrays;
