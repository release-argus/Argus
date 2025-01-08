import {
	FieldValues,
	UseFormClearErrors,
	UseFormGetValues,
	UseFormSetError,
} from 'react-hook-form';

/**
 * Checks whether the value is a number.
 *
 * @param value - The value to test.
 * @param use - Whether to use this test.
 * @returns - An error message if the value is not a number.
 */
export const numberTest = (value: string, use?: boolean) => {
	if (!value || !use) return true;

	if (isNaN(Number(value))) return 'Must be a number';

	return true;
};

/**
 * Checks whether the value is a number within a range.
 *
 * @param value - The value to test.
 * @param min - The minimum value.
 * @param max - The maximum value.
 * @returns - An error message if the value is not a number within a range.
 */
export const numberRangeTest = (value: string, min: number, max: number) => {
	if (!value) return true;

	if (isNaN(Number(value))) return 'Must be a number';

	if (Number(value) < min || Number(value) > max)
		return `Must be between ${min} and ${max}`;

	return true;
};

/**
 * Checks whether the value is valid RegEx.
 *
 * @param value - The value to test.
 * @param use - Whether to use this test.
 * @returns - An error message if the value is not a valid RegEx.
 */
export const regexTest = (value: string, use?: boolean) => {
	if (!value || !use) return true;

	try {
		new RegExp(value);
	} catch (error) {
		return 'Invalid RegEx';
	}

	return true;
};

/**
 * Checks whether the value follows the Git repository format.
 *
 * @param value - The value to test.
 * @param use - Whether to use this test.
 * @returns - An error message if the value is not a valid Git repository.
 */
export const repoTest = (value: string, use?: boolean) => {
	if (!value || !use) return true;

	if (/^[\w.-]+\/[\w.-]+$/g.test(value)) {
		return true;
	}

	return 'Must be in the format OWNER/REPO';
};

/**
 * Checks whether the value is empty.
 *
 * @param value - The value to test.
 * @param name - The name of the field.
 * @param setError - The function to set an error.
 * @param clearErrors - The function to clear errors.
 * @param use - Whether to use this test.
 * @returns - An error message if the value is required and clears any errors if the value is non-empty.
 */
export const requiredTest = (
	value: string,
	name: string,
	setError: UseFormSetError<FieldValues>,
	clearErrors: UseFormClearErrors<FieldValues>,
	use?: boolean | string,
) => {
	if (!use) return true;

	if (/.+/.test(value)) {
		clearErrors(name);
		return true;
	}
	const msg = use === true ? 'Required' : use;
	setError(name, {
		type: 'required',
		message: msg,
	});

	return msg;
};

/**
 * Checks whether the value is a unique child of the parent.
 *
 * @param value - The value to test.
 * @param name - The name of the field.
 * @param getValues - The function to get the values of the form.
 * @param use - Whether to use this test.
 * @returns - An error message if the value is not a unique child of the parent.
 */
export const uniqueTest = (
	value: string,
	name: string,
	getValues: UseFormGetValues<FieldValues>,
	use?: boolean,
) => {
	if (!value || !use) return true;

	const parts = name.split('.');
	const parent = parts.slice(0, parts.length - 2).join('.');
	const values = getValues(parent);
	const uniqueName = parts[parts.length - 1];
	const unique: boolean =
		values &&
		values
			.map((item: { [x: string]: string }) => item[uniqueName])
			// <=1 in case of default value.
			.filter((item: string) => item === value).length <= 1;

	return unique || 'Must be unique';
};

/**
 * Checks whether the value follows the URL format.
 *
 * @param value - The value to test.
 * @param use - Whether to use this test.
 * @returns An error message if the value is not a valid URL.
 */
export const urlTest = (value?: string, use?: boolean) => {
	if (!value || !use) return true;

	try {
		const parsedURL = new URL(value);
		if (!['http:', 'https:'].includes(parsedURL.protocol))
			throw new Error('Invalid protocol');
	} catch (error) {
		if (/^https?:\/\//.test(value)) {
			return 'Invalid URL';
		}
		return 'Invalid URL - http(s):// prefix required';
	}

	return true;
};
