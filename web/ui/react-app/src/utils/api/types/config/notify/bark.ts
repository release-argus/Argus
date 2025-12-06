export const BARK_SCHEME = {
	HTTP: { label: 'HTTP', value: 'http' },
	HTTPS: { label: 'HTTPS', value: 'https' },
} as const;
export type BarkScheme =
	(typeof BARK_SCHEME)[keyof typeof BARK_SCHEME]['value'];
export const barkSchemeOptions = Object.values(BARK_SCHEME);

// biome-ignore assist/source/useSortedKeys: default on top.
export const BARK_SOUND = {
	DEFAULT: { label: '\u00A0', value: '' },
	ALARM: { label: 'Alarm', value: 'alarm' },
	ANTICIPATE: { label: 'Anticipate', value: 'anticipate' },
	BELL: { label: 'Bell', value: 'bell' },
	BIRDSONG: { label: 'Birdsong', value: 'birdsong' },
	BLOOM: { label: 'Bloom', value: 'bloom' },
	CALYPSO: { label: 'Calypso', value: 'calypso' },
	CHIME: { label: 'Chime', value: 'chime' },
	CHOO: { label: 'Choo', value: 'choo' },
	DESCENT: { label: 'Descent', value: 'descent' },
	ELECTRONIC: { label: 'Electronic', value: 'electronic' },
	FANFARE: { label: 'Fanfare', value: 'fanfare' },
	GLASS: { label: 'Glass', value: 'glass' },
	GO_TO_SLEEP: { label: 'GoToSleep', value: 'gotosleep' },
	HEALTH_NOTIFICATION: {
		label: 'HealthNotification',
		value: 'healthnotification',
	},
	HORN: { label: 'Horn', value: 'horn' },
	LADDER: { label: 'Ladder', value: 'ladder' },
	MAIL_SENT: { label: 'MailSent', value: 'mailsent' },
	MINUET: { label: 'Minuet', value: 'minuet' },
	MULTI_WAY_INVITATION: {
		label: 'MultiWayInvitation',
		value: 'multiwayinvitation',
	},
	NEW_MAIL: { label: 'NewMail', value: 'newmail' },
	NEWS_FLASH: { label: 'NewsFlash', value: 'newsflash' },
	NOIR: { label: 'Noir', value: 'noir' },
	PAYMENT_SUCCESS: { label: 'PaymentSuccess', value: 'paymentsuccess' },
	SHAKE: { label: 'Shake', value: 'shake' },
	SHERWOOD_FOREST: { label: 'SherwoodForest', value: 'sherwoodforest' },
	SILENCE: { label: 'Silence', value: 'silence' },
	SPELL: { label: 'Spell', value: 'spell' },
	SUSPENSE: { label: 'Suspense', value: 'suspense' },
	TELEGRAPH: { label: 'Telegraph', value: 'telegraph' },
	TIPTOES: { label: 'Tiptoes', value: 'tiptoes' },
	TYPEWRITERS: { label: 'Typewriters', value: 'typewriters' },
	UPDATE: { label: 'Update', value: 'update' },
} as const;
export type BarkSound = (typeof BARK_SOUND)[keyof typeof BARK_SOUND]['value'];
export const barkSoundOptions = Object.values(BARK_SOUND);
