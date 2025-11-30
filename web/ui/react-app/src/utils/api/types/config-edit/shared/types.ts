import type { ZodType, z } from 'zod';

/* The response from a schema builder function. */
export type BuilderResponse<T extends ZodType, U extends ZodType = T> = {
	/* The schema used for validation. */
	schema: T;
	/* Validated data for the schema. */
	schemaData: z.infer<T>;
	/* Validated defaults for the schema. */
	schemaDataDefaults: z.infer<U>;
};
