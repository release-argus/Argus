import {
	GroupBase,
	MultiValue,
	MultiValueProps,
	MultiValueRemoveProps,
	SingleValue,
	components,
} from 'react-select';

import { CSS } from '@dnd-kit/utilities';
import { OptionType } from 'types/util';
import { useSortable } from '@dnd-kit/sortable';

/**
 * Custom styles for a React Select component.
 */
export const customStyles = {
	control: (provided: any, state: { isFocused: any }) => ({
		...provided,
		borderRadius: 'var(--bs-border-radius)',
		borderColor: state.isFocused
			? 'var(--bs-secondary-color)'
			: 'var(--bs-border-color)',
		':hover': {
			borderColor: 'var(--bs-secondary-color)',
			boxShadow: 'none',
		},
		backgroundColor: state.isFocused
			? 'var(--bs-secondary-bg-subtle)'
			: 'var(--bs-body-bg)',
		boxShadow: 'none',
	}),
	input: (provided: any) => ({
		...provided,
		color: 'var(--bs-body-color)',
	}),
	indicatorSeparator: (provided: any) => ({
		...provided,
		backgroundColor: 'var(--bs-border-color)',
	}),
	option: (provided: any, state: { isSelected: any; isFocused: any }) => ({
		...provided,
		backgroundColor: state.isSelected
			? 'var(--bs-border-color)'
			: 'transparent',
		color: state.isSelected
			? 'var(--bs-secondary-color)'
			: 'var(--bs-body-color)',
		cursor: 'pointer',
		padding: '0.25rem 0.5rem',
		':hover': {
			backgroundColor: state.isFocused
				? state.isSelected
					? 'var(--bs-secondary)'
					: 'var(--bs-secondary-bg)'
				: 'var(--bs-border-color)',
		},
	}),
	menu: (provided: any) => ({
		...provided,
		marginTop: '5px',
		boxShadow: '0 2px 5px var(--bs-secondary)',
		borderRadius: 'var(--bs-border-radius)',
		borderColor: 'var(--bs-border-color)',
		backgroundColor: 'var(--bs-body-bg)',
		color: 'var(--bs-border-color)',
		zIndex: 10,
	}),
	singleValue: (provided: any) => ({
		...provided,
		backgroundColor: 'transparent',
		color: 'var(--bs-body-color)',
	}),
	multiValue: (provided: any) => ({
		...provided,
		backgroundColor: 'var(--bs-secondary-bg)',
		color: 'var(--bs-body-color)',
	}),
	multiValueLabel: (provided: any) => ({
		...provided,
		color: 'var(--bs-body-color)',
	}),
	multiValueRemove: (provided: any) => ({
		...provided,
		color: 'var(--bs-body-color)',
		':hover': {
			backgroundColor: 'var(--bs-secondary)',
			color: 'var(--bs-body-color)',
		},
	}),
};

/**
 * Custom styles for a fixed height form select component.
 *
 * This object extends the base `customStyles` object and overrides
 * specific style properties to enforce a maximum height of 2.25rem
 * for the control and value container elements.
 */
export const customStylesFixedHeight = {
	...customStyles,
	control: (provided: any, state: { isFocused: any }) => ({
		...customStyles.control(provided, state),
		maxHeight: '2.25rem',
	}),
	valueContainer: (provided: any) => ({
		...provided,
		maxHeight: '2.25rem',
	}),
};

/**
 * Custom components for rendering selected values in a form select component.
 *
 * Replaces the default `SingleValue` and `MultiValueLabel` components with custom
 * components that render the option value, rather than the label.
 */
export const customComponents = {
	/** Custom rendering for single-selected value */
	SingleValue: (props: any) => (
		<components.SingleValue {...props}>
			<span>{props.data.value}</span>
		</components.SingleValue>
	),

	/** Custom rendering for multi-selected value */
	MultiValueLabel: (props: any) => (
		<components.MultiValueLabel {...props}>
			<span>{props.data.value}</span>
		</components.MultiValueLabel>
	),
};

