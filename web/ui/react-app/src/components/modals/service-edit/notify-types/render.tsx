import type { FC, ReactElement } from 'react';
import {
	BARK,
	DISCORD,
	GENERIC,
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
} from '@/components/modals/service-edit/notify-types';
import type { NotifyTypeSchema } from '@/utils/api/types/config-edit/notify/schemas';

type NotifyComponentProps<K extends keyof NotifyTypeSchema> = {
	name: string;
	main?: NotifyTypeSchema[K];
	defaults?: NotifyTypeSchema[K];
	hard_defaults?: NotifyTypeSchema[K];
};

const RENDER_TYPE_COMPONENTS = {
	bark: BARK,
	discord: DISCORD,
	generic: GENERIC,
	googlechat: GOOGLE_CHAT,
	gotify: GOTIFY,
	ifttt: IFTTT,
	join: JOIN,
	matrix: MATRIX,
	mattermost: MATTERMOST,
	ntfy: NTFY,
	opsgenie: OPSGENIE,
	pushbullet: PUSHBULLET,
	pushover: PUSHOVER,
	rocketchat: ROCKET_CHAT,
	slack: SLACK,
	smtp: SMTP,
	teams: TEAMS,
	telegram: TELEGRAM,
	zulip: ZULIP,
} satisfies {
	[K in keyof NotifyTypeSchema]: FC<NotifyComponentProps<K>>;
};

/**
 * The type-specific component for this 'notify'.
 *
 * @template T A key of `NotifyTypesMap`, representing the type of 'notify' to render.
 *
 * @param props The properties object to customise the notification rendering.
 * @param props.name The name identifier.
 * @param props.type The type.
 * @param props.main The `main` values.
 * @param props.defaults The `default` values.
 * @param props.hard_defaults The `hard default` values.
 *
 * @returns The rendered 'notify' accordion with the corresponding field values.
 */
const RenderNotify = <T extends keyof NotifyTypeSchema>(props: {
	name: string;
	type: T;
	main?: NotifyTypeSchema[T];
	defaults?: NotifyTypeSchema[T];
	hardDefaults?: NotifyTypeSchema[T];
}): ReactElement => {
	const { type, ...rest } = props;
	const Component = RENDER_TYPE_COMPONENTS[type] as FC<NotifyComponentProps<T>>;
	return <Component {...rest} />;
};

export default RenderNotify;
