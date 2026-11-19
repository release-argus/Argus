import { z } from 'zod';
import { NOTIFY_TYPE_MAP } from '@/utils/api/types/config/notify/all-types';
import { BarkSoundEnum } from '@/utils/api/types/config-edit/notify/types/bark';
import { GenericRequestMethodZodEnum } from '@/utils/api/types/config-edit/notify/types/generic';
import {
	gotifyExtrasSchema,
	preprocessGotifyExtrasToStringWithDefaults,
} from '@/utils/api/types/config-edit/notify/types/gotify.ts';
import {
	IFTTTMessageValueEnum,
	IFTTTTitleValueEnum,
} from '@/utils/api/types/config-edit/notify/types/ifttt';
import { MQTTQoSZodEnum } from '@/utils/api/types/config-edit/notify/types/mqtt';
import {
	NtfyPriorityZodEnum,
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
	PagerDutyActionZodEnum,
	PagerDutySeverityZodEnum,
} from '@/utils/api/types/config-edit/notify/types/pagerduty';
import { SignalTextModeZodEnum } from '@/utils/api/types/config-edit/notify/types/signal';
import { SignalgridTypeZodEnum } from '@/utils/api/types/config-edit/notify/types/signalgrid';
import {
	SMTPAuthEnum,
	SMTPEncryptionEnum,
} from '@/utils/api/types/config-edit/notify/types/smtp';
import { TelegramParseModeEnum } from '@/utils/api/types/config-edit/notify/types/telegram';
import { ZulipTypeZodEnum } from '@/utils/api/types/config-edit/notify/types/zulip';
import {
	headersSchema,
	preprocessStringFromHeaderArrayWithDefaults,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import {
	preprocessBooleanFromString,
	preprocessStringFromBoolean,
	preprocessStringFromZodEnum,
	stringDefault,
} from '@/utils/api/types/config-edit/shared/preprocess';
import { atLeastTwo } from '@/utils/api/types/config-edit/util';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

/* Notify 'Options' Schema */
const notifyOptionsSchema = z
	.object({
		delay: stringDefault,
		max_tries: stringDefault,
		message: stringDefault,
	})
	.default({
		delay: '',
		max_tries: '',
		message: '',
	});
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
			badge: stringDefault,
			copy: stringDefault,
			disabletls: preprocessBooleanFromString,
			group: stringDefault,
			icon: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			sound: BarkSoundEnum.or(z.literal(nullString)).default(nullString),
			title: stringDefault,
			url: stringDefault,
		})
		.default({
			badge: '',
			copy: '',
			disabletls: null,
			group: '',
			icon: '',
			skiptlsverify: null,
			sound: nullString,
			title: '',
			url: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.BARK.value),
	url_fields: z
		.object({
			devicekey: stringDefault, // Required.
			host: stringDefault, // Required.
			path: stringDefault,
			port: stringDefault, // Required.
		})
		.default({
			devicekey: '',
			host: '',
			path: '',
			port: '',
		}),
});
export type NotifyBarkSchema = z.infer<typeof notifyBarkSchema>;
export const notifyBarkSchemaOutgoing = notifyBarkSchema.extend({
	params: notifyBarkSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
		sound: preprocessStringFromZodEnum(BarkSoundEnum),
	}),
});

