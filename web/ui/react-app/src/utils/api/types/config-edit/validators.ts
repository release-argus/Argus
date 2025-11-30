import {
	type RefinementCtx,
	type ZodDefault,
	type ZodSafeParseResult,
	type ZodString,
	type ZodType,
	z,
} from 'zod';
import { isEmptyOrNull } from '@/utils';

/* Field validation */

export const CUSTOM_ISSUE_CODE = 'custom';
export const REQUIRED_MESSAGE = 'Required.';
export const NUMBER_REQUIRED_MESSAGE = 'Number required.';
export const INVALID_GITHUB_REPO_MESSAGE = 'Invalid GitHub repository.';
export const INVALID_URL_MESSAGE =
	"Invalid URL (Must start with 'http://' or 'https://').";
export const UNIQUE_MESSAGE = 'Must be unique.';

const GITHUB_REPO_REGEX = /^[a-zA-Z0-9-_.]+\/[a-zA-Z0-9-_.]+$/;
/**
 * Validates that the input is a valid GitHub repository.
 *
 * @param arg - The input to validate.
 * @param ctx - The Zod refinement context.
 * @param path - The path to the input in the object.
 */
export const validateGitHubRepo = ({ arg, ctx, path }: FieldValidatorProps) => {
	if (!GITHUB_REPO_REGEX.test(arg as string)) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: INVALID_GITHUB_REPO_MESSAGE,
			path: path,
		});
	}
};

export const SUPPORTED_URL_PROTOCOLS = ['http:', 'https:'] as const;
type SupportedURLProtocol = (typeof SUPPORTED_URL_PROTOCOLS)[number];

/**
 * Checks whether the input is a valid URL.
 *
 * @param arg - The input to validate.
 * @returns `true` if the input is a valid URL, `false` otherwise.
 */
export const isValidURL = (arg: unknown) => {
	if (typeof arg !== 'string') return false;

	// Must start with 'http(s)://'.
	if (!/^https?:\/\//.test(arg)) return false;

	try {
		const url = new URL(arg);
		return SUPPORTED_URL_PROTOCOLS.includes(
			url.protocol as SupportedURLProtocol,
		);
	} catch {
		return false;
	}
};

/**
 * Validates that the input is a valid URL.
 *
 * @param arg - The input to validate.
 * @param ctx - The Zod refinement context.
 * @param path - The path to the input in the object.
 */
export const validateURL = ({ arg, ctx, path }: FieldValidatorProps) => {
	if (!isValidURL(arg)) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: INVALID_URL_MESSAGE,
			path,
		});
	}
};

/* RegEx for a hex string representing a colour */
const hexStringRegEx = new RegExp(/^[\da-f]{6}$|^$/i);

/**
 * Validates that the input is a valid hexadecimal string.
 *
 * @param arg - The input to validate.
 * @param ctx - The Zod refinement context.
 * @param path - The path to the input in the object.
 */
export const validateHexString = ({ arg, ctx, path }: FieldValidatorProps) => {
	if (arg && typeof arg === 'string' && !hexStringRegEx.exec(arg)) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: 'Invalid hexadecimal. Must be 6 characters, 0-9 and A-F.',
			path: path,
		});
	}
};

/**
 * Validates that the input is a number.
 *
 * @param arg - The input to validate.
 * @param ctx - The Zod refinement context.
 * @param path - The path to the input in the object.
 */
export const validateNumberString = ({
	arg,
	ctx,
	path,
}: FieldValidatorProps) => {
	if (arg && Number.isNaN(Number.parseFloat(arg as string))) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: NUMBER_REQUIRED_MESSAGE,
			path: path,
		});
	}
};

/**
 * Creates a field validator that ensures a numeric value is within a specified range.
 *
 * @param min - Minimum allowed value (inclusive).
 * @param max - Maximum allowed value (inclusive).
 * @returns A validator function that checks if `arg` is a number within [min, max]
 *          and adds a custom Zod issue if out of range.
 */
