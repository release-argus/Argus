export const getBasename = () => {
  let basename = window.location.pathname;
  const paths = ["/approvals", "/status", "/flags", "/config"];

  if (basename.endsWith("/")) {
    basename = basename.slice(0, -1);
  }

  if (basename.length > 1) {
    for (let i = 0; i < paths.length; i++) {
      if (basename.endsWith(paths[i])) {
        basename = basename.slice(0, basename.length - paths[i].length);
        return basename;
      }
    }
  }
  return basename;
};
