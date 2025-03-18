import {
	DeployedVersionLookupType,
	LatestVersionLookupType,
	ServiceOptionsType,
} from 'types/config';
import {
	ServiceRefreshType,
	TemplateAPIResponseType,
} from 'types/service-edit';
import { convertToQueryParams, getChanges } from './query-params';

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

type parseTemplateProps = {
	serviceID: string;
	template: string;
	extraParams?: Record<string, string | undefined>;
};

export const parseTemplate = async ({
	serviceID,
	template,
	extraParams,
}: parseTemplateProps): Promise<string> => {
	const queryParamStr = convertToQueryParams(
		Object.fromEntries(
			Object.entries({
				service_id: serviceID,
				template: template,
				params: extraParams ? JSON.stringify(extraParams) : undefined,
			}).filter(([_, v]) => v !== undefined),
		),
	);
	return await fetchJSON<TemplateAPIResponseType>({
		url: `api/v1/template${queryParamStr}`,
		method: 'GET',
	}).then((data) => data.parsed);
};
