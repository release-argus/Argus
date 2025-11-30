import { z } from 'zod';
import { preprocessToHeadersArray } from '@/utils/api/types/config-edit/shared/header/preprocess';
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

/* Array of Header objects (min length 1 on key and value) */
export const headersSchema = preprocessToHeadersArray(headerSchema);
/* Array of Header objects (no validation) */
export const headersSchemaDefaults =
	preprocessToHeadersArray(headerSchemaDefaults);
