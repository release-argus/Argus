import { FieldText, FieldTextArea } from '@/components/generic/field';
import { FieldLegend, FieldSet } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import type { NotifyOptionsSchema } from '@/utils/api/types/config-edit/notify/schemas';

type HeadingProps = {
	title: string;
};

export const Heading = ({ title }: HeadingProps) => (
	<div className="col-span-full flex flex-col gap-1 pt-4">
		<FieldLegend className="text-lg underline">{title}</FieldLegend>
		<Separator className="col-span-full" />
	</div>
);

/**
 * Form fields for the `notify.X.options` section.
 *
 * @param name - The path to the parent `notify` in the form.
 * @param defaults - The default values.
 * @returns The form fields for the `options` section of this `Notify`.
 */
export const NotifyOptions = ({
	name,

	defaults,
}: {
	name: string;

	defaults?: NotifyOptionsSchema;
}) => {
	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<Heading title="Options" />
			<FieldText
				colSize={{ xs: 6 }}
				defaultVal={defaults?.delay}
				label="Delay"
				name={`${name}.options.delay`}
				tooltip={{
					content: 'e.g. 1h2m3s = 1 hour, 2 minutes and 3 seconds',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ xs: 6 }}
				defaultVal={defaults?.max_tries}
				label="Max tries"
				name={`${name}.options.max_tries`}
			/>
			<FieldTextArea
				colSize={{ sm: 12 }}
				defaultVal={defaults?.message}
				label="Message"
				name={`${name}.options.message`}
				rows={3}
			/>
		</FieldSet>
	);
};
