/**
 * Auth API request payloads. Entity/response shapes live in `@/types/auth`.
 */

import type { Grant } from '@/types/auth';

export type LoginRequest = {
	username: string;
	password: string;
};

/** GET /auth/setup - whether first-run setup is still pending. */
export type SetupState = {
	setup_required: boolean;
};

/** POST /auth/setup - the first administrator's account details. */
export type SetupRequest = {
	username: string;
	display_name?: string;
	password: string;
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
