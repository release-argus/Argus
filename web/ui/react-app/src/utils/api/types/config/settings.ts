export type SettingsLog = {
	level?: string;
	timestamps?: boolean;
};

export type SettingsData = {
	database_file?: string;
	readonly?: boolean;
};

export type SettingsWeb = {
	cert_file: string;
	pkey_file: string;

	listen_host: string;
	listen_port: string;

	trusted_proxies?: string[];
};

export type SettingsAuthSession = {
	lifetime?: string;
	idle_timeout?: string;
};

export type SettingsAuthLocal = {
	enabled?: boolean;
};

export type SettingsAuth = {
	enabled?: boolean;
	session?: SettingsAuthSession;
	local?: SettingsAuthLocal;
};

export type Settings = {
	data?: SettingsData;
	log?: SettingsLog;
	web?: SettingsWeb;
	auth?: SettingsAuth;
};
