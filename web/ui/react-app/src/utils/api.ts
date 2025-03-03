import {
	DeployedVersionLookupType,
	LatestVersionLookupType,
	ServiceOptionsType,
} from 'types/config';
import { convertToQueryParams, getChanges } from './query-params';

import { ServiceRefreshType } from 'types/service-edit';
import fetchJSON from './fetch-json';

type fetchVersionJSONProps = {
	serviceID?: string;
	dataTarget: 'latest_version' | 'deployed_version';
	semanticVersioning?: boolean;
	options?: ServiceOptionsType;
	data?: LatestVersionLookupType | DeployedVersionLookupType;
	original: LatestVersionLookupType | DeployedVersionLookupType | null;
};

export const fetchVersionJSON = ({
	serviceID,
	dataTarget,
	semanticVersioning,
	options,
	data,
	original,
}: fetchVersionJSONProps) => {
	let semantic_versioning;
	if ((semanticVersioning ?? '') !== (options?.semantic_versioning ?? '')) {
		if (semanticVersioning === null) {
			semantic_versioning = 'null';
		} else {
			semantic_versioning = `${semanticVersioning}`;
		}
	}
	const overrides = data
		? getChanges({
				params: data,
				defaults: original,
				target: dataTarget,
		  })
		: '';
	return fetchJSON<ServiceRefreshType>({
		url: `api/v1/${dataTarget}/refresh${
			serviceID ? `/${encodeURIComponent(serviceID)}` : ''
		}${convertToQueryParams({
			overrides,
			semantic_versioning,
		})}`,
	});
};
