import type { FieldError, FieldErrors } from 'react-hook-form';
import type { StringStringMap } from '@/types/util';
import { isEmptyObject } from '@/utils/is-empty';

export class FetchTimeoutError extends Error {
	constructor(timeout: number) {
		super(`Request timed out after ${timeout}ms`);
		this.name = 'FetchTimeoutError';
	}
}

export class APIError extends Error {
	name: string;
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.name = 'APIError';
		this.status = status;
	}
}

export const createTimeoutPromise = (timeout: number): Promise<never> => {
	return new Promise((_, reject) =>
		setTimeout(() => {
			reject(new FetchTimeoutError(timeout));
		}, timeout),
	);
};

export const handleResponseError = async (
	response: Response,
): Promise<never> => {
	const errorData = (await response.json()) as { message: string };
	const error = new APIError(
		errorData.message || `Request failed with status ${response.status}`,
		response.status,
	);
	console.error(`API Error: ${error.message}`);
	throw error;
};

/**
 * Extracts and flattens errors from a react-hook-form errors object
 *
 * e.g.
 *
 * {
 *   first: {
 *     second: [
 *       {item1: {message: "reason"}},
 *       {item2: {message: "otherReason"}}
 *     ]
 *   }
 * }
 *
 * becomes
 *
 * {
 *   first.second.0.item1: "reason",
 *   first.second.1.item2: "otherReason"
 * }.
 *
 * @param errors - The react-hook-form errors object.
 * @param path - The path to limit the errors to.
 * @returns The flattened errors object for the provided path.
 */
export const extractErrors = (
	errors: FieldErrors,
	path = '',
): StringStringMap | undefined => {
	const flatErrors: StringStringMap = {};

	const traverse = (prefix: string, obj: unknown) => {
		if (typeof obj !== 'object' || obj === null) return;

		for (const key of Object.keys(obj)) {
			const value = (obj as Record<string, unknown>)[key];
			if (value == null) continue;

			const currentPath = prefix ? `${prefix}.${key}` : key;
			const atOrBelowPath = currentPath.startsWith(path); // `path` in the key.
			const onPathToTarget = path.startsWith(currentPath); // Building to `path`.
			if (!onPathToTarget && !atOrBelowPath) continue;

			// A leaf has a ref and no children to recurse into.
			const isLeafNode = typeof value === 'object' && 'ref' in value;
			// A branch node has children we need to recurse into.
			const isBranchNode = typeof value === 'object' && !isLeafNode;

			if (isBranchNode) {
				traverse(currentPath, value);
			} else if (isLeafNode && atOrBelowPath) {
				// Only extract errors at or below the target path.
				const trimmedPath = path
					? currentPath.substring(path.length + 1)
					: currentPath;
				// 'message' optional on FieldError.
				const fieldError = value as FieldError;
				if (fieldError.message) flatErrors[trimmedPath] = fieldError.message;
			}
		}
	};

	traverse('', errors);
	return isEmptyObject(flatErrors) ? undefined : flatErrors;
};

export const getErrorMessage = (error: unknown): string => {
	if (error instanceof APIError || error instanceof FetchTimeoutError)
		return error.message;
	if (error instanceof Error) return error.message;

	return 'An unknown error occurred';
};
