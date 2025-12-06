import type { Command } from '@/utils/api/types/config/shared';
import type { NullString } from '@/utils/api/types/config-edit/shared/null-string';

export const LATEST_VERSION_LOOKUP_TYPE = {
	GITHUB: { label: 'GitHub', value: 'github' },
	URL: { label: 'URL', value: 'url' },
} as const;
export type LatestVersionLookupType =
	(typeof LATEST_VERSION_LOOKUP_TYPE)[keyof typeof LATEST_VERSION_LOOKUP_TYPE]['value'];
export const latestVersionLookupTypeOptions = Object.values(
	LATEST_VERSION_LOOKUP_TYPE,
);

export type LatestVersionLookup =
	| LatestVersionLookupGitHub
	| LatestVersionLookupURL;

export type LatestVersionLookupBase = {
	url?: string;
	url_commands?: URLCommand[];
	require?: LatestVersionRequire;
};

export type LatestVersionLookupDefaults = {
	type?: LatestVersionLookupType | null;
	url?: string;
	url_commands?: URLCommand[];
	require?: LatestVersionRequireDefaults;
	access_token?: string;
	use_prerelease?: boolean;
	allow_invalid_certs?: boolean | null;
};

/* URL Command */
export const LATEST_VERSION__URL_COMMAND_TYPE = {
	REGEX: { label: 'Regex', value: 'regex' },
	REPLACE: { label: 'Replace', value: 'replace' },
	SPLIT: { label: 'Split', value: 'split' },
} as const;
export type LatestVersionURLCommandType =
	(typeof LATEST_VERSION__URL_COMMAND_TYPE)[keyof typeof LATEST_VERSION__URL_COMMAND_TYPE]['value'];
export const latestVersionURLCommandTypeOptions = Object.values(
	LATEST_VERSION__URL_COMMAND_TYPE,
);
export type URLCommand = URLCommandRegex | URLCommandReplace | URLCommandSplit;
export type FormURLCommand =
	| FormURLCommandRegex
	| URLCommandReplace
	| FormURLCommandSplit;

type URLCommandRegex = {
	type: typeof LATEST_VERSION__URL_COMMAND_TYPE.REGEX.value;

	regex: string;
	index?: string | number | null;
	template?: string;
};
type FormURLCommandRegex = URLCommandRegex & {
	index: number | null;
	template_toggle: boolean;
};
type URLCommandReplace = {
	type: typeof LATEST_VERSION__URL_COMMAND_TYPE.REPLACE.value;

	old: string;
	new: string;
};
type URLCommandSplit = {
	type: typeof LATEST_VERSION__URL_COMMAND_TYPE.SPLIT.value;

	text: string;
	index?: string | number | null;
};
type FormURLCommandSplit = URLCommandSplit & {
	index: number | null;
};

/* Require */
export const LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE = {
	DOCKER_HUB: { label: 'Docker Hub', value: 'hub' },
	GHCR: { label: 'GHCR', value: 'ghcr' },
	QUAY: { label: 'Quay', value: 'quay' },
} as const;
export type DockerFilterType =
	| (typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE)[keyof typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE]['value']
	| NullString;
export const latestVersionRequireDockerTypeOptions = Object.values(
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE,
);

type DockerFilterBase = {
	type:
		| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.GHCR.value
		| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.QUAY.value
		| null;
	image: string;
	tag: string;
	token: string;
	username: string;
};
export type DockerFilterUsername = DockerFilterBase & {
	type:
		| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value
		| null;
	username: string;
};
export type DockerFilterUsernameDefaults = Partial<DockerFilterUsername>;
export type DockerType =
	| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.GHCR.value
	| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.QUAY.value
	| typeof LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value
	| NullString;

export type DockerFilter = DockerFilterBase | DockerFilterUsername;
export type DockerFilterDefaults =
	| Partial<DockerFilterBase>
	| DockerFilterUsernameDefaults;

export type RequireDockerFilterDefaults = {
	type?: DockerFilterType;

	ghcr?: DockerFilterDefaults;
	hub?: DockerFilterUsernameDefaults;
	quay?: DockerFilterDefaults;
};

export type LatestVersionRequire = {
	regex_content?: string;
	regex_version?: string;
	command?: Command;
	docker?: DockerFilter;
};
export type LatestVersionRequireDefaults = LatestVersionRequire & {
	docker?: RequireDockerFilterDefaults;
};

/* Type: github */
export type LatestVersionLookupGitHub = LatestVersionLookupBase & {
	type: typeof LATEST_VERSION_LOOKUP_TYPE.GITHUB.value | null;
	access_token?: string;
	use_prerelease?: boolean;
};

/* Type: url */
export type LatestVersionLookupURL = LatestVersionLookupBase & {
	type: typeof LATEST_VERSION_LOOKUP_TYPE.URL.value | null;
	allow_invalid_certs?: boolean;
};
