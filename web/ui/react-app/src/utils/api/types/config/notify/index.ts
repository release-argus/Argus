import type {
	NOTIFY_TYPE_MAP,
	NotifyType,
} from '@/utils/api/types/config/notify/all-types';
import type {
	BarkScheme,
	BarkSound,
} from '@/utils/api/types/config/notify/bark';
import type { GenericRequestMethod } from '@/utils/api/types/config/notify/generic';
import type {
	NTFYPriority,
	NTFYScheme,
} from '@/utils/api/types/config/notify/ntfy';
import type {
	OpsGenieAction,
	OpsGenieTargets,
} from '@/utils/api/types/config/notify/opsgenie';
import type {
	smtpAuthOptions,
	smtpEncryptionOptions,
} from '@/utils/api/types/config/notify/smtp';
import type { TelegramParsemode } from '@/utils/api/types/config/notify/telegram';
import type {
	CustomHeaders,
	EmptyObject,
} from '@/utils/api/types/config/shared';

export type NotifyTypesMap = {
	bark: NotifyBark;
	discord: NotifyDiscord;
	smtp: NotifySMTP;
	googlechat: NotifyGoogleChat;
	gotify: NotifyGotify;
	ifttt: NotifyIFTTT;
	join: NotifyJoin;
	mattermost: NotifyMatterMost;
	matrix: NotifyMatrix;
	ntfy: NotifyNtfy;
	opsgenie: NotifyOpsGenie;
	pushbullet: NotifyPushbullet;
	pushover: NotifyPushover;
	rocketchat: NotifyRocketChat;
	slack: NotifySlack;
	teams: NotifyTeams;
	telegram: NotifyTelegram;
	zulip: NotifyZulip;
	generic: NotifyGeneric;
};

export type NotifyTypesValues = NotifyTypesMap[NotifyType];
export type NotifyMap = Record<string, NotifyTypesValues>;

export type NotifyOptions = {
	delay?: string;
	max_tries?: number;
	message?: string;
};

export type NotifyBase = {
	/* Previous name */
	old_index?: string;
	/* Current name */
	name: string;

	type?: NotifyType;
	options?: NotifyOptions;
	url_fields?: EmptyObject;
	params?: EmptyObject;
};

/* Bark */
export type NotifyBark = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.BARK.value;
	url_fields?: {
		devicekey?: string;
		host?: string;
		port?: string;
		path?: string;
	};
	params?: {
		badge?: number;
		copy?: string;
		group?: string;
		icon?: string;
		scheme?: BarkScheme;
		sound?: BarkSound;
		title?: string;
		url?: string;
	};
};

/* Discord */
export type NotifyDiscord = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.DISCORD.value;
	url_fields?: {
		token?: string;
		webhookid?: string;
	};
	params?: {
		avatar?: string;
		title?: string;
		username?: string;
		splitlines?: string;
	};
};

/* SMTP */
export type NotifySMTP = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.SMTP.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		username?: string;
	};
	params?: {
		auth?: (typeof smtpAuthOptions)[number]['value'];
		clienthost?: string;
		encryption?: (typeof smtpEncryptionOptions)[number]['value'];
		fromaddress?: string;
		fromname?: string;
		subject?: string;
		toaddresses?: string;
		usehtml?: string;
		usestarttls?: string;
	};
};

/* Google Chat */
export type NotifyGoogleChat = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.GOOGLE_CHAT.value;
	url_fields?: {
		raw?: string;
	};
};

/* Gotify */
export type NotifyGotify = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.GOTIFY.value;
	url_fields?: {
		host?: string;
		port?: string;
		path?: string;
		token?: string;
	};
	params?: {
		disabletls?: string;
		priority?: string;
		title?: string;
	};
};

/* IFTTT */
export type NotifyIFTTT = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.IFTTT.value;
	url_fields?: {
		usemessageasvalue?: string;
		webhookid?: string;
	};
	params?: {
		events?: string;
		title?: string;
		usemessageasvalue?: string;
		usetitleasvalue?: string;
		value1?: string;
		value2?: string;
		value3?: string;
	};
};

