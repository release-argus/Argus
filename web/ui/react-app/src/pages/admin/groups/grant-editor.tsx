import { Plus } from 'lucide-react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { RequiredMark } from '@/components/generic/field-label';
import HelpTooltip from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import {
	actionsFor as catalogueActionsFor,
	defaultActionFor as catalogueDefaultActionFor,
	scopesForAction as catalogueScopesForAction,
} from '@/pages/admin/groups/catalogue';
import { type GrantDiff, grantKey } from '@/pages/admin/groups/grant-diff';
import { GrantRow, LAYOUT } from '@/pages/admin/groups/grant-row';
import { PermissionDescription } from '@/pages/admin/groups/permission-description';
import type { GroupFormValues } from '@/pages/admin/groups/schema';
import type {
	Action,
	Grant,
	Resource,
	ResourcePermissions,
	ScopeType,
} from '@/types/auth';

type GrantEditorProps = {
	catalogue: ResourcePermissions[];
	/** Which rows differ from the saved grants, and which are duplicates. */
	diff: GrantDiff;
};

/**
 * Edits the enclosing group form's permission grants against the catalogue:
 * one row per grant of resource, action, scope, and (when scoped) a ref.
 */
export const GrantEditor = ({ catalogue, diff }: GrantEditorProps) => {
	const { control, setValue } = useFormContext<GroupFormValues>();
	const { fields, append, remove } = useFieldArray({
		control,
		name: 'permissions',
	});
	const grants = useWatch({ control, name: 'permissions' }) ?? [];

	const { modifiedRows, duplicateKeys } = diff;

	const anyModified = modifiedRows.some(Boolean);
	const layout = anyModified ? LAYOUT.marked : LAYOUT.plain;

	const setGrant = (index: number, grant: Grant) =>
		setValue(`permissions.${index}`, grant, { shouldDirty: true });

	const actionsFor = (resource: Resource) =>
		catalogueActionsFor(catalogue, resource);
	const scopesForAction = (resource: Resource, action: Action) =>
		catalogueScopesForAction(catalogue, resource, action);
	const defaultActionFor = (resource: Resource) =>
		catalogueDefaultActionFor(catalogue, resource);

	const changeResource = (index: number, resource: Resource) => {
		const grant = grants[index];
		const actions = actionsFor(resource);
		// Keep the selected action when the new resource offers it too.
		const action = actions.some((a) => a.action === grant.action)
			? grant.action
			: defaultActionFor(resource);
		// Keep the scope only while the new pair still supports it.
		const scope = scopesForAction(resource, action).includes(grant.scope.type)
			? grant.scope
			: { type: 'global' as const };
		setGrant(index, { action, resource, scope });
	};

	const changeAction = (index: number, action: Action) => {
		const grant = grants[index];
		if (!actionsFor(grant.resource).some((a) => a.action === action)) return;
		// Reset to global when the new action doesn't support the current scope
		const scope = scopesForAction(grant.resource, action).includes(
			grant.scope.type,
		)
			? grant.scope
			: { type: 'global' as const };
		setGrant(index, { ...grant, action, scope });
	};

	const changeScope = (index: number, type: ScopeType) => {
		const grant = grants[index];
		if (!scopesForAction(grant.resource, grant.action).includes(type)) return;
		setGrant(index, {
			...grant,
			scope: type === 'global' ? { type } : { ref: '', type },
		});
	};

	const addGrant = () => {
		const resource = catalogue[0]?.name ?? 'service';
		append({
			action: defaultActionFor(resource),
			resource,
			scope: { type: 'global' },
		});
	};

	return (
		<fieldset className="grid min-w-0 gap-2">
			<legend className="flex items-center pb-1 font-medium text-sm">
				Permissions
				<HelpTooltip
					ariaLabel="What each permission allows"
					content={
						<dl className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1 text-left">
							{catalogue.flatMap((resource) =>
								resource.actions.map((action) => (
									<PermissionDescription
										action={action.action}
										key={`${resource.name}:${action.action}`}
										resource={resource.name}
									/>
								)),
							)}
						</dl>
					}
					size="sm"
					type="element"
				/>
			</legend>
			{fields.length > 0 && (
				<div
					className={cn(
						'hidden gap-2 text-muted-foreground text-xs sm:grid sm:items-center',
						layout.row,
					)}
				>
					{anyModified && <span />}
					<span>Resource</span>
					<span>Action</span>
					<span>Scope</span>
					<span>
						Target
						<RequiredMark />
					</span>
					<span />
					<span />
				</div>
			)}
			{fields.map((field, index) => {
				const grant = grants[index];
				if (!grant) return null;
				return (
					<GrantRow
						catalogue={catalogue}
						duplicate={duplicateKeys.has(grantKey(grant))}
						grant={grant}
						index={index}
						key={field.id}
						layout={layout}
						modified={modifiedRows[index]}
						onActionChange={(action) => changeAction(index, action)}
						onRemove={() => remove(index)}
						onResourceChange={(resource) => changeResource(index, resource)}
						onScopeChange={(type) => changeScope(index, type)}
						showMarker={anyModified}
					/>
				);
			})}
			<Button
				className="w-fit"
				onClick={addGrant}
				size="sm"
				type="button"
				variant="outline"
			>
				<Plus aria-hidden /> Add permission
			</Button>
		</fieldset>
	);
};
