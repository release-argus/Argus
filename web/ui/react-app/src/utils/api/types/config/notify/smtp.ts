// biome-ignore assist/source/useSortedKeys: encryption ordering.
export const SMTP_AUTH = {
	UNKNOWN: { label: 'Unknown', value: 'Unknown' },
	NONE: { label: 'None', value: 'None' },
	PLAIN: { label: 'Plain', value: 'Plain' },
	CRAM_MD5: { label: 'CRAM-MD5', value: 'CRAMMD5' },
	OAUTH2: { label: 'OAuth2', value: 'OAuth2' },
} as const;
export type SMTPAuth = (typeof SMTP_AUTH)[keyof typeof SMTP_AUTH]['value'];
export const smtpAuthOptions = Object.values(SMTP_AUTH);

export const SMTP_ENCRYPTION = {
	AUTO: { label: 'Auto', value: 'Auto' },
	EXPLICIT_TLS: { label: 'ExplicitTLS', value: 'ExplicitTLS' },
	IMPLICIT_TLS: { label: 'ImplicitTLS', value: 'ImplicitTLS' },
	NONE: { label: 'None', value: 'None' },
} as const;
export type SMTPEncryption =
	(typeof SMTP_ENCRYPTION)[keyof typeof SMTP_ENCRYPTION]['value'];
export const smtpEncryptionOptions = Object.values(SMTP_ENCRYPTION);
