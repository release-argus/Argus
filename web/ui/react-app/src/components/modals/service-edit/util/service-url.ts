import {
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupForgejoHostDefaults,
	type LatestVersionLookupType,
} from '@/utils/api/types/config/service/latest-version';
import type { NullString } from '@/utils/api/types/config-edit/shared/null-string';

/**
 * Resolves a forge `host` to the root URL of the instance, defaulting the scheme to HTTPS.
 */
const forgeInstanceURL = (host?: string) => {
	const trimmed = (host ?? '').trim().replace(/\/+$/, '');
	if (trimmed === '') return '';

	return trimmed.includes('://') ? trimmed : `https://${trimmed}`;
};

/* scheme://hostname[/path][?query][#fragment] */
const forgeHostParts = /^([a-z][a-z\d+.-]*):\/\/([^/?#]*)([^?#]*)/i;

/**
 * Resolves a forge `host` to the instance it addresses.
 *
 * A host may name a defaults entry or be a URL itself; a name wins, compared
 * case-insensitively.
 */
export const resolveForgeHost = (
	host: string,
	hosts?: Record<string, LatestVersionLookupForgejoHostDefaults>,
) => {
	const entry = Object.entries(hosts ?? {}).find(
		([name]) => name.toLowerCase() === host.toLowerCase(),
	);

	return entry?.[1].url || host;
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

	const parts = forgeHostParts.exec(instance);
	if (!parts) return instance.toLowerCase();

	const scheme = parts[1].toLowerCase();
	let hostname = parts[2].toLowerCase();
	const port = /:(\d+)$/.exec(hostname)?.[1];
	if (
		port &&
		((scheme === 'https' && port === '443') ||
			(scheme === 'http' && port === '80'))
	)
		hostname = hostname.slice(0, -(port.length + 1));

	return `${scheme}://${hostname}${parts[3].replace(/\/+$/, '')}`;
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
