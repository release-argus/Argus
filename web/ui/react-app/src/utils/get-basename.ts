const APP_PATHS = ['/approvals', '/status', '/flags', '/config'] as const;

/**
 * @returns The path prefix for this app.
 */
const getBasename = () => {
	const normalisedPath = removeTrailingSlash(globalThis.location.pathname);

	if (normalisedPath.length <= 1) return normalisedPath;

	return removeKnownAppPath(normalisedPath) ?? normalisedPath;
};

/**
 * Removes trailing slash from a string if present.
 */
const removeTrailingSlash = (str: string): string => {
	return str.endsWith('/') ? str.slice(0, -1) : str;
};

/**
 * Removes known app path from the end of the given path if present.
 */
const removeKnownAppPath = (path: string): string | null => {
	const matchedPath = APP_PATHS.find((appPath) => path.endsWith(appPath));
	if (matchedPath) return path.slice(0, path.length - matchedPath.length);
	return null;
};

export default getBasename;
