import { Plus, X } from 'lucide-react';
import { Fragment, type ReactNode, useMemo } from 'react';
import { useFieldArray, useFormContext, useWatch } from 'react-hook-form';
import { RequiredMark } from '@/components/generic/field-label';
import HelpTooltip from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { describePermission } from '@/pages/admin/groups/permissions';
import type { GroupFormValues } from '@/pages/admin/groups/schema';
import type {
	Action,
	ActionPermission,
	Grant,
	Resource,
	ResourcePermissions,
	ScopeType,
} from '@/types/auth';

/**
 * Grid placement shared between the header and every grant row.
 */
const LAYOUT = {
	marked: {
		help: 'col-start-4 row-start-1 sm:col-start-auto sm:row-start-auto',
		remove: 'col-start-4 row-start-2 sm:col-start-auto sm:row-start-auto',
		row: 'grid-cols-[0.75rem_minmax(0,1fr)_minmax(0,1fr)_auto] sm:grid-cols-[0.75rem_minmax(6rem,1.1fr)_minmax(5rem,0.9fr)_minmax(6rem,1fr)_minmax(5rem,1fr)_auto_auto]',
		scope: 'col-start-2 row-start-2 sm:col-start-auto sm:row-start-auto',
		target: 'col-start-3 row-start-2 sm:col-start-auto sm:row-start-auto',
	},
	plain: {
		help: 'col-start-3 row-start-1 sm:col-start-auto sm:row-start-auto',
		remove: 'col-start-3 row-start-2 sm:col-start-auto sm:row-start-auto',
		row: 'grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] sm:grid-cols-[minmax(6rem,1.1fr)_minmax(5rem,0.9fr)_minmax(6rem,1fr)_minmax(5rem,1fr)_auto_auto]',
		scope: 'col-start-1 row-start-2 sm:col-start-auto sm:row-start-auto',
		target: 'col-start-2 row-start-2 sm:col-start-auto sm:row-start-auto',
	},
} as const;

/**
 * Alignment for the row items that are shorter than a control: hold them on the
 * control line rather than centring them in a row grown by an error message.
 */
const CONTROL_LINE = 'mb-auto flex items-center sm:h-9';

/** The change marker sits beside both halves of a stacked row. */
const MARKER_ALIGN = cn(CONTROL_LINE, 'row-span-2 sm:row-span-1');

// grantKey identifies a grant by value, for comparison against the saved set.
const grantKey = (grant: Grant) =>
	`${grant.resource}:${grant.action}:${grant.scope.type}:${grant.scope.ref ?? ''}`;

/**
 * A labelled control. The label shows only while the row is stacked.
 * `mb-auto` holds it on the control line when a sibling cell grows an error.
 */
const Cell = ({
	label,
	className,
	required,
	children,
}: {
	label: string;
	className?: string;
	required?: boolean;
	children: ReactNode;
}) => (
	<div className={cn('mb-auto grid min-w-0 gap-1', className)}>
		<span className="text-muted-foreground text-xs sm:hidden">
			{label}
			{required && <RequiredMark />}
		</span>
		{children}
	</div>
);

type GrantEditorProps = {
	catalogue: ResourcePermissions[];
	/**
	 * The saved group's grants. When given, rows that differ from it are
	 * marked; omit when creating, where every row is new.
	 */
	savedGrants?: Grant[];
};

/**
 * Edits the enclosing group form's permission grants against the catalogue:
 * one row per grant of resource, action, scope, and (when scoped) a ref.
 * Reads and writes `permissions` via form context.
 */