/**
 * A component that renders a sortable multi-value item for a form select input.
 * This component uses the `useSortable` hook to enable drag-and-drop sorting functionality.
 *
 * @param props - The properties for the sortable multi-value component.
 * @param props.data - The data for the option being rendered.
 * @param props.data.value - The unique identifier for the option.
 *
 * @returns A JSX element representing the sortable multi-value item.
 */
export const sortableMultiValue = (
	props: MultiValueProps<OptionType, true, GroupBase<OptionType>>,
) => {
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id: props.data.value });

	const style = {
		transform: CSS.Translate.toString(transform),
		transition,
		cursor: isDragging ? 'grabbing' : 'grab',
		opacity: isDragging ? 0.5 : 1,
	};

	const onMouseDown = (e: {
		preventDefault: () => void;
		stopPropagation: () => void;
	}) => {
		e.preventDefault();
		e.stopPropagation();
	};

	return (
		<div
			ref={setNodeRef}
			style={style}
			{...attributes}
			{...listeners}
			onMouseDown={onMouseDown}
		>
			<components.MultiValue {...props}>
				<span>{props.data.value}</span>
			</components.MultiValue>
		</div>
	);
};

/**
 * A custom component that wraps the `MultiValueRemove` component from `react-select`.
 * It prevents the default propagation of the `onPointerDown` event.
 *
 * @param props - The properties passed to the `MultiValueRemove` component.
 * @returns A JSX element that renders the `MultiValueRemove` component with modified `innerProps`.
 */
export const MultiValueRemove = (props: MultiValueRemoveProps<OptionType>) => {
	return (
		<components.MultiValueRemove
			{...props}
			innerProps={{
				onPointerDown: (e) => e.stopPropagation(),
				...props.innerProps,
			}}
		/>
	);
};

/**
 * Converts a `string[]` to a `OptionType[]`,
 * or returns the input as is if it's not a `string[]`.
 *
 * @param input - The input value to check and convert.
 * @returns A list of `OptionType[]` objects if the input is a string array, or the original input.
 */
export const convertStringArrayToOptionTypeArray = (
	input: string[] | OptionType[],
	sort?: boolean,
): OptionType[] => {
	// Check whether input is an array of strings.
	if (Array.isArray(input) && input.every((item) => typeof item === 'string')) {
		// Convert to a list of `Option` objects.
		if (sort) {
			return input
				.toSorted((a, b) => a.localeCompare(b))
				.map((opt) => createOption(opt));
		}
		return input.map((opt) => createOption(opt));
	}

	// Already a list of `Option` objects, return as is.
	if (sort) return input.toSorted((a, b) => a.label.localeCompare(b.label));
	return input;
};

// Create an OptionType object from a string.
export const createOption = (
	inputValue: string,
	count?: number,
): OptionType => ({
	label: count === undefined ? inputValue : `'${inputValue}' (${count})`,
	value: inputValue,
});

/**
 * Handles the change in selected options for a form select component.
 *
 * @template T - The type of the selected value(s), either a string or an array of strings.
 *
 * @param hadValue - The previously selected value(s).
 * @param newValue - The newly selected value(s). Can be a single value or multiple values, depending on `T`.
 * @param allOptions - The list of options available.
 * @param useCounts - Optional. Whether to update the counts on the option labels. Defaults to `false`.
 * @param setSelectableOptions - Optional. A setter for the options available to select.
 * @param setSelectedOptions - Optional. A setter for the options that have been selected.
 *
 * @returns The updated list of options with counts adjusted if `useCounts` is `true`.
 */
