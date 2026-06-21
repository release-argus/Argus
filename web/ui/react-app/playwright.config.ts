import { defineConfig, devices } from '@playwright/test';

/**
 * Read environment variables from file.
 * https://github.com/motdotla/dotenv
 */
// import dotenv from 'dotenv';
// import path from 'path';
// dotenv.config({ path: path.resolve(__dirname, '.env') });

/**
 * Specs that create/delete services must run one-at-a-time per browser.
 * They suffix their service IDs per project to avoid cross-browser collisions,
 * but they all drive a single shared backend and a dashboard that re-renders on
 * every service add/remove. Running them concurrently makes the stateful flows
 * (modal create/save, the ordering drag) race that churn and time out, so each
 * `-mutating` project below caps itself to one worker.
 */
const SERIAL_SPECS = [
	'create-service.spec.ts',
	'webhook.spec.ts',
	'service-secret-inheritance.spec.ts',
	'service-ordering.spec.ts',
];

const BROWSERS = {
	chromium: devices['Desktop Chrome'],
	firefox: devices['Desktop Firefox'],
	webkit: devices['Desktop Safari'],
};

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
	/* Fail the build on CI if you accidentally left test.only in the source code. */
	forbidOnly: !!process.env.CI,
	/* Run tests in files in parallel */
	fullyParallel: true,

	/* Hard ceiling on the whole run so a stuck backend (an unreachable endpoint
	 * or an exhausted API rate limit) can't hang the suite for very long. */
	globalTimeout: process.env.CI ? 30 * 60_000 : 15 * 60_000,

	/* One project per browser, split into a parallel-safe project and a
	 * single-worker `-mutating` project. */
	projects: [
		...Object.entries(BROWSERS).flatMap(([name, use]) => [
			{
				name,
				testIgnore: SERIAL_SPECS,
				use,
			},
			{
				name: `${name}-mutating`,
				testMatch: SERIAL_SPECS,
				use,
				/* Mutating specs run one-at-a-time per browser. */
				workers: 1,
			},
		]),

		/* Mobile viewports run the read-only/validation specs only, to avoid
		 * cross-project contention. Skip navigation since navbar differs on mobile */
		{
			name: 'mobile--google-chrome',
			testIgnore: [...SERIAL_SPECS, 'navigation.spec.ts'],
			use: devices['Pixel 8'],
		},
		{
			name: 'mobile--safari',
			testIgnore: [...SERIAL_SPECS, 'navigation.spec.ts'],
			use: devices['iPhone 16'],
		},
	],
	/* Reporter to use. See https://playwright.dev/docs/test-reporters */
	reporter: process.env.CI
		? [['github'], ['html'], ['json', { outputFile: 'results.json' }]]
		: [['html']],
	/* Retry on CI only */
	retries: process.env.CI ? 2 : 0,
	testDir: './tests',
	/* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
	use: {
		/* Bound individual interactions/navigations so a single stuck action
		 * fails fast rather than riding out the (tripled, via `test.slow()`)
		 * test timeout. */
		actionTimeout: 15_000,
		/* Base URL to use in actions like `await page.goto('')`. */
		baseURL: process.env.BASE_URL ?? 'http://localhost:8080',
		navigationTimeout: 30_000,

		/* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
		trace: 'on-first-retry',
	},
	/* Total worker pool. The `-mutating` projects cap themselves to 1 worker
	 * each regardless of this value; everything else uses the full pool. */
	workers: process.env.CI ? '50%' : undefined,
});
