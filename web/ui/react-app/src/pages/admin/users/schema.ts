import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

export const userSchema = z.object({
	display_name: z.string(),
	email: z.string(),
	enabled: z.boolean(),
	groups: z.array(z.string()),
	password: z.string(),
	username: z.string(),
});

/** Editing leaves the username alone and takes a blank password as 'keep'. */
export const userCreateSchema = userSchema.extend({
	password: z.string().min(1, REQUIRED_MESSAGE),
	username: z.string().min(1, REQUIRED_MESSAGE),
});

export type UserFormValues = z.infer<typeof userSchema>;
