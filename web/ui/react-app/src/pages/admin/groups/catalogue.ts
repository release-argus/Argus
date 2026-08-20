import type {
	Action,
	ActionPermission,
	Resource,
	ResourcePermissions,
	ScopeType,
} from '@/types/auth';

/** actionsFor returns the resource's actions, in catalogue order. */
export const actionsFor = (
	catalogue: ResourcePermissions[],
	resource: Resource,
): ActionPermission[] =>
	catalogue.find((entry) => entry.name === resource)?.actions ?? [];

/** scopesForAction returns the scopes the (resource, action) pair supports. */
export const scopesForAction = (
	catalogue: ResourcePermissions[],
	resource: Resource,
	action: Action,
): ScopeType[] =>
	actionsFor(catalogue, resource).find((a) => a.action === action)?.scopes ?? [
		'global',
	];

/** defaultActionFor picks the least-privileged action the resource offers. */
export const defaultActionFor = (
	catalogue: ResourcePermissions[],
	resource: Resource,
): Action => {
	const actions = actionsFor(catalogue, resource);
	return (
		actions.find((a) => a.action === 'read')?.action ??
		actions[0]?.action ??
		'read'
	);
};
