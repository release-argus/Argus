import {
	ButtonGroup,
	Dropdown,
	OverlayTrigger,
	Tooltip,
} from 'react-bootstrap';
import { FC, useMemo } from 'react';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { URL_PARAMS } from 'constants/toolbar';
import { faEye } from '@fortawesome/free-solid-svg-icons';

export enum HideValue {
	UpToDate = 0,
	Updatable = 1,
	Skipped = 2,
	Inactive = 3,
}

export const HIDE_OPTIONS = [
	{ key: 'upToDate', label: 'Hide up-to-date', value: HideValue.UpToDate },
	{ key: 'updatable', label: 'Hide updatable', value: HideValue.Updatable },
	{ key: 'skipped', label: 'Hide skipped', value: HideValue.Skipped },
	{ key: 'inactive', label: 'Hide inactive', value: HideValue.Inactive },
] as const;

export const DEFAULT_HIDE_VALUE = [HideValue.Inactive];

type Props = {
	values: number[];
	setValue: (
		key: (typeof URL_PARAMS)[keyof typeof URL_PARAMS],
		value: number[],
	) => void;
};

const FilterDropdown: FC<Props> = ({ values, setValue }) => {
	const optionsMap = useMemo(
		() => ({
			upToDate: () =>
				setValue(URL_PARAMS.HIDE, toggleHideValue(HideValue.UpToDate)),
			updatable: () =>
				setValue(URL_PARAMS.HIDE, toggleHideValue(HideValue.Updatable)),
			skipped: () =>
				setValue(URL_PARAMS.HIDE, toggleHideValue(HideValue.Skipped)),
			inactive: () =>
				setValue(URL_PARAMS.HIDE, toggleHideValue(HideValue.Inactive)),
			reset: () => setValue(URL_PARAMS.HIDE, [HideValue.Inactive]),
			flipAllHideOptions: () =>
				setValue(URL_PARAMS.HIDE, toggleAllHideValues()),
		}),
		[values],
	);

	const toggleHideValue = (value: number) =>
		values.includes(value)
			? values.filter((v) => v !== value)
			: [...values, value];

	const toggleAllHideValues = () =>
		[
			HideValue.UpToDate,
			HideValue.Updatable,
			HideValue.Skipped,
			HideValue.Inactive,
		].filter((n) => !(n !== HideValue.Inactive && values.includes(n)));

	const handleOption = (option: string) => {
		const hideUpdatable = values.includes(HideValue.Updatable);
		const hideUpToDate = values.includes(HideValue.UpToDate);
		const hideSkipped = values.includes(HideValue.Skipped);
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
				{HIDE_OPTIONS.map(({ key, label, value }) => (
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
