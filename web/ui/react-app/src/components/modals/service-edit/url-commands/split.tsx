import { FormItem } from "components/generic/form";

/**
 * Returns the form fields for a `Split` url_command
 *
 * @param name - The name of the field in the form
 * @returns The form fields for this Split url_command
 */
const SPLIT = ({ name }: { name: string }) => (
  <>
    <FormItem
      key="text"
      name={`${name}.text`}
      required
      col_xs={5}
      col_sm={6}
      label="Text"
      smallLabel
      position="middle"
    />
    <FormItem
      key="index"
      name={`${name}.index`}
      required
      col_xs={2}
      col_sm={2}
      label="Index"
      smallLabel
      isNumber
      position="right"
    />
  </>
);

export default SPLIT;
