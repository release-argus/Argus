import { z } from 'zod';
import { durationString } from '@/utils/api/types/config-edit/validators.ts';

export const serviceOptionsSchema = z.object({
	active: z.boolean().default(true),
	interval: durationString(),
	semantic_versioning: z.boolean().nullable().default(null),
});

export const serviceOptionsSchemaDefaults = z
	.object({
		active: z.boolean().optional(),
		interval: z.string().optional(),
		semantic_versioning: z.boolean().nullable().optional(),
	})
	.optional();
