import { z } from 'zod';
import { NOTIFY_TYPE_MAP } from '@/utils/api/types/config/notify/all-types';
import {
	BarkSchemeEnum,
	BarkSoundEnum,
} from '@/utils/api/types/config-edit/notify/types/bark';
import { GenericRequestMethodZodEnum } from '@/utils/api/types/config-edit/notify/types/generic';
import {
	gotifyExtrasSchema,
	preprocessGotifyExtrasToStringWithDefaults,
} from '@/utils/api/types/config-edit/notify/types/gotify.ts';
import {
	NtfyPriorityZodEnum,
	NtfySchemeZodEnum,
	ntfyActionsSchema,
	preprocessStringFromNtfyActionsWithDefaults,
} from '@/utils/api/types/config-edit/notify/types/ntfy';
import {
	opsGenieActionsSchema,
	opsGenieTargetsSchema,
	preprocessOpsGenieTargetsToStringWithDefaults,
	preprocessStringFromOpsGenieActionsWithDefaults,
} from '@/utils/api/types/config-edit/notify/types/opsgenie';
import {
	SMTPAuthEnum,
	SMTPEncryptionEnum,
} from '@/utils/api/types/config-edit/notify/types/smtp';
import { TelegramParseModeEnum } from '@/utils/api/types/config-edit/notify/types/telegram';
import {
	headersSchema,
	preprocessStringFromHeaderArrayWithDefaults,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import {
	preprocessBooleanFromString,
	preprocessStringFromBoolean,
	preprocessStringFromZodEnum,
} from '@/utils/api/types/config-edit/shared/preprocess';
import { atLeastTwo } from '@/utils/api/types/config-edit/util';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/* Notify 'Options' Schema */
const notifyOptionsSchema = z
	.object({
		delay: z.string().default(''),
		max_tries: z.string().default(''),
		message: z.string().default(''),
	})
	.default({ delay: '', max_tries: '', message: '' });
export type NotifyOptionsSchema = z.infer<typeof notifyOptionsSchema>;

/* Base Notify Schema */
const notifyBaseSchema = z.object({
	name: z.string().min(1, REQUIRED_MESSAGE).default(''),
	old_index: z.string().nullable().default(null),
	options: notifyOptionsSchema,
	params: z.object({}).default({}),
	url_fields: z.object({}).default({}),
});

/* Bark */
export const notifyBarkSchema = notifyBaseSchema.extend({
	params: z
		.object({
			badge: z.string().default(''),
			copy: z.string().default(''),
			group: z.string().default(''),
			icon: z.string().default(''),
			scheme: BarkSchemeEnum.or(z.literal(nullString)).default(nullString),
			sound: BarkSoundEnum.or(z.literal(nullString)).default(nullString),
			title: z.string().default(''),
			url: z.string().default(''),
		})
		.default({
			badge: '',
			copy: '',
			group: '',
			icon: '',
			scheme: nullString,
			sound: nullString,
			title: '',
			url: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.BARK.value),
	url_fields: z
		.object({
			devicekey: z.string().default(''), // Required.
			host: z.string().default(''), // Required.
			path: z.string().optional(),
			port: z.string().default(''), // Required.
		})
		.default({ devicekey: '', host: '', path: '', port: '' }),
});
export type NotifyBarkSchema = z.infer<typeof notifyBarkSchema>;
export const notifyBarkSchemaOutgoing = notifyBarkSchema.extend({
	params: notifyBarkSchema.shape.params.unwrap().extend({
		scheme: preprocessStringFromZodEnum(BarkSchemeEnum),
		sound: preprocessStringFromZodEnum(BarkSoundEnum),
	}),
});

/* Discord */
export const notifyDiscordSchema = notifyBaseSchema.extend({
	params: z
		.object({
			avatar: z.string().default(''),
			splitlines: preprocessBooleanFromString,
			threadid: z.string().default(''),
			title: z.string().default(''),
			username: z.string().default(''),
		})
		.default({
			avatar: '',
			splitlines: null,
			threadid: '',
			title: '',
			username: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.DISCORD.value),
	url_fields: z
		.object({
			token: z.string().default(''), // Required.
			webhookid: z.string().default(''), // Required.
		})
		.default({ token: '', webhookid: '' }),
});
export type NotifyDiscordSchema = z.infer<typeof notifyDiscordSchema>;
export const notifyDiscordSchemaOutgoing = notifyDiscordSchema.extend({
	params: notifyDiscordSchema.shape.params.unwrap().extend({
		splitlines: preprocessStringFromBoolean,
	}),
});

/* SMTP */
export const notifySMTPSchema = notifyBaseSchema.extend({
	params: z
		.object({
			auth: SMTPAuthEnum.or(z.literal(nullString)).default(nullString),
			clienthost: z.string().default(''),
			encryption: SMTPEncryptionEnum.or(z.literal(nullString)).default(
				nullString,
			),
			fromaddress: z.string().default(''), // Required.
			fromname: z.string().default(''), // Required.
			requirestarttls: preprocessBooleanFromString,
			skiptlsverification: preprocessBooleanFromString,
			subject: z.string().default(''),
			timeout: z.string().default(''),
			toaddresses: z.string().default(''),
			usehtml: preprocessBooleanFromString,
			usestarttls: preprocessBooleanFromString,
		})
		.default({
			auth: nullString,
			clienthost: '',
			encryption: nullString,
			fromaddress: '',
			fromname: '',
			requirestarttls: null,
			skiptlsverification: null,
			subject: '',
			timeout: '',
			toaddresses: '',
			usehtml: null,
			usestarttls: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.SMTP.value),
	url_fields: z
		.object({
			host: z.string().default(''), // Required.
			password: z.string().default(''),
			port: z.string().default(''), // Required.
			username: z.string().default(''),
		})
		.default({ host: '', password: '', port: '', username: '' }),
});
export type NotifySMTPSchema = z.infer<typeof notifySMTPSchema>;
export const notifySMTPSchemaOutgoing = notifySMTPSchema.extend({
	params: notifySMTPSchema.shape.params.unwrap().extend({
		auth: preprocessStringFromZodEnum(SMTPAuthEnum),
		encryption: preprocessStringFromZodEnum(SMTPEncryptionEnum),
		requirestarttls: preprocessStringFromBoolean,
		skiptlsverification: preprocessStringFromBoolean,
		usehtml: preprocessStringFromBoolean,
		usestarttls: preprocessStringFromBoolean,
	}),
});

/* Google Chat */
export const notifyGoogleChatSchema = notifyBaseSchema.extend({
	type: z.literal(NOTIFY_TYPE_MAP.GOOGLE_CHAT.value),
	url_fields: z
		.object({
			raw: z.string().default(''),
		})
		.default({ raw: '' }),
});
export type NotifyGoogleChatSchema = z.infer<typeof notifyGoogleChatSchema>;

/* Gotify */
export const notifyGotifySchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			extras: gotifyExtrasSchema,
			insecureskipverify: preprocessBooleanFromString,
			priority: z.string().default(''),
			title: z.string().default(''),
			useheader: preprocessBooleanFromString,
		})
		.default({
			disabletls: null,
			extras: [],
			insecureskipverify: null,
			priority: '',
			title: '',
			useheader: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.GOTIFY.value),
	url_fields: z
		.object({
			host: z.string().default(''), // Required.
			path: z.string().default(''),
			port: z.string().default(''),
			token: z.string().default(''),
		})
		.default({
			host: '',
			path: '',
			port: '',
			token: '',
		}),
});
export type NotifyGotifySchema = z.infer<typeof notifyGotifySchema>;
export const notifyGotifySchemaOutgoing = notifyGotifySchema.extend({
	params: notifyGotifySchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		insecureskipverify: preprocessStringFromBoolean,
		useheader: preprocessStringFromBoolean,
	}),
});

