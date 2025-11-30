import { useCallback, useState } from 'react';
import { useFormContext } from 'react-hook-form';

/**
 * Values in the form, with a function to refetch.
 *
 * @param name - The name of the field in the form.
 * @param undefinedInitially - Whether the value is undefined initially.
 * @returns The data in the form at `name`, and a function to refetch the data.
 */
const useValuesRefetch = <T = unknown>(
	name: string,
	undefinedInitially?: boolean,
) => {
	const { getValues } = useFormContext();
	const [data, setData] = useState<T | undefined>(
		undefinedInitially ? undefined : (getValues(name) as T),
	);

	const refetchData = useCallback(() => {
		const values = getValues(name) as T;
		setData(values);
	}, [getValues, name]);

	return { data, refetchData };
};

export default useValuesRefetch;
