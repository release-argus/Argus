import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/* Command */
/* Array of arguments for a command (min length 1 on argument text). */
export const commandSchema = z
	.array(
		z.object({
			arg: z.string().min(1, REQUIRED_MESSAGE).default(''),
		}),
	)
	.default([]);
export type CommandSchema = z.infer<typeof commandSchema>;
/* Array of arguments for a command (no validation). */
export const commandDefaultSchema = z.array(
	z.object({
		arg: z.string().default(''),
	}),
);

/* Commands */
/* Array of commands (min length 1 on argument text). */
export const commandsSchema = z.array(commandSchema).default([]);
export type CommandsSchema = z.infer<typeof commandsSchema>;
/* Array of commands (no validation). */
export const commandsDefaultSchema = z.array(commandDefaultSchema).default([]);

/* API Outgoing requests */

export const commandSchemaOutgoing = z.array(z.string());
export type CommandSchemaOutgoing = z.infer<typeof commandSchemaOutgoing>;
export const commandsSchemaOutgoing = z
	.array(commandSchemaOutgoing)
	.nullable()
	.default(null);
export type CommandsSchemaOutgoing = z.infer<typeof commandsSchemaOutgoing>;
