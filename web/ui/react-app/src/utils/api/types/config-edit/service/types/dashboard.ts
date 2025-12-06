import { z } from 'zod';

export const serviceDashboardOptionsSchema = z
	.object({
		auto_approve: z.boolean().nullable().default(null),
		icon: z.string().default(''),
		icon_link_to: z.string().default(''),
		tags: z.array(z.string()).default([]),
		web_url: z.string().default(''),
	})
	.default({
		auto_approve: null,
		icon: '',
		icon_link_to: '',
		tags: [],
		web_url: '',
	});
export type ServiceDashboardOptionsSchema = z.infer<
	typeof serviceDashboardOptionsSchema
>;
