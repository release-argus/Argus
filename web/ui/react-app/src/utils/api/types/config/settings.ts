export type SettingsLog = {
	level?: string;
	timestamps?: boolean;
};

export type SettingsWeb = {
	cert_file: string;
	pkey_file: string;

	listen_host: string;
	listen_port: string;
};

export type Settings = {
	log?: SettingsLog;
	web?: SettingsWeb;
};
