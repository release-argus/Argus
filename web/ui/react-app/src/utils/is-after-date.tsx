const dateIsAfterNow = (dateStr: string) => {
  const then = new Date(dateStr);
  const now = new Date();
  return then > now;
};

export default dateIsAfterNow;
