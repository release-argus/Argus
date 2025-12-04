import type { z } from 'zod';
import type { Command, Commands } from '@/utils/api/types/config/shared';
import {
	mapAPIPayloadToCommandSchema,
	mapAPIPayloadToCommandsSchema,
} from '@/utils/api/types/config-edit/command/api/conversions';
import {
	commandDefaultSchema,
	commandSchema,
	commandsDefaultSchema,
} from '@/utils/api/types/config-edit/command/schemas';
import { addZodIssuesToContext } from '@/utils/api/types/config-edit/shared/add-issues';
import { overrideSchemaDefault } from '@/utils/api/types/config-edit/shared/override-schema-default';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import {
	safeParseListOfListWithSchemas,
	safeParseListWithSchemas,
} from '@/utils/api/types/config-edit/validators';

/**
 * Builds a schema for a command.
 *
 * @param data - The current value from the API.
 * @param defaults - The default value.
 * @param hardDefaults - The hard default value.
 */
export const buildCommandSchemaWithFallbacks = (
	data?: Command,
	defaults?: Command,
	hardDefaults?: Command,
): BuilderResponse<typeof commandSchema> & {
	schemaDataDefaultsHollow: z.infer<typeof commandSchema>;
} => {
	const path = 'command';
	const dataConverted = mapAPIPayloadToCommandSchema(data);
	const defaultItem = defaults ?? hardDefaults ?? [];
	const defaultValue = mapAPIPayloadToCommandSchema(defaultItem);
	const defaultValueHollow = defaultItem.map((_) => ({ arg: '' }));

	// Command schema.
	const schema = commandSchema.superRefine((arg, ctx) => {
		if (arg.length === 0) return;

		// Schema validation.
		const { result: schemaResult } = safeParseListWithSchemas({
			arg: arg,
			defaultSchema: commandDefaultSchema,
			defaultValue: defaultValueHollow,
			schema: commandSchema,
		});

		if (!schemaResult.success) {
			addZodIssuesToContext({ ctx, error: schemaResult.error });
		}
	});

	// Initial schema data.
	const schemaData = safeParse({
		data: dataConverted,
		fallback: [],
		path: path,
		schema: schema,
	});

	// Default schema data.
	const schemaDataDefaults = safeParse({
		data: defaultValue,
		fallback: [],
		path: `${path} (defaults)`,
		schema,
	});

	return {
		schema: overrideSchemaDefault(schema, schemaData),
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
		schemaDataDefaultsHollow: defaultValueHollow,
	};
};

/**
 * Builds a schema for a list of commands.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildCommandsSchemaWithFallbacks = (
	data?: Commands,
	defaults?: Commands,
	hardDefaults?: Commands,
) => {
	const path = 'command';
	const dataConverted = mapAPIPayloadToCommandsSchema(data);
	const defaultItem = defaults ?? hardDefaults ?? [];
	const defaultValue = mapAPIPayloadToCommandsSchema(defaultItem);
	const defaultValueHollow = defaultItem.map((command) =>
		command.map(() => ({ arg: '' })),
	);

	// Commands schema.
	const schema = commandsDefaultSchema
		.superRefine((arg, ctx) => {
			if (arg.length === 0) return;

			const { result } = safeParseListOfListWithSchemas({
				arg: arg,
				defaultSchema: commandDefaultSchema,
				defaultValue: defaultValueHollow,
				schema: commandSchema,
			});

			if (!result.success) {
				addZodIssuesToContext({ ctx, error: result.error });
			}
		})
		.default(defaultValueHollow);

	// Initial schema data.
	let schemaData;
	if (data) {
		schemaData = safeParse({
			data: dataConverted,
			fallback: defaultValueHollow,
			path: path,
			schema: schema,
		});
	} else {
		schemaData = defaultValueHollow;
	}

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: defaultValue,
		fallback: [],
		path: `${path} (defaults)`,
		schema: schema,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
		schemaDataDefaultsHollow: defaultValueHollow,
	};
};
