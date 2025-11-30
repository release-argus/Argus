import { type FC, useEffect, useMemo } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { convertStringArrayToOptionTypeArray } from '@/components/generic/field-select-shared';
import SelectTriState from '@/components/ui/react-select/select-tri-state';
import { useWebSocket } from '@/contexts/websocket';
import type { TagsTriType } from '@/types/util';

/**
 * TagSelect
 *
 * Select and manage service tags with tri-state logic.
 * Prunes tags not present in `monitorData` and sync with the toolbar state.
 */
const TagSelect: FC = () => {
	const { monitorData } = useWebSocket();
	const { values, setTags } = useToolbar();
	const tags: TagsTriType = values.tags ?? { exclude: [], include: [] };

	const tagOptions = useMemo(
		() =>
			convertStringArrayToOptionTypeArray(
				Array.from(monitorData.tags ?? []),
				true,
			),
		[monitorData.tags],
	);

	// Prune unknown tags.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setTags stable.
	useEffect(() => {
		if (!monitorData.tagsLoaded || !monitorData.tags) return;

		const prune = (arr: string[]) =>
			arr.filter((tag) => monitorData.tags?.has(tag));

		const newTags = {
			exclude: prune(tags.exclude),
			include: prune(tags.include),
		};

		if (JSON.stringify(tags) !== JSON.stringify(newTags)) {
			setTags(newTags);
		}
	}, [monitorData.tags, monitorData.tagsLoaded, tags]);

	if (tagOptions.length === 0) return null;

	return (
		<div className="w-80 md:w-120 2xl:w-150">
			<SelectTriState
				aria-label="Select tags to filter services on"
				closeMenuOnSelect={false}
				fixedHeight
				hideSelectedOptions={false}
				isMulti
				onChange={(newValue) => {
					const include = newValue
						.filter((item) => item.state === 'include')
						.map((item) => item.value);

					const exclude = newValue
						.filter((item) => item.state === 'exclude')
						.map((item) => item.value);

					setTags({ exclude, include });
				}}
				options={tagOptions}
				placeholder="Tags..."
				value={[
					...tags.include.map((value) => ({
						state: 'include' as const,
						value,
					})),
					...tags.exclude.map((value) => ({
						state: 'exclude' as const,
						value,
					})),
				]}
			/>
		</div>
	);
};

export default TagSelect;
