import { FC, memo, useMemo } from 'react';
import { FormText, FormTextWithPreview } from 'components/generic/form';

import { Accordion } from 'react-bootstrap';
import { BooleanWithDefault } from 'components/generic';
import FormSelectCreatableSortable from 'components/generic/form-select-creatable-sortable';
import { ServiceDashboardOptionsType } from 'types/config';
import { createOption } from 'components/generic/form-select-shared';
import { firstNonDefault } from 'utils';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	originals?: ServiceDashboardOptionsType;
	defaults?: ServiceDashboardOptionsType;
	hard_defaults?: ServiceDashboardOptionsType;
}

/**
 * The `dashboard` form fields.
 *
 * @param originals - The original values of the form.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for the `dashboard` options.
 */
const EditServiceDashboard: FC<Props> = ({
	originals,
	defaults,
	hard_defaults,
}) => {
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
				<FormText
					key="icon_link_to"
					name="dashboard.icon_link_to"
					col_sm={12}
					label="Icon link to"
					tooltip="Where the Icon will redirect when clicked"
					defaultVal={convertedDefaults.icon_link_to}
					isURL
				/>
				<FormText
					key="web_url"
					name="dashboard.web_url"
					col_sm={12}
					label="Web URL"
					tooltip="Where the 'Service name' will redirect when clicked"
					defaultVal={convertedDefaults.web_url}
					isURL
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
