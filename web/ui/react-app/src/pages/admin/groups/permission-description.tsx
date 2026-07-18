import { describePermission } from '@/pages/admin/groups/permissions';
import type { Action, Resource } from '@/types/auth';

/**
 * The `<dt>/<dd>` pair describing one `resource:action` permission.
 * Falls back to a placeholder when no description exists.
 */
export const PermissionDescription = ({
	resource,
	action,
}: {
	resource: Resource;
	action: Action;
}) => (
	<>
		<dt className="font-mono">
			{resource}:{action}
		</dt>
		<dd>
			{describePermission(resource, action) || 'No description available.'}
		</dd>
	</>
);