export const validateNumberInRange =
	({ min, max }: { min: number; max: number }) =>
	({ arg, ctx, path }: FieldValidatorProps) => {
		if (
			(arg &&
				!Number.isNaN(Number.parseFloat(arg as string)) &&
				Number(arg) < min) ||
			Number(arg) > max
		) {
			ctx.addIssue({
				code: CUSTOM_ISSUE_CODE,
				message: `Must be between ${min} and ${max}.`,
				path: path,
			});
		}
	};

/**
 * Validates that the input is not empty.
 *
 * @param arg - The input to validate.
 * @param ctx - The Zod refinement context.
 * @param path - The path to the input in the object.
 */
export const validateRequired = ({ arg, ctx, path }: FieldValidatorProps) => {
	if ((arg ?? '') === '') {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: REQUIRED_MESSAGE,
			path: path,
		});
	}
};

/**
 * Recursively checks if an object or array contains any non-empty string values,
 * ignoring specified keys.
 *
 * @param obj - The object to check.
 * @param ignoredKeys - Optional list of keys to ignore.
 * @returns `true` if the object has at least one non-empty string value-field, `false` otherwise.
 */
const hasNonEmptyField = (
	obj: unknown,
	ignoredKeys: string[] = [],
): boolean => {
	if (obj == null) return false;

	if (typeof obj === 'string') return obj.trim() !== '';
	if (typeof obj !== 'object') return false;

	if (Array.isArray(obj)) {
		return obj.some((v) => hasNonEmptyField(v, ignoredKeys));
	}

	for (const key of Object.keys(obj)) {
		if (ignoredKeys.includes(key)) continue;
		if (hasNonEmptyField((obj as Record<string, unknown>)[key], ignoredKeys)) {
			return true;
		}
	}
	return false;
};

/**
 * Creates a Zod string schema that validates a value as a regular expression,
 * using optional fallback defaults and enforcing requirement if specified.
 *
 * @param required - Whether a value must be provided.
 * @param defaults - Optional fallback strings to use if the field value is empty.
 * @returns A Zod string schema that checks for a valid RegExp and adds custom issues on failure.
 */
export const regexStringWithFallback = (
	required: boolean,
	...defaults: (string | undefined)[]
) =>
	z
		.string()
		.default('')
		.superRefine((arg, ctx) => {
			const value = arg || defaults.find((d) => d?.trim());
			// No value.
			if (!value) {
				if (required) {
					ctx.addIssue({
						code: CUSTOM_ISSUE_CODE,
						message: REQUIRED_MESSAGE,
					});
				}
				return;
			}

			// RegEx validation.
			try {
				new RegExp(value);
			} catch {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: 'Invalid regular expression.',
				});
			}
		});

/* RegEx validation for a field */
type RegexValidation<T extends boolean = true> = T extends true
	? { regex: RegExp; message: string }
	: { regex: undefined; message: undefined };

/**
 * Creates a Zod string schema that validates a value as a string,
 * optionally using fallback defaults and enforcing requirement if specified.
 *
 * @param validation - Optional validation rehular expression.
 * @param required - Whether a value must be provided.
 * @param defaults - Optional fallback strings to use if the field value is empty.
 */
export const stringWithFallback = (
	validation?: RegexValidation,
	required = true,
	...defaults: (string | undefined)[]
): ZodDefault<ZodString> =>
	z
		.string()
		.default('')
		.superRefine((arg, ctx) => {
			const value = arg || defaults.find((d) => d?.trim());

			if (!value) {
				if (required) {
					ctx.addIssue({
						code: CUSTOM_ISSUE_CODE,
						message: REQUIRED_MESSAGE,
					});
				}
				return;
			}

			if (validation?.regex && !validation.regex.test(value)) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: validation.message,
				});
			}
		});
/* URL prefix validator. */
export const urlPrefixValidator: RegexValidation = {
	message: INVALID_URL_MESSAGE,
	regex: /^https?:\/\//,
};

type IsUsingDefaultsParams<T> = {
	/* The input data */
	arg: T;
	/* Default value for this field */
	defaultValue: T;
	/* Fields starting with these values that must much the defaults. */
	matchingFieldsStartsWiths?: string[];
	/* Fields ending with these values that must much the defaults. */
	matchingFieldsEndsWiths?: string[];
	/* The key path of the current `fieldValues`. */
	key?: string;
};