/* IFTTT */
export const notifyIFTTTSchema = notifyBaseSchema.extend({
	params: z
		.object({
			events: z.string().default(''), // Required.
			title: z.string().default(''),
			usemessageasvalue: z.string().default(''),
			usetitleasvalue: z.string().default(''),
			value1: z.string().default(''),
			value2: z.string().default(''),
			value3: z.string().default(''),
		})
		.default({
			events: '',
			title: '',
			usemessageasvalue: '',
			usetitleasvalue: '',
			value1: '',
			value2: '',
			value3: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.IFTTT.value),
	url_fields: z
		.object({
			webhookid: z.string().default(''), // Required.
		})
		.default({ webhookid: '' }),
});
export type NotifyIFTTTSchema = z.infer<typeof notifyIFTTTSchema>;

/* Join */
export const notifyJoinSchema = notifyBaseSchema.extend({
	params: z
		.object({
			devices: z.string().default(''), // Required.
			icon: z.string().default(''),
			title: z.string().default(''),
		})
		.default({ devices: '', icon: '', title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.JOIN.value),
	url_fields: z
		.object({
			apikey: z.string().default(''), // Required.
		})
		.default({ apikey: '' }),
});
export type NotifyJoinSchema = z.infer<typeof notifyJoinSchema>;

/* MatterMost */
export const notifyMatterMostSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			icon: z.string().default(''),
		})
		.default({ disabletls: null, icon: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.MATTERMOST.value),
	url_fields: z
		.object({
			channel: z.string().default(''),
			host: z.string().default(''), // Required.
			password: z.string().default(''),
			path: z.string().default(''), // Required.
			port: z.string().default(''),
			token: z.string().default(''), // Required.
			username: z.string().default(''),
		})
		.default({
			channel: '',
			host: '',
			password: '',
			path: '',
			port: '',
			token: '',
			username: '',
		}),
});
export type NotifyMatterMostSchema = z.infer<typeof notifyMatterMostSchema>;
export const notifyMatterMostSchemaOutgoing = notifyMatterMostSchema.extend({
	params: notifyMatterMostSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
	}),
});

