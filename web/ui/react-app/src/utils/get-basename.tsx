/**
 * @returns The path prefix for this app.
 */
const getBasename = () => {
	let basename = window.location.pathname;
	const paths = ['/approvals', '/status', '/flags', '/config'];

	if (basename.endsWith('/')) basename = basename.slice(0, -1);

	if (basename.length > 1)
		for (const path of paths) {
			if (basename.endsWith(path)) {
				basename = basename.slice(0, basename.length - path.length);
				return basename;
			}
		}
	return basename;
};

export default getBasename;
