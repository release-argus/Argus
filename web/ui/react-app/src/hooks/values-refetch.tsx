import { useFormContext } from "react-hook-form";
import { useState } from "react";

/**
 * Returns the data in the form and a function to refetch this data
 * @param name - The name of the field in the form
 * @param undefinedInitially - Whether the value should be undefined initially
 * @returns The data and a function to refetch the data
 */
const useValuesRefetch = (name: string, undefinedInitially?: boolean) => {
  const { getValues } = useFormContext();
  const [data, setData] = useState(
    undefinedInitially ? undefined : getValues(name)
  );
  const refetchData = () => {
    const values = getValues(name);
    setData(values);
  };

  return { data, refetchData };
};

export default useValuesRefetch;