/* Matrix */
export const notifyMatrixSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			rooms: z.string().default(''),
			title: z.string().default(''),
		})
		.default({ disabletls: null, rooms: '', title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.MATRIX.value),
	url_fields: z
		.object({
			host: z.string().default(''), // Required.
			password: z.string().default(''), // Required.
			port: z.string().default(''), // Required.
			username: z.string().default(''),
		})
		.default({ host: '', password: '', port: '', username: '' }),
});
export type NotifyMatrixSchema = z.infer<typeof notifyMatrixSchema>;
export const notifyMatrixSchemaOutgoing = notifyMatrixSchema.extend({
	params: notifyMatrixSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
	}),
});

/* NTFY */
export const notifyNtfySchema = notifyBaseSchema.extend({
	params: z
		.object({
			actions: ntfyActionsSchema,
			attach: z.string().default(''),
			cache: preprocessBooleanFromString,
			click: z.string().default(''),
			delay: z.string().default(''),
			disabletls: preprocessBooleanFromString,
			email: z.string().default(''),
			filename: z.string().default(''),
			firebase: preprocessBooleanFromString,
			icon: z.string().default(''),
			priority: NtfyPriorityZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			scheme: NtfySchemeZodEnum.or(z.literal(nullString)).default(nullString),
			tags: z.string().default(''),
			title: z.string().default(''),
		})
		.default({
			actions: [],
			attach: '',
			cache: null,
			click: '',
			delay: '',
			disabletls: null,
			email: '',
			filename: '',
			firebase: null,
			icon: '',
			priority: nullString,
			scheme: nullString,
			tags: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.NTFY.value),
	url_fields: z
		.object({
			host: z.string().default(''),
			password: z.string().default(''),
			port: z.string().default(''),
			topic: z.string().default(''), // Required.
			username: z.string().default(''),
		})
		.default({ host: '', password: '', port: '', topic: '', username: '' }),
});
export type NotifyNtfySchema = z.infer<typeof notifyNtfySchema>;
const notifyNtfySchemaOutgoing = notifyNtfySchema.extend({
	params: notifyNtfySchema.shape.params.unwrap().extend({
		cache: preprocessStringFromBoolean,
		disabletls: preprocessStringFromBoolean,
		firebase: preprocessStringFromBoolean,
		priority: preprocessStringFromZodEnum(NtfyPriorityZodEnum),
		scheme: preprocessStringFromZodEnum(NtfySchemeZodEnum),
	}),
});

/* OpsGenie */
export const notifyOpsGenieSchema = notifyBaseSchema.extend({
	params: z
		.object({
			actions: opsGenieActionsSchema,
			alias: z.string().default(''),
			description: z.string().default(''),
			details: headersSchema,
			entity: z.string().default(''),
			note: z.string().default(''),
			priority: z.string().default(''),
			responders: opsGenieTargetsSchema,
			source: z.string().default(''),
			tags: z.string().default(''),
			title: z.string().default(''),
			user: z.string().default(''),
			visibleto: opsGenieTargetsSchema,
		})
		.default({
			actions: [],
			alias: '',
			description: '',
			details: [],
			entity: '',
			note: '',
			priority: '',
			responders: [],
			source: '',
			tags: '',
			title: '',
			user: '',
			visibleto: [],
		}),
	type: z.literal(NOTIFY_TYPE_MAP.OPSGENIE.value),
	url_fields: z
		.object({
			apikey: z.string().default(''), // Required.
			host: z.string().default(''),
			port: z.string().default(''),
		})
		.default({ apikey: '', host: '', port: '' }),
});
export type NotifyOpsGenieSchema = z.infer<typeof notifyOpsGenieSchema>;

/* Pushbullet */
export const notifyPushbulletSchema = notifyBaseSchema.extend({
	params: z
		.object({
			title: z.string().default(''),
		})
		.default({ title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.PUSHBULLET.value),
	url_fields: z
		.object({
			targets: z.string().default(''), // Required.
			token: z.string().default(''), // Required.
		})
		.default({ targets: '', token: '' }),
});
export type NotifyPushbulletSchema = z.infer<typeof notifyPushbulletSchema>;

/* Pushover */
export const notifyPushoverSchema = notifyBaseSchema.extend({
	params: z
		.object({
			devices: z.string().default(''),
			priority: z.string().default(''),
			title: z.string().default(''),
		})
		.default({ devices: '', priority: '', title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.PUSHOVER.value),
	url_fields: z
		.object({
			token: z.string().default(''), // Required.
			user: z.string().default(''), // Required.
		})
		.default({ token: '', user: '' }),
});
export type NotifyPushoverSchema = z.infer<typeof notifyPushoverSchema>;

/* Rocket.Chat */
export const notifyRocketChatSchema = notifyBaseSchema.extend({
	type: z.literal(NOTIFY_TYPE_MAP.ROCKET_CHAT.value),
	url_fields: z
		.object({
			channel: z.string().default(''), // Required.
			host: z.string().default(''), // Required.
			path: z.string().default(''), // Required.
			port: z.string().default(''),
			tokena: z.string().default(''), // Required.
			tokenb: z.string().default(''), // Required.
			username: z.string().default(''),
		})
		.default({
			channel: '',
			host: '',
			path: '',
			port: '',
			tokena: '',
			tokenb: '',
			username: '',
		}),
});
export type NotifyRocketChatSchema = z.infer<typeof notifyRocketChatSchema>;

/* Slack */
export const notifySlackSchema = notifyBaseSchema.extend({
	params: z
		.object({
			botname: z.string().default(''),
			color: z.string().default(''),
			icon: z.string().default(''),
			title: z.string().default(''),
		})
		.default({ botname: '', color: '', icon: '', title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.SLACK.value),
	url_fields: z
		.object({
			channel: z.string().default(''), // Required.
			token: z.string().default(''), // Required.
		})
		.default({ channel: '', token: '' }),
});
export type NotifySlackSchema = z.infer<typeof notifySlackSchema>;

/* Teams */
export const notifyTeamsSchema = notifyBaseSchema.extend({
	params: z
		.object({
			color: z.string().default(''),
			host: z.string().default(''),
			title: z.string().default(''),
		})
		.default({ color: '', host: '', title: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.TEAMS.value),
	url_fields: z
		.object({
			altid: z.string().default(''),
			extraid: z.string().default(''),
			group: z.string().default(''),
			groupowner: z.string().default(''),
			tenant: z.string().default(''),
		})
		.default({ altid: '', extraid: '', group: '', groupowner: '', tenant: '' }),
});
export type NotifyTeamsSchema = z.infer<typeof notifyTeamsSchema>;

/* Telegram */
export const notifyTelegramSchema = notifyBaseSchema.extend({
	params: z
		.object({
			chats: z.string().default(''), // Required.
			notification: preprocessBooleanFromString,
			parsemode: TelegramParseModeEnum.or(z.literal(nullString)).default(
				nullString,
			),
			preview: preprocessBooleanFromString,
			title: z.string().default(''),
		})
		.default({
			chats: '',
			notification: null,
			parsemode: nullString,
			preview: null,
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.TELEGRAM.value),
	url_fields: z
		.object({
			token: z.string().default(''), // Required.
		})
		.default({ token: '' }),
});
export type NotifyTelegramSchema = z.infer<typeof notifyTelegramSchema>;
export const notifyTelegramSchemaOutgoing = notifyTelegramSchema.extend({
	params: notifyTelegramSchema.shape.params.unwrap().extend({
		notification: preprocessStringFromBoolean,
		parsemode: preprocessStringFromZodEnum(TelegramParseModeEnum),
		preview: preprocessStringFromBoolean,
	}),
});

/* Zulip */
export const notifyZulipSchema = notifyBaseSchema.extend({
	params: z
		.object({
			stream: z.string().default(''),
			topic: z.string().default(''),
		})
		.default({ stream: '', topic: '' }),
	type: z.literal(NOTIFY_TYPE_MAP.ZULIP.value),
	url_fields: z
		.object({
			botkey: z.string().default(''), // Required.
			botmail: z.string().default(''), // Required.
			host: z.string().default(''), // Required.
		})
		.default({ botkey: '', botmail: '', host: '' }),
});
export type NotifyZulipSchema = z.infer<typeof notifyZulipSchema>;

/* Generic */
export const notifyGenericSchema = notifyBaseSchema.extend({
	params: z
		.object({
			contenttype: z.string().default(''),
			disabletls: preprocessBooleanFromString,
			messagekey: z.string().default(''),
			requestmethod: GenericRequestMethodZodEnum.or(
				z.literal(nullString),
			).default(nullString),
			template: z.string().default(''),
			title: z.string().default(''),
			titlekey: z.string().default(''),
		})
		.default({
			contenttype: '',
			disabletls: null,
			messagekey: '',
			requestmethod: nullString,
			template: '',
			title: '',
			titlekey: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.GENERIC.value),
	url_fields: z
		.object({
			custom_headers: headersSchema,
			host: z.string().default(''), // Required.
			json_payload_vars: headersSchema,
			path: z.string().default(''),
			port: z.string().default(''),
			query_vars: headersSchema,
		})
		.default({
			custom_headers: [],
			host: '',
			json_payload_vars: [],
			path: '',
			port: '',
			query_vars: [],
		}),
});
export type NotifyGenericSchema = z.infer<typeof notifyGenericSchema>;
const notifyGenericSchemaOutgoing = notifyGenericSchema.extend({
	params: notifyGenericSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		requestmethod: preprocessStringFromZodEnum(GenericRequestMethodZodEnum),
	}),
});

/* All */
export const notifySchemaMap = {
	bark: notifyBarkSchema,
	discord: notifyDiscordSchema,
	generic: notifyGenericSchema,
	googlechat: notifyGoogleChatSchema,
	gotify: notifyGotifySchema,
	ifttt: notifyIFTTTSchema,
	join: notifyJoinSchema,
	matrix: notifyMatrixSchema,
	mattermost: notifyMatterMostSchema,
	ntfy: notifyNtfySchema,
	opsgenie: notifyOpsGenieSchema,
	pushbullet: notifyPushbulletSchema,
	pushover: notifyPushoverSchema,
	rocketchat: notifyRocketChatSchema,
	slack: notifySlackSchema,
	smtp: notifySMTPSchema,
	teams: notifyTeamsSchema,
	telegram: notifyTelegramSchema,
	zulip: notifyZulipSchema,
} as const;
export type NotifyTypeSchema = {
	bark: NotifyBarkSchema;
	discord: NotifyDiscordSchema;
	smtp: NotifySMTPSchema;
	googlechat: NotifyGoogleChatSchema;
	gotify: NotifyGotifySchema;
	ifttt: NotifyIFTTTSchema;
	join: NotifyJoinSchema;
	mattermost: NotifyMatterMostSchema;
	matrix: NotifyMatrixSchema;
	ntfy: NotifyNtfySchema;
	opsgenie: NotifyOpsGenieSchema;
	pushbullet: NotifyPushbulletSchema;
	pushover: NotifyPushoverSchema;
	rocketchat: NotifyRocketChatSchema;
	slack: NotifySlackSchema;
	teams: NotifyTeamsSchema;
	telegram: NotifyTelegramSchema;
	zulip: NotifyZulipSchema;
	generic: NotifyGenericSchema;
};
export type NotifySchemaKeys = keyof NotifyTypeSchema;
export type NotifySchemaValues = z.infer<
	(typeof notifySchemaMap)[NotifySchemaKeys]
>;
export type NotifySchemaRecord = Record<string, NotifySchemaValues>;
export const notifySchema = z.discriminatedUnion(
	'type',
	atLeastTwo(Object.values(notifySchemaMap)),
);

export const notifyTypes: NotifySchemaKeys[] = Object.values(
	NOTIFY_TYPE_MAP,
).map((option) => option.value);
export const isNotifyType = (value: string): value is NotifySchemaKeys =>
	notifyTypes.includes(value as NotifySchemaKeys);

export const notifiersSchema = z.array(notifySchema);
export type NotifiersSchema = z.infer<typeof notifiersSchema>;

/* API Outgoing requests */

export const notifySchemaMapOutgoing = {
	...notifySchemaMap,
	bark: notifyBarkSchemaOutgoing,
	discord: notifyDiscordSchemaOutgoing,
	gotify: notifyGotifySchemaOutgoing,
	matrix: notifyMatrixSchemaOutgoing,
	mattermost: notifyMatterMostSchemaOutgoing,
	smtp: notifySMTPSchemaOutgoing,
	telegram: notifyTelegramSchemaOutgoing,
} as const;
export const notifySchemaOutgoing = z.discriminatedUnion(
	'type',
	atLeastTwo(Object.values(notifySchemaMapOutgoing)),
);
export type NotifySchemaOutgoing = z.infer<typeof notifySchemaOutgoing>;
export const notifiersSchemaOutgoing = z
	.array(notifySchemaOutgoing)
	.nullable()
	.default(null);
export type NotifiersSchemaOutgoing = z.infer<typeof notifiersSchemaOutgoing>;

/**
 * Outgoing schemas that are defaults-aware for list-like fields.
 *
 * @returns a per-type schema with the provided defaults where
 * preprocessors can null fields that match the defaults.
 */
export const notifySchemaMapOutgoingWithDefaults = (
	defaults: NotifySchemaValues,
) => {
	switch (defaults.type) {
		// Generic WebHook.
		case NOTIFY_TYPE_MAP.GENERIC.value:
			return notifyGenericSchemaOutgoing.extend({
				url_fields: notifyGenericSchemaOutgoing.shape.url_fields
					.unwrap()
					.extend({
						custom_headers: preprocessStringFromHeaderArrayWithDefaults(
							defaults?.url_fields?.custom_headers,
						),
						json_payload_vars: preprocessStringFromHeaderArrayWithDefaults(
							defaults?.url_fields?.json_payload_vars,
						),
						query_vars: preprocessStringFromHeaderArrayWithDefaults(
							defaults?.url_fields?.query_vars,
						),
					}),
			});
		// 	Gotify
		case NOTIFY_TYPE_MAP.GOTIFY.value:
			return notifyGotifySchemaOutgoing.extend({
				params: notifyGotifySchemaOutgoing.shape.params.extend({
					extras: preprocessGotifyExtrasToStringWithDefaults(
						defaults?.params?.extras,
					),
				}),
			});
		// ntfy.
		case NOTIFY_TYPE_MAP.NTFY.value:
			return notifyNtfySchemaOutgoing.extend({
				params: notifyNtfySchemaOutgoing.shape.params.extend({
					actions: preprocessStringFromNtfyActionsWithDefaults(
						defaults?.params?.actions,
					),
				}),
			});
		// OpsGenie.
		case NOTIFY_TYPE_MAP.OPSGENIE.value:
			return notifyOpsGenieSchema.extend({
				params: notifyOpsGenieSchema.shape.params.unwrap().extend({
					actions: preprocessStringFromOpsGenieActionsWithDefaults(
						defaults?.params?.actions,
					),
					details: preprocessStringFromHeaderArrayWithDefaults(
						defaults?.params?.details,
					),
					responders: preprocessOpsGenieTargetsToStringWithDefaults(
						defaults?.params?.responders,
					),
					visibleto: preprocessOpsGenieTargetsToStringWithDefaults(
						defaults?.params?.visibleto,
					),
				}),
			});
		default:
			// For types without list defaults handling, reuse existing outgoing schema.
			return notifySchemaMapOutgoing[defaults.type];
	}
};
