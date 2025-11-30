import { Moon, Sun, SunMoon } from 'lucide-react';
import { type Theme, useTheme } from '@/components/theme-provider';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';

/**
 * ThemeModeToggle provides a toggle interface for switching between different theme modes:
 * light, dark, and system (browser-preferred).
 */
export const ThemeModeToggle = () => {
	const { setTheme, theme } = useTheme();

	return (
		<ToggleGroup
			className="space-x-1"
			onValueChange={(val: Theme | '') => {
				if (val) setTheme(val);
			}}
			type="single"
			value={theme}
		>
			<ToggleGroupItem
				aria-label="Use light theme"
				title="Use light theme"
				value="light"
			>
				<Sun />
			</ToggleGroupItem>
			<ToggleGroupItem
				aria-label="Use dark theme"
				title="Use dark theme"
				value="dark"
			>
				<Moon />
			</ToggleGroupItem>
			<ToggleGroupItem
				aria-label="Use browser-preferred theme"
				title="Use browser-preferred theme"
				value="system"
			>
				<SunMoon />
			</ToggleGroupItem>
		</ToggleGroup>
	);
};
