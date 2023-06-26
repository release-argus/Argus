import { FormItem } from "components/generic/form";

const REGEX = ({ name }: { name: string }) => (
  <>
    <FormItem
      name={`${name}.regex`}
      label="RegEx"
      smallLabel
      col_sm={6}
      col_xs={6}
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
  </>
);

export default REGEX;
