import { z } from 'zod';

export const serviceOptionsSchema = z.object({
	active: z.boolean().default(true),
	interval: z.string().default(''),
	semantic_versioning: z.boolean().nullable().default(null),
});

export const serviceOptionsSchemaDefaults = z
	.object({
		active: z.boolean().optional(),
		interval: z.string().optional(),
		semantic_versioning: z.boolean().nullable().optional(),
	})
	.optional();
