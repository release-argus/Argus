import { ButtonGroup, Form } from 'react-bootstrap';
import React, { FC, memo } from 'react';

import { ApprovalsToolbarOptions } from 'types/util';
import EditModeToggle from './edit-mode-toggle';
import FilterDropdown from './filter-dropdown';
import SearchBar from './search-bar';
import TagSelect from './tag-select';

type Props = {
	values: ApprovalsToolbarOptions;
	setValues: React.Dispatch<React.SetStateAction<ApprovalsToolbarOptions>>;

	// Sorting.
	onEditModeToggle: (value: boolean) => void;
	onSaveOrder: () => void;
	hasOrderChanged: boolean;
};

const ApprovalsToolbar: FC<Props> = ({
	values,
	setValues,

	onEditModeToggle,
	onSaveOrder,
	hasOrderChanged,
}) => {
	const setValue = (key: keyof typeof values, value: any) => {
		setValues((prev) => ({ ...prev, [key]: value }));
	};

	// Edit mode.
	const toggleEditMode = () => {
		const newValue = !values.editMode;
		setValue('editMode', newValue);
		onEditModeToggle(newValue);
	};

	return (
		<Form className="mb-3 gap-2 gap-md-3" style={{ display: 'flex' }}>
			<SearchBar
				search={values.search ?? ''}
				setSearch={(value) => setValue('search', value)}
			/>
			<TagSelect
				tags={values.tags ?? []}
				setTags={(tags) => setValue('tags', tags)}
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