/**
 * Checks if an array of items is using the default values.
 *
 * Compares each item in `arg` against the corresponding item in `defaultValue`.
 * If any item has a non-empty field or differs in any of the `matchingFieldsStartsWiths`/`matchingFieldsEndsWiths`,
 * this returns false. Otherwise, returns true.
 *
 * @param arg - The array of items to check.
 * @param defaultValue - The array of default values to compare against.
 * @param matchingFieldsStartsWiths - Optional list of fields starting with these values that must much the defaults.
 * @param matchingFieldsEndsWiths - Optional list of fields ending with these values that must much the defaults.
 * @param key - The key path of the current `fieldValues`.
 * @returns `true` if `arg` matches `defaultValue` on the specified fields and
 *          all other items are empty; otherwise `false`.
 */
export const isUsingDefaults = <T>({
	arg,
	defaultValue,
	matchingFieldsStartsWiths,
	matchingFieldsEndsWiths,
	key,
}: IsUsingDefaultsParams<T>): boolean => {
	// No value.
	if (isEmptyOrNull(arg))
		return key == null ? !isEmptyOrNull(defaultValue) : true;
	// No defaults.
	if (isEmptyOrNull(defaultValue)) return false;

	// Arrays.
	if (Array.isArray(arg)) {
		// Type mismatch.
		if (!Array.isArray(defaultValue)) return false;
		// Length mismatch.
		if (arg.length !== defaultValue.length) {
			// Give defaults when empty.
			if (arg.length === 0) return true;

			// Allow length mismatch on 'id' fields.
			const argHasID =
				arg.length > 0 &&
				typeof arg[0] === 'object' &&
				arg[0] != null &&
				'id' in arg[0];
			const defaultsHasID =
				defaultValue.length > 0 &&
				typeof defaultValue[0] === 'object' &&
				defaultValue[0] != null &&
				'id' in defaultValue[0];
			if (
				Math.abs(arg.length - defaultValue.length) !==
				(argHasID === defaultsHasID ? 0 : 1)
			)
				return false;
		}
		// Compare each element recursively.
		return arg.every((val, i) =>
			isUsingDefaults({
				arg: val,
				defaultValue: defaultValue[i],
				key: key ? `${key}.${i}` : String(i),
				matchingFieldsEndsWiths: matchingFieldsEndsWiths,
				matchingFieldsStartsWiths: matchingFieldsStartsWiths,
			}),
		);
	}

	// Objects.
	if (typeof arg === 'object' && typeof defaultValue === 'object') {
		const keys = Array.from(
			new Set([
				...Object.keys(arg as Record<string, unknown>),
				...Object.keys(defaultValue as Record<string, unknown>),
			]),
		);
		// Recursively check each key in the object.
		return keys.every((k) =>
			isUsingDefaults({
				arg: (arg as Record<string, unknown>)[k],
				defaultValue: (defaultValue as Record<string, unknown>)[k],
				key: key ? `${key}.${k}` : k,
				matchingFieldsEndsWiths: matchingFieldsEndsWiths,
				matchingFieldsStartsWiths: matchingFieldsStartsWiths,
			}),
		);
	}

	// Booleans.
	if (typeof arg === 'boolean' || typeof defaultValue === 'boolean')
		return !!arg === !!defaultValue;

	// Strings, numbers, etc.
	if (
		key &&
		(matchingFieldsEndsWiths?.some((k) => key.endsWith(k)) ||
			matchingFieldsStartsWiths?.some((k) => key.startsWith(k)))
	) {
		return arg === defaultValue;
	}

	// Non-matching fields: empty/undefined values considered 'defaults'.
	if (isEmptyOrNull(arg)) return true;

	// Fallback: values must match exactly.
	return arg === defaultValue;
};

/**
 * Type guard that checks if a value is an array of objects.
 *
 * @param val - The value to check.
 * @returns `true` if `val` is an array where every item is a non-null object (not an array); otherwise `false`.
 */
const isRecordArray = (val: unknown): val is Record<string, unknown>[] =>
	Array.isArray(val) &&
	val.every(
		(item) => typeof item === 'object' && item !== null && !Array.isArray(item),
	);

