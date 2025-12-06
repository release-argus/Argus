import z from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { OPSGENIE_TARGET_TYPE } from '@/utils/api/types/config/notify/opsgenie';
import { makeDefaultsAwareListPreprocessor } from '@/utils/api/types/config-edit/shared/preprocess';

/* Actions */

export const opsGenieActionSchema = z.object({
	arg: z.string(),
});
export const opsGenieActionsSchema = z.preprocess((arg) => {
	if (typeof arg === 'string') {
		try {
			const list = JSON.parse(arg) as unknown;
			if (list && Array.isArray(list)) {
				return list.map((item: unknown) => ({
					arg: typeof item === 'string' ? item : JSON.stringify(item),
				}));
			}
			return [];
		} catch {
			return arg; // zod validation fail
		}
	}
	return arg;
}, z.array(opsGenieActionSchema).default([]));
export type OpsGenieActionsSchema = z.infer<typeof opsGenieActionsSchema>;

/**
 *  Converts the OpsGenie actions from an array of objects to a JSON string.
 *
 *  @param obj - The `NotifyOpsGenieAction[]` to convert.
 *  @returns A JSON string of the actions.
 */
export const preprocessStringFromOpsGenieActions = z.preprocess(
	(val: unknown) => {
		if (!val || !Array.isArray(val) || val.length === 0) return '';

		const formatted = val.map((item: { arg: string }) => item.arg);
		// Using defaults if any action empty.
		if (formatted.includes('')) return '';

		return JSON.stringify(formatted);
	},
	z.string(),
);

/**
 * Defaults-aware variant of OpsGenie actions -> string preprocessor.
 * - Empty array -> null.
 * - Matches defaults -> null.
 *
 * @param defaults - The default values for the actions.
 */
export const preprocessStringFromOpsGenieActionsWithDefaults = (
	defaults?: OpsGenieActionsSchema,
) =>
	makeDefaultsAwareListPreprocessor(
		preprocessStringFromOpsGenieActions.nullable(),
		{
			defaults: defaults,
			matchingFields: [],
		},
	);

/* Target */
const opsGenieTargetTeamSubtypeValues = toZodEnumTuple(
	Object.values(OPSGENIE_TARGET_TYPE.TEAM.SUB_TYPE),
);
const opsGenieTargetTeamSchema = z.object({
	sub_type: z.enum(opsGenieTargetTeamSubtypeValues),
	type: z.literal(OPSGENIE_TARGET_TYPE.TEAM.value),
	value: z.string().default(''),
});

const opsGenieTargetUserSubtypeValues = toZodEnumTuple(
	Object.values(OPSGENIE_TARGET_TYPE.USER.SUB_TYPE),
);
const opsGenieTargetUserSchema = z.object({
	sub_type: z.enum(opsGenieTargetUserSubtypeValues),
	type: z.literal(OPSGENIE_TARGET_TYPE.USER.value),
	value: z.string().default(''),
});
export const opsGenieTargetSchema = z.discriminatedUnion('type', [
	opsGenieTargetTeamSchema,
	opsGenieTargetUserSchema,
]);
export type OpsGenieTargetSchema = z.infer<typeof opsGenieTargetSchema>;
/* Preprocess OpsGenie targets from a string to an array of objects. */
export const opsGenieTargetsSchema = z.preprocess((arg) => {
	if (typeof arg === 'string') {
		try {
			const list = JSON.parse(arg) as unknown;
			if (list && Array.isArray(list)) {
				return list.map((item: object) => {
					const itemType =
						'type' in item ? item.type : OPSGENIE_TARGET_TYPE.USER.value;
					const itemSubType = 'name' in item ? 'name' : 'id';
					let itemValue: unknown = '';
					if ('id' in item) {
						itemValue = item.id;
					} else if ('name' in item) {
						itemValue = item.name;
					} else if ('username' in item) {
						itemValue = item.username;
					}

					return {
						sub_type: itemSubType,
						type: itemType,
						value: itemValue,
					};
				});
			}
			return [];
		} catch {
			return arg; // Zod validation fail.
		}
	}
	return arg;
}, z.array(opsGenieTargetSchema).default([]));
export type OpsGenieTargetsSchema = z.infer<typeof opsGenieTargetsSchema>;

/**
 * Converts the OpsGenie targets from an array of objects to a JSON string.
 *
 * @param obj - The `NotifyOpsGenieTarget[]` to convert.
 * @returns A JSON string of the targets.
 */
export const preprocessOpsGenieTargetsToString = z.preprocess(
	(val: unknown) => {
		if (!val || !Array.isArray(val) || val.length === 0) return '';

		let usingDefaults = false;
		const formatted = val.map((item) => {
			usingDefaults = usingDefaults || item.value === '';
			const { type, sub_type: subType, value } = item;
			return {
				[subType]: value,
				sub_type: subType,
				type: type,
			};
		});
		// Using defaults if any target empty.
		if (usingDefaults) return '';

		return JSON.stringify(formatted);
	},
	z.string(),
);

/**
 * Defaults-aware variant of OpsGenie targets -> string preprocessor.
 * - Empty array -> null.
 * - Matches defaults -> null.
 */
export const preprocessOpsGenieTargetsToStringWithDefaults = (
	defaults?: OpsGenieTargetsSchema,
) =>
	makeDefaultsAwareListPreprocessor(
		preprocessOpsGenieTargetsToString.nullable(),
		{
			defaults: defaults,
			matchingFields: ['type', 'sub_type'],
		},
	);
