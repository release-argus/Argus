import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/* Header object (no validation) */
export const headerSchema = z.object({
	key: z.string(),
	old_index: z.number().nullable().default(null),
	value: z.string(),
});
/* Header object (min length 1 on key and value) */
export const headerSchemaWithValidation = headerSchema.extend({
	key: z.string().min(1, REQUIRED_MESSAGE).default(''),
	value: z.string().min(1, REQUIRED_MESSAGE).default(''),
});

/* JSON object entry, keeping the key it loaded with so the API can restore a renamed secret. */
export const mapEntrySchema = headerSchema.extend({
	old_key: z.string().nullable().default(null),
});
/* JSON object entry (min length 1 on key and value) */
export const mapEntrySchemaWithValidation = headerSchemaWithValidation.extend({
	old_key: z.string().nullable().default(null),
});
