import { useFormContext } from 'react-hook-form';
import { useState } from 'react';

/**
 * Values in the form, with a function to refetch.
 *
 * @param name - The name of the field in the form.
 * @param undefinedInitially - Whether the value is undefined initially.
 * @returns The data in the form at name, and a function to refetch the data.
 */
const useValuesRefetch = (name: string, undefinedInitially?: boolean) => {
	const { getValues } = useFormContext();
	const [data, setData] = useState(
		undefinedInitially ? undefined : getValues(name),
	);
	const refetchData = () => {
		const values = getValues(name);
		setData(values);
	};

	return { data, refetchData };
};

export default useValuesRefetch;
