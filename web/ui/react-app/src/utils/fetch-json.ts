import {
	APIError,
	createTimeoutPromise,
	FetchTimeoutError,
	handleResponseError,
} from '@/utils/errors';

type Props = {
	url: string;
	method?: 'DELETE' | 'GET' | 'POST' | 'PUT';
	headers?: Record<string, string>;
	body?: string;
	timeout?: number;
};

/**
 * Fetch JSON from a URL.
 *
 * @param method - The HTTP method to use.
 * @param headers - Optional headers to include in the request.
 * @param url - The URL to fetch data from.
 * @param body - Optional request body, applicable for POST requests.
 * @param timeout - Optional timeout value in milliseconds for the request.
 * @returns The JSON data returned from the request to the server.
 */
const fetchJSON = async <T>({
	method = 'GET',
	headers = { 'Content-Type': 'application/json' },
	url,
	body,
	timeout = 10000,
}: Props): Promise<T> => {
	try {
		const response = await Promise.race([
			fetch(url, { body, headers, method }),
			createTimeoutPromise(timeout),
		]);

		if (!response.ok) await handleResponseError(response);
		const text = await response.text();
		if (!text) return null as T;

		return JSON.parse(text) as T;
	} catch (error) {
		if (error instanceof FetchTimeoutError || error instanceof APIError) {
			throw error;
		}
		console.error('Network Error:', error);
		throw new Error('Failed to fetch data');
	}
};

export default fetchJSON;
