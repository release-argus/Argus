/**
 * @returns The path prefix this app is served under, read from the `<base
 * href>` the server writes into index.html - stated by the server rather than
 * inferred from the URL. Empty when served from the root.
 */
const getBasename = () => {
	const { pathname } = new URL(document.baseURI);

	return removeTrailingSlash(pathname);
};

/**
 * Removes trailing slash from a string if present.
 */
const removeTrailingSlash = (str: string): string => {
	return str.endsWith('/') ? str.slice(0, -1) : str;
};

export default getBasename;
