import type { Action, Resource } from '@/types/auth';

/**
 * What each (resource, action) pairing of the RBAC catalogue lets a member do.
 * Mirrors the routes guarded in `web/api/v1/http.go`.
 */
const DESCRIPTIONS: Partial<Record<`${Resource}:${Action}`, string>> = {
	'config:read':
		'View the running config, the flags it started with, and build info.',
	'metric:read': 'Scrape Prometheus metrics and read the dashboard counts.',
	'notify:execute': 'Send test notifications.',
	'service_action:execute':
		"View a service's webhooks and commands, approve or skip releases, and re-run actions.",
	'service_order:update': 'Reorder the services on the dashboard.',
	'service:create': 'Add services, and test their lookups before saving.',
	'service:delete': 'Delete services.',
	'service:read': 'View services and their versions.',
	'service:update':
		'Edit existing services, including the commands they run on a new release.',
	'version_refresh:execute':
		'Re-query the latest and deployed versions on demand. Refreshing with overrides can execute commands, so this is equivalent to shell access on the Argus host.',
};

/** describePermission returns what the pairing enables ('' when unknown). */
export const describePermission = (
	resource: Resource,
	action: Action,
): string => DESCRIPTIONS[`${resource}:${action}`] ?? '';
