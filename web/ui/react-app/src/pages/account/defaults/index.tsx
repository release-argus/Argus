import { useMutation, useQueryClient } from '@tanstack/react-query';
import { type ReactElement, useEffect } from 'react';
import { Controller } from 'react-hook-form';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { TextOrLoading } from '@/components/ui/loading-ellipsis';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import {
	approvalsToolbarViewOptions,
	DEFAULT_CARD_TIMESTAMPS,
	DEFAULT_HIDE_VALUE,
	DEFAULT_VIEW_VALUE,
	hideNamesFromValues,
	type ToolbarViewOption,
	toolbarHideOptions,
	toolbarTimestampOptions,
} from '@/constants/toolbar';
import { useAuth } from '@/contexts/auth';
import useZodForm from '@/hooks/use-zod-form';
import { QUERY_KEYS } from '@/lib/query-keys';
import {
	type DefaultsFormValues,
	defaultsSchema,
} from '@/pages/account/defaults/schema';
import { GroupHeading, GroupRule } from '@/pages/account/group';
import type { AuthMe, DashboardPreferences } from '@/types/auth';
import * as authAPI from '@/utils/api/auth';
import { clearCardTimestamps } from '@/utils/card-timestamps';
import { getErrorMessage } from '@/utils/errors';

const valuesFor = (preferences?: DashboardPreferences): DefaultsFormValues => ({
	hide: preferences?.hide ?? hideNamesFromValues(DEFAULT_HIDE_VALUE),
	timestamps: preferences?.timestamps ?? [...DEFAULT_CARD_TIMESTAMPS],
	view: preferences?.view ?? DEFAULT_VIEW_VALUE,
});

/** Adds value, or drops it if already there. */
const toggled = <T extends string>(list: T[], value: T, on: boolean): T[] =>
	on ? [...list, value] : list.filter((item) => item !== value);

type ToggleRowProps = {
	id: string;
	label: string;
	checked: boolean;
	onCheckedChange: (checked: boolean) => void;
};

const ToggleRow = ({ id, label, checked, onCheckedChange }: ToggleRowProps) => (
	<div className="flex items-center gap-2">
		<Checkbox
			checked={checked}
			id={id}
			onCheckedChange={(next) => onCheckedChange(next === true)}
		/>
		<Label className="font-normal" htmlFor={id}>
			{label}
		</Label>
	</div>
);

/**
 * The signed-in user's dashboard preferences.
 */
export const Defaults = (): ReactElement => {
	const { preferences } = useAuth();
	const queryClient = useQueryClient();
	const form = useZodForm({
		defaultValues: valuesFor(preferences),
		schema: defaultsSchema,
	});
	const { isDirty } = form.formState;

	// Re-seed only an untouched form.
	// biome-ignore lint/correctness/useExhaustiveDependencies: form is stable.
	useEffect(() => {
		if (form.formState.isDirty) return;
		form.reset(valuesFor(preferences));
	}, [preferences]);

	const applied = (me: AuthMe, message: string) => {
		queryClient.setQueryData(QUERY_KEYS.AUTH.ME(), me);
		form.reset(valuesFor(me.preferences));
		clearCardTimestamps();
		toast.success(message);
	};

	const save = useMutation({
		mutationFn: (values: DefaultsFormValues) =>
			authAPI.updatePreferences(values),
		onError: (error) =>
			toast.error('Failed to save defaults', {
				description: getErrorMessage(error),
			}),
		onSuccess: (me: AuthMe) => applied(me, 'Dashboard defaults saved'),
	});

	const reset = useMutation({
		mutationFn: authAPI.resetPreferences,
		onError: (error) =>
			toast.error('Failed to reset defaults', {
				description: getErrorMessage(error),
			}),
		onSuccess: (me: AuthMe) => applied(me, 'Dashboard defaults reset'),
	});

	const pending = save.isPending || reset.isPending;
	const inertStyle =
		'aria-disabled:pointer-events-none aria-disabled:opacity-50';
	const saveInert = pending || !isDirty;
	const resetInert = pending || !preferences;

	return (
		<form
			className="flex flex-col gap-6"
			onSubmit={form.handleSubmit((values) => {
				if (saveInert) return;
				save.mutate(values);
			})}
		>
			<Card>
				<CardContent className="grid gap-6">
					<section className="grid gap-4">
						<GroupHeading
							description="Which services the dashboard shows when you open it without filters in the URL. A link carrying filters still wins."
							id="defaults-filters-heading"
							title="Filters"
						/>
						<Controller
							control={form.control}
							name="hide"
							render={({ field }) => (
								<fieldset
									aria-labelledby="defaults-filters-heading"
									className="grid gap-3 sm:grid-cols-2"
								>
									{toolbarHideOptions.map(({ key, label }) => (
										<ToggleRow
											checked={field.value.includes(key)}
											id={`defaults-hide-${key}`}
											key={key}
											label={label}
											onCheckedChange={(on) =>
												field.onChange(toggled(field.value, key, on))
											}
										/>
									))}
								</fieldset>
							)}
						/>
					</section>

					<GroupRule />

					<section className="grid gap-4">
						<GroupHeading
							description="The layout the dashboard opens in."
							id="defaults-layout-heading"
							title="Layout"
						/>
						<Controller
							control={form.control}
							name="view"
							render={({ field }) => (
								<ToggleGroup
									aria-labelledby="defaults-layout-heading"
									className="w-fit"
									onValueChange={(view: ToolbarViewOption | '') => {
										if (!view) return;
										field.onChange(view);
									}}
									type="single"
									value={field.value}
									variant="outline"
								>
									{approvalsToolbarViewOptions.map(
										({ icon: Icon, label, value }) => (
											<ToggleGroupItem
												aria-label={label}
												key={value}
												value={value}
											>
												<Icon className="h-4 w-4" />
												{label}
											</ToggleGroupItem>
										),
									)}
								</ToggleGroup>
							)}
						/>
					</section>

					<GroupRule />

					<section className="grid gap-4">
						<GroupHeading
							description="The timestamps shown under each service card, in the grid layout only."
							id="defaults-timestamps-heading"
							title="Timestamps"
						/>
						<Controller
							control={form.control}
							name="timestamps"
							render={({ field }) => (
								<fieldset
									aria-labelledby="defaults-timestamps-heading"
									className="grid gap-3 sm:grid-cols-2"
								>
									{toolbarTimestampOptions.map(({ key, label }) => (
										<ToggleRow
											checked={field.value.includes(key)}
											id={`defaults-timestamp-${key}`}
											key={key}
											label={label}
											onCheckedChange={(on) =>
												field.onChange(toggled(field.value, key, on))
											}
										/>
									))}
								</fieldset>
							)}
						/>
					</section>
				</CardContent>
			</Card>

			<div className="flex gap-2">
				<Button aria-disabled={saveInert} className={inertStyle} type="submit">
					<TextOrLoading loading={save.isPending} text="Save changes" />
				</Button>
				<Button
					aria-disabled={resetInert}
					className={inertStyle}
					onClick={() => {
						if (resetInert) return;
						reset.mutate();
					}}
					type="button"
					variant="outline"
				>
					<TextOrLoading
						loading={reset.isPending}
						text="Reset to Argus defaults"
					/>
				</Button>
			</div>
		</form>
	);
};
