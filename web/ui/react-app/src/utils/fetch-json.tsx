type Props = {
	// The URL to fetch data from.
	url: string;
	// The HTTP method to use, either GET or POST.
	method?: 'GET' | 'POST' | 'PUT';
	// Optional headers to include in the request.
	headers?: Record<string, string>;
	// Optional request body, applicable for POST requests.
	body?: string;
};

/**
 * Fetch JSON from a URL.
 *
 * @param url - The URL to fetch data from.
 * @param method - The HTTP method to use.
 * @param headers - Optional headers to include in the request.
 * @param body - Optional request body, applicable for POST requests.
 * @returns The JSON data returned from the request to the server.
 */
const fetchJSON = async <T,>({
	url,
	method = 'GET',
	headers = { 'Content-Type': 'application/json' },
	body,
}: Props): Promise<T> => {
	const response = await Promise.race([
		fetch(url, {
			method: method,
			headers: headers,
			body: body,
		}),
		new Promise<Response>((_, reject) =>
			setTimeout(() => reject(new Error('Timeout')), 10000),
		),
	]);
	return await response.json();
};

export default fetchJSON;
