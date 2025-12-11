import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/* Header object (min length 1 on key and value) */
export const headerSchema = z.object({
	key: z.string().min(1, REQUIRED_MESSAGE).default(''),
	old_index: z.number().nullable().default(null),
	value: z.string().min(1, REQUIRED_MESSAGE).default(''),
});
/* Header object (no validation) */
export const headerSchemaDefaults = z.object({
	key: z.string(),
	old_index: z.number().nullable().default(null),
	value: z.string(),
});
