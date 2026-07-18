/**
 * Auth types, mirroring the API shapes in `web/api/types/auth.go`.
 */

// Single source of truth for the RBAC catalogue, so the zod schema and the
// types stay in lockstep (mirrors the Go catalogue in auth/rbac).
export const RESOURCES = [
	'service',
	'service_order',
	'service_action',
	'version_refresh',
	'notify',
	'config',
] as const;
export const ACTIONS = ['read', 'create', 'update', 'delete', 'execute'] as const;
export const SCOPE_TYPES = ['global', 'service', 'service_tag'] as const;

export type Resource = (typeof RESOURCES)[number];
export type Action = (typeof ACTIONS)[number];
export type ScopeType = (typeof SCOPE_TYPES)[number];

export type Scope = {
	type: ScopeType;
	ref?: string;
};

export type Grant = {
	resource: Resource;
	action: Action;
	scope: Scope;
};

export type AuthUser = {
	id: string;
	username: string;
	display_name?: string;
	email?: string;
	enabled: boolean;
	groups: string[] | null;
	created_at: string;
	updated_at: string;
};

export type AuthMe = {
	user: AuthUser;
	permissions: Grant[] | null;
};

export type AuthGroup = {
	id: string;
	name: string;
	description?: string;
	system: boolean;
	/** Immutable seed identity; empty/absent for user-created groups. */
	seed_key?: string;
	members: number;
	permissions: Grant[] | null;
	created_at: string;
	updated_at: string;
};

export type ActionPermission = {
	action: Action;
	scopes: ScopeType[];
};

export type ResourcePermissions = {
	name: Resource;
	actions: ActionPermission[];
};

export type PermissionCatalogue = {
	resources: ResourcePermissions[];
};

export type APIToken = {
	id: string;
	user_id: string;
	name: string;
	prefix: string;
	created_at: string;
	expires_at?: string;
	last_used_at?: string;
};

export type APITokenCreated = APIToken & {
	/** The plaintext token - shown once at creation, never again. */
	token: string;
};
