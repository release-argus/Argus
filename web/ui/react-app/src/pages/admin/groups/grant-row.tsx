import { X } from 'lucide-react';
import type { ReactNode } from 'react';
import { useFormContext } from 'react-hook-form';
import { RequiredMark } from '@/components/generic/field-label';
import HelpTooltip from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { Field, FieldError, FieldSet } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import {
	actionsFor as catalogueActionsFor,
	scopesForAction as catalogueScopesForAction,
} from '@/pages/admin/groups/catalogue';
import { describePermission } from '@/pages/admin/groups/permissions';
import type { GroupFormValues } from '@/pages/admin/groups/schema';
import type {
	Action,
	Grant,
	Resource,
	ResourcePermissions,
	ScopeType,
} from '@/types/auth';

/**
 * Grid placement shared between the header and every grant row.
 */
export const LAYOUT = {
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

export type RowLayout = (typeof LAYOUT)[keyof typeof LAYOUT];

/** Alignment for row items that are shorter than a control. */
const CONTROL_LINE = 'mb-auto flex items-center sm:h-9';

/** The change marker sits beside both halves of a stacked row. */
const MARKER_ALIGN = cn(CONTROL_LINE, 'row-span-2 sm:row-span-1');

/**
 * A labelled control. The label shows only while the row is stacked.
 * `mb-auto` holds it on the control line when a sibling cell grows an error.
 */
const Cell = ({
	label,
	className,
	required,
	invalid,
	children,
}: {
	label: string;
	className?: string;
	required?: boolean;
	invalid?: boolean;
	children: ReactNode;
}) => (
	<Field
		className={cn('mb-auto min-w-0 gap-1', className)}
		data-invalid={invalid}
	>
		<span className="text-muted-foreground text-xs sm:hidden">
			{label}
			{required && <RequiredMark />}
		</span>
		{children}
	</Field>
);

type GrantRowProps = {
	index: number;
	grant: Grant;
	catalogue: ResourcePermissions[];
	layout: RowLayout;
	/** Reserve the marker column; set while any row in the editor is modified. */
	showMarker: boolean;
	modified: boolean;
	duplicate: boolean;
	onResourceChange: (resource: Resource) => void;
	onActionChange: (action: Action) => void;
	onScopeChange: (type: ScopeType) => void;
	onRemove: () => void;
};

/**
 * One grant of the enclosing group form:
 * resource, action, scope, and (when scoped) a ref.
 */
export const GrantRow = ({
	index,
	grant,
	catalogue,
	layout,
	showMarker,
	modified,
	duplicate,
	onResourceChange,
	onActionChange,
	onScopeChange,
	onRemove,
}: GrantRowProps) => {
	const {
		formState: { errors },
		register,
	} = useFormContext<GroupFormValues>();

	const description = describePermission(grant.resource, grant.action);
	const refError = errors.permissions?.[index]?.scope?.ref;

	return (
		<FieldSet
			className={cn(
				'grid items-start gap-2 rounded-md border p-2 sm:items-center sm:border-0 sm:p-0',
				layout.row,
				duplicate
					? 'border-destructive bg-destructive/10 sm:border sm:border-destructive'
					: modified && 'bg-primary/5 sm:rounded-sm',
			)}
		>
			{duplicate && <span className="sr-only">Duplicate permission</span>}
			{showMarker &&
				(modified ? (
					<span
						className={cn(
							MARKER_ALIGN,
							'mx-auto text-xs',
							duplicate ? 'text-destructive' : 'text-primary',
						)}
					>
						<span aria-hidden>●</span>
						<span className="sr-only">Unsaved change</span>
					</span>
				) : (
					<span className={MARKER_ALIGN} />
				))}
			<Cell label="Resource">
				<Select
					onValueChange={(value) => onResourceChange(value as Resource)}
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
					onValueChange={(value) => onActionChange(value as Action)}
					value={grant.action}
				>
					<SelectTrigger
						aria-label={`Grant ${index} action`}
						className="w-full"
					>
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{catalogueActionsFor(catalogue, grant.resource).map((action) => (
							<SelectItem key={action.action} value={action.action}>
								{action.action}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</Cell>
			<Cell className={layout.scope} label="Scope">
				<Select
					onValueChange={(value) => onScopeChange(value as ScopeType)}
					value={grant.scope.type}
				>
					<SelectTrigger aria-label={`Grant ${index} scope`} className="w-full">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{catalogueScopesForAction(
							catalogue,
							grant.resource,
							grant.action,
						).map((scope) => (
							<SelectItem key={scope} value={scope}>
								{scope}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</Cell>
			{grant.scope.type === 'global' ? (
				<span className="hidden sm:block" />
			) : (
				<Cell
					className={layout.target}
					invalid={!!refError}
					label="Target"
					required
				>
					<Input
						aria-describedby={refError ? `grant-${index}-ref-error` : undefined}
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
					<FieldError errors={[refError]} id={`grant-${index}-ref-error`} />
				</Cell>
			)}
			<span
				className={cn(CONTROL_LINE, 'justify-self-center *:mb-0', layout.help)}
			>
				<HelpTooltip
					ariaLabel={`What ${grant.resource}:${grant.action} allows`}
					content={
						<div className="grid gap-1 text-left">
							<dl className="grid gap-1">
								<dt className="font-mono">
									{grant.resource}:{grant.action}
								</dt>
								<dd>{description || 'No description available.'}</dd>
							</dl>
							{duplicate && (
								<span className="text-destructive">
									This grant is listed more than once.
								</span>
							)}
						</div>
					}
					size="sm"
					type="element"
				/>
			</span>
			<Button
				aria-label={`Remove grant ${index}`}
				className={cn('mb-auto', layout.remove)}
				onClick={onRemove}
				size="icon-md"
				type="button"
				variant="ghost"
			>
				<X aria-hidden />
			</Button>
		</FieldSet>
	);
};
