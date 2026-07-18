import { z } from 'zod';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

export const EXPIRY_OPTIONS = [
	{ label: 'Never', value: 'never' },
	{ label: '7 days', value: '168h' },
	{ label: '30 days', value: '720h' },
	{ label: '90 days', value: '2160h' },
	{ label: 'Custom', value: 'custom' },
] as const;

// Go duration (time.ParseDuration): numbers with unit suffixes, e.g. 1h30m.
const DURATION_RE = /^(\d+(\.\d+)?(ns|us|µs|ms|s|m|h))+$/;

/** Whether value is a valid Go duration string (e.g. 48h, 90m, 1h30m). */
export const isValidDuration = (value: string) =>
	DURATION_RE.test(value.trim());

export const tokenSchema = z.object({
	customExpiry: z.string(),
	expiry: z.string(),
	name: z.string().min(1, REQUIRED_MESSAGE),
});

export type TokenFormValues = z.infer<typeof tokenSchema>;