export const GrantEditor = ({ catalogue, savedGrants }: GrantEditorProps) => {
	const {
		control,
		formState: { errors },
		register,
		setValue,
	} = useFormContext<GroupFormValues>();
	const { fields, append, remove } = useFieldArray({
		control,
		name: 'permissions',
	});
	const grants = useWatch({ control, name: 'permissions' }) ?? [];

	// Matches rows against the saved set by value, not by position.
	const { modifiedRows, duplicateKeys } = useMemo(() => {
		// How many times each grant appears in the saved set, claimed as rows match it.
		const unclaimed = new Map<string, number>();
		for (const grant of savedGrants ?? []) {
			const key = grantKey(grant);
			unclaimed.set(key, (unclaimed.get(key) ?? 0) + 1);
		}

		const seen = new Set<string>();
		const duplicates = new Set<string>();
		const modified = grants.map((grant) => {
			const key = grantKey(grant);
			if (seen.has(key)) duplicates.add(key);
			seen.add(key);

			// A row is new or edited when no saved grant matches it.
			const left = unclaimed.get(key) ?? 0;
			if (left === 0) return savedGrants !== undefined;
			unclaimed.set(key, left - 1);
			return false;
		});
		return { duplicateKeys: duplicates, modifiedRows: modified };
	}, [grants, savedGrants]);

	const anyModified = modifiedRows.some(Boolean);
	const layout = anyModified ? LAYOUT.marked : LAYOUT.plain;

	const setGrant = (index: number, grant: Grant) =>
		setValue(`permissions.${index}`, grant, { shouldDirty: true });

	// forResource returns the resource permissions from the catalogue.
	const forResource = (resource: Resource) =>
		catalogue.find((entry) => entry.name === resource);

	// actionsFor returns the resource's actions, in catalogue order.
	const actionsFor = (resource: Resource): ActionPermission[] =>
		forResource(resource)?.actions ?? [];

	// scopesForAction returns the scopes the (resource, action) pair supports.
	const scopesForAction = (resource: Resource, action: Action): ScopeType[] =>
		forResource(resource)?.actions.find((a) => a.action === action)?.scopes ?? [
			'global',
		];

	// defaultActionFor picks the least-privileged action the resource offers.
	const defaultActionFor = (resource: Resource): Action => {
		const actions = actionsFor(resource);
		return (
			actions.find((a) => a.action === 'read')?.action ??
			actions[0]?.action ??
			'read'
		);
	};

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
						<span className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1 text-left">
							{catalogue.flatMap((resource) =>
								resource.actions.map((action) => (
									<Fragment key={`${resource.name}:${action.action}`}>
										<span className="font-mono">
											{resource.name}:{action.action}
										</span>
										<span>
											{describePermission(resource.name, action.action) ||
												'No description available.'}
										</span>
									</Fragment>
								)),
							)}
						</span>
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
				const modified = modifiedRows[index];
				const duplicate = duplicateKeys.has(grantKey(grant));
				const description = describePermission(grant.resource, grant.action);
				const refError = errors.permissions?.[index]?.scope?.ref;
				return (
					<div
						className={cn(
							'grid items-start gap-2 rounded-md border p-2 sm:items-center sm:border-0 sm:p-0',
							layout.row,
							duplicate
								? 'border-destructive bg-destructive/10 sm:border sm:border-destructive'
								: modified && 'bg-primary/5 sm:rounded-sm',
						)}
						key={field.id}
					>
						{duplicate && <span className="sr-only">Duplicate permission</span>}
						{anyModified &&
							(modified ? (
								<span
									className={cn(MARKER_ALIGN, 'mx-auto text-primary text-xs')}
								>
									<span aria-hidden>●</span>
									<span className="sr-only">Unsaved change</span>
								</span>
							) : (
								<span className={MARKER_ALIGN} />
							))}
						<Cell label="Resource">
							<Select
								onValueChange={(value) =>
									changeResource(index, value as Resource)
								}
								value={grant.resource}
							>
								<SelectTrigger
									aria-label={`Grant ${index} resource`}
									className="w-full"
								>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{catalogue.map((resource) => (
										<SelectItem key={resource.name} value={resource.name}>
											{resource.name}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</Cell>
						<Cell label="Action">
							<Select
								onValueChange={(value) => changeAction(index, value as Action)}
								value={grant.action}
							>
								<SelectTrigger
									aria-label={`Grant ${index} action`}
									className="w-full"
								>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{actionsFor(grant.resource).map((action) => (
										<SelectItem key={action.action} value={action.action}>
											{action.action}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</Cell>
						<Cell className={layout.scope} label="Scope">
							<Select
								onValueChange={(value) =>
									changeScope(index, value as ScopeType)
								}
								value={grant.scope.type}
							>
								<SelectTrigger
									aria-label={`Grant ${index} scope`}
									className="w-full"
								>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{scopesForAction(grant.resource, grant.action).map(
										(scope) => (
											<SelectItem key={scope} value={scope}>
												{scope}
											</SelectItem>
										),
									)}
								</SelectContent>
							</Select>
						</Cell>
						{grant.scope.type === 'global' ? (
							<span className="hidden sm:block" />
						) : (
							<Cell className={layout.target} label="Target" required>
								<Input
									aria-invalid={!!refError}
									aria-label={`Grant ${index} scope ref`}
									aria-required
									placeholder={
										grant.scope.type === 'service'
											? 'e.g. release-argus/Argus'
											: 'e.g. tag_name'
									}
									{...register(`permissions.${index}.scope.ref`)}
								/>
								<FieldError errors={[refError]} />
							</Cell>
						)}
						<span
							className={cn(
								CONTROL_LINE,
								'justify-self-center *:mb-0',
								layout.help,
							)}
						>
							<HelpTooltip
								ariaLabel={`What ${grant.resource}:${grant.action} allows`}
								content={
									<span className="grid gap-1 text-left">
										<span className="font-mono">
											{grant.resource}:{grant.action}
										</span>
										<span>{description || 'No description available.'}</span>
										{duplicate && (
											<span className="text-destructive">
												This grant is listed more than once.
											</span>
										)}
									</span>
								}
								size="sm"
								type="element"
							/>
						</span>
						<Button
							aria-label={`Remove grant ${index}`}
							className={cn('mb-auto', layout.remove)}
							onClick={() => remove(index)}
							size="icon-md"
							type="button"
							variant="ghost"
						>
							<X aria-hidden />
						</Button>
					</div>
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