/* Join */
export type NotifyJoin = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.JOIN.value;
	url_fields?: {
		apikey?: string;
	};
	params?: {
		devices?: string;
		icon?: string;
		title?: string;
	};
};

export type NotifyMatterMost = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.MATTERMOST.value;
	url_fields?: {
		channel?: string;
		host?: string;
		password?: string;
		path?: string;
		port?: string;
		token?: string;
		username?: string;
	};
	params?: {
		icon?: string;
	};
};

/* Matrix */
export type NotifyMatrix = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.MATRIX.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		username?: string;
	};
	params?: {
		disabletls?: string;
		rooms?: string;
		title?: string;
	};
};

/* NTFY */
export type NotifyNtfy = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.NTFY.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		topic?: string;
		username?: string;
	};
	params?: {
		actions?: string;
		attach?: string;
		cache?: string;
		click?: string;
		delay?: string;
		email?: string;
		filename?: string;
		firebase?: string;
		icon?: string;
		priority?: NTFYPriority;
		scheme?: NTFYScheme;
		tags?: string;
		title?: string;
	};
};

/* OpsGenie */
export type NotifyOpsGenie = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.OPSGENIE.value;
	url_fields?: {
		apikey?: string;
		host?: string;
		port?: string;
	};
	params?: {
		actions?: OpsGenieAction;
		alias?: string;
		description?: string;
		details?: CustomHeaders;
		entity?: string;
		note?: string;
		priority?: string;
		responders?: OpsGenieTargets;
		source?: string;
		tags?: string;
		title?: string;
		user?: string;
		visibleto?: OpsGenieTargets;
	};
};

/* Pushbullet */
export type NotifyPushbullet = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.PUSHBULLET.value;
	url_fields?: {
		targets?: string;
		token?: string;
	};
	params?: {
		title?: string;
	};
};

/* Pushover */
export type NotifyPushover = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.PUSHOVER.value;
	url_fields?: {
		token?: string;
		user?: string;
	};
	params?: {
		devices?: string;
		priority?: string;
		title?: string;
	};
};

/* Rocket.Chat */
export type NotifyRocketChat = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.ROCKET_CHAT.value;
	url_fields?: {
		channel?: string;
		host?: string;
		path?: string;
		port?: string;
		tokena?: string;
		tokenb?: string;
		username?: string;
	};
};

/* Slack */
export type NotifySlack = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.SLACK.value;
	url_fields?: {
		channel?: string;
		token?: string;
	};
	params?: {
		botname?: string;
		color?: string;
		icon?: string;
		title?: string;
	};
};

/* Teams */
export type NotifyTeams = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.TEAMS.value;
	url_fields?: {
		altid?: string;
		group?: string;
		groupowner: string;
		tenant?: string;
	};
	params?: {
		color?: string;
		host?: string;
		title?: string;
	};
};

/* Telegram */
export type NotifyTelegram = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.TELEGRAM.value;
	url_fields?: {
		token?: string;
	};
	params?: {
		chats?: string;
		notification?: string;
		parsemode?: TelegramParsemode;
		preview?: string;
		title?: string;
	};
};

/* Zulip */
export type NotifyZulip = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.ZULIP.value;
	url_fields?: {
		botkey?: string;
		botmail?: string;
		host?: string;
	};
	params?: {
		stream?: string;
		topic?: string;
	};
};

/* Generic */
export type NotifyGeneric = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.GENERIC.value;
	url_fields?: {
		host?: string;
		port?: string;
		path?: string;
		custom_headers?: string;
		json_payload_vars?: string;
		query_vars?: string;
	};
	params?: {
		contenttype?: string;
		disabletls?: string;
		messagekey?: string;
		requestmethod?: GenericRequestMethod;
		template?: string;
		title?: string;
		titlekey?: string;
	};
};