/**
 * Recursively adds REQUIRED issues for any empty string fields.
 *
 * Supports the following shapes:
 * - A single record (object).
 * - A list of records.
 * - Nested lists (e.g. list of lists of records).
 *
 * @param arg - The object or array to check.
 * @param ctx - The Zod refinement context.
 * @param path - Optional path in the object for error reporting.
 * @param notRequired - Optional list of fields that may be empty.
 */
export const addIssuesForEmptyStringFields = ({
	arg,
	ctx,
	path,
	notRequired,
}: {
	arg: unknown;
	ctx: RefinementCtx;
	path?: (string | number)[];
	notRequired?: string[];
}): void => {
	const withPath = (key: string | number) => (path ? [...path, key] : [key]);

	// Arrays.
	if (Array.isArray(arg)) {
		for (let i = 0; i < arg.length; i++) {
			addIssuesForEmptyStringFields({
				arg: arg[i],
				ctx: ctx,
				notRequired: notRequired,
				path: withPath(i),
			});
		}
		return;
	}

	// Objects.
	if (arg && typeof arg === 'object') {
		for (const [key, value] of Object.entries(arg)) {
			const nextPath = withPath(key);

			if (typeof value === 'string') {
				if (value.trim() === '' && !notRequired?.includes(key)) {
					ctx.addIssue({
						code: CUSTOM_ISSUE_CODE,
						message: REQUIRED_MESSAGE,
						path: nextPath,
					});
				}
			} else if (value && typeof value === 'object') {
				addIssuesForEmptyStringFields({
					arg: value,
					ctx: ctx,
					path: nextPath,
				});
			}
		}
	}
};

type ValidateListWithSchemasProps = {
	/* Fields that must match between `arg` and `defaultValue` for `defaultSchema` to be applicable. */
	matchingFields?: string[];
	/* Fields that may be empty for validation. */
	notRequired?: string[];
};

/**
 * Validates an array of objects against provided schemas and default values.
 *
 * Skips validation if the array exactly matches the default values and has no non-empty fields
 * outside the specified `matchingFields`, which must match.
 *
 * @param arg - The array of objects to validate.
 * @param ctx - Zod validation context for adding issues.
 * @param path - Path in the object for error reporting.
 * @param defaultValue - Array of default objects to compare against.
 * @param matchingFields - Keys to ignore when checking against defaults.
 */
export const validateListWithSchemas = ({
	arg,
	ctx,
	path,
	defaultValue,
	props,
}: ArrayFieldValidatorProps<ValidateListWithSchemasProps>): void => {
	if (!isRecordArray(arg)) return;
	const matchingFields = props?.matchingFields;
	const notRequired = props?.notRequired;

	// Defaults used if the length matches, and fields in matchingFields match, and all other fields empty.
	const usingDefaults = isUsingDefaults({
		arg: arg,
		defaultValue: defaultValue,
		matchingFieldsStartsWiths: matchingFields,
	});
	if (usingDefaults) return;

	// Otherwise, add issues for any empty string fields in the list of records.
	addIssuesForEmptyStringFields({ arg, ctx, notRequired, path });
};

/**
 * Validates that a list of objects has unique keys.
 *
 * @param arg - The list of objects to validate.
 * @param ctx - The Zod validation context.
 * @param path - The path to the list in the object.
 */
export const validateListUniqueKeys = ({
	arg,
	ctx,
	path,
}: FieldValidatorProps): void => {
	if (!Array.isArray(arg)) return;

	// Map of value -> indexes.
	const seen = new Map<unknown, number>();

	for (let index = 0; index < arg.length; index++) {
		const item = arg[index] as Record<string, unknown>;
		const value = item?.key;

		if (value == null || value === '') continue;

		const firstIndex = seen.get(value);
		if (firstIndex === undefined) {
			// Record first occurrence.
			seen.set(value, index);
		} else {
			// Add an issue for this duplicate name, and the first occurrence.
			for (const i of [firstIndex, index]) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: UNIQUE_MESSAGE,
					path: [...(path ?? []), i, 'key'],
				});
			}
		}
	}
};

