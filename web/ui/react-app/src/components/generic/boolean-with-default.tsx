import { CircleCheck, CircleX } from 'lucide-react';
import type { FC } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';

const options = [
	{
		color: 'text-success',
		icon: CircleCheck,
		text: 'Yes',
		value: 'true',
	},
	{
		color: 'text-destructive',
		icon: CircleX,
		text: 'No',
		value: 'false',
	},
];

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* The form label to display. */
	label: string;
	/* The default value of the field. */
	defaultValue?: boolean | null;
};

type BooleanWithDefaultProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A labelled form field with buttons to select between true, false and default.
 *
 * @param name - The name of the field.
 * @param label - The form label to display.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip media type, either a string or a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaultValue - The default value of the field.
 */
const BooleanWithDefault: FC<BooleanWithDefaultProps> = ({
	name,
	label,
	defaultValue,
	tooltip,
}) => {
	const { control } = useFormContext();

	const optionsDefault = {
		color: defaultValue ? 'text-success' : 'text-destructive',
		icon: defaultValue ? CircleCheck : CircleX,
		text: 'Default:',
		value: 'null',
	};

	return (
		<Controller
			control={control}
			name={name}
			render={({ field: { value, onChange }, fieldState }) => {
				const typedValue = value as string | boolean | undefined;
				if (typedValue === '') onChange(null);
				const strValue =
					(typedValue ?? null) === null ? 'null' : String(typedValue);

				return (
					<Field
						className="col-span-full flex flex-row justify-between gap-2 py-1"
						data-invalid={fieldState.invalid}
						orientation="responsive"
					>
						{label && (
							<FieldLabelWithTooltip
								htmlFor={name}
								text={label}
								tooltip={tooltip}
							/>
						)}

						<ToggleGroup
							aria-describedby={tooltip && `${name}-tooltip`}
							aria-labelledby={`${name}-label`}
							className="justify-end"
							onValueChange={(val) => {
								if (val === 'null') {
									onChange(null);
								} else if (val === 'true') {
									onChange(true);
								} else if (val === 'false') {
									onChange(false);
								}
							}}
							type="single"
							value={strValue}
							variant="outline"
						>
							{options.map((opt) => (
								<ToggleGroupItem key={opt.value} value={opt.value}>
									<p>{opt.text}</p>
									<opt.icon className={opt.color} />
								</ToggleGroupItem>
							))}

							<Separator className="mx-1" orientation="vertical" />

							<ToggleGroupItem
								className="border-l-1!"
								value={optionsDefault.value}
							>
								<p>{optionsDefault.text}</p>
								<optionsDefault.icon className={optionsDefault.color} />
							</ToggleGroupItem>
						</ToggleGroup>
					</Field>
				);
			}}
		/>
	);
};

export default BooleanWithDefault;
