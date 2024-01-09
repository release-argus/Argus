import { FormItem } from "components/generic/form";

const REGEX = ({ name }: { name: string }) => (
  <>
    <FormItem
      key="old"
      name={`${name}.old`}
      label="Replace"
      smallLabel
      required
      col_xs={7}
      col_sm={4}
      onRight
    />
    <FormItem
      key="new"
      name={`${name}.new`}
      label="With"
      smallLabel
      col_xs={12}
      col_sm={4}
      onRight
    />
  </>
);

export default REGEX;
