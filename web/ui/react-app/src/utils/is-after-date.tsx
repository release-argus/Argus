// isAfterDate checks whether thenString is after the current Date
const isAfterDate = (thenString: string) => {
  const then = new Date(thenString);
  const now = new Date();
  return then <= now;
};

export default isAfterDate;
