import type { Settings } from '@/utils/api/types/config/settings';
import type { DefaultsSchema } from '@/utils/api/types/config-edit/defaults';
import type { NotifySchemaValues } from '@/utils/api/types/config-edit/notify/schemas';
import type { ServiceSchema } from '@/utils/api/types/config-edit/service/schemas';
import type { WebHookSchema } from '@/utils/api/types/config-edit/webhook/schemas';

export type ConfigType = {
	settings?: Settings;
	defaults?: DefaultsSchema;
	notify?: Record<string, NotifySchemaValues>;
	webhook?: Record<string, WebHookSchema>;
	service?: Record<string, ServiceSchema>;
	order?: string[];
};