export const handleSelectedChange = <T extends string | string[]>(
	hadValue: T,
	newValue: T extends string ? SingleValue<OptionType> : MultiValue<OptionType>,
	allOptions: OptionType[],
	useCount?: boolean,
	setSelectableOptions?: (value: OptionType[]) => void,
	setSelectedOptions?: (value: string[]) => void,
) => {
	if (newValue === null) return [];

	// Helper: Determine if an option needs to be updated.
	const calcDelta = (option: OptionType): number => {
		const isInHadValues = hadValue.includes(option.value);
		const isInNewValues = (newValue as MultiValue<OptionType>).some(
			(nv) => nv.value === option.value,
		);

		// State unchanged.
		if (isInHadValues === isInNewValues) return 0;
		// Value is newly selected.
		if (isInNewValues) return 1;
		// Value is deselected.
		return -1;
	};

	// Helper: Calculate the new count.
	const calculateNewCount = (option: OptionType, delta: number): number => {
		const oldCount = extractNumber(option.label) ?? 0;
		return Math.max(oldCount + delta, 0); // Prevent negative counts.
	};

	// Multi-select scenario.
	if (Array.isArray(hadValue)) {
		// Update selected options.
		setSelectedOptions &&
			setSelectedOptions(optionsToValues(newValue as OptionType[]));

		// Map and update creatableOptions.
		useCount &&
			setSelectableOptions &&
			setSelectableOptions(
				allOptions.map((option) => {
					const delta = calcDelta(option);
					if (delta) {
						const newCount = calculateNewCount(option, delta);
						return createOption(option.value, newCount);
					}
					return option; // No change needed.
				}),
			);

		return;
	}

	// Single-select scenario.
	const newOption = newValue as SingleValue<OptionType>;
	if (newOption && hadValue !== newOption.value) {
		// Find and update the option in creatableOptions.
		useCount &&
			setSelectableOptions &&
			setSelectableOptions(
				allOptions.map((option) => {
					if (option.value === hadValue || option.value === newOption.value) {
						const newCount = Math.max(
							(extractNumber(option.label) || 0) + calcDelta(newOption),
							0,
						);
						return createOption(option.value, newCount);
					}
					return option; // Don't change counts for other options.
				}),
			);
	}
};

/**
 * Extracts a number from a given string if it is enclosed in parentheses.
 *
 * @param input - The input string to extract the number from (e.g. 'Option (1)').
 * @returns The extracted number, otherwise `undefined`.
 */
export const extractNumber = (input: string): number | undefined => {
	// Match a number inside parentheses.
	const match = RegExp(/\((\d+)\)/).exec(input);
	// Convert the number, or undefined if no match.
	return match ? parseInt(match[1], 10) : undefined;
};

/**
 * Converts an array of options to an array of their values.
 */
export const optionsToValues = (options: OptionType[]) =>
	options.map((option) => option.value);

/**
 * Props for handling onChange events conditionally based on the `isMulti` flag.
 *
 * If `isMulti` is `true`, the `onChange` callback will receive a `MultiValue` object.
 *
 * If `isMulti` is `false`, the `onChange` callback will receive a `SingleValue` object.
 */
export type ConditionalOnChangeProps =
	| {
			isMulti: true;
			onChange?: (newValue: MultiValue<OptionType>) => void;
	  }
	| {
			isMulti?: false;
			onChange?: (newValue: SingleValue<OptionType>) => void;
	  };

/**
 * Handles the change event for a form select component, supporting both single and multi-select modes.
 *
 * @param newValue - The new value selected. Can be a single value or multiple values.
 * @param param1 - An object containing the following properties:
 * @param param1.isMulti - A boolean indicating whether the select component supports multiple selections.
 * @param param1.onChange - A callback function to handle the change event.
 */
export const customOnChange = (
	newValue: SingleValue<OptionType> | MultiValue<OptionType>,
	{ isMulti, onChange }: ConditionalOnChangeProps,
) => {
	if (!onChange) return;

	if (isMulti) {
		onChange((newValue ?? []) as MultiValue<OptionType>);
	} else {
		onChange(newValue as SingleValue<OptionType>);
	}
};
