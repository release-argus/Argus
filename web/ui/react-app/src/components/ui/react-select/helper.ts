// SOURCE: https://gist.github.com/ilkou/7bf2dbd42a7faf70053b43034fc4b5a4

import type { FocusEvent } from 'react';
import type { ClassNamesConfig, GroupBase, StylesConfig } from 'react-select';
import { cn } from '@/lib/utils';

/**
 * styles that align with shadcn/ui
 */
const controlStyles = {
	base: 'flex !min-h-9 w-full rounded-md border border-input dark:bg-input/30 bg-transparent pl-3 py-1 pr-1 gap-1 text-sm shadow-xs transition-colors hover:cursor-pointer',
	disabled: 'cursor-not-allowed opacity-50',
	focus: 'outline-none border-ring ring-ring/50 ring-[3px] caret-transparent',
};
const placeholderStyles = 'text-sm text-muted-foreground';
const valueContainerStyles = 'gap-1';
const multiValueStyles =
	'inline-flex items-center gap-2 rounded-md border border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80 px-1.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2';
const indicatorsContainerStyles = 'gap-1';
const clearIndicatorStyles = 'p-1 rounded-md';
const indicatorSeparatorStyles = 'bg-border';
const dropdownIndicatorStyles = 'p-1 rounded-md';
const menuStyles =
	'p-1 mt-1 border bg-popover shadow-md rounded-md text-popover-foreground !z-50';
const groupHeadingStyles =
	'py-2 px-1 text-secondary-foreground text-sm font-semibold';
export const optionStyles = {
	base: 'hover:cursor-pointer hover:bg-accent hover:text-accent-foreground px-2 py-1.5 rounded-sm !text-sm !cursor-default !select-none !outline-none font-sans',
	disabled: 'pointer-events-none opacity-50',
	focus: 'active:bg-accent/90 bg-accent text-accent-foreground',
	selected: 'bg-accent/50',
	selectedFocus: 'bg-accent/90',
};
const noOptionsMessageStyles =
	'text-accent-foreground p-2 bg-accent border border-dashed border-border rounded-sm';
const loadingIndicatorStyles =
	'flex items-center justify-center h-4 w-4 opacity-50';
const loadingMessageStyles = 'text-accent-foreground p-2 bg-accent';

export const createClassNames = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option> = GroupBase<Option>,
>(
	classNames?: ClassNamesConfig<Option, IsMulti, Group>,
): ClassNamesConfig<Option, IsMulti, Group> => {
	return {
		clearIndicator: (state) =>
			cn(clearIndicatorStyles, classNames?.clearIndicator?.(state)),
		container: (state) => cn(classNames?.container?.(state)),
		control: (state) =>
			cn(
				controlStyles.base,
				state.isDisabled && controlStyles.disabled,
				state.isFocused && controlStyles.focus,
				classNames?.control?.(state),
			),
		dropdownIndicator: (state) =>
			cn(dropdownIndicatorStyles, classNames?.dropdownIndicator?.(state)),
		group: (state) => cn(classNames?.group?.(state)),
		groupHeading: (state) =>
			cn(groupHeadingStyles, classNames?.groupHeading?.(state)),
		indicatorSeparator: (state) =>
			cn(indicatorSeparatorStyles, classNames?.indicatorSeparator?.(state)),
		indicatorsContainer: (state) =>
			cn(indicatorsContainerStyles, classNames?.indicatorsContainer?.(state)),
		input: (state) => cn(classNames?.input?.(state)),
		loadingIndicator: (state) =>
			cn(loadingIndicatorStyles, classNames?.loadingIndicator?.(state)),
		loadingMessage: (state) =>
			cn(loadingMessageStyles, classNames?.loadingMessage?.(state)),
		menu: (state) => cn(menuStyles, classNames?.menu?.(state)),
		menuList: (state) => cn(classNames?.menuList?.(state)),
		menuPortal: (state) => cn(classNames?.menuPortal?.(state)),
		multiValue: (state) =>
			cn(multiValueStyles, classNames?.multiValue?.(state)),
		multiValueLabel: (state) => cn(classNames?.multiValueLabel?.(state)),
		multiValueRemove: (state) => cn(classNames?.multiValueRemove?.(state)),
		noOptionsMessage: (state) =>
			cn(noOptionsMessageStyles, classNames?.noOptionsMessage?.(state)),
		option: (state) =>
			cn(
				optionStyles.base,
				state.isFocused && optionStyles.focus,
				state.isDisabled && optionStyles.disabled,
				state.isSelected && optionStyles.selected,
				state.isSelected && state.isFocused && optionStyles.selectedFocus,
				classNames?.option?.(state),
			),
		placeholder: (state) =>
			cn(placeholderStyles, classNames?.placeholder?.(state)),
		singleValue: (state) => cn(classNames?.singleValue?.(state)),
		valueContainer: (state) =>
			cn(valueContainerStyles, classNames?.valueContainer?.(state)),
	};
};

export const createStyles = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option>,
>(): StylesConfig<Option, IsMulti, Group> => ({
	control: (base) => ({
		...base,
		transition: 'none',
		// minHeight: '2.25rem', // we used !min-h-9 instead
	}),
	input: (base) => ({
		...base,
	}),
	menuList: (base) => ({
		...base,
		'::-webkit-scrollbar': {
			background: 'transparent',
		},
		'::-webkit-scrollbar-thumb': {
			background: 'hsl(var(--border))',
		},
		'::-webkit-scrollbar-thumb:hover': {
			background: 'transparent',
		},
		'::-webkit-scrollbar-track': {
			background: 'transparent',
		},
	}),
	multiValueLabel: (base) => ({
		...base,
		overflow: 'visible',
		whiteSpace: 'normal',
	}),
});

export const onBlurWorkaround = (event: FocusEvent<HTMLInputElement>) => {
	const element = event.relatedTarget;
	if (
		element &&
		(element.tagName === 'A' ||
			element.tagName === 'BUTTON' ||
			element.tagName === 'INPUT')
	) {
		(element as HTMLElement).focus();
	}
};

export const getSelectDetails = <
	Option,
	IsMulti extends boolean,
	Group extends GroupBase<Option> = GroupBase<Option>,
>(params: {
	classNames?: ClassNamesConfig<Option, IsMulti, Group>;
	styles?: StylesConfig<Option, IsMulti, Group>;
	fixedHeight?: boolean;
}): {
	classNames: ClassNamesConfig<Option, IsMulti, Group>;
	styles: StylesConfig<Option, IsMulti, Group>;
} => {
	const classNames = createClassNames<Option, IsMulti, Group>(
		params.classNames,
	);
	const styles = params.styles ?? createStyles<Option, IsMulti, Group>();
	if (params.fixedHeight) {
		const originalControl = styles.control;
		styles.control = (base, state) => {
			const existingStyles = originalControl
				? originalControl(base, state)
				: base;
			return {
				...existingStyles,
				height: '36px',
				overflow: 'hidden',
			};
		};
		const originalValueContainer = styles.valueContainer;
		styles.valueContainer = (base, state) => {
			const existingStyles = originalValueContainer
				? originalValueContainer(base, state)
				: base;
			return {
				...existingStyles,
				height: '26px',
			};
		};
	}
	return { classNames, styles };
};
