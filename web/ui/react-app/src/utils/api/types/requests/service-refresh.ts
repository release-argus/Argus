import type { z } from 'zod';
import type { buildDeployedVersionLookupSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--deployed-version';
import type { buildLatestVersionLookupSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--latest-version';

type RequestBuilderBase = {
	/* The service ID. */
	serviceID?: string | null;

	/* The current `semantic_version` value. */
	dataSemanticVersioning: boolean | null;
	/* The original `semantic_version` value. */
	originalSemanticVersioning: boolean | null;
};

type LatestVersionType = ReturnType<
	typeof buildLatestVersionLookupSchemaWithFallbacks
>['schemaData'];

type DeployedVersionType = ReturnType<
	typeof buildDeployedVersionLookupSchemaWithFallbacks
>['schemaData'];

type RefreshMap = {
	latest_version: LatestVersionType;
	deployed_version: DeployedVersionType;
};

export type ServiceRefreshRequestBuilder = {
	[K in keyof RefreshMap]: RequestBuilderBase & {
		/* Refresh this target version of the service. */
		dataTarget: K;
		/* The current schema data. */
		data?: z.infer<RefreshMap[K]>;
		/* The original schema data. */
		original: z.infer<RefreshMap[K]> | null;
	};
}[keyof RefreshMap];

export type ServiceRefreshRequest = {
	queryParams: {
		/* The ID of the service to refresh. */
		service_id?: string | null;
		/* Values to override the current schema data with. */
		overrides?: string;
		/* The semantic versioning override. */
		semantic_versioning?: boolean | string;
	};
};

export type ServiceRefreshResponse = {
	/* The version of the service. */
	version?: string;
	/* The error message. */
	error?: string;
	/* The timestamp of the refresh. */
	timestamp: string;

	/* 'Route disabled' message. */
	message?: string;
};
