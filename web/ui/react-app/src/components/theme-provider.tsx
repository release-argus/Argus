import {
	createContext,
	type FC,
	type ReactNode,
	use,
	useEffect,
	useMemo,
	useState,
} from 'react';

export type Theme = 'dark' | 'light' | 'system';

type ThemeProviderProps = {
	/* The content to wrap. */
	children: ReactNode;
	/* The default theme to use if no user preference is set. */
	defaultTheme?: Theme;
	/* The key used to store the theme preference in localStorage. */
	storageKey?: string;
};

type ThemeProviderState = {
	/* The current theme. */
	theme: Theme;
	/* Function to set the theme. */
	setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
	setTheme: () => null,
	theme: 'system',
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

/**
 * ThemeProvider manages the theme of the app.
 *
 * @param children - The content to wrap.
 * @param defaultTheme - The default theme to use if no user preference is set.
 * @param storageKey - The key used to store the theme preference in localStorage.
 * @param props - Additional props to pass to the ThemeProviderContext.
 */
export const ThemeProvider: FC<ThemeProviderProps> = ({
	children,
	defaultTheme = 'system',
	storageKey = 'user-prefers-color-scheme',
	...props
}) => {
	const [theme, setTheme] = useState<Theme>(
		() =>
			(localStorage.getItem(storageKey) as Theme | undefined) ?? defaultTheme,
	);

	// biome-ignore lint/correctness/useExhaustiveDependencies: storageKey stable.
	const value = useMemo(
		() => ({
			setTheme: (theme: Theme) => {
				localStorage.setItem(storageKey, theme);
				setTheme(theme);
			},
			theme,
		}),
		[theme],
	);

	// biome-ignore lint/correctness/useExhaustiveDependencies: setTheme stable.
	useEffect(() => {
		const root = globalThis.document.documentElement;
		root.classList.remove('light', 'dark');

		// TODO: Remove 0.28+
		if ((theme as unknown) === '"auto"') value.setTheme('system');

		if (theme === 'system') {
			const systemTheme = globalThis.matchMedia('(prefers-color-scheme: dark)')
				.matches
				? 'dark'
				: 'light';

			root.classList.add(systemTheme);
			return;
		}

		root.classList.add(theme);
	}, [theme]);

	return (
		<ThemeProviderContext {...props} value={value}>
			{children}
		</ThemeProviderContext>
	);
};

/**
 * useTheme retrieves the current theme context value from the ThemeProviderContext.
 */
export const useTheme = () => {
	const context = use(ThemeProviderContext);

	if ((context as ThemeProviderState | undefined) === undefined)
		throw new Error('useTheme must be used within a ThemeProvider');

	return context;
};
