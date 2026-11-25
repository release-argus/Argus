import type { FC, ReactElement } from 'react';
import {
	BARK,
	DISCORD,
	GENERIC,
	GOOGLE_CHAT,
	GOTIFY,
	HOME_ASSISTANT,
	IFTTT,
	JOIN,
	LARK,
	MATRIX,
	MATTERMOST,
	MQTT,
	NOTIFIARR,
	NTFY,
	OPSGENIE,
	PAGERDUTY,
	PUSHBULLET,
	PUSHOVER,
	ROCKET_CHAT,
	SHOUTRRR,
	SIGNAL,
	SIGNALGRID,
	SLACK,
	SMTP,
	TEAMS,
	TELEGRAM,
	TWILIO,
	WECOM,
	XMPP,
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
	homeassistant: HOME_ASSISTANT,
	ifttt: IFTTT,
	join: JOIN,
	lark: LARK,
	matrix: MATRIX,
	mattermost: MATTERMOST,
	mqtt: MQTT,
	notifiarr: NOTIFIARR,
	ntfy: NTFY,
	opsgenie: OPSGENIE,
	pagerduty: PAGERDUTY,
	pushbullet: PUSHBULLET,
	pushover: PUSHOVER,
	rocketchat: ROCKET_CHAT,
	shoutrrr: SHOUTRRR,
	signal: SIGNAL,
	signalgrid: SIGNALGRID,
	slack: SLACK,
	smtp: SMTP,
	teams: TEAMS,
	telegram: TELEGRAM,
	twilio: TWILIO,
	wecom: WECOM,
	xmpp: XMPP,
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
