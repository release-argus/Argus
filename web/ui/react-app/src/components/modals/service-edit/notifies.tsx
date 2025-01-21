import { Accordion, Button, Stack } from 'react-bootstrap';
import { Dict, NotifyTypes, NotifyTypesValues } from 'types/config';
import { FC, memo, useCallback, useMemo } from 'react';

import Notify from './notify';
import { NotifyEditType } from 'types/service-edit';
import { createOption } from 'components/generic/form-select-shared';
import { isEmptyArray } from 'utils';
import { useFieldArray } from 'react-hook-form';

interface Props {
	serviceID: string;

	originals?: NotifyEditType[];
	mains?: Dict<NotifyTypesValues>;
	defaults?: NotifyTypes;
	hard_defaults?: NotifyTypes;
	loading: boolean;
}

/**
 * The form fields for a mutable list of notifiers.
 *
 * @param serviceID - The ID of the service.
 * @param originals - The original values in the form.
 * @param mains - The main Notifiers.
 * @param defaults - The default values for each `notify` types.
 * @param hard_defaults - The hard default values for each `notify` types.
 * @param loading - Whether the modal is loading.
 * @returns The form fields for a mutable list of notifiers.
 */
const EditServiceNotifies: FC<Props> = ({
	serviceID,

	originals,
	mains,
	defaults,
	hard_defaults,
	loading,
}) => {
	const { fields, append, remove } = useFieldArray({
		name: 'notify',
	});
	const addItem = useCallback(() => {
		append(
			{
				type: 'discord',
				name: '',
				options: {},
				url_fields: {},
				params: { avatar: '', color: '', icon: '' },
			},
			{ shouldFocus: false },
		);
	}, []);

	const globalNotifyOptions = useMemo(
		() => [
			{ label: '--Not global--', value: '' },
			...Object.keys(mains ?? []).map((n) => createOption(n)),
		],
		[mains],
	);

	return (
		<Accordion>
			<Accordion.Header>Notify:</Accordion.Header>
			<Accordion.Body>
				<Stack gap={2}>
					{fields.map(({ id }, index) => (
						<Notify
							key={id}
							name={`notify.${index}`}
							removeMe={() => remove(index)}
							serviceID={serviceID}
							originals={originals}
							globalOptions={globalNotifyOptions}
							mains={mains}
							defaults={defaults}
							hard_defaults={hard_defaults}
						/>
					))}
					<Button
						className={isEmptyArray(fields) ? 'mt-2' : ''}
						variant="secondary"
						style={{ width: '100%', marginTop: '1rem' }}
						onClick={addItem}
						disabled={loading}
					>
						Add Notify
					</Button>
				</Stack>
			</Accordion.Body>
		</Accordion>
	);
};

export default memo(EditServiceNotifies);
