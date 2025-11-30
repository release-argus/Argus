import { Trash2 } from 'lucide-react';
import { type FC, memo } from 'react';
import { FieldText } from '@/components/generic/field';
import type { HeaderPlaceholders } from '@/components/generic/field-shared';
import { Button } from '@/components/ui/button';
import { FieldGroup } from '@/components/ui/field';
import { cn } from '@/lib/utils';
import type { CustomHeader } from '@/utils/api/types/config/shared';

type FieldKeyValProps = {
	/* The name of the field. */
	name: string;
	/* The column width on XS+ screens. */
	colSpan: number;
	/* The default values for the field. */
	defaults?: CustomHeader;
	/* The function to remove the field. */
	removeMe: () => void;
	/* Optional placeholders for the key and value fields. */
	placeholders?: HeaderPlaceholders;
};

/**
 * The form fields for a key-value pair.
 *
 * @param name - The name of the field in the form.
 * @param colSpan - The width of the field.
 * @param defaults - The default values for the field.
 * @param removeMe - The function to remove the field.
 * @param placeholders - Optional placeholders for the key/value fields.
 * @returns The form fields for a key-value pair at 'name'.
 */
const FieldKeyVal: FC<FieldKeyValProps> = ({
	name,
	colSpan,
	defaults,
	removeMe,
	placeholders,
}) => {
	const keyColSpan = Math.round(colSpan / 2);
	const valueColSpan = Math.floor(colSpan / 2);

	return (
		<>
			<div className="col-span-1 py-1 md:pe-2">
				<Button
					aria-label="Delete this key-value pair"
					className="size-full"
					onClick={removeMe}
					size="icon-md"
					variant="outline"
				>
					<Trash2 />
				</Button>
			</div>
			<FieldGroup
				className={cn(`col-span-${colSpan}`, 'grid grid-cols-subgrid gap-x-2')}
			>
				<FieldText
					colSize={{ sm: keyColSpan, xs: keyColSpan }}
					defaultVal={defaults?.key}
					name={`${name}.key`}
					placeholder={placeholders?.key ?? 'e.g. X-Header'}
					required
				/>
				<FieldText
					colSize={{ sm: valueColSpan, xs: valueColSpan }}
					defaultVal={defaults?.value}
					name={`${name}.value`}
					placeholder={placeholders?.value ?? 'e.g. value'}
					required
				/>
			</FieldGroup>
		</>
	);
};

export default memo(FieldKeyVal);
