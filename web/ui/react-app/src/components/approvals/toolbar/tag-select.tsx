import { FC, useEffect, useMemo } from 'react';
import Select, { MultiValue } from 'react-select';
import {
	convertStringArrayToOptionTypeArray,
	customComponents,
	customStylesFixedHeight,
} from 'components/generic/form-select-shared';

import { Col } from 'react-bootstrap';
import { OptionType } from 'types/util';
import { useWebSocket } from 'contexts/websocket';

type Props = {
	tags: string[];
	setTags: (tags: string[]) => void;
};

const TagSelect: FC<Props> = ({ tags, setTags }) => {
	const { monitorData } = useWebSocket();

	const tagOptions = useMemo(
		() =>
			convertStringArrayToOptionTypeArray(
				Array.from(monitorData.tags ?? []),
				true,
			),
		[monitorData.tags],
	);

	// useEffect as cannot change state during rendering (which happens in useMemo).
	// useEffect ensures the state is updated after the render.
	useEffect(() => {
		if (
			monitorData.tags === undefined ||
			Object.values(monitorData.service).find((service) => service.loading)
		)
			return;
		// Ensure selected tags exist.
		const newTags = tags.filter((tag) => monitorData.tags?.has(tag));
		if (tags.length !== newTags.length) setTags(newTags);
	}, [monitorData.tags, tags, setTags]);

	if (tagOptions.length === 0) return null;

	return (
		<Col xs={3}>
			<Select
				className="form-select"
				options={tagOptions}
				value={tagOptions.filter((option) => tags.includes(option.value))}
				onChange={(newValue: MultiValue<OptionType>) =>
					setTags(newValue.map((option) => option.value))
				}
				isMulti
				placeholder="Tags..."
				closeMenuOnSelect={false}
				hideSelectedOptions={false}
				noOptionsMessage={() => 'No matches'}
				components={customComponents}
				styles={customStylesFixedHeight}
				aria-label="Select tag to filter services by"
			/>
		</Col>
	);
};

export default TagSelect;
