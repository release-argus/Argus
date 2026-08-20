import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/** Editing leaves the username alone and takes a blank password as 'keep'. */
export const userSchema = z.object({
	display_name: z.string(),
	email: z.string(),
	enabled: z.boolean(),
	groups: z.array(z.string()),
	password: z
		.string()
		.refine(
			(value) => value === '' || value.length >= 8,
			'Must be at least 8 characters',
		),
	username: z.string(),
});

export const userCreateSchema = userSchema.extend({
	password: z.string().min(8, 'Must be at least 8 characters'),
	username: z.string().min(1, REQUIRED_MESSAGE),
});

export type UserFormValues = z.infer<typeof userSchema>;
