import { z } from 'zod';
import { stringDefault } from '@/utils/api/types/config-edit/shared/preprocess';

export const serviceDashboardOptionsSchema = z
	.object({
		auto_approve: z.boolean().nullable().default(null),
		icon: stringDefault,
		icon_link_to: stringDefault,
		tags: z.array(z.string()).default([]),
		web_url: stringDefault,
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
