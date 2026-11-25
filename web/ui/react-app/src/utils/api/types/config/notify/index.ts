import type {
	NOTIFY_TYPE_MAP,
	NotifyType,
} from '@/utils/api/types/config/notify/all-types';
import type { BarkSound } from '@/utils/api/types/config/notify/bark';
import type { GenericRequestMethod } from '@/utils/api/types/config/notify/generic';
import type {
	IFTTTMessageValue,
	IFTTTTitleValue,
} from '@/utils/api/types/config/notify/ifttt';
import type { MQTTQoS } from '@/utils/api/types/config/notify/mqtt';
import type { NTFYPriority } from '@/utils/api/types/config/notify/ntfy';
import type {
	OpsGenieAction,
	OpsGenieTargets,
} from '@/utils/api/types/config/notify/opsgenie';
import type {
	PagerDutyAction,
	PagerDutySeverity,
} from '@/utils/api/types/config/notify/pagerduty';
import type { SignalTextMode } from '@/utils/api/types/config/notify/signal';
import type { SignalgridType } from '@/utils/api/types/config/notify/signalgrid';
import type {
	smtpAuthOptions,
	smtpEncryptionOptions,
} from '@/utils/api/types/config/notify/smtp';
import type { TelegramParsemode } from '@/utils/api/types/config/notify/telegram';
import type { ZulipType } from '@/utils/api/types/config/notify/zulip';
import type { EmptyObject, Headers } from '@/utils/api/types/config/shared';

export type NotifyTypesMap = {
	bark: NotifyBark;
	discord: NotifyDiscord;
	googlechat: NotifyGoogleChat;
	gotify: NotifyGotify;
	homeassistant: NotifyHomeAssistant;
	ifttt: NotifyIFTTT;
	join: NotifyJoin;
	lark: NotifyLark;
	matrix: NotifyMatrix;
	mattermost: NotifyMatterMost;
	mqtt: NotifyMQTT;
	notifiarr: NotifyNotifiarr;
	ntfy: NotifyNtfy;
	opsgenie: NotifyOpsGenie;
	pagerduty: NotifyPagerDuty;
	pushbullet: NotifyPushbullet;
	pushover: NotifyPushover;
	rocketchat: NotifyRocketChat;
	signal: NotifySignal;
	signalgrid: NotifySignalgrid;
	slack: NotifySlack;
	smtp: NotifySMTP;
	teams: NotifyTeams;
	telegram: NotifyTelegram;
	twilio: NotifyTwilio;
	wecom: NotifyWeCom;
	xmpp: NotifyXMPP;
	zulip: NotifyZulip;
	generic: NotifyGeneric;
	shoutrrr: NotifyShoutrrr;
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
		disabletls?: string;
		group?: string;
		icon?: string;
		skiptlsverify?: string;
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
		color?: string;
		colordebug?: string;
		colorerror?: string;
		colorinfo?: string;
		colorwarn?: string;
		json?: string;
		mentions?: string;
		splitlines?: string;
		thread_id?: string;
		thread_name?: string;
		title?: string;
		username?: string;
	};
};

/* Generic */
export type NotifyGeneric = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.GENERIC.value;
	url_fields?: {
		host?: string;
		port?: string;
		path?: string;
		username?: string;
		password?: string;
		headers?: string;
		json_payload_vars?: string;
		query_vars?: string;
	};
	params?: {
		contenttype?: string;
		disabletls?: string;
		messagekey?: string;
		method?: GenericRequestMethod;
		skiptlsverify?: string;
		template?: string;
		title?: string;
		titlekey?: string;
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
		extras?: string;
		priority?: string;
		skiptlsverify?: string;
		title?: string;
		useheader?: string;
	};
};

/* Home Assistant */
export type NotifyHomeAssistant = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.HOME_ASSISTANT.value;
	url_fields?: {
		host?: string;
		port?: string;
		path?: string;
		token?: string;
	};
	params?: {
		disabletls?: string;
		nid?: string;
		service?: string;
		skiptlsverify?: string;
		targets?: string;
		title?: string;
	};
};

/* IFTTT */
export type NotifyIFTTT = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.IFTTT.value;
	url_fields?: {
		webhookid?: string;
	};
	params?: {
		events?: string;
		messagevalue?: IFTTTMessageValue;
		title?: string;
		titlevalue?: IFTTTTitleValue;
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

/* Lark */
export type NotifyLark = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.LARK.value;
	url_fields?: {
		host?: string;
		token?: string;
	};
	params?: {
		link?: string;
		secret?: string;
		title?: string;
	};
};