type SafeParseListWithSchemasParams<
	T extends Record<string, unknown>[] | Record<string, unknown>[][],
> = {
	/* The input data */
	arg: T;
	/* Default value for this field */
	defaultValue: T;
	/* Schema to validate with */
	schema: ZodType<Record<string, unknown>[]>;
	/* Schema to use if `arg` is a hollow variant of `defaultValue`. */
	defaultSchema: ZodType<Record<string, unknown>[]>;
	/* Fields that must match between `arg` and `defaultValue` for `defaultSchema` to be applicable. */
	matchingFields?: string[];
};

type SafeParseWithDefaultsResult<
	T extends Record<string, unknown>[] | Record<string, unknown>[][],
> = {
	/* Result of safeParse. */
	result: ZodSafeParseResult<T>;
	/* Whether the result is using defaults. */
	usingDefaults?: boolean;
};

/**
 * Parses a list of records using a schema.
 * - If `arg` is a hollow variant of `defaultValue` (where matchingFields match), uses `defaultSchema`.
 * - Otherwise, uses `schema`.
 *
 * @param arg - The input data: list of records
 * @param defaultValue - Default value for this field.
 * @param schema - Schema to validate with.
 * @param defaultSchema - Schema to use if `arg` is a hollow variant of `defaultValue`.
 * @param matchingFields - Fields that must match between `arg` and `defaultValue` for `defaultSchema` to be applicable.
 */
export const safeParseListWithSchemas = ({
	arg,
	defaultValue,
	schema,
	defaultSchema,
	matchingFields = [],
}: SafeParseListWithSchemasParams<
	Record<string, unknown>[]
>): SafeParseWithDefaultsResult<Record<string, unknown>[]> => {
	const usingDefaults =
		defaultValue.length > 0 &&
		defaultValue.length === arg.length &&
		!arg.some(
			(item, i) =>
				hasNonEmptyField(item, matchingFields) ||
				matchingFields.some((key) => item[key] !== defaultValue[i][key]),
		);

	const result = (usingDefaults ? defaultSchema : schema).safeParse(arg);

	return { result, usingDefaults };
};

/**
 * Parses a list of list of records using a schema.
 * - If `arg` is a hollow variant of `defaultValue`, uses `defaultSchema`.
 * - Otherwise, uses `schema`.
 *
 * @param arg - The input data: list of list of records
 * @param defaultValue - Default value for this field.
 * @param schema - Schema to validate each element with.
 * @param defaultSchema - Schema to use if `arg` is a hollow variant of `defaultValue`.
 * @param matchingFields - Fields that must match between `arg` and `defaultValue` for `defaultSchema` to be applicable.
 */
export const safeParseListOfListWithSchemas = ({
	arg,
	defaultValue,
	schema,
	defaultSchema,
	matchingFields = [],
}: SafeParseListWithSchemasParams<
	Record<string, unknown>[][]
>): SafeParseWithDefaultsResult<Record<string, unknown>[][]> => {
	// Cannot use defaults if lists differ in length.
	const couldBeDefaults = arg.length === defaultValue.length;

	const results = arg.map((subList, i) =>
		safeParseListWithSchemas({
			arg: subList,
			defaultSchema: defaultSchema,
			defaultValue: couldBeDefaults ? (defaultValue[i] ?? []) : [],
			matchingFields: matchingFields,
			schema: schema,
		}),
	);

	// Using defaults if all sub-lists use defaults.
	const usingDefaults = results.every((r) => r.usingDefaults);

	// Group success/issues.
	const allOk = results.every((r) => r.result.success);
	const issues = allOk
		? []
		: results.flatMap((r, i) =>
				r.result.success
					? []
					: r.result.error.issues.map((issue) => ({
							...issue,
							path: [i, ...issue.path],
						})),
			);
	const error = new z.ZodError(issues);

	const data: Record<string, unknown>[][] = results.map(
		(r) => (r.result as z.ZodSafeParseSuccess<Record<string, unknown>[]>).data,
	);

	return {
		result: allOk
			? {
					data: data,
					success: true,
				}
			: {
					error: error as z.ZodError<Record<string, unknown>[][]>,
					success: false,
				},
		usingDefaults: usingDefaults,
	};
};

