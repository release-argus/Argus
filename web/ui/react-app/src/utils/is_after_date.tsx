export const isAfterDate = (thenString: string) => {
  const then = new Date(thenString);
  const now = new Date();
  return then <= now;
};
