import { FormItem } from "components/generic/form";

const REGEX = ({ name }: { name: string }) => (
  <>
    <FormItem
      name={`${name}.regex`}
      label="RegEx"
      smallLabel
      col_sm={8}
      col_xs={8}
      isRegex
      onRight
    />
  </>
);

export default REGEX;
