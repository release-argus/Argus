import type { NotifyTypesValues } from '@/utils/api/types/config/notify';
import type { ServiceDashboardOptions } from '@/utils/api/types/config/service/dashboard';
import type { DeployedVersionLookup } from '@/utils/api/types/config/service/deployed-version';
import type {
	LatestVersionLookup,
	LatestVersionLookupDefaults,
} from '@/utils/api/types/config/service/latest-version';
import type { ServiceOptions } from '@/utils/api/types/config/service/options';
import type { Command } from '@/utils/api/types/config/shared';
import type { WebHook } from '@/utils/api/types/config/webhook';

export type Services = Record<string, Service>;

export type ServiceDefault = {
	options?: ServiceOptions;
	latest_version?: LatestVersionLookupDefaults;
	deployed_version?: DeployedVersionLookup;
	notify?: Record<string, unknown>;
	command?: Command[];
	webhook?: Record<string, unknown>;
	dashboard?: ServiceDashboardOptions;
};

export type Service = {
	id?: string;
	name?: string;

	comment?: string;
	options?: ServiceOptions;
	latest_version?: LatestVersionLookup;
	deployed_version?: DeployedVersionLookup;
	command?: Command[];
	webhook?: WebHook[];
	notify?: NotifyTypesValues[];
	dashboard?: ServiceDashboardOptions;
};
