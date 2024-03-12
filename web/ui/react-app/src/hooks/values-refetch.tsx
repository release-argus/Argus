import { useFormContext } from "react-hook-form";
import { useState } from "react";

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
