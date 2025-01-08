import { Dispatch, SetStateAction, useEffect, useState } from 'react';

/**
 * The value from local storage, and a setter.
 *
 * @param localStorageKey - The key to store the value in local storage.
 * @param initialState - The initial state of the value.
 * @returns The value and a function to set the value.
 */
export function useLocalStorage<S>(
	localStorageKey: string,
	initialState: S,
): [S, Dispatch<SetStateAction<S>>] {
	const localStorageState = JSON.parse(
		localStorage.getItem(localStorageKey) || JSON.stringify(initialState),
	);
	const [value, setValue] = useState(localStorageState);

	useEffect(() => {
		const serializedState = JSON.stringify(value);
		localStorage.setItem(localStorageKey, serializedState);
	}, [localStorageKey, value]);

	return [value, setValue];
}

export default useLocalStorage;
