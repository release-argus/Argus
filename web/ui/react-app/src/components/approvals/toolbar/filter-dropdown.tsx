import {
	ButtonGroup,
	Dropdown,
	OverlayTrigger,
	Tooltip,
} from 'react-bootstrap';
import { FC, useMemo } from 'react';

import { ApprovalsToolbarOptions } from 'types/util';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faEye } from '@fortawesome/free-solid-svg-icons';

type Props = {
	values: number[];
	setValue: (key: keyof ApprovalsToolbarOptions, value: number[]) => void;
};

const FilterDropdown: FC<Props> = ({ values, setValue }) => {
	const hideOptions = [
		{ key: 'upToDate', label: 'Hide up-to-date', value: 0 },
		{ key: 'updatable', label: 'Hide updatable', value: 1 },
		{ key: 'skipped', label: 'Hide skipped', value: 2 },
		{ key: 'inactive', label: 'Hide inactive', value: 3 },
	];

	const optionsMap = useMemo(
		() => ({
			upToDate: () => setValue('hide', toggleHideValue(0)),
			updatable: () => setValue('hide', toggleHideValue(1)),
			skipped: () => setValue('hide', toggleHideValue(2)),
			inactive: () => setValue('hide', toggleHideValue(3)),
			reset: () => setValue('hide', [3]),
			flipAllHideOptions: () => setValue('hide', toggleAllHideValues()),
		}),
		[values],
	);

	const toggleHideValue = (value: number) =>
		values.includes(value)
			? values.filter((v) => v !== value)
			: [...values, value];

	const toggleAllHideValues = () =>
		[0, 1, 2, 3].filter((n) => !(n !== 3 && values.includes(n)));

	const handleOption = (option: string) => {
		const hideUpdatable = values.includes(0);
		const hideUpToDate = values.includes(1);
		const hideSkipped = values.includes(2);
		switch (option) {
			case 'upToDate': // 0
				hideUpToDate && hideSkipped // 1 && 2
					? optionsMap.flipAllHideOptions()
					: optionsMap.upToDate();
				break;
			case 'updatable': // 1
				hideUpdatable && hideSkipped // 0 && 2
					? optionsMap.flipAllHideOptions()
					: optionsMap.updatable();
				break;
			case 'skipped': // 2
				hideUpdatable && hideUpToDate // 0 && 1
					? optionsMap.flipAllHideOptions()
					: optionsMap.skipped();
				break;
			case 'inactive': // 3
				optionsMap.inactive();
				break;
			case 'reset':
				optionsMap.reset();
				break;
		}
	};
	const filterButtonTooltip = 'Filter shown services';

	return (
		<Dropdown as={ButtonGroup}>
			<OverlayTrigger
				delay={{ show: 500, hide: 500 }}
				overlay={<Tooltip id="tooltip-help">{filterButtonTooltip}</Tooltip>}
			>
				<Dropdown.Toggle
					variant="secondary"
					className="border-0"
					aria-label={filterButtonTooltip}
				>
					<FontAwesomeIcon icon={faEye} />
				</Dropdown.Toggle>
			</OverlayTrigger>
			<Dropdown.Menu>
				{hideOptions.map(({ key, label, value }) => (
					<Dropdown.Item
						key={key}
						eventKey={key}
						active={values.includes(value)}
						onClick={() => handleOption(key)}
					>
						{label}
					</Dropdown.Item>
				))}
				<Dropdown.Divider />
				<Dropdown.Item eventKey="reset" onClick={() => handleOption('reset')}>
					Reset
				</Dropdown.Item>
			</Dropdown.Menu>
		</Dropdown>
	);
};

export default FilterDropdown;
