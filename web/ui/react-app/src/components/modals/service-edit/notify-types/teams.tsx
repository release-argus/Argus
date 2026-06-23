import { useMemo } from 'react';
import { FieldColour, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyTeamsSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Teams` (Power Automate) notifier.
 *
 * @param name - The path to this `Teams` in the form.
 * @param main - The main values.
 */
const TEAMS = ({ name, main }: { name: string; main?: NotifyTeamsSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.teams,
		[main, typeDataDefaults?.notify.teams],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.host}
					label="Host"
					name={`${name}.params.host`}
					required
					tooltip={{
						ariaLabel: 'Full Power Automate incoming webhook URL',
						content: (
							<>
								<span className="text-muted-foreground">
									{'Full Power Automate workflow webhook URL, e.g. '}
								</span>
								<span className="break-all">
									{'https://prod-00.westus.logic.azure.com:443/workflows/...'}
								</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldColour
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
					tooltip={{
						content:
							'Accepts Hex (e.g. #FFA500), ' +
							'Adaptive Card Color names (e.g. accent, good, warning, attention, dark, light, default), ' +
							'or Common Color names (e.g. red, green, blue, yellow, orange).',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default TEAMS;