/**
 * Retrieves a nested value from an object using a path of keys.
 *
 * @param obj - The object to traverse.
 * @param path - An array of keys representing the path to the nested value.
 * @returns The value at the nested path, or `null` if any key is missing.
 */
export const getNested = (
	obj: object | null,
	path: readonly string[],
): unknown => {
	let current: unknown = obj;
	for (const key of path) {
		if (current && typeof current === 'object' && key in current) {
			current = (current as Record<string, unknown>)[key];
		} else {
			return null;
		}
	}
	return current;
};

type FieldValidatorProps = {
	/* Value to validate. */
	arg: unknown;
	/* Zod refinement context. */
	ctx: RefinementCtx;
	/* Path to the value in the object. */
	path?: string[];
};
export type ArrayFieldValidatorProps<P = unknown> = FieldValidatorProps & {
	/* Default value to use if value empty. */
	defaultValue: Record<string, unknown>[];
	/* Extra props passed into validators. */
	props?: P;
};

/**
 * Validation rule for a field or array field.
 * - For single fields: specifies the path, and a validator function.
 * - For array fields: specifies the path, kind as "array", validator functions, and extra props for the validator functions.
 */
export type FieldValidation<P = unknown> =
	| {
			/* Validation for a single value. */
			kind?: 'field';
			/* Schema path. */
			path: string[];

			/* Validator function. */
			validator: (props: FieldValidatorProps) => void;
	  }
	| {
			/* Validation for an array. */
			kind: 'array';
			/* Schema path. */
			path: string[];
			/* Extra props passed into validators. */
			props?: P[];
			/* Validator functions. */
			validator: {
				[K in keyof P]: (props: ArrayFieldValidatorProps<P[K]>) => void;
			};
	  };

/**
 * Validates fields of an object using provided validation rules.
 *
 * @template T - The type of the object to validate (must have a `name` property).
 * @param arg - The object to validate.
 * @param mainValues - Optional main values to use as fallback for arg.
 * @param defaults - Default values to use as fallback for mainValues.
 * @param fields - Array of field validation rules.
 * @param ctx - Zod refinement context for reporting validation issues.
 */
export const validateFields = <T extends { name: string }>(
	arg: T,
	mainValues: T | null,
	defaults: T,
	fields: FieldValidation[],
	ctx: z.RefinementCtx,
) => {
	for (const field of fields) {
		const value = getNested(arg, field.path);
		const fallback =
			getNested(mainValues, field.path) || getNested(defaults, field.path);

		const path = [...field.path];
		if (field.kind === 'array') {
			const fallbackArray = isRecordArray(fallback) ? fallback : [];
			const validators = Array.isArray(field.validator)
				? field.validator
				: Object.values(field.validator);

			for (const validator of validators) {
				const index = validators.indexOf(validator);
				const propValue =
					Array.isArray(field.props) && index < field.props.length
						? field.props[index]
						: null;

				validator({
					arg: value,
					ctx: ctx,
					defaultValue: fallbackArray,
					path: path,
					props: propValue,
				});
			}
		} else {
			// Kind: field.
			field.validator({
				arg: value || fallback,
				ctx: ctx,
				path: path,
			});
		}
	}
};

/**
 * Validates that the `type` property of `arg` matches the `type` property of `mainValues`.
 * If they do not match, adds a custom issue to the Zod refinement context.
 *
 * @template T - Object type with `name` and `type` properties.
 * @param arg - The object to validate.
 * @param mainValues - The main values object to compare against.
 * @param ctx - Zod refinement context for reporting validation issues.
 */
export const validateMainTypeMatch = <T extends { name: string; type: string }>(
	arg: T,
	mainValues: T | null,
	ctx: z.RefinementCtx,
) => {
	if (mainValues?.type && arg.type !== mainValues.type) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message:
				`${arg.type} does not match the global for "${arg.name}" of ${mainValues.type}. ` +
				`Change the type to match that, or choose a new name`,
			path: ['type'],
		});
	}
};
