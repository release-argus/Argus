import { FormCheck, FormItem } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { useEffect } from "react";

const REGEX = ({ name }: { name: string }) => {
  const { setValue } = useFormContext();

  // Template toggle
  const templateToggle = useWatch({ name: `${name}.template_toggle` });
  useEffect(() => {
    // Clear the template if the toggle is false
    if (templateToggle === false) {
      setValue(`${name}.template`, "");
      setValue(`${name}.template_toggle`, false);
    }
  }, [templateToggle]);

  return (
    <>
      <FormItem
        name={`${name}.regex`}
        required
        label="RegEx"
        smallLabel
        col_sm={5}
        col_xs={7}
        isRegex
        onRight
      />
      <FormItem
        type="number"
        name={`${name}.index`}
        label="Index"
        smallLabel
        placeholder="0"
        col_sm={2}
        col_xs={2}
        isRegex
        onRight
      />
      <FormCheck
        name={`${name}.template_toggle`}
        size="lg"
        label="T"
        tooltip="Use the RegEx to create a template ($1,$2,etc to reference capture groups)"
        smallLabel
        col_sm={1}
        col_xs={2}
        onRight
      />
      {templateToggle && (
        <FormItem
          name={`${name}.template`}
          label="Template"
          smallLabel
          col_sm={12}
          col_xs={12}
        />
      )}
    </>
  );
};

export default REGEX;
