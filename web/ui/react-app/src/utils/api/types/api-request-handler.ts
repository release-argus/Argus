import { fetchJSON } from '@/utils';
import { API_BASE, RequestMap } from '@/utils/api/types/api-request';
import type { RequestType } from '@/utils/api/types/api-request-types';

/**
 * Makes a request to the API and returns the response.
 *
 * @param requestType - The endpoint to make the request to.
 * @param input - Request params.
 * @returns The response from the API.
 */
export const mapRequest = async <T extends keyof RequestType>(
	requestType: T,
	input: RequestType[T]['input'],
): Promise<RequestType[T]['response']> => {
	// Get method, endpoint, etc from RequestMap
	const props = RequestMap[requestType](input);
	const { method, headers, endpoint, body, timeout, reducer } = props;

	// Add `API_BASE` to the endpoint.
	const url = `${API_BASE}/${endpoint}`;

	const data = await fetchJSON({
		body:
			body ??
			(method.toUpperCase() === 'GET' ? undefined : JSON.stringify(input)),
		headers: headers,
		method: method,
		timeout: timeout,
		url: url,
	});

	// Apply reducer if present.
	return reducer ? reducer(data) : (data as RequestType[T]['response']);
};
