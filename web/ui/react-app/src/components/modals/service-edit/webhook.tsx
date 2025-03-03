import { Accordion, Button, Row } from 'react-bootstrap';
import { Dict, WebHookType } from 'types/config';
import { FC, useEffect, useMemo } from 'react';
import { FormKeyValMap, FormSelect, FormText } from 'components/generic/form';
import { firstNonDefault, firstNonEmpty } from 'utils';
import { useFormContext, useWatch } from 'react-hook-form';

import { BooleanWithDefault } from 'components/generic';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { OptionType } from 'types/util';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import { numberRangeTest } from 'components/generic/form-validate';

interface Props {
	name: string;
	removeMe: () => void;

	globalOptions: OptionType[];
	mains?: Dict<WebHookType>;
	defaults?: WebHookType;
	hard_defaults?: WebHookType;
}

/**
 * The form fields for a WebHook.
 *
 * @param name - The name of the field in the form.
 * @param removeMe - The function to remove this WebHook.
 * @param globalOptions - The options for the global WebHooks.
 * @param mains - The main WebHooks.
 * @param defaults - The default values for a WebHook.
 * @param hard_defaults - The hard default values for a WebHook.
 * @returns The form fields for this WebHook.
 */
const EditServiceWebHook: FC<Props> = ({
	name,
	removeMe,

	globalOptions,
	mains,
	defaults,
	hard_defaults,
}) => {
	const webHookTypeOptions: {
		label: string;
		value: NonNullable<WebHookType['type']>;
	}[] = [
		{ label: 'GitHub', value: 'github' },
		{ label: 'GitLab', value: 'gitlab' },
	];

	const { setValue, trigger } = useFormContext();

	const itemName: string = useWatch({ name: `${name}.name` });
	const itemType: WebHookType['type'] = useWatch({ name: `${name}.type` });
	const main = mains?.[itemName];
	useEffect(() => {
		main?.type && setValue(`${name}.type`, main.type);
	}, [main]);
	useEffect(() => {
		if (mains?.[itemName]?.type !== undefined)
			setValue(`${name}.type`, mains[itemName].type);
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

	const convertedDefaults = useMemo(
		() => ({
			allow_invalid_certs:
				main?.allow_invalid_certs ??
				defaults?.allow_invalid_certs ??
				hard_defaults?.allow_invalid_certs,
			custom_headers: firstNonEmpty(
				main?.custom_headers,
				defaults?.custom_headers,
				hard_defaults?.custom_headers,
			),
			delay: firstNonDefault(
				main?.delay,
				defaults?.delay,
				hard_defaults?.delay,
			),
			desired_status_code: firstNonDefault(
				main?.desired_status_code,
				defaults?.desired_status_code,
				hard_defaults?.desired_status_code,
			),
			max_tries: firstNonDefault(
				main?.max_tries,
				defaults?.max_tries,
				hard_defaults?.max_tries,
			),
			secret: firstNonDefault(
				main?.secret,
				defaults?.secret,
				hard_defaults?.secret,
			),
			silent_fails:
				main?.silent_fails ??
				defaults?.silent_fails ??
				hard_defaults?.silent_fails,
			type: firstNonDefault(main?.type, defaults?.type, hard_defaults?.type),
			url: firstNonDefault(main?.url, defaults?.url, hard_defaults?.url),
		}),
		[main, defaults, hard_defaults],
	);

	return (
		<Accordion>
			<div style={{ display: 'flex', alignItems: 'center' }}>
				<Button
					aria-label="Delete this WebHook"
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
					<FormSelect
						name={`${name}.name`}
						col_xs={6}
						label="Global?"
						tooltip="Use this WebHook as a base"
						options={globalOptions}
					/>
					<FormSelect
						name={`${name}.type`}
						customValidation={(value: string) => {
							if (
								itemType !== undefined &&
								mains?.[itemName]?.type &&
								itemType !== mains?.[itemName]?.type
							) {
								return `${value} does not match the global for "${itemName}" of ${mains?.[itemName]?.type}. Either change the type to match that, or choose a new name`;
							}
							return true;
						}}
						col_xs={6}
						label="Type"
						tooltip="Style of WebHook to emulate"
						options={webHookTypeOptions}
						positionXS="right"
					/>
					<FormText
						name={`${name}.name`}
						required
						unique
						col_sm={12}
						label={'Name'}
					/>
					<FormText
						name={`${name}.url`}
						required
						col_sm={12}
						type="text"
						label="Target URL"
						tooltip="Where to send the WebHook"
						defaultVal={convertedDefaults.url}
						isURL
					/>
					<BooleanWithDefault
						name={`${name}.allow_invalid_certs`}
						label="Allow Invalid Certs"
						defaultValue={convertedDefaults.allow_invalid_certs}
					/>
					<FormText
						name={`${name}.secret`}
						required
						col_sm={12}
						label="Secret"
						defaultVal={convertedDefaults.secret}
					/>
					<FormKeyValMap
						name={`${name}.custom_headers`}
						defaults={convertedDefaults.custom_headers}
					/>
					<FormText
						name={`${name}.desired_status_code`}
						col_xs={6}
						label="Desired Status Code"
						tooltip="Treat the WebHook as successful when this status code is received (0=2XX)"
						isNumber
						defaultVal={convertedDefaults.desired_status_code}
					/>
					<FormText
						name={`${name}.max_tries`}
						col_xs={6}
						label="Max tries"
						isNumber
						validationFunc={(value: string) => numberRangeTest(value, 0, 255)}
						defaultVal={convertedDefaults.max_tries}
						positionXS="right"
					/>
					<FormText
						name={`${name}.delay`}
						col_sm={12}
						label="Delay"
						tooltip="Delay sending by this duration"
						defaultVal={convertedDefaults.delay}
						positionXS="right"
					/>
					<BooleanWithDefault
						name={`${name}.silent_fails`}
						label="Silent fails"
						tooltip="Notify if WebHook fails max tries times"
						defaultValue={convertedDefaults.silent_fails}
					/>
				</Row>
			</Accordion.Body>
		</Accordion>
	);
};

export default EditServiceWebHook;