/* Discord */
export const notifyDiscordSchema = notifyBaseSchema.extend({
	params: z
		.object({
			avatar: stringDefault,
			color: stringDefault,
			colordebug: stringDefault,
			colorerror: stringDefault,
			colorinfo: stringDefault,
			colorwarn: stringDefault,
			json: preprocessBooleanFromString,
			mentions: stringDefault,
			splitlines: preprocessBooleanFromString,
			thread_id: stringDefault,
			thread_name: stringDefault,
			title: stringDefault,
			username: stringDefault,
		})
		.default({
			avatar: '',
			color: '',
			colordebug: '',
			colorerror: '',
			colorinfo: '',
			colorwarn: '',
			json: null,
			mentions: '',
			splitlines: null,
			thread_id: '',
			thread_name: '',
			title: '',
			username: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.DISCORD.value),
	url_fields: z
		.object({
			token: stringDefault, // Required.
			webhookid: stringDefault, // Required.
		})
		.default({
			token: '',
			webhookid: '',
		}),
});
export type NotifyDiscordSchema = z.infer<typeof notifyDiscordSchema>;
export const notifyDiscordSchemaOutgoing = notifyDiscordSchema.extend({
	params: notifyDiscordSchema.shape.params.unwrap().extend({
		json: preprocessStringFromBoolean,
		splitlines: preprocessStringFromBoolean,
	}),
});

/* Generic */
export const notifyGenericSchema = notifyBaseSchema.extend({
	params: z
		.object({
			contenttype: stringDefault,
			disabletls: preprocessBooleanFromString,
			messagekey: stringDefault,
			method: GenericRequestMethodZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			skiptlsverify: preprocessBooleanFromString,
			template: stringDefault,
			title: stringDefault,
			titlekey: stringDefault,
		})
		.default({
			contenttype: '',
			disabletls: null,
			messagekey: '',
			method: nullString,
			skiptlsverify: null,
			template: '',
			title: '',
			titlekey: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.GENERIC.value),
	url_fields: z
		.object({
			headers: headersSchema,
			host: stringDefault, // Required.
			json_payload_vars: headersSchema,
			path: stringDefault,
			port: stringDefault,
			query_vars: headersSchema,
		})
		.default({
			headers: [],
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
		method: preprocessStringFromZodEnum(GenericRequestMethodZodEnum),
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* Google Chat */
export const notifyGoogleChatSchema = notifyBaseSchema.extend({
	type: z.literal(NOTIFY_TYPE_MAP.GOOGLE_CHAT.value),
	url_fields: z
		.object({
			raw: stringDefault,
		})
		.default({
			raw: '',
		}),
});
export type NotifyGoogleChatSchema = z.infer<typeof notifyGoogleChatSchema>;

/* Gotify */
export const notifyGotifySchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			extras: gotifyExtrasSchema,
			priority: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			title: stringDefault,
			useheader: preprocessBooleanFromString,
		})
		.default({
			disabletls: null,
			extras: [],
			priority: '',
			skiptlsverify: null,
			title: '',
			useheader: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.GOTIFY.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			path: stringDefault,
			port: stringDefault,
			token: stringDefault,
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
		skiptlsverify: preprocessStringFromBoolean,
		useheader: preprocessStringFromBoolean,
	}),
});

/* Home Assistant */
export const notifyHomeAssistantSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			nid: stringDefault,
			service: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			targets: stringDefault,
			title: stringDefault,
		})
		.default({
			disabletls: null,
			nid: '',
			service: '',
			skiptlsverify: null,
			targets: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.HOME_ASSISTANT.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			path: stringDefault,
			port: stringDefault,
			token: stringDefault, // Required.
		})
		.default({
			host: '',
			path: '',
			port: '',
			token: '',
		}),
});
export type NotifyHomeAssistantSchema = z.infer<
	typeof notifyHomeAssistantSchema
>;
export const notifyHomeAssistantSchemaOutgoing =
	notifyHomeAssistantSchema.extend({
		params: notifyHomeAssistantSchema.shape.params.unwrap().extend({
			disabletls: preprocessStringFromBoolean,
			skiptlsverify: preprocessStringFromBoolean,
		}),
	});

