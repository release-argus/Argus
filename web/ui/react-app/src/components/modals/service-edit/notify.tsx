import { Accordion, Button, Col, Form, FormGroup, Row } from 'react-bootstrap';
import {
	Dict,
	LatestVersionLookupType,
	NotifyTypes,
	NotifyTypesConst,
	NotifyTypesKeys,
	NotifyTypesValues,
} from 'types/config';
import { FC, JSX, memo, useEffect, useMemo } from 'react';
import { FormLabel, FormSelect, FormText } from 'components/generic/form';
import { NotifyEditType, ServiceEditOtherData } from 'types/service-edit';
import {
	convertNotifyParams,
	convertNotifyURLFields,
} from 'components/modals/service-edit/util';
import { useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import RenderNotify from './notify-types/render';
import { TYPE_OPTIONS } from './notify-types/types';
import TestNotify from 'components/modals/service-edit/test-notify';
import { faTrash } from '@fortawesome/free-solid-svg-icons';

interface Props {
	name: string;
	removeMe: () => void;

	serviceID: string;
	originals?: NotifyEditType[];
	globalOptions: JSX.Element;
	mains?: Dict<NotifyTypesValues>;
	defaults?: NotifyTypes;
	hard_defaults?: NotifyTypes;
}

/**
 * The form fields for a notifier.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this Notify.
 * @param serviceID - The ID of the service.
 * @param originals - The original values for the Notify.
 * @param globalOptions - The options for the global Notifiers.
 * @param mains - The main Notifiers.
 * @param defaults - The default values for all Notify types.
 * @param hard_defaults - The hard default values for all Notify types.
 * @returns The form fields for this Notify.
 */
const Notify: FC<Props> = ({
	name,
	removeMe,

	serviceID,
	originals,
	globalOptions,
	mains,
	defaults,
	hard_defaults,
}) => {
	const { setValue, trigger } = useFormContext();

	const itemName: string = useWatch({ name: `${name}.name` });
	const itemType: NotifyTypesKeys = useWatch({ name: `${name}.type` });
	const lvType: LatestVersionLookupType['type'] = useWatch({
		name: 'latest_version.type',
	});
	const lvURL: string | undefined = useWatch({ name: 'latest_version.url' });
	const webURL: string | undefined = useWatch({ name: 'dashboard.web_url' });
	useEffect(() => {
		// Set Type to that of the global for the new name if it exists.
		if (mains?.[itemName]?.type) setValue(`${name}.type`, mains[itemName].type);
		else if (itemType && (NotifyTypesConst as string[]).includes(itemName))
			setValue(`${name}.type`, itemName);
		// Trigger validation on name/type.
		const timeout = setTimeout(() => {
			if (itemName !== '') trigger(`${name}.name`);
			trigger(`${name}.type`);
		}, 25);
		return () => clearTimeout(timeout);
	}, [itemName]);
	const header = useMemo(
		() => `${name.split('.').slice(-1)}: (${itemType}) ${itemName}`,
		[name, itemName, itemType],
	);

	const original: NotifyEditType = useMemo(() => {
		const original = originals?.find((o) => o.oldIndex === itemName);
		return (
			original ?? { type: 'discord', options: {}, url_fields: {}, params: {} }
		);
	}, [originals]);
	const serviceURL =
		lvType === 'github' && (lvURL?.match(/\//g) ?? []).length == 1
			? `https://github.com/${lvURL}`
			: lvURL;

	const onChangeNotifyType = (
		newType: NotifyTypesKeys,
		original: NotifyEditType,
		otherOptionsData: ServiceEditOtherData,
	) => {
		// Reset to original type.
		if (newType === original?.type) {
			setValue(`${name}.url_fields`, original.url_fields);
			setValue(`${name}.params`, original.params);
			return;
		}

		// Set the default values for the selected type.
		setValue(
			`${name}.url_fields`,
			convertNotifyURLFields(name, newType, undefined, otherOptionsData),
		);
		setValue(
			`${name}.params`,
			convertNotifyParams(name, newType, undefined, otherOptionsData),
		);
	};

	return (
		<Accordion>
			<div style={{ display: 'flex', alignItems: 'center' }}>
				<Button
					className="btn-unchecked"
					variant="secondary"
					onClick={removeMe}
				>
					<FontAwesomeIcon icon={faTrash} />
				</Button>
				<Accordion.Button className="p-2">{header}</Accordion.Button>
			</div>

			<Accordion.Body>
				<Row xs={12}>
					<Col xs={6} className={`pe-2 pt-1 pb-1`}>
						<FormGroup className="mb-2">
							<FormLabel text="Global?" tooltip="Use this Notify as a base" />
							<Form.Select
								value={
									mains && Object.keys(mains).indexOf(itemName) !== -1
										? itemName
										: ''
								}
								onChange={(e) => setValue(`${name}.name`, e.target.value)}
							>
								{globalOptions}
							</Form.Select>
						</FormGroup>
					</Col>
					<FormSelect
						name={`${name}.type`}
						customValidation={(value) => {
							if (
								itemType !== undefined &&
								mains?.[itemName]?.type &&
								itemType !== mains?.[itemName]?.type
							) {
								return `${value} does not match the global for "${itemName}" of ${mains?.[itemName]?.type}. Either change the type to match that, or choose a new name`;
							}
							return true;
						}}
						onChange={(e) => {
							const newType = e.target.value as NotifyTypesKeys;
							const otherOptionsData: ServiceEditOtherData = {
								notify: mains,
								defaults: { notify: defaults },
								hard_defaults: { notify: hard_defaults },
							};
							onChangeNotifyType(newType, original, otherOptionsData);
							setValue(`${name}.type`, newType);
						}}
						col_xs={6}
						label="Type"
						options={TYPE_OPTIONS}
						position="right"
					/>
					<FormText
						name={`${name}.name`}
						required
						unique
						col_sm={12}
						label="Name"
					/>
					<RenderNotify
						name={name}
						type={itemType}
						main={mains?.[itemName]}
						defaults={defaults?.[itemType]}
						hard_defaults={hard_defaults?.[itemType]}
					/>
					<TestNotify
						path={name}
						original={original}
						extras={{
							service_id_previous: serviceID,
							service_url: serviceURL,
							web_url: webURL,
						}}
					/>
				</Row>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(Notify);
