import { z } from 'zod';
import { ACTIONS, RESOURCES, SCOPE_TYPES } from '@/types/auth';
import {
	CUSTOM_ISSUE_CODE,
	REQUIRED_MESSAGE,
} from '@/utils/api/types/config-edit/validators';

const grantSchema = z
	.object({
		action: z.enum(ACTIONS),
		resource: z.enum(RESOURCES),
		scope: z.object({
			ref: z.string().optional(),
			type: z.enum(SCOPE_TYPES),
		}),
	})
	.superRefine((arg, ctx) => {
		// Every scope but `global` targets something named.
		if (arg.scope.type !== 'global' && !arg.scope.ref?.trim()) {
			ctx.addIssue({
				code: CUSTOM_ISSUE_CODE,
				message: REQUIRED_MESSAGE,
				path: ['scope', 'ref'],
			});
		}
	});

export const groupSchema = z.object({
	description: z.string(),
	name: z.string().min(1, REQUIRED_MESSAGE),
	permissions: z.array(grantSchema),
});

export type GroupFormValues = z.infer<typeof groupSchema>;
