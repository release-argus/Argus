/**
 * Auth API request payloads. Entity/response shapes live in `@/types/auth`.
 */

import type {
	CardTimestampType,
	HideName,
	ToolbarViewOption,
} from '@/constants/toolbar';
import type { Grant } from '@/types/auth';

export type LoginRequest = {
	username: string;
	password: string;
};

/**
 * GET /auth/setup - whether first-run setup is still pending, and any demo
 * credentials to prefill the login form with.
 */
export type SetupState = {
	setup_required: boolean;
	demo?: {
		username: string;
		password: string;
	};
};

/** POST /auth/setup - the first administrator's account details. */
export type SetupRequest = {
	username: string;
	display_name?: string;
	password: string;
};

/**
 * PATCH /auth/me - the signed-in user changing their own account. The current
 * password is always required; omitted fields stay unchanged.
 */
export type AccountUpdateRequest = {
	current_password: string;
	display_name?: string;
	email?: string;
	new_password?: string;
};

export type UserCreateRequest = {
	username: string;
	password: string;
	display_name?: string;
	email?: string;
	groups?: string[];
};

export type UserPatchRequest = {
	display_name?: string;
	email?: string;
	enabled?: boolean;
	groups?: string[];
	password?: string;
};

export type GroupCreateRequest = {
	name: string;
	description?: string;
	permissions?: Grant[];
};

export type GroupPatchRequest = {
	name?: string;
	description?: string;
	permissions?: Grant[];
};

export type APITokenCreateRequest = {
	name: string;
	expires_in?: string;
};

/**
 * PUT /auth/me/preferences - the signed-in user's dashboard preferences. An
 * omitted field inherits the built-in default; an empty list is an explicit
 * none.
 */
export type PreferencesUpdateRequest = {
	hide?: HideName[];
	view?: ToolbarViewOption;
	timestamps?: CardTimestampType[];
};
