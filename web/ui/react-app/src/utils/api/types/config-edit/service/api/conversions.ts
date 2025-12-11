import { removeEmptyValues } from '@/utils';
import { DEPLOYED_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/deployed-version';
import { mapCommandsSchemaToAPIPayload } from '@/utils/api/types/config-edit/command/api/conversions';
import { mapNotifiersSchemaToAPIPayload } from '@/utils/api/types/config-edit/notify/api/conversions';
import type { buildServiceSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder';
import type {
	ServiceSchema,
	ServiceSchemaDefault,
	ServiceSchemaOutgoing,
} from '@/utils/api/types/config-edit/service/schemas';
import { mapWebHooksSchemaToAPIPayload } from '@/utils/api/types/config-edit/webhook/api/conversions';

/**
 * Converts a `ServiceSchema` object to a `ServiceSchemaOutgoing` object.
 *
 * @param data - The `ServiceSchema` data to map.
 * @param defaults - The default values to compare against (and omit where all defaults used and unmodified).
 * @param mainDefaults - The notify/webhook globals.
 * @param typeDefaults - Type-specific notify/webhook form data.
 */
export const mapServiceToAPIRequest = (
	data: ServiceSchema,
	defaults: ServiceSchemaDefault | null,
	mainDefaults: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['mainDataDefaults'],
	typeDefaults: ReturnType<
		typeof buildServiceSchemaWithFallbacks
	>['typeDataDefaults'],
): ServiceSchemaOutgoing => {
	const dv = data.deployed_version;
	let deployedVersion = null;
	// Have a deployed version lookup.
	if (
		(dv.type === DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value && dv.version) ||
		(dv.type === DEPLOYED_VERSION_LOOKUP_TYPE.URL.value && dv.url)
	) {
		deployedVersion = dv;
	}

	return removeEmptyValues({
		...data,
		command: mapCommandsSchemaToAPIPayload(data.command, defaults?.command),
		deployed_version: deployedVersion,
		id_name_separator: null,
		notify: mapNotifiersSchemaToAPIPayload(
			data.notify,
			defaults?.notify,
			mainDefaults?.notify,
			typeDefaults?.notify,
		),
		webhook: mapWebHooksSchemaToAPIPayload(
			data.webhook,
			defaults?.webhook,
			mainDefaults?.webhook,
			typeDefaults?.webhook,
		),
	}) as ServiceSchemaOutgoing;
};
