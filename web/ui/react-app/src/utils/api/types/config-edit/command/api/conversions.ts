import type {
	CommandSchema,
	CommandSchemaOutgoing,
	CommandsSchema,
	CommandsSchemaOutgoing,
} from '@/utils/api/types/config-edit/command/schemas';

/**
 * Converts a `CommandsSchema` object to a `CommandsSchemaOutgoing` object.
 *
 * @param data - The `CommandsSchema` data to map.
 * @param defaultValue - The default values to compare against (and omit if all defaults used and unmodified).
 * @returns A `CommandsSchemaOutgoing` representing the `CommandsSchema`.
 */
export const mapCommandsSchemaToAPIPayload = (
	data: CommandsSchema,
	defaultValue?: CommandsSchema,
): CommandsSchemaOutgoing => {
	// Omit if defaults used (lengths match, and all args empty on all commands).
	if (
		defaultValue &&
		data.length == defaultValue.length &&
		data.every((cmd, cmdIndex) => {
			if (defaultValue[cmdIndex].length != cmd.length) return false;
			return Object.values(cmd).every(({ arg }) => !arg);
		})
	) {
		return null;
	}

	// Flatten commands.
	return data.map((cmd) => Object.values(cmd).map(({ arg }) => arg));
};

/**
 * Converts a `CommandSchemaOutgoing` (API Payload) object to a `CommandSchema` object.
 *
 * @param data - The input payload to be mapped to the command schema.
 * @returns The transformed command schema, or null if input is null/undefined/empty.
 */
export const mapAPIPayloadToCommandSchema = (
	data?: CommandSchemaOutgoing | null,
): CommandSchema =>
	data && data.length > 0 ? data.map((arg) => ({ arg })) : [];

/**
 * Converts a `CommandsSchemaOutgoing` (API Payload) object to a `CommandsSchema` object.
 *
 * @param data - The input payload to be mapped to the commands schema.
 * @returns The transformed commands schema, or null if input is null/undefined/empty.
 */
export const mapAPIPayloadToCommandsSchema = (
	data?: CommandsSchemaOutgoing,
): CommandsSchema =>
	data && data.length > 0
		? data.map((cmd) => mapAPIPayloadToCommandSchema(cmd) ?? [])
		: [];
