import { z } from 'zod';
import { MIN_PASSWORD_LENGTH, PASSWORD_LENGTH_MESSAGE } from '@/types/auth';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/** A blank password leaves the current one in place. */
export const accountSchema = z.object({
	confirmPassword: z.string(),
	currentPassword: z.string().min(1, REQUIRED_MESSAGE),
	display_name: z.string(),
	email: z.string(),
	password: z
		.string()
		.refine(
			(value) => value === '' || value.length >= MIN_PASSWORD_LENGTH,
			PASSWORD_LENGTH_MESSAGE,
		),
});

export type AccountFormValues = z.infer<typeof accountSchema>;