/* IFTTT */
export const notifyIFTTTSchema = notifyBaseSchema.extend({
	params: z
		.object({
			events: stringDefault, // Required.
			messagevalue: IFTTTMessageValueEnum.or(z.literal(nullString)).default(
				nullString,
			),
			title: stringDefault,
			titlevalue: IFTTTTitleValueEnum.or(z.literal(nullString)).default(
				nullString,
			),
			value1: stringDefault,
			value2: stringDefault,
			value3: stringDefault,
		})
		.default({
			events: '',
			messagevalue: nullString,
			title: '',
			titlevalue: nullString,
			value1: '',
			value2: '',
			value3: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.IFTTT.value),
	url_fields: z
		.object({
			webhookid: stringDefault, // Required.
		})
		.default({
			webhookid: '',
		}),
});
export type NotifyIFTTTSchema = z.infer<typeof notifyIFTTTSchema>;
export const notifyIFTTTSchemaOutgoing = notifyIFTTTSchema.extend({
	params: notifyIFTTTSchema.shape.params.unwrap().extend({
		messagevalue: preprocessStringFromZodEnum(IFTTTMessageValueEnum),
		titlevalue: preprocessStringFromZodEnum(IFTTTTitleValueEnum),
	}),
});

/* Join */
export const notifyJoinSchema = notifyBaseSchema.extend({
	params: z
		.object({
			devices: stringDefault, // Required.
			icon: stringDefault,
			title: stringDefault,
		})
		.default({
			devices: '',
			icon: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.JOIN.value),
	url_fields: z
		.object({
			apikey: stringDefault, // Required.
		})
		.default({
			apikey: '',
		}),
});
export type NotifyJoinSchema = z.infer<typeof notifyJoinSchema>;

/* Lark */
export const notifyLarkSchema = notifyBaseSchema.extend({
	params: z
		.object({
			link: stringDefault,
			secret: stringDefault,
			title: stringDefault,
		})
		.default({
			link: '',
			secret: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.LARK.value),
	url_fields: z
		.object({
			host: stringDefault,
			token: stringDefault, // Required.
		})
		.default({
			host: '',
			token: '',
		}),
});
export type NotifyLarkSchema = z.infer<typeof notifyLarkSchema>;

/* Matrix */
export const notifyMatrixSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			rooms: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			title: stringDefault,
		})
		.default({
			disabletls: null,
			rooms: '',
			skiptlsverify: null,
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.MATRIX.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			password: stringDefault, // Required.
			port: stringDefault,
			user: stringDefault,
		})
		.default({
			host: '',
			password: '',
			port: '',
			user: '',
		}),
});
export type NotifyMatrixSchema = z.infer<typeof notifyMatrixSchema>;
export const notifyMatrixSchemaOutgoing = notifyMatrixSchema.extend({
	params: notifyMatrixSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* MatterMost */
export const notifyMatterMostSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			icon: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			title: stringDefault,
		})
		.default({
			disabletls: null,
			icon: '',
			skiptlsverify: null,
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.MATTERMOST.value),
	url_fields: z
		.object({
			channel: stringDefault,
			host: stringDefault, // Required.
			password: stringDefault,
			port: stringDefault,
			token: stringDefault, // Required.
			username: stringDefault,
		})
		.default({
			channel: '',
			host: '',
			password: '',
			port: '',
			token: '',
			username: '',
		}),
});
export type NotifyMatterMostSchema = z.infer<typeof notifyMatterMostSchema>;
export const notifyMatterMostSchemaOutgoing = notifyMatterMostSchema.extend({
	params: notifyMatterMostSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* MQTT */
export const notifyMQTTSchema = notifyBaseSchema.extend({
	params: z
		.object({
			cleansession: preprocessBooleanFromString,
			clientid: stringDefault,
			disabletls: preprocessBooleanFromString,
			qos: MQTTQoSZodEnum.or(z.literal(nullString)).default(nullString),
			retained: preprocessBooleanFromString,
			skiptlsverify: preprocessBooleanFromString,
		})
		.default({
			cleansession: null,
			clientid: '',
			disabletls: null,
			qos: nullString,
			retained: null,
			skiptlsverify: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.MQTT.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			password: stringDefault,
			port: stringDefault,
			topic: stringDefault, // Required.
			username: stringDefault,
		})
		.default({
			host: '',
			password: '',
			port: '',
			topic: '',
			username: '',
		}),
});
export type NotifyMQTTSchema = z.infer<typeof notifyMQTTSchema>;
export const notifyMQTTSchemaOutgoing = notifyMQTTSchema.extend({
	params: notifyMQTTSchema.shape.params.unwrap().extend({
		cleansession: preprocessStringFromBoolean,
		disabletls: preprocessStringFromBoolean,
		qos: preprocessStringFromZodEnum(MQTTQoSZodEnum),
		retained: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* Notifiarr */
export const notifyNotifiarrSchema = notifyBaseSchema.extend({
	params: z
		.object({
			channel: stringDefault,
			color: stringDefault,
			image: stringDefault,
			name: stringDefault,
			thumbnail: stringDefault,
		})
		.default({
			channel: '',
			color: '',
			image: '',
			name: '',
			thumbnail: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.NOTIFIARR.value),
	url_fields: z
		.object({
			apikey: stringDefault, // Required.
		})
		.default({
			apikey: '',
		}),
});
export type NotifyNotifiarrSchema = z.infer<typeof notifyNotifiarrSchema>;

/* NTFY */
export const notifyNtfySchema = notifyBaseSchema.extend({
	params: z
		.object({
			actions: ntfyActionsSchema,
			attach: stringDefault,
			cache: preprocessBooleanFromString,
			click: stringDefault,
			delay: stringDefault,
			disabletls_: preprocessBooleanFromString,
			email: stringDefault,
			filename: stringDefault,
			firebase: preprocessBooleanFromString,
			icon: stringDefault,
			markdown: preprocessBooleanFromString,
			priority: NtfyPriorityZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			skiptlsverify: preprocessBooleanFromString,
			tags: stringDefault,
			title: stringDefault,
			token: stringDefault,
		})
		.default({
			actions: [],
			attach: '',
			cache: null,
			click: '',
			delay: '',
			disabletls_: null,
			email: '',
			filename: '',
			firebase: null,
			icon: '',
			markdown: null,
			priority: nullString,
			skiptlsverify: null,
			tags: '',
			title: '',
			token: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.NTFY.value),
	url_fields: z
		.object({
			host: stringDefault,
			password: stringDefault,
			port: stringDefault,
			topic: stringDefault, // Required.
			username: stringDefault,
		})
		.default({
			host: '',
			password: '',
			port: '',
			topic: '',
			username: '',
		}),
});
export type NotifyNtfySchema = z.infer<typeof notifyNtfySchema>;
export const notifyNtfySchemaOutgoing = notifyNtfySchema.extend({
	params: notifyNtfySchema.shape.params.unwrap().extend({
		cache: preprocessStringFromBoolean,
		disabletls_: preprocessStringFromBoolean,
		firebase: preprocessStringFromBoolean,
		markdown: preprocessStringFromBoolean,
		priority: preprocessStringFromZodEnum(NtfyPriorityZodEnum),
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* OpsGenie */
export const notifyOpsGenieSchema = notifyBaseSchema.extend({
	params: z
		.object({
			actions: opsGenieActionsSchema,
			alias: stringDefault,
			description: stringDefault,
			details: headersSchema,
			entity: stringDefault,
			note: stringDefault,
			priority: stringDefault,
			responders: opsGenieTargetsSchema,
			source: stringDefault,
			tags: stringDefault,
			title: stringDefault,
			user: stringDefault,
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
			apikey: stringDefault, // Required.
			host: stringDefault,
			port: stringDefault,
		})
		.default({
			apikey: '',
			host: '',
			port: '',
		}),
});
export type NotifyOpsGenieSchema = z.infer<typeof notifyOpsGenieSchema>;

/* PagerDuty */
export const notifyPagerDutySchema = notifyBaseSchema.extend({
	params: z
		.object({
			action: PagerDutyActionZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			client: stringDefault,
			client_url: stringDefault,
			contexts: stringDefault,
			dedup_key: stringDefault,
			details: stringDefault,
			severity: PagerDutySeverityZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			source: stringDefault,
			title: stringDefault,
		})
		.default({
			action: nullString,
			client: '',
			client_url: '',
			contexts: '',
			dedup_key: '',
			details: '',
			severity: nullString,
			source: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.PAGERDUTY.value),
	url_fields: z
		.object({
			host: stringDefault,
			integrationkey: stringDefault, // Required.
			port: stringDefault,
		})
		.default({
			host: '',
			integrationkey: '',
			port: '',
		}),
});
export type NotifyPagerDutySchema = z.infer<typeof notifyPagerDutySchema>;
export const notifyPagerDutySchemaOutgoing = notifyPagerDutySchema.extend({
	params: notifyPagerDutySchema.shape.params.unwrap().extend({
		action: preprocessStringFromZodEnum(PagerDutyActionZodEnum),
		severity: preprocessStringFromZodEnum(PagerDutySeverityZodEnum),
	}),
});

/* Pushbullet */
export const notifyPushbulletSchema = notifyBaseSchema.extend({
	params: z
		.object({
			title: stringDefault,
		})
		.default({
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.PUSHBULLET.value),
	url_fields: z
		.object({
			targets: stringDefault, // Required.
			token: stringDefault, // Required.
		})
		.default({
			targets: '',
			token: '',
		}),
});
export type NotifyPushbulletSchema = z.infer<typeof notifyPushbulletSchema>;

/* Pushover */
export const notifyPushoverSchema = notifyBaseSchema.extend({
	params: z
		.object({
			devices: stringDefault,
			encryptionkey: stringDefault,
			expire: stringDefault,
			priority: stringDefault,
			retry: stringDefault,
			timestamp: stringDefault,
			title: stringDefault,
		})
		.default({
			devices: '',
			encryptionkey: '',
			expire: '',
			priority: '',
			retry: '',
			timestamp: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.PUSHOVER.value),
	url_fields: z
		.object({
			token: stringDefault, // Required.
			user: stringDefault, // Required.
		})
		.default({
			token: '',
			user: '',
		}),
});
export type NotifyPushoverSchema = z.infer<typeof notifyPushoverSchema>;

/* Rocket.Chat */
export const notifyRocketChatSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			skiptlsverify: preprocessBooleanFromString,
		})
		.default({
			disabletls: null,
			skiptlsverify: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.ROCKET_CHAT.value),
	url_fields: z
		.object({
			channel: stringDefault,
			host: stringDefault, // Required.
			port: stringDefault,
			tokena: stringDefault, // Required.
			tokenb: stringDefault, // Required.
			username: stringDefault,
		})
		.default({
			channel: '',
			host: '',
			port: '',
			tokena: '',
			tokenb: '',
			username: '',
		}),
});
export type NotifyRocketChatSchema = z.infer<typeof notifyRocketChatSchema>;
export const notifyRocketChatSchemaOutgoing = notifyRocketChatSchema.extend({
	params: notifyRocketChatSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* Signal */
export const notifySignalSchema = notifyBaseSchema.extend({
	params: z
		.object({
			attachments: stringDefault,
			disabletls: preprocessBooleanFromString,
			notifyself: preprocessBooleanFromString,
			skiptlsverify: preprocessBooleanFromString,
			textmode: SignalTextModeZodEnum.or(z.literal(nullString)).default(
				nullString,
			),
			title: stringDefault,
			token: stringDefault,
		})
		.default({
			attachments: '',
			disabletls: null,
			notifyself: null,
			skiptlsverify: null,
			textmode: nullString,
			title: '',
			token: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.SIGNAL.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			password: stringDefault,
			port: stringDefault,
			recipients: stringDefault, // Required.
			source: stringDefault, // Required.
			user: stringDefault,
		})
		.default({
			host: '',
			password: '',
			port: '',
			recipients: '',
			source: '',
			user: '',
		}),
});
export type NotifySignalSchema = z.infer<typeof notifySignalSchema>;
export const notifySignalSchemaOutgoing = notifySignalSchema.extend({
	params: notifySignalSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		notifyself: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
		textmode: preprocessStringFromZodEnum(SignalTextModeZodEnum),
	}),
});

/* Signalgrid */
export const notifySignalgridSchema = notifyBaseSchema.extend({
	params: z
		.object({
			critical: preprocessBooleanFromString,
			title: stringDefault,
			type: SignalgridTypeZodEnum.or(z.literal(nullString)).default(nullString),
		})
		.default({
			critical: null,
			title: '',
			type: nullString,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.SIGNALGRID.value),
	url_fields: z
		.object({
			channel: stringDefault, // Required.
			clientkey: stringDefault, // Required.
		})
		.default({
			channel: '',
			clientkey: '',
		}),
});
export type NotifySignalgridSchema = z.infer<typeof notifySignalgridSchema>;
export const notifySignalgridSchemaOutgoing = notifySignalgridSchema.extend({
	params: notifySignalgridSchema.shape.params.unwrap().extend({
		critical: preprocessStringFromBoolean,
		type: preprocessStringFromZodEnum(SignalgridTypeZodEnum),
	}),
});

/* Slack */
export const notifySlackSchema = notifyBaseSchema.extend({
	params: z
		.object({
			botname: stringDefault,
			color: stringDefault,
			icon: stringDefault,
			thread_ts: stringDefault,
			title: stringDefault,
		})
		.default({
			botname: '',
			color: '',
			icon: '',
			thread_ts: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.SLACK.value),
	url_fields: z
		.object({
			channel: stringDefault, // Required.
			token: stringDefault, // Required.
		})
		.default({
			channel: '',
			token: '',
		}),
});
export type NotifySlackSchema = z.infer<typeof notifySlackSchema>;

/* SMTP */
export const notifySMTPSchema = notifyBaseSchema.extend({
	params: z
		.object({
			auth: SMTPAuthEnum.or(z.literal(nullString)).default(nullString),
			clienthost: stringDefault,
			clientid: stringDefault,
			clientsecret: stringDefault,
			encryption: SMTPEncryptionEnum.or(z.literal(nullString)).default(
				nullString,
			),
			fromaddress: stringDefault, // Required.
			fromname: stringDefault,
			refreshtoken: stringDefault,
			requirestarttls: preprocessBooleanFromString,
			scopes: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			subject: stringDefault,
			timeout: stringDefault,
			toaddresses: stringDefault, // Required.
			tokenurl: stringDefault,
			usehtml: preprocessBooleanFromString,
			usestarttls: preprocessBooleanFromString,
		})
		.default({
			auth: nullString,
			clienthost: '',
			clientid: '',
			clientsecret: '',
			encryption: nullString,
			fromaddress: '',
			fromname: '',
			refreshtoken: '',
			requirestarttls: null,
			scopes: '',
			skiptlsverify: null,
			subject: '',
			timeout: '',
			toaddresses: '',
			tokenurl: '',
			usehtml: null,
			usestarttls: null,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.SMTP.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			password: stringDefault,
			port: stringDefault, // Required.
			username: stringDefault,
		})
		.default({
			host: '',
			password: '',
			port: '',
			username: '',
		}),
});
export type NotifySMTPSchema = z.infer<typeof notifySMTPSchema>;
export const notifySMTPSchemaOutgoing = notifySMTPSchema.extend({
	params: notifySMTPSchema.shape.params.unwrap().extend({
		auth: preprocessStringFromZodEnum(SMTPAuthEnum),
		encryption: preprocessStringFromZodEnum(SMTPEncryptionEnum),
		requirestarttls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
		usehtml: preprocessStringFromBoolean,
		usestarttls: preprocessStringFromBoolean,
	}),
});

/* Teams (Power Automate) */
export const notifyTeamsSchema = notifyBaseSchema.extend({
	params: z
		.object({
			color: stringDefault,
			host: stringDefault, // Required.
			title: stringDefault,
		})
		.default({
			color: '',
			host: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.TEAMS.value),
});
export type NotifyTeamsSchema = z.infer<typeof notifyTeamsSchema>;

/* Telegram */
export const notifyTelegramSchema = notifyBaseSchema.extend({
	params: z
		.object({
			apiurl: stringDefault,
			chats: stringDefault, // Required.
			notification: preprocessBooleanFromString,
			parsemode: TelegramParseModeEnum.or(z.literal(nullString)).default(
				nullString,
			),
			preview: preprocessBooleanFromString,
			skiptlsverify: preprocessBooleanFromString,
			title: stringDefault,
		})
		.default({
			apiurl: '',
			chats: '',
			notification: null,
			parsemode: nullString,
			preview: null,
			skiptlsverify: null,
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.TELEGRAM.value),
	url_fields: z
		.object({
			token: stringDefault, // Required.
		})
		.default({
			token: '',
		}),
});
export type NotifyTelegramSchema = z.infer<typeof notifyTelegramSchema>;
export const notifyTelegramSchemaOutgoing = notifyTelegramSchema.extend({
	params: notifyTelegramSchema.shape.params.unwrap().extend({
		notification: preprocessStringFromBoolean,
		parsemode: preprocessStringFromZodEnum(TelegramParseModeEnum),
		preview: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* Twilio */
export const notifyTwilioSchema = notifyBaseSchema.extend({
	params: z
		.object({
			apikey: stringDefault,
			title: stringDefault,
		})
		.default({
			apikey: '',
			title: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.TWILIO.value),
	url_fields: z
		.object({
			accountsid: stringDefault, // Required.
			authtoken: stringDefault, // Required.
			fromnumber: stringDefault, // Required.
			tonumbers: stringDefault, // Required.
		})
		.default({
			accountsid: '',
			authtoken: '',
			fromnumber: '',
			tonumbers: '',
		}),
});
export type NotifyTwilioSchema = z.infer<typeof notifyTwilioSchema>;

/* WeCom */
export const notifyWeComSchema = notifyBaseSchema.extend({
	params: z
		.object({
			mentioned_list: stringDefault,
			mentioned_mobile_list: stringDefault,
		})
		.default({
			mentioned_list: '',
			mentioned_mobile_list: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.WECOM.value),
	url_fields: z
		.object({
			key: stringDefault, // Required.
		})
		.default({
			key: '',
		}),
});
export type NotifyWeComSchema = z.infer<typeof notifyWeComSchema>;

/* XMPP */
export const notifyXMPPSchema = notifyBaseSchema.extend({
	params: z
		.object({
			disabletls: preprocessBooleanFromString,
			nick: stringDefault,
			roompassword: stringDefault,
			rooms: stringDefault,
			skiptlsverify: preprocessBooleanFromString,
			title: stringDefault,
			to: stringDefault,
		})
		.default({
			disabletls: null,
			nick: '',
			roompassword: '',
			rooms: '',
			skiptlsverify: null,
			title: '',
			to: '',
		}),
	type: z.literal(NOTIFY_TYPE_MAP.XMPP.value),
	url_fields: z
		.object({
			host: stringDefault, // Required.
			password: stringDefault, // Required.
			port: stringDefault,
			user: stringDefault, // Required.
		})
		.default({
			host: '',
			password: '',
			port: '',
			user: '',
		}),
});
export type NotifyXMPPSchema = z.infer<typeof notifyXMPPSchema>;
export const notifyXMPPSchemaOutgoing = notifyXMPPSchema.extend({
	params: notifyXMPPSchema.shape.params.unwrap().extend({
		disabletls: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
	}),
});

/* Zulip */
export const notifyZulipSchema = notifyBaseSchema.extend({
	params: z
		.object({
			read_by_sender: preprocessBooleanFromString,
			skiptlsverify: preprocessBooleanFromString,
			stream: stringDefault,
			title: stringDefault,
			to: stringDefault,
			topic: stringDefault,
			type: ZulipTypeZodEnum.or(z.literal(nullString)).default(nullString),
		})
		.default({
			read_by_sender: null,
			skiptlsverify: null,
			stream: '',
			title: '',
			to: '',
			topic: '',
			type: nullString,
		}),
	type: z.literal(NOTIFY_TYPE_MAP.ZULIP.value),
	url_fields: z
		.object({
			botkey: stringDefault, // Required.
			botmail: stringDefault, // Required.
			host: stringDefault, // Required.
			port: stringDefault,
		})
		.default({
			botkey: '',
			botmail: '',
			host: '',
			port: '',
		}),
});
export type NotifyZulipSchema = z.infer<typeof notifyZulipSchema>;
export const notifyZulipSchemaOutgoing = notifyZulipSchema.extend({
	params: notifyZulipSchema.shape.params.unwrap().extend({
		read_by_sender: preprocessStringFromBoolean,
		skiptlsverify: preprocessStringFromBoolean,
		type: preprocessStringFromZodEnum(ZulipTypeZodEnum),
	}),
});

/* Shoutrrr (raw URL — hidden/deprecated, for legacy configs) */
export const notifyShoutrrrSchema = notifyBaseSchema.extend({
	type: z.literal(NOTIFY_TYPE_MAP.SHOUTRRR.value),
	url_fields: z
		.object({
			raw: stringDefault, // Required: full shoutrrr URL.
		})
		.default({
			raw: '',
		}),
});
export type NotifyShoutrrrSchema = z.infer<typeof notifyShoutrrrSchema>;

/* All */
export const notifySchemaMap = {
	bark: notifyBarkSchema,
	discord: notifyDiscordSchema,
	generic: notifyGenericSchema,
	googlechat: notifyGoogleChatSchema,
	gotify: notifyGotifySchema,
	homeassistant: notifyHomeAssistantSchema,
	ifttt: notifyIFTTTSchema,
	join: notifyJoinSchema,
	lark: notifyLarkSchema,
	matrix: notifyMatrixSchema,
	mattermost: notifyMatterMostSchema,
	mqtt: notifyMQTTSchema,
	notifiarr: notifyNotifiarrSchema,
	ntfy: notifyNtfySchema,
	opsgenie: notifyOpsGenieSchema,
	pagerduty: notifyPagerDutySchema,
	pushbullet: notifyPushbulletSchema,
	pushover: notifyPushoverSchema,
	rocketchat: notifyRocketChatSchema,
	shoutrrr: notifyShoutrrrSchema,
	signal: notifySignalSchema,
	signalgrid: notifySignalgridSchema,
	slack: notifySlackSchema,
	smtp: notifySMTPSchema,
	teams: notifyTeamsSchema,
	telegram: notifyTelegramSchema,
	twilio: notifyTwilioSchema,
	wecom: notifyWeComSchema,
	xmpp: notifyXMPPSchema,
	zulip: notifyZulipSchema,
} as const;
export type NotifyTypeSchema = {
	bark: NotifyBarkSchema;
	discord: NotifyDiscordSchema;
	generic: NotifyGenericSchema;
	googlechat: NotifyGoogleChatSchema;
	gotify: NotifyGotifySchema;
	homeassistant: NotifyHomeAssistantSchema;
	ifttt: NotifyIFTTTSchema;
	join: NotifyJoinSchema;
	lark: NotifyLarkSchema;
	matrix: NotifyMatrixSchema;
	mattermost: NotifyMatterMostSchema;
	mqtt: NotifyMQTTSchema;
	notifiarr: NotifyNotifiarrSchema;
	ntfy: NotifyNtfySchema;
	opsgenie: NotifyOpsGenieSchema;
	pagerduty: NotifyPagerDutySchema;
	pushbullet: NotifyPushbulletSchema;
	pushover: NotifyPushoverSchema;
	rocketchat: NotifyRocketChatSchema;
	shoutrrr: NotifyShoutrrrSchema;
	signal: NotifySignalSchema;
	signalgrid: NotifySignalgridSchema;
	slack: NotifySlackSchema;
	smtp: NotifySMTPSchema;
	teams: NotifyTeamsSchema;
	telegram: NotifyTelegramSchema;
	twilio: NotifyTwilioSchema;
	wecom: NotifyWeComSchema;
	xmpp: NotifyXMPPSchema;
	zulip: NotifyZulipSchema;
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

export const notifyTypes: NotifySchemaKeys[] = Object.keys(
	notifySchemaMap,
) as NotifySchemaKeys[];
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
	homeassistant: notifyHomeAssistantSchemaOutgoing,
	ifttt: notifyIFTTTSchemaOutgoing,
	matrix: notifyMatrixSchemaOutgoing,
	mattermost: notifyMatterMostSchemaOutgoing,
	mqtt: notifyMQTTSchemaOutgoing,
	ntfy: notifyNtfySchemaOutgoing,
	pagerduty: notifyPagerDutySchemaOutgoing,
	rocketchat: notifyRocketChatSchemaOutgoing,
	signal: notifySignalSchemaOutgoing,
	signalgrid: notifySignalgridSchemaOutgoing,
	smtp: notifySMTPSchemaOutgoing,
	telegram: notifyTelegramSchemaOutgoing,
	xmpp: notifyXMPPSchemaOutgoing,
	zulip: notifyZulipSchemaOutgoing,
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
		// Generic Webhook.
		case NOTIFY_TYPE_MAP.GENERIC.value:
			return notifyGenericSchemaOutgoing.extend({
				url_fields: notifyGenericSchemaOutgoing.shape.url_fields
					.unwrap()
					.extend({
						headers: preprocessStringFromHeaderArrayWithDefaults(
							defaults?.url_fields?.headers,
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
