export const NOTIFY_TYPE_MAP = {
	BARK: { label: 'Bark', value: 'bark' },
	DISCORD: { label: 'Discord', value: 'discord' },
	GENERIC: { label: 'Generic WebHook', value: 'generic' },
	GOOGLE_CHAT: { label: 'Google Chat', value: 'googlechat' },
	GOTIFY: { label: 'Gotify', value: 'gotify' },
	IFTTT: { label: 'IFTTT', value: 'ifttt' },
	JOIN: { label: 'Join', value: 'join' },
	MATRIX: { label: 'Matrix', value: 'matrix' },
	MATTERMOST: { label: 'MatterMost', value: 'mattermost' },
	NTFY: { label: 'Ntfy', value: 'ntfy' },
	OPSGENIE: { label: 'OpsGenie', value: 'opsgenie' },
	PUSHBULLET: { label: 'PushBullet', value: 'pushbullet' },
	PUSHOVER: { label: 'PushOver', value: 'pushover' },
	ROCKET_CHAT: { label: 'Rocket.Chat', value: 'rocketchat' },
	SLACK: { label: 'Slack', value: 'slack' },
	SMTP: { label: 'Email (SMTP)', value: 'smtp' },
	TEAMS: { label: 'Teams', value: 'teams' },
	TELEGRAM: { label: 'Telegram', value: 'telegram' },
	ZULIP: { label: 'Zulip Chat', value: 'zulip' },
} as const;
export type NotifyType =
	(typeof NOTIFY_TYPE_MAP)[keyof typeof NOTIFY_TYPE_MAP]['value'];
export const notifyTypeOptions = Object.values(NOTIFY_TYPE_MAP);
