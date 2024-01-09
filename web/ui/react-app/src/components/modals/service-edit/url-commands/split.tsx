import { FormItem } from "components/generic/form";

const REGEX = ({ name }: { name: string }) => (
  <>
    <FormItem
      key="text"
      name={`${name}.text`}
      label="Text"
      smallLabel
      required
      col_xs={5}
      col_sm={6}
      onRight
    />
    <FormItem
      key="index"
      name={`${name}.index`}
      label="Index"
      smallLabel
      required
      col_xs={2}
      col_sm={2}
      type="number"
      onRight
    />
  </>
);

export default REGEX;
