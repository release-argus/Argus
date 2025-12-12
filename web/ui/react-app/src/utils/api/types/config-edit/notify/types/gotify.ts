import z from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	GOTIFY_EXTRA_NAMESPACE,
	gotifyExtraClientDisplayContentTypeOptions,
} from '@/utils/api/types/config/notify/gotify.ts';
import { makeDefaultsAwareListPreprocessor } from '@/utils/api/types/config-edit/shared/preprocess.ts';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators.ts';

/* Extras */
// client::display - contentType.
const gotifyExtraClientDisplayContentTypeEnum = z.enum(
	toZodEnumTuple(gotifyExtraClientDisplayContentTypeOptions),
);

// namespace: android::action.
const gotifyExtraAndroidActionSchema = z.object({
	namespace: z.literal(GOTIFY_EXTRA_NAMESPACE.ANDROID_ACTION.value),
	onReceive: z
		.object({
			intentUrl: z.string().default(''),
		})
		.default({ intentUrl: '' }),
});
const gotifyExtraAndroidActionSchemaWithValidation =
	gotifyExtraAndroidActionSchema.extend({
		onReceive: z.object({
			intentUrl: z.string().min(1, REQUIRED_MESSAGE).default(''),
		}),
	});

// namespace: client::display.
const gotifyExtraClientDisplaySchema = z.object({
	contentType: gotifyExtraClientDisplayContentTypeEnum,
	namespace: z.literal(GOTIFY_EXTRA_NAMESPACE.CLIENT_DISPLAY.value),
});

// namespace: client::notification.
const gotifyExtraClientNotificationSchema = z.object({
	bigImageUrl: z.string().default(''),
	click: z
		.object({
			url: z.string().default(''),
		})
		.default({ url: '' }),
	namespace: z.literal(GOTIFY_EXTRA_NAMESPACE.CLIENT_NOTIFICATION.value),
});
const gotifyExtraClientNotificationSchemaWithValidation =
	gotifyExtraClientNotificationSchema.extend({
		bigImageUrl: z.string().min(1, REQUIRED_MESSAGE).default(''),
		click: z
			.object({
				url: z.string().min(1, REQUIRED_MESSAGE).default(''),
			})
			.default({ url: '' }),
	});

// namespace: other.
const gotifyExtraOtherSchema = z.object({
	_namespace: z.string().default(''),
	namespace: z.literal(GOTIFY_EXTRA_NAMESPACE.OTHER.value),
	value: z.string().default(''),
});
const gotifyExtraOtherSchemaWithValidation = gotifyExtraOtherSchema.extend({
	_namespace: z.string().min(1, REQUIRED_MESSAGE).default(''),
	value: z.string().min(1, REQUIRED_MESSAGE).default(''),
});

const gotifyExtraSchema = z.discriminatedUnion('namespace', [
	gotifyExtraAndroidActionSchema,
	gotifyExtraClientDisplaySchema,
	gotifyExtraClientNotificationSchema,
	gotifyExtraOtherSchema,
]);
export type GotifyExtraSchema = z.infer<typeof gotifyExtraSchema>;
const gotifyExtraSchemaWithValidation = z.discriminatedUnion('namespace', [
	gotifyExtraAndroidActionSchemaWithValidation,
	gotifyExtraClientDisplaySchema,
	gotifyExtraClientNotificationSchemaWithValidation,
	gotifyExtraOtherSchemaWithValidation,
]);

export const knownExtraNamespaces = new Set<string>([
	GOTIFY_EXTRA_NAMESPACE.ANDROID_ACTION.value,
	GOTIFY_EXTRA_NAMESPACE.CLIENT_DISPLAY.value,
	GOTIFY_EXTRA_NAMESPACE.CLIENT_NOTIFICATION.value,
]);

// Preprocess Gotify extras from a string to the array of objects schema.
const preprocessGotifyExtras = (arg: unknown) => {
	if (typeof arg === 'string') {
		try {
			const obj = JSON.parse(arg) as unknown;
			if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
				return Object.entries(obj as Record<string, unknown>).map(
					([namespace, value]) => {
						if (!knownExtraNamespaces.has(namespace)) {
							// Map other namespaces to 'OTHER' and preserve the custom namespace key.
							return {
								_namespace: namespace,
								namespace: GOTIFY_EXTRA_NAMESPACE.OTHER.value,
								value: value ?? '',
							};
						}

						// Spread known namespaces, preserving the namespace key.
						if (value && typeof value === 'object' && !Array.isArray(value)) {
							return { namespace: namespace, ...value };
						}

						// Fallback.
						return { namespace: namespace };
					},
				);
			}
			return [];
		} catch {
			return arg; // zod validation fail
		}
	}
	return arg;
};

export const gotifyExtrasSchema = z.preprocess(
	preprocessGotifyExtras,
	z.array(gotifyExtraSchema).default([]),
);

export const gotifyExtrasSchemaWithValidation = z.preprocess(
	preprocessGotifyExtras,
	z.array(gotifyExtraSchemaWithValidation).default([]),
);
export type GotifyExtrasSchema = z.infer<
	typeof gotifyExtrasSchemaWithValidation
>;

/**
 * Converts the Gotify extras from an array of objects to a JSON string.
 *
 * @param obj - The `GotifyExtrasSchema` to convert.
 * @returns A JSON string of the extras.
 */
export const preprocessGotifyExtrasToString = z.preprocess(
	(val: GotifyExtrasSchema) => {
		if (!val || !Array.isArray(val) || val.length === 0) return '';

		const result: Record<string, unknown> = {};
		for (const item of val) {
			if (item.namespace === GOTIFY_EXTRA_NAMESPACE.OTHER.value) {
				result[item._namespace] = item.value;
			} else {
				// Flatten the item and remove 'namespace' from the object.
				result[item.namespace] = Object.fromEntries(
					Object.entries(item).filter(
						([key]) => key !== 'namespace' && key !== '_namespace',
					),
				);
			}
		}

		return Object.keys(result).length === 0 ? '' : JSON.stringify(result);
	},
	z.string(),
);

/**
 * Defaults-aware variant of Gotify extras -> string preprocessor.
 * - Empty array -> null.
 * - Matches defaults -> null.
 */
export const preprocessGotifyExtrasToStringWithDefaults = (
	defaults?: GotifyExtrasSchema,
) =>
	makeDefaultsAwareListPreprocessor(preprocessGotifyExtrasToString.nullable(), {
		defaults: defaults,
		matchingFields: ['namespace', 'contentType'],
	});
