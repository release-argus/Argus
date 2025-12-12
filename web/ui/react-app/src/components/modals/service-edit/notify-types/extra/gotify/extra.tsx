import { Trash2 } from 'lucide-react';
import { type FC, memo, useEffect } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { FieldSelect } from '@/components/generic/field';
import { AndroidAction } from '@/components/modals/service-edit/notify-types/extra/gotify/types/android-action.tsx';
import { ClientDisplay } from '@/components/modals/service-edit/notify-types/extra/gotify/types/client-display.tsx';
import { ClientNotification } from '@/components/modals/service-edit/notify-types/extra/gotify/types/client-notification.tsx';
import { Other } from '@/components/modals/service-edit/notify-types/extra/gotify/types/other.tsx';
import { Button } from '@/components/ui/button';
import {
	GOTIFY_EXTRA_NAMESPACE,
	type GotifyExtraNamespace,
	gotifyExtraNamespaceOptions,
} from '@/utils/api/types/config/notify/gotify.ts';
import type { GotifyExtraSchema } from '@/utils/api/types/config-edit/notify/types/gotify.ts';

const widerNamespaceLabels = new Set<GotifyExtraNamespace>([
	GOTIFY_EXTRA_NAMESPACE.ANDROID_ACTION.value,
	GOTIFY_EXTRA_NAMESPACE.CLIENT_NOTIFICATION.value,
]);

type ExtraProps = {
	/* The name of the field in the form. */
	name: string;
	/* The function to remove this extra. */
	removeMe: () => void;

	/* The default values for the extra. */
	defaults?: GotifyExtraSchema;
};

/**
 * Extra renders the fields for a single Extra item.
 *
 * @param name - Base path of this Extra in the form.
 * @param removeMe - Callback to remove this Extra.
 * @param defaults - Default values (if using defaults mode upstream).
 */
const Extra: FC<ExtraProps> = ({ name, removeMe, defaults }) => {
	const { setValue } = useFormContext();
	const namespace = useWatch({
		name: `${name}.namespace`,
	}) as GotifyExtraNamespace;

	// biome-ignore lint/correctness/useExhaustiveDependencies: name stable.
	useEffect(() => {
		if (namespace === GOTIFY_EXTRA_NAMESPACE.CLIENT_DISPLAY.value)
			setValue(`${name}.bigImageUrl`, '');
	}, [namespace]);

	return (
		<div className="col-span-full grid grid-cols-subgrid gap-x-2">
			<div className="col-span-1 py-1 md:pe-2">
				<Button
					aria-label="Remove this extra"
					className="size-full"
					onClick={removeMe}
					size="icon-md"
					variant="outline"
				>
					<Trash2 />
				</Button>
			</div>
			<div className="col-span-11 grid grid-cols-subgrid gap-x-2">
				<FieldSelect
					colSize={{
						md: 4,
						sm: widerNamespaceLabels.has(namespace) ? 12 : 5,
						xs: widerNamespaceLabels.has(namespace) ? 12 : 5,
					}}
					label="Namespace"
					name={`${name}.namespace`}
					options={gotifyExtraNamespaceOptions}
					tooltip={{
						ariaLabel: 'Namespace used by the client',
						content: (
							<ol>
								<li>
									client::display - Changes how the client displays information
								</li>
								<li>client::notification - Customises the notification</li>
								<li>android::action - React to events</li>
								<li>other - Defined by end-users</li>
							</ol>
						),
						type: 'element',
					}}
				/>

				{/* Namespace specific fields */}
				{namespace === 'client::display' && <ClientDisplay name={name} />}
				{namespace === 'client::notification' && (
					<ClientNotification defaults={defaults} name={name} />
				)}
				{namespace === 'android::action' && (
					<AndroidAction defaults={defaults} name={name} />
				)}
				{namespace === 'other' && <Other defaults={defaults} name={name} />}
			</div>
		</div>
	);
};

export default memo(Extra);
