import { IdCard, LoaderCircle } from 'lucide-react';
import { type FC, useCallback, useEffect, useState } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { FieldText } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import Tip from '@/components/ui/tip';
import { Toggle } from '@/components/ui/toggle';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useDelayedRender } from '@/hooks/use-delayed-render';

type AdvancedToggleProps = {
	name: string;
	onClick: () => void;
};

const AdvancedToggle: FC<AdvancedToggleProps> = ({ name, onClick }) => {
	const state = useWatch({ name });

	return (
		<Tip
			className="absolute top-1 right-0"
			content="Toggle to separate ID (service key) and Name in the config YAML"
			delayDuration={500}
			touchDelayDuration={250}
		>
			<Toggle className="h-6 w-10" onClick={onClick} pressed={!!state}>
				<IdCard />
			</Toggle>
		</Tip>
	);
};

type EditServiceRootProps = {
	loading: boolean;
};

/**
 * The form fields for the root values of a Service.
 *
 * @param loading - Indicates whether the modal shows a loading state.
 * @returns The form fields for the root values of a Service.
 */
const EditServiceRoot: FC<EditServiceRootProps> = ({ loading }) => {
	const delayedRender = useDelayedRender(500);
	const { serviceID, schemaData } = useSchemaContext();
	const { setValue, resetField } = useFormContext();

	const [separateName, setSeparateName] = useState(false);

	const originalName = schemaData?.name;

	// biome-ignore lint/correctness/useExhaustiveDependencies: separateName stable.
	useEffect(() => {
		if (Boolean(originalName) !== separateName) {
			setSeparateName(!!originalName);
		}
	}, [originalName]);

	// Handle toggling of id/name separation.
	const idNameSeparatorFormName = 'id_name_separator';
	const toggleOnClick = useCallback(() => {
		setValue(idNameSeparatorFormName, !separateName);
		if (separateName) {
			if (originalName) {
				setValue('name', '', { shouldDirty: true });
			} else {
				resetField('name');
			}
			// Making ID and name separate.
		} else if (originalName) {
			resetField('name');
		} else {
			setValue('name', serviceID, { shouldDirty: true });
		}
		setSeparateName((prev) => !prev);
	}, [serviceID, originalName, resetField, separateName, setValue]);

	const tooltip: TooltipWithAriaProps | undefined = separateName
		? {
				ariaLabel: 'Format: services.ID.NAME=service_name',
				content: (
					<pre className="m-0 whitespace-pre-wrap text-left font-mono text-xs">
						{'services:\n  '}
						<span className="font-semibold underline">ID</span>
						{':\n    '}
						<span className="font-semibold underline">NAME</span>
						{': service_name\n    latest_version: ...'}
					</pre>
				),
				type: 'element',
			}
		: undefined;

	return (
		<div className="relative mb-2 grid grid-cols-12 gap-2">
			<FieldText
				colSize={{ sm: separateName ? 6 : 12, xs: 12 }}
				label={separateName ? 'ID' : 'Name'}
				name="id"
				required
				tooltip={tooltip}
			/>
			<AdvancedToggle name={idNameSeparatorFormName} onClick={toggleOnClick} />
			{separateName && (
				<FieldText
					colSize={{ sm: 6 }}
					label="Name"
					name="name"
					required
					tooltip={{
						content: 'Name shown in the UI',
						type: 'string',
					}}
				/>
			)}
			<FieldText colSize={{ sm: 12 }} label="Comment" name="comment" />
			{loading &&
				delayedRender(() => (
					<div className="col-span-full flex flex-row items-center">
						<LoaderCircle className="h-full animate-spin" />
						<span className="pl-2">Loading...</span>
					</div>
				))}
		</div>
	);
};

export default EditServiceRoot;
