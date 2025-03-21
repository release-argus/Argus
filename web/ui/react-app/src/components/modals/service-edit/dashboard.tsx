import { FC, memo, useMemo } from 'react';
import {
	FormSelectCreatableSortable,
	FormTextWithButton,
	FormTextWithPreview,
} from 'components/generic/form';
import { firstNonDefault, parseTemplate, removeEmptyValues } from 'utils';

import { Accordion } from 'react-bootstrap';
import { BooleanWithDefault } from 'components/generic';
import { ServiceDashboardOptionsType } from 'types/config';
import { StatusSummaryType } from 'types/summary';
import { createOption } from 'components/generic/form-select-shared';
import { faLink } from '@fortawesome/free-solid-svg-icons';
import { useFormContext } from 'react-hook-form';
import useValuesRefetch from 'hooks/values-refetch';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	serviceID: string;
	originals?: ServiceDashboardOptionsType;
	defaults?: ServiceDashboardOptionsType;
	hard_defaults?: ServiceDashboardOptionsType;
	serviceStatus?: StatusSummaryType;
}

/**
 * The `dashboard` form fields.
 *
 * @param serviceID - The ID of the service.
 * @param originals - The original values of the form.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for the `dashboard` options.
 */
const EditServiceDashboard: FC<Props> = ({
	serviceID,
	originals,
	defaults,
	hard_defaults,
	serviceStatus,
}) => {
	const { setError } = useFormContext();

	const { data: latest_version, refetchData: refetchLatestVersion } =
		useValuesRefetch<string>('latest_version.version', true);
	const convertedDefaults = useMemo(
		() => ({
			auto_approve: defaults?.auto_approve ?? hard_defaults?.auto_approve,
			icon: firstNonDefault(defaults?.icon, hard_defaults?.icon),
			icon_link_to: firstNonDefault(
				defaults?.icon_link_to,
				hard_defaults?.icon_link_to,
			),
			web_url: firstNonDefault(defaults?.web_url, hard_defaults?.web_url),
		}),
		[defaults, hard_defaults],
	);
	const { monitorData } = useWebSocket();
	const tagOptions = useMemo(
		() =>
			Array.from(monitorData.tags ?? []).map((value) =>
				createOption(
					value,
					Object.values(monitorData.service).reduce(
						(count, service) => count + (service.tags?.includes(value) ? 1 : 0),
						0,
					),
				),
			),
		[monitorData.service, monitorData.tags],
	);
	const handleTemplateClick = (fieldName: string, template: string) => {
		if (!template.includes('{{')) {
			window.open(template, '_blank');
		}

		refetchLatestVersion();
		// setTimeout to allow time for refetch setStates ^
		const timeout = setTimeout(() => {
			const extraParams = removeEmptyValues({
				latest_version: latest_version ?? serviceStatus?.latest_version,
			});
			parseTemplate({
				serviceID,
				template,
				extraParams,
			})
				.then((parsed) => window.open(parsed, '_blank'))
				.catch((error) => setError(fieldName, { message: error.message }));
		});
		return () => clearTimeout(timeout);
	};

	return (
		<Accordion>
			<Accordion.Header>Dashboard:</Accordion.Header>
			<Accordion.Body>
				<BooleanWithDefault
					name="dashboard.auto_approve"
					label="Auto-approve"
					tooltip="Send all commands/webhooks when a new release is found"
					defaultValue={convertedDefaults.auto_approve}
				/>
				<FormTextWithPreview
					name="dashboard.icon"
					label="Icon"
					tooltip="e.g. https://example.com/icon.png"
					defaultVal={convertedDefaults.icon}
				/>
				<FormTextWithButton
					key="icon_link_to"
					name="dashboard.icon_link_to"
					col_sm={12}
					label="Icon link to"
					tooltip="Where the Icon will redirect when clicked"
					defaultVal={convertedDefaults.icon_link_to}
					isURL
					buttonIcon={faLink}
					buttonAriaLabel="Open icon link"
					buttonOnClick={(tpl) =>
						handleTemplateClick('dashboard.icon_link_to', tpl)
					}
				/>
				<FormTextWithButton
					key="web_url"
					name="dashboard.web_url"
					col_sm={12}
					label="Web URL"
					tooltip="Where the 'Service name' will redirect when clicked"
					defaultVal={convertedDefaults.web_url}
					isURL
					buttonIcon={faLink}
					buttonAriaLabel="Open web URL"
					buttonOnClick={(tpl) => handleTemplateClick('dashboard.web_url', tpl)}
				/>
				<FormSelectCreatableSortable
					name="dashboard.tags"
					col_sm={12}
					label="Tags"
					placeholder=""
					initialValue={originals?.tags}
					options={tagOptions}
					isClearable
					noOptionsMessage="No other tags in use. Type to create a new one."
					optionCounts
					dynamicHeight={true}
					positionXS="right"
				/>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceDashboard);
