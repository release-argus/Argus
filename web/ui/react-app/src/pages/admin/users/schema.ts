import { z } from 'zod';
import { MIN_PASSWORD_LENGTH, PASSWORD_LENGTH_MESSAGE } from '@/types/auth';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/** Editing leaves the username alone and takes a blank password as 'keep'. */
export const userSchema = z.object({
	/** Only collected when setting your own password. */
	confirmPassword: z.string(),
	display_name: z.string(),
	email: z.string(),
	enabled: z.boolean(),
	groups: z.array(z.string()),
	password: z
		.string()
		.refine(
			(value) => value === '' || value.length >= MIN_PASSWORD_LENGTH,
			PASSWORD_LENGTH_MESSAGE,
		),
	username: z.string(),
});

export const userCreateSchema = userSchema.extend({
	password: z.string().min(MIN_PASSWORD_LENGTH, PASSWORD_LENGTH_MESSAGE),
	username: z.string().min(1, REQUIRED_MESSAGE),
});

export type UserFormValues = z.infer<typeof userSchema>;
