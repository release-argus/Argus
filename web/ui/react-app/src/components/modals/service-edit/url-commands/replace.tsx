import { FormItem } from "components/generic/form";

const REGEX = ({ name }: { name: string }) => (
  <>
    <FormItem
      key="old"
      name={`${name}.old`}
      label="Replace"
      smallLabel
      required
      col_xs={4}
      col_sm={4}
      onRight
    />
    <FormItem
      key="new"
      name={`${name}.new`}
      label="With"
      smallLabel
      required
      col_xs={4}
      col_sm={4}
      onRight
    />
  </>
);

export default REGEX;
