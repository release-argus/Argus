import { Accordion, Button, Stack } from 'react-bootstrap';
import { Dict, WebHookType } from 'types/config';
import { FC, useCallback, useMemo } from 'react';
import { firstNonEmpty, isEmptyArray } from 'utils';

import EditServiceWebHook from 'components/modals/service-edit/webhook';
import { createOption } from 'components/generic/form-select-shared';
import { useFieldArray } from 'react-hook-form';

interface Props {
	mains?: Dict<WebHookType>;
	defaults?: WebHookType;
	hard_defaults?: WebHookType;
	loading: boolean;
}

/**
 * The form fields for s service's webhooks.
 *
 * @param mains - The global WebHooks.
 * @param defaults - The default values for a WebHook.
 * @param hard_defaults - The hard default values for a WebHook.
 * @param loading - Whether the modal is loading.
 * @returns The form fields for a service's webhooks.
 */
const EditServiceWebHooks: FC<Props> = ({
	mains,
	defaults,
	hard_defaults,
	loading,
}) => {
	const { fields, append, remove } = useFieldArray({
		name: 'webhook',
	});
	const convertedDefaults = useMemo(
		() => ({
			custom_headers: firstNonEmpty(
				defaults?.custom_headers,
				hard_defaults?.custom_headers,
			).map(() => ({ key: '', item: '' })),
		}),
		[defaults, hard_defaults],
	);
	const addItem = useCallback(() => {
		append(
			{
				type: 'github',
				name: '',
				custom_headers: convertedDefaults.custom_headers,
			},
			{ shouldFocus: false },
		);
	}, []);

	const globalWebHookOptions = useMemo(
		() => [
			{ label: '--Not global--', value: '' },
			...Object.keys(mains ?? []).map((n) => createOption(n)),
		],
		[mains],
	);

	return (
		<Accordion>
			<Accordion.Header>WebHook:</Accordion.Header>
			<Accordion.Body>
				<Stack gap={2}>
					{fields.map(({ id }, index) => (
						<EditServiceWebHook
							key={id}
							name={`webhook.${index}`}
							removeMe={() => remove(index)}
							globalOptions={globalWebHookOptions}
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
						Add WebHook
					</Button>
				</Stack>
			</Accordion.Body>
		</Accordion>
	);
};

export default EditServiceWebHooks;
