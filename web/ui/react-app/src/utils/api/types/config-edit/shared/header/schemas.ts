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
