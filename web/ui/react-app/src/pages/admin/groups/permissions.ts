import type { Action, Resource } from '@/types/auth';

/**
 * What each (resource, action) pairing of the RBAC catalogue lets a member do.
 * Mirrors the routes guarded in `web/api/v1/http.go`.
 */
const DESCRIPTIONS: Partial<Record<`${Resource}:${Action}`, string>> = {
	'config:read':
		'View the running config, the flags it started with, and build info.',
	'notify:execute': 'Send test notifications.',
	'service_action:execute':
		'Approve or skip a release, and re-run its WebHooks and commands.',
	'service_order:update': 'Reorder the services on the dashboard.',
	'service:create': 'Add services, and test their lookups before saving.',
	'service:delete': 'Delete services.',
	'service:read': 'View services, their versions, and the actions they define.',
	'service:update': 'Edit existing services.',
	'version_refresh:execute':
		'Re-query the latest and deployed versions on demand.',
};

/** describePermission returns what the pairing enables ('' when unknown). */
export const describePermission = (
	resource: Resource,
	action: Action,
): string => DESCRIPTIONS[`${resource}:${action}`] ?? '';
