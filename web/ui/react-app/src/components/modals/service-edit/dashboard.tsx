import { useQueryClient } from '@tanstack/react-query';
import { Link } from 'lucide-react';
import { type FC, useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldSelectCreatableSortable,
	FieldTextWithButton,
	FieldTextWithPreview,
} from '@/components/generic/field';
import {
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useTags } from '@/hooks/use-tags';
import useValuesRefetch from '@/hooks/values-refetch';
import { QUERY_KEYS } from '@/lib/query-keys';
import { removeEmptyValues } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

/**
 * @returns The form fields for the `dashboard` options.
 */
const EditServiceDashboard: FC = () => {
	const queryClient = useQueryClient();
	const name = 'dashboard';
	const { serviceID, schemaDataDefaults } = useSchemaContext();
	const { setError } = useFormContext();

	const { data: latestVersion, refetchData: refetchLatestVersion } =
		useValuesRefetch<string>('latest_version.version', true);

	// Options, and their respective use-counts on other services.
	const { tags: allTags, counts } = useTags(serviceID);
	const tagState = useMemo(
		() => ({
			counts: counts,
			options: allTags.map((value) => ({ label: value, value })),
		}),
		[allTags, counts],
	);

	// Apply the template and navigate to the address.
	const handleTemplateClick = (fieldName: string, template: string) => {
		if (!template.includes('{{')) {
			window.open(template, '_blank');
		}

		refetchLatestVersion();
		// setTimeout to give time for refetch setStates ^
		const timeout = setTimeout(() => {
			const serviceStatusLV = serviceID
				? queryClient.getQueryData<ServiceSummary>(
						QUERY_KEYS.SERVICE.SUMMARY_ITEM(serviceID),
					)?.status?.latest_version
				: null;
			const extraParams = removeEmptyValues({
				latest_version: latestVersion ?? serviceStatusLV,
			});

			mapRequest('TEMPLATE_PARSE', {
				extraParams: extraParams,
				serviceID: serviceID as string,
				template: template,
			})
				.then((parsed) => parsed && window.open(parsed, '_blank'))
				.catch((error: unknown) => {
					const message =
						error instanceof Error
							? error.message
							: 'An unknown error occurred';
					setError(fieldName, { message });
				});
		});
		return () => {
			clearTimeout(timeout);
		};
	};

	const defaults = schemaDataDefaults?.dashboard;

	return (
		<AccordionItem value="dashboard">
			<AccordionTrigger>Dashboard:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<BooleanWithDefault
					defaultValue={defaults?.auto_approve}
					label="Auto-approve"
					name={`${name}.auto_approve`}
					tooltip={{
						content: 'Send all commands/webhooks when a new release is found',
						type: 'string',
					}}
				/>
				<FieldTextWithPreview
					defaultVal={defaults?.icon}
					label="Icon"
					name={`${name}.icon`}
					tooltip={{
						content: 'e.g. https://example.com/icon.png',
						type: 'string',
					}}
				/>
				<FieldTextWithButton
					button={{
						ariaLabel: 'Open icon link',
						Icon: Link,
						kind: 'click',
						onClick: (tpl) => handleTemplateClick(`${name}.icon_link_to`, tpl),
					}}
					colSize={{ sm: 12 }}
					defaultVal={defaults?.icon_link_to}
					key="icon_link_to"
					label="Icon link to"
					name={`${name}.icon_link_to`}
					tooltip={{
						content: 'Where the Icon will redirect when clicked',
						type: 'string',
					}}
				/>
				<FieldTextWithButton
					button={{
						ariaLabel: 'Open web URL',
						Icon: Link,
						kind: 'click',
						onClick: (tpl) => handleTemplateClick(`${name}.web_url`, tpl),
					}}
					colSize={{ sm: 12 }}
					defaultVal={defaults?.web_url}
					key="web_url"
					label="Web URL"
					name={`${name}.web_url`}
					tooltip={{
						content: "Where the 'Service name' will redirect when clicked",
						type: 'string',
					}}
				/>
				<FieldSelectCreatableSortable
					colSize={{ sm: 12 }}
					counts={tagState.counts}
					isClearable
					label="Tags"
					name={`${name}.tags`}
					noOptionsMessage="No other tags in use. Type to create a new one."
					options={tagState.options}
					placeholder=""
				/>
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceDashboard;
