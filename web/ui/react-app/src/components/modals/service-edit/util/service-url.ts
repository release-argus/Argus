import {
	LATEST_VERSION_LOOKUP_TYPE,
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