/* Matrix */
export type NotifyMatrix = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.MATRIX.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		user?: string;
	};
	params?: {
		disabletls?: string;
		rooms?: string;
		skiptlsverify?: string;
		title?: string;
	};
};

/* MatterMost */
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
		disabletls?: string;
		icon?: string;
		skiptlsverify?: string;
		title?: string;
	};
};

/* MQTT */
export type NotifyMQTT = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.MQTT.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		topic?: string;
		username?: string;
	};
	params?: {
		cleansession?: string;
		clientid?: string;
		disabletls?: string;
		qos?: MQTTQoS;
		retained?: string;
		skiptlsverify?: string;
	};
};

/* Notifiarr */
export type NotifyNotifiarr = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.NOTIFIARR.value;
	url_fields?: {
		apikey?: string;
	};
	params?: {
		channel?: string;
		color?: string;
		image?: string;
		name?: string;
		thumbnail?: string;
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
		disabletls_?: string;
		email?: string;
		filename?: string;
		firebase?: string;
		icon?: string;
		markdown?: string;
		priority?: NTFYPriority;
		skiptlsverify?: string;
		tags?: string;
		title?: string;
		token?: string;
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
		details?: Headers;
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

/* PagerDuty */
export type NotifyPagerDuty = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.PAGERDUTY.value;
	url_fields?: {
		host?: string;
		integrationkey?: string;
		port?: string;
	};
	params?: {
		action?: PagerDutyAction;
		client?: string;
		client_url?: string;
		contexts?: string;
		dedup_key?: string;
		details?: string;
		severity?: PagerDutySeverity;
		source?: string;
		title?: string;
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
		encryptionkey?: string;
		expire?: string;
		priority?: string;
		retry?: string;
		timestamp?: string;
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
	params?: {
		disabletls?: string;
		skiptlsverify?: string;
	};
};

/* Signal */
export type NotifySignal = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.SIGNAL.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		recipients?: string;
		source?: string;
		user?: string;
	};
	params?: {
		attachments?: string;
		disabletls?: string;
		notifyself?: string;
		skiptlsverify?: string;
		textmode?: SignalTextMode;
		title?: string;
		token?: string;
	};
};

/* Signalgrid */
export type NotifySignalgrid = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.SIGNALGRID.value;
	url_fields?: {
		channel?: string;
		clientkey?: string;
	};
	params?: {
		critical?: string;
		title?: string;
		type?: SignalgridType;
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
		thread_ts?: string;
		title?: string;
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
		clientid?: string;
		clientsecret?: string;
		encryption?: (typeof smtpEncryptionOptions)[number]['value'];
		fromaddress?: string;
		fromname?: string;
		refreshtoken?: string;
		requirestarttls?: string;
		scopes?: string;
		skiptlsverify?: string;
		subject?: string;
		timeout?: string;
		toaddresses?: string;
		tokenurl?: string;
		usehtml?: string;
		usestarttls?: string;
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
		apiurl?: string;
		chats?: string;
		notification?: string;
		parsemode?: TelegramParsemode;
		preview?: string;
		skiptlsverify?: string;
		title?: string;
	};
};

/* Twilio */
export type NotifyTwilio = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.TWILIO.value;
	url_fields?: {
		accountsid?: string;
		authtoken?: string;
		fromnumber?: string;
		tonumbers?: string;
	};
	params?: {
		apikey?: string;
		title?: string;
	};
};

/* WeCom */
export type NotifyWeCom = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.WECOM.value;
	url_fields?: {
		key?: string;
	};
	params?: {
		mentioned_list?: string;
		mentioned_mobile_list?: string;
	};
};

/* XMPP */
export type NotifyXMPP = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.XMPP.value;
	url_fields?: {
		host?: string;
		password?: string;
		port?: string;
		user?: string;
	};
	params?: {
		disabletls?: string;
		nick?: string;
		roompassword?: string;
		rooms?: string;
		skiptlsverify?: string;
		title?: string;
		to?: string;
	};
};

/* Zulip */
export type NotifyZulip = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.ZULIP.value;
	url_fields?: {
		botkey?: string;
		botmail?: string;
		host?: string;
		port?: string;
	};
	params?: {
		read_by_sender?: string;
		skiptlsverify?: string;
		stream?: string;
		title?: string;
		to?: string;
		topic?: string;
		type?: ZulipType;
	};
};

/* Shoutrrr (raw URL passthrough) */
export type NotifyShoutrrr = NotifyBase & {
	type: typeof NOTIFY_TYPE_MAP.SHOUTRRR.value;
	url_fields?: {
		raw?: string;
	};
};
