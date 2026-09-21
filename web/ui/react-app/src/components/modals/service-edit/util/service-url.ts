import {
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupForgejoHostDefaults,
	type LatestVersionLookupType,
} from '@/utils/api/types/config/service/latest-version';
import type { NullString } from '@/utils/api/types/config-edit/shared/null-string';

/**
 * Returns the defaults entry `host` names, compared case-insensitively.
 *
 * A host names an entry or is a URL itself.
 */
export const forgeHostEntry = (
	host?: string,
	hosts?: Record<string, LatestVersionLookupForgejoHostDefaults>,
) => {
	if (!host) return undefined;

	return Object.entries(hosts ?? {}).find(
		([name]) => name.toLowerCase() === host.toLowerCase(),
	)?.[1];
};

/**
 * Resolves a forge `host` to the instance it addresses.
 *
 * A host may name a defaults entry or be a URL itself; a name wins.
 */
export const resolveForgeHost = (
	host: string,
	hosts?: Record<string, LatestVersionLookupForgejoHostDefaults>,
) => forgeHostEntry(host, hosts)?.url || host;

/**
 * Resolves a forge `host` to the root URL of the instance, defaulting the scheme to
 * HTTPS and dropping any query or fragment in it.
 */
const forgeInstanceURL = (host?: string) => {
	const trimmed = (host ?? '')
		.trim()
		.replace(/[?#].*$/, '')
		.replace(/\/+$/, '');
	if (trimmed === '') return '';

	return trimmed.includes('://') ? trimmed : `https://${trimmed}`;
};

/**
 * Canonicalises a forge `host`, so that every spelling of one instance compares equal.
 *
 * Lowercased scheme and hostname, HTTPS by default, a port that is the default for
 * the scheme trimmed, and no trailing '/'.
 */
export const canonicalForgeHost = (host?: string) => {
	const instance = forgeInstanceURL(host);
	if (instance === '') return '';

	try {
		const url = new URL(instance);
		return `${url.protocol}//${url.host}${url.pathname.replace(/\/+$/, '')}`;
	} catch {
		return instance.toLowerCase();
	}
};

/**
 * Builds the URL of the source a `latest_version` lookup monitors.
 *
 * @param type - The 'type' of the lookup.
 * @param url - The lookup's URL - 'owner/repo' for the forge types.
 * @param host - The instance the repository is hosted on (Forgejo only).
 */
export const getServiceURL = (
	type: LatestVersionLookupType | NullString | undefined,
	url: string,
	host?: string,
) => {
	switch (type) {
		case LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value: {
			const instance = forgeInstanceURL(host);
			return instance === '' ? url : `${instance}/${url}`;
		}
		case LATEST_VERSION_LOOKUP_TYPE.GITHUB.value:
			return `https://github.com/${url}`;
		default:
			return url;
	}
};
