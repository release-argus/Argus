import type { FC } from 'react';
import BROADCAST from '@/components/modals/service-edit/notify-types/extra/ntfy/actions/types/broadcast';
import HTTP from '@/components/modals/service-edit/notify-types/extra/ntfy/actions/types/http';
import VIEW from '@/components/modals/service-edit/notify-types/extra/ntfy/actions/types/view';
import type { NtfyActionType } from '@/utils/api/types/config/notify/ntfy';
import type { NtfyActionSchema } from '@/utils/api/types/config-edit/notify/types/ntfy';

type RenderTypeProps = {
	/* The name of the field in the form. */
	name: string;
	/* Specifies the 'ntfy' field type. */
	targetType: NtfyActionType;
	/* The default values for the field. */
	defaults?: NtfyActionSchema;
};

type RenderTypeComponentsProps = Omit<RenderTypeProps, 'targetType'>;

/**
 * Mapping of 'ntfy' action types to their corresponding components.
 */
const RENDER_TYPE_COMPONENTS: Record<
	NtfyActionType,
	FC<RenderTypeComponentsProps>
> = {
	broadcast: BROADCAST,
	http: HTTP,
	view: VIEW,
};

/**
 *
 * @param name - The name of the field in the form.
 * @param targetType - Specifies the field type.
 * @param defaults - The default values for the field.
 * @returns The form fields for ntfy.params.actions.
 */
const RenderAction: FC<RenderTypeProps> = ({ name, targetType, defaults }) => {
	const RenderTypeComponent = RENDER_TYPE_COMPONENTS[targetType] as FC<{
		name: string;
		defaults?: Extract<NtfyActionSchema, { type: NtfyActionSchema }>;
	}>;

	return (
		<RenderTypeComponent
			defaults={
				defaults as Extract<NtfyActionSchema, { type: NtfyActionSchema }>
			}
			name={name}
		/>
	);
};

export default RenderAction;
