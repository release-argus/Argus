import {
	BARK,
	DISCORD,
	GOOGLE_CHAT,
	GOTIFY,
	IFTTT,
	JOIN,
	MATRIX,
	MATTERMOST,
	NTFY,
	OPSGENIE,
	PUSHBULLET,
	PUSHOVER,
	ROCKET_CHAT,
	SLACK,
	SMTP,
	TEAMS,
	TELEGRAM,
	ZULIP,
} from 'components/modals/service-edit/notify-types';
import { FC, memo } from 'react';
import { NotifyTypesKeys, NotifyTypesValues } from 'types/config';

import GENERIC from './generic';

interface RenderTypeProps {
	name: string;
	type: NotifyTypesKeys;
	main?: NotifyTypesValues;
	defaults?: NotifyTypesValues;
	hard_defaults?: NotifyTypesValues;
}

const RENDER_TYPE_COMPONENTS: {
	[key in NotifyTypesKeys]: FC<{
		name: string;
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		main: any;
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		defaults: any;
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		hard_defaults: any;
	}>;
} = {
	bark: BARK,
	discord: DISCORD,
	smtp: SMTP,
	googlechat: GOOGLE_CHAT,
	gotify: GOTIFY,
	ifttt: IFTTT,
	join: JOIN,
	mattermost: MATTERMOST,
	matrix: MATRIX,
	ntfy: NTFY,
	opsgenie: OPSGENIE,
	pushbullet: PUSHBULLET,
	pushover: PUSHOVER,
	rocketchat: ROCKET_CHAT,
	slack: SLACK,
	teams: TEAMS,
	telegram: TELEGRAM,
	zulip: ZULIP,
	generic: GENERIC,
};

/**
 * The type-specific component for this notify.
 *
 * @param name - The name of the notifier in the form.
 * @param type - The type of the notifier.
 * @param main - The main notify options.
 * @param defaults - The default values for all notify types.
 * @param hard_defaults - The hard default values for all notify types.
 * @returns The type-specific component for this notify type.
 */
const RenderNotify: FC<RenderTypeProps> = ({
	name,
	type,
	main,
	defaults,
	hard_defaults,
}) => {
	const RenderTypeComponent = RENDER_TYPE_COMPONENTS[type || 'discord'];

	// Unknown type
	if (!RenderTypeComponent) return null;

	return (
		<RenderTypeComponent
			name={name}
			main={main}
			defaults={defaults}
			hard_defaults={hard_defaults}
		/>
	);
};

export default memo(RenderNotify);
