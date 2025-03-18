type Props = {
	url: string;
	method?: 'GET' | 'POST' | 'PUT';
	headers?: Record<string, string>;
	body?: string;
	timeout?: number;
};

/**
 * Fetch JSON from a URL.
 *
 * @param url - The URL to fetch data from.
 * @param method - The HTTP method to use.
 * @param headers - Optional headers to include in the request.
 * @param body - Optional request body, applicable for POST requests.
 * @param timeout - Optional timeout value in milliseconds for the request.
 * @returns The JSON data returned from the request to the server.
 */
const fetchJSON = async <T,>({
	url,
	method = 'GET',
	headers = { 'Content-Type': 'application/json' },
	body,
	timeout = 5000,
}: Props): Promise<T> => {
	let loggedError = false;
	try {
		const response = await Promise.race([
			fetch(url, {
				method: method,
				headers: headers,
				body: body,
			}),
			new Promise<Response>((_, reject) =>
				setTimeout(() => reject(new Error('Timeout')), timeout),
			),
		]);

		if (response.ok) return await response.json();

		const errorData = await response.json();
		const error = new Error(errorData.message || 'Request failed');
		console.error(error.message);
		loggedError = true;
		throw error;
	} catch (error) {
		if (!loggedError) console.error(error);
		throw error;
	}
};

export default fetchJSON;
