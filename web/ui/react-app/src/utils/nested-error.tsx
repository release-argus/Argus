// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getNestedError = (errors: any, key: string) =>
  key
    .split(".")
    .reduce((acc, key) => (acc && acc[key] ? acc[key] : undefined), errors);

export default getNestedError;
