/**
 * Returns true if the two arrays are equal, false otherwise.
 *
 * @param arr1 - The first array
 * @param arr2 - The second array
 * @returns True if the arrays are equal, false otherwise
 */
const compareStringArrays = (arr1: string[], arr2: string[]): boolean => {
  // Check if the lengths are different
  if (arr1.length !== arr2.length) {
    return false;
  }

  // Sort both arrays to ensure order doesn't affect comparison
  const sortedArr1 = arr1.slice().sort();
  const sortedArr2 = arr2.slice().sort();

  // Compare each element
  for (let i = 0; i < sortedArr1.length; i++) {
    if (sortedArr1[i] !== sortedArr2[i]) {
      return false;
    }
  }

  // All elements are equal
  return true;
};

export default compareStringArrays;
