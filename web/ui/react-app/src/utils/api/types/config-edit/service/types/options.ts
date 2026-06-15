import { z } from 'zod';
import { stringDefault } from '@/utils/api/types/config-edit/shared/preprocess.ts';
import { durationString } from '@/utils/api/types/config-edit/validators.ts';

export const serviceOptionsSchema = z.object({
	active: z.boolean().default(true),
	interval: durationString(),
	semantic_versioning: z.boolean().nullable().default(null),
});

export const serviceOptionsSchemaDefaults = z
	.object({
		active: z.boolean().optional(),
		interval: stringDefault,
		semantic_versioning: z.boolean().nullable().optional(),
	})
	.optional();
