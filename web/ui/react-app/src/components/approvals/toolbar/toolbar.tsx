import { ButtonGroup, Form } from 'react-bootstrap';
import { FC, memo, useCallback } from 'react';
import FilterDropdown, { DEFAULT_HIDE_VALUE } from './filter-dropdown';

import { ApprovalsToolbarOptions } from 'types/util';
import EditModeToggle from './edit-mode-toggle';
import SearchBar from './search-bar';
import TagSelect from './tag-select';
import { URL_PARAMS } from 'constants/toolbar';
import { useSearchParams } from 'react-router-dom';

type Props = {
	values: ApprovalsToolbarOptions;

	// Sorting.
	onEditModeToggle: (value: boolean) => void;
	onSaveOrder: () => void;
	hasOrderChanged: boolean;
};

const ApprovalsToolbar: FC<Props> = ({
	values,

	onEditModeToggle,
	onSaveOrder,
	hasOrderChanged,
}) => {
	const [_, setSearchParams] = useSearchParams();

	const updateURLParam = useCallback(
		(
			key: (typeof URL_PARAMS)[keyof typeof URL_PARAMS],
			value: boolean | number[] | string | string[],
			defaultValue: boolean | number[] | string | string[],
		) => {
			const newSearchParams = new URLSearchParams(window.location.search);

			if (Array.isArray(value)) {
				if (JSON.stringify(value) === JSON.stringify(defaultValue)) {
					newSearchParams.delete(key);
				} else {
					newSearchParams.set(key, JSON.stringify(value));
				}
			} else if (value === defaultValue) {
				newSearchParams.delete(key);
			} else {
				newSearchParams.set(key, value.toString());
			}

			setSearchParams(newSearchParams);
		},
		[setSearchParams],
	);

	const setValue = useCallback(
		(key: (typeof URL_PARAMS)[keyof typeof URL_PARAMS], value: any) => {
			switch (key) {
				case URL_PARAMS.SEARCH:
					updateURLParam(URL_PARAMS.SEARCH, value, '');
					break;
				case URL_PARAMS.TAGS:
					updateURLParam(URL_PARAMS.TAGS, value, []);
					break;
				case URL_PARAMS.EDIT_MODE:
					updateURLParam(URL_PARAMS.EDIT_MODE, value, false);
					break;
				case URL_PARAMS.HIDE:
					updateURLParam(URL_PARAMS.HIDE, value, DEFAULT_HIDE_VALUE);
					break;
			}
		},
		[updateURLParam],
	);

	// Edit mode.
	const toggleEditMode = () => {
		const newValue = !values.editMode;
		updateURLParam(URL_PARAMS.EDIT_MODE, newValue, false);
		onEditModeToggle(newValue);
	};

	return (
		<Form className="mb-3 gap-2 gap-md-3" style={{ display: 'flex' }}>
			<SearchBar
				search={values.search ?? ''}
				setSearch={(value) => setValue(URL_PARAMS.SEARCH, value)}
			/>
			<TagSelect
				tags={values.tags ?? []}
				setTags={(tags) => setValue(URL_PARAMS.TAGS, tags)}
			/>
			<ButtonGroup className="gap-1 gap-md-2">
				<FilterDropdown values={values.hide} setValue={setValue} />
				<EditModeToggle
					editMode={values.editMode}
					toggleEditMode={toggleEditMode}
					onSaveOrder={onSaveOrder}
					hasOrderChanged={hasOrderChanged}
				/>
			</ButtonGroup>
		</Form>
	);
};

export default memo(ApprovalsToolbar);
