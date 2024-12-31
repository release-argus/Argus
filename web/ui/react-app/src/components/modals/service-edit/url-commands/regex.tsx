import { FormCheck, FormItem } from "components/generic/form";
import { useFormContext, useWatch } from "react-hook-form";

import { useEffect } from "react";

/**
 * The form fields for a `RegEx` url_command
 *
 * @param name - The name of the field in the form
 * @returns The form fields for this RegEx url_command
 */
const REGEX = ({ name }: { name: string }) => {
  const { setValue } = useFormContext();

  // Template toggle.
  const templateToggle: boolean | undefined = useWatch({
    name: `${name}.template_toggle`,
  });
  useEffect(() => {
    // Clear the template if the toggle is false.
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
        position="middle"
        positionXS="right"
      />
      <FormItem
        name={`${name}.index`}
        label="Index"
        smallLabel
        isNumber
        placeholder="0"
        col_sm={2}
        col_xs={2}
        isRegex
        position="middle"
        positionXS="left"
      />
      <FormCheck
        name={`${name}.template_toggle`}
        size="lg"
        label="T"
        tooltip="Use the RegEx to create a template"
        smallLabel
        col_sm={1}
        col_xs={2}
        position="right"
        positionXS="middle"
      />
      {templateToggle && (
        <FormItem
          name={`${name}.template`}
          label="Template"
          tooltip="e.g. RegEx of 'v(\d)-(\d)-(\d)' on 'v4-0-1' with template '$1.$2.$3' would give '4.0.1'"
          smallLabel
          col_sm={12}
          col_xs={8}
          position="right"
        />
      )}
    </>
  );
};

export default REGEX;
