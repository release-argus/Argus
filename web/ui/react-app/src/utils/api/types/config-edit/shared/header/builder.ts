import { firstNonEmpty } from '@/utils';
import type { Headers } from '@/utils/api/types/config/shared';
import { addZodIssuesToContext } from '@/utils/api/types/config-edit/shared/add-issues';
import {
	headersSchema,
	headersSchemaWithValidation,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import { overrideSchemaDefault } from '@/utils/api/types/config-edit/shared/override-schema-default';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import {
	safeParseListWithSchemas,
	validateListUniqueKeys,
} from '@/utils/api/types/config-edit/validators';

/**
 * Builds a schema for HTTP headers.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 */
export const buildHeadersSchemaWithFallbacks = (
	data?: Headers,
	...defaults: (Headers | undefined)[]
): BuilderResponse<typeof headersSchema> => {
	const path = 'headers';
	const dataDefaulted =
		(data ?? []).map((h, i) => ({ ...h, old_index: i })) ?? [];
	const defaultItem = defaults.find((arr) => arr?.length) ?? [];
	const defaultValue: Headers = defaultItem.map((_) => ({
		key: '',
		old_index: null,
		value: '',
	}));

	// Schema.
	const schemaFinal = overrideSchemaDefault(
		headersSchema,
		defaultValue,
	).superRefine((arg, ctx) => {
		if (arg.length === 0) return;

		// Schema validation.
		const { result: schemaResult } = safeParseListWithSchemas({
			arg: arg,
			defaultSchema: headersSchema,
			defaultValue: defaultValue,
			schema: headersSchemaWithValidation,
		});

		// Unique key validation.
		validateListUniqueKeys({
			arg: arg,
			ctx: ctx,
			path: [],
		});

		if (!schemaResult.success) {
			addZodIssuesToContext({ ctx, error: schemaResult.error });
		}
	});

	// Initial schema data.
	const schemaData = safeParse({
		data: firstNonEmpty(dataDefaulted, defaultValue),
		fallback: [],
		path: path,
		schema: schemaFinal,
	});

	// Defaults for the schema.
	const schemaDataDefaults =
		dataDefaulted.length === 0
			? schemaData
			: safeParse({
					data: defaultValue,
					fallback: [],
					path: `${path} (defaults)`,
					schema: schemaFinal,
				});

	return {
		schema: schemaFinal,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};
