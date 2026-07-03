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
};

export type Settings = {
	data?: SettingsData;
	log?: SettingsLog;
	web?: SettingsWeb;
};
