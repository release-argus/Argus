import { FC, memo, useMemo } from 'react';
import { FormText, FormTextWithPreview } from 'components/generic/form';

import { Accordion } from 'react-bootstrap';
import { BooleanWithDefault } from 'components/generic';
import { ServiceDashboardOptionsType } from 'types/config';
import { firstNonDefault } from 'utils';

interface Props {
	defaults?: ServiceDashboardOptionsType;
	hard_defaults?: ServiceDashboardOptionsType;
}

/**
 * The `dashboard` form fields.
 *
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for the `dashboard` options.
 */
const EditServiceDashboard: FC<Props> = ({ defaults, hard_defaults }) => {
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

	return (
		<Accordion>
			<Accordion.Header>Dashboard:</Accordion.Header>
			<Accordion.Body>
				<BooleanWithDefault
					name={'dashboard.auto_approve'}
					label="Auto-approve"
					tooltip="Send all commands/webhooks when a new release is found"
					defaultValue={convertedDefaults.auto_approve}
				/>
				<FormTextWithPreview
					name={'dashboard.icon'}
					label="Icon"
					tooltip="e.g. https://example.com/icon.png"
					defaultVal={convertedDefaults.icon}
				/>
				<FormText
					key="icon_link_to"
					name={'dashboard.icon_link_to'}
					col_sm={12}
					label="Icon link to"
					tooltip="Where the Icon will redirect when clicked"
					defaultVal={convertedDefaults.icon_link_to}
					isURL
				/>
				<FormText
					key="web_url"
					name={'dashboard.web_url'}
					col_sm={12}
					label="Web URL"
					tooltip="Where the 'Service name' will redirect when clicked"
					defaultVal={convertedDefaults.web_url}
					isURL
				/>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceDashboard);
