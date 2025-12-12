import type { GotifyExtraSchema } from '@/utils/api/types/config-edit/notify/types/gotify.ts';

export type GotifyExtraProps = {
	/* The name of the field in the form. */
	name: string;

	/* The default values for the target. */
	defaults?: GotifyExtraSchema;
};
