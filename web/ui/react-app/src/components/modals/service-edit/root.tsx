import {
	Button,
	Container,
	FormGroup,
	OverlayTrigger,
	Row,
	Tooltip,
} from 'react-bootstrap';
import { FC, memo, useEffect, useState } from 'react';
import { faCircleNotch, faGears } from '@fortawesome/free-solid-svg-icons';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormText } from 'components/generic/form';
import { useDelayedRender } from 'hooks/delayed-render';
import { useFormContext } from 'react-hook-form';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	id: string;
	name?: string;
	original_name?: string;
	loading: boolean;
}

/**
 * The form fields for the root values of a Service.
 *
 * @param id - The ID of the Service.
 * @param name - The name of the Service.
 * @param original_name - The original name of the Service.
 * @param loading - Whether the modal is loading.
 * @returns The form fields for the root values of a Service.
 */
const EditServiceRoot: FC<Props> = ({ id, name, original_name, loading }) => {
	const delayedRender = useDelayedRender(500);
	const [separateName, setSeparateName] = useState(false);
	const { monitorData } = useWebSocket();
	const { setValue, resetField } = useFormContext();

	useEffect(() => {
		if (!original_name === separateName) {
			setSeparateName(!!original_name);
		}
	}, [original_name]);

	const advancedToggle = (
		<OverlayTrigger
			placement="top"
			delay={{ show: 500, hide: 500 }}
			overlay={
				<Tooltip id="name-toggle-tooltip">
					Toggle to separate ID (service key) and Name in the config YAML
				</Tooltip>
			}
		>
			<Button
				aria-describedby="name-toggle-tooltip"
				name="separate_name_toggle"
				id="separate_name_toggle"
				className={`btn-border btn-${separateName ? '' : 'un'}checked pad-no`}
				style={{
					height: '1.5rem',
					width: '2.5rem',
					marginBottom: '0.25rem',
					position: 'absolute',
					right: 0,
				}}
				onClick={() => {
					if (separateName) {
						name
							? setValue('name', null, { shouldDirty: true })
							: resetField('name');
					} else {
						name
							? resetField('name')
							: setValue('name', id, { shouldDirty: true });
					}
					setSeparateName((prev) => !prev);
				}}
				variant="secondary"
			>
				<FontAwesomeIcon
					icon={faGears}
					style={{
						height: '1rem',
					}}
				/>
			</Button>
		</OverlayTrigger>
	);

	const valueNotInOtherNameOrID = (value: string) => {
		return (
			value === id ||
			value === name ||
			(!monitorData.order.includes(value) && !monitorData.names.has(value))
		);
	};

	const getTooltipProps = () => {
		if (!separateName) return {};
		return {
			tooltip: (
				<pre
					style={{
						margin: 0,
						textAlign: 'left',
						whiteSpace: 'pre-wrap',
					}}
				>
					{'services:\n  '}
					<span className="bold-underline">ID</span>
					{':\n    '}
					<span className="bold-underline">NAME</span>
					{': service_name\n    latest_version: ...'}
				</pre>
			),
			tooltipAriaLabel: 'Format: services.ID.NAME=service_name',
		};
	};

	return (
		<FormGroup className="mb-2">
			<Row style={{ position: 'relative' }}>
				{advancedToggle}
				<FormText
					name="id"
					required
					col_xs={12}
					col_sm={separateName ? 6 : 12}
					registerParams={{
						validate: (value: string) => {
							const validation =
								value === ''
									? false
									: // ID hasn't changed, or not in use.
									  valueNotInOtherNameOrID(value);
							return (
								validation || (value === '' ? 'Required' : 'Must be unique')
							);
						},
					}}
					label={separateName ? 'ID' : 'Name'}
					{...getTooltipProps()}
				/>
				{separateName && (
					<FormText
						name="name"
						required
						col_sm={6}
						registerParams={{
							validate: (value: string) => {
								const validation =
									value === ''
										? false
										: // Name hasn't changed, or not in use.
										  valueNotInOtherNameOrID(value);
								return (
									validation || (value === '' ? 'Required' : 'Must be unique')
								);
							},
						}}
						label="Name"
						tooltip="Name shown in the UI"
						positionXS="right"
					/>
				)}
			</Row>
			<FormText name="comment" col_sm={12} label="Comment" />
			{loading &&
				delayedRender(() => (
					<Container className="empty">
						<FontAwesomeIcon icon={faCircleNotch} className={'fa-spin'} />
						<span style={{ paddingLeft: '0.5rem' }}>Loading...</span>
					</Container>
				))}
		</FormGroup>
	);
};

export default memo(EditServiceRoot);
