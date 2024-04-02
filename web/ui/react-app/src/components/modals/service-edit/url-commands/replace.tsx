import { FormItem } from "components/generic/form";

/**
 * Returns the form fields for the `Replace` url_command
 *
 * @param name - The name of the field in the form
 * @returns The form fields for this Replace url_command
 */
const REPLACE = ({ name }: { name: string }) => (
  <>
    <FormItem
      key="old"
      name={`${name}.old`}
      label="Replace"
      smallLabel
      required
      col_xs={7}
      col_sm={4}
      position="middle"
      positionXS="right"
    />
    <FormItem
      key="new"
      name={`${name}.new`}
      label="With"
      smallLabel
      col_sm={4}
      position="right"
    />
  </>
);

export default REPLACE;
