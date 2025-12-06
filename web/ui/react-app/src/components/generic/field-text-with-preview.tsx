import { type FC, useCallback, useEffect, useState } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupInput,
} from '@/components/ui/input-group';
import { cn } from '@/lib/utils';
import { isValidURL } from '@/utils/api/types/config-edit/validators';

type BaseProps = {
	/* The name of the field. */
	name: string;

	/* The label of the field. */
	label: string;
	/* The default value of the field. */
	defaultVal?: string;
	/* The placeholder of the field. */
	placeholder?: string;
};

type FieldTextWithPreviewProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A field item with a preview image
 *
 * @param name - The name of the form item.
 * @param label - The label of the form item.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 * @returns A form item at `name` with a preview image, label and tooltip.
 */
const FieldTextWithPreview: FC<FieldTextWithPreviewProps> = ({
	name,

	label,
	tooltip,
	defaultVal,
	placeholder,
}) => {
	const { control } = useFormContext();
	const formValue = useWatch({ name: name }) as string | null;

	// The preview image address, or undefined if invalid.
	const [previewURL, setPreviewURL] = useState(
		isValidURL(formValue) ? formValue : null,
	);
	// Set the preview image.
	const setPreview = useCallback(
		(url: string | null) => {
			const previewSource = url || defaultVal;
			if (previewSource && isValidURL(previewSource)) {
				setPreviewURL(previewSource);
			} else {
				setPreviewURL(null);
			}
		},
		[defaultVal],
	);

	// Wait for a pause in typing to set the preview.
	useEffect(() => {
		const timer = setTimeout(() => {
			setPreview(formValue);
		}, 750);
		return () => {
			clearTimeout(timer);
		};
	}, [formValue, setPreview]);

	return (
		<FieldGroup className={cn('col-span-full py-1')}>
			<Controller
				control={control}
				name={name}
				render={({ field, fieldState }) => (
					<Field data-invalid={fieldState.invalid}>
						<FieldLabelWithTooltip
							htmlFor={name}
							text={label}
							tooltip={tooltip}
						/>
						<InputGroup>
							<InputGroupInput
								{...field}
								aria-describedby={cn(tooltip && `${name}-tooltip`)}
								aria-label={`Value field for ${label}`}
								id={name}
								onBlur={() => setPreview(formValue)}
								placeholder={placeholder ?? defaultVal}
								type="text"
							/>

							{previewURL && (
								<InputGroupAddon
									align="inline-end"
									aria-label="Preview of the image"
									className="max-w-12"
								>
									<img
										alt="Icon preview"
										className="h-8 w-auto text-xs"
										src={previewURL}
									/>
								</InputGroupAddon>
							)}
						</InputGroup>
						{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
					</Field>
				)}
			/>
		</FieldGroup>
	);
};

export default FieldTextWithPreview;
