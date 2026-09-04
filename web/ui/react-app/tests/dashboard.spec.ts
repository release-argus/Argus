import { expect, type Page, test } from '@playwright/test';
import { expectServiceLoaded, serviceCard } from './fixtures/dashboard';

/**
 * The service seeded from config.yml.example.
 * It's a GitHub lookup with no `deployed_version`, so the first
 * successful query gives it a `deployed_version_timestamp` as
 * well as a `latest_version_timestamp` and a `last_queried`.
 */
const EXAMPLE_SERVICE_ID = 'release-argus/Argus';

/** `aria-label` of the toolbar button holding the filter/timestamp options. */
const FILTER_BUTTON_NAME = 'Filter shown services';

/**
 * Each timestamp's menu label, and the prefix its card line renders with.
 */
const TIMESTAMP = {
	deployed: { label: 'Show deployed', prefix: 'deployed' },
	found: { label: 'Show found', prefix: 'found' },
	queried: { label: 'Show queried', prefix: 'queried' },
} as const;

type TimestampName = keyof typeof TIMESTAMP;

/** The fixed order the card renders the timestamps in. */
const CARD_ORDER: TimestampName[] = ['deployed', 'found', 'queried'];

/**
 * Locates the timestamp lines under the example service's card.
 *
 * @param page - The dashboard page.
 * @returns The `[data-timestamp]` per shown timestamp, in render order.
 */
const timestampLines = (page: Page) =>
	serviceCard(page, EXAMPLE_SERVICE_ID).locator('[data-timestamp]');

/**
 * Asserts exactly which timestamps the example service's card shows, and in
 * which order.
 *
 * @param page - The dashboard page.
 * @param expected - The expected timestamps, in render order.
 */
const expectTimestamps = async (
	page: Page,
	expected: readonly TimestampName[],
) => {
	const lines = timestampLines(page);
	await expect(lines).toHaveCount(expected.length);
	for (const [index, name] of expected.entries()) {
		await expect(lines.nth(index)).toHaveAttribute('data-timestamp', name);
		await expect(lines.nth(index)).toHaveText(
			new RegExp(`^${TIMESTAMP[name].prefix} .+`),
		);
	}
};

/**
 * Opens the filter dropdown.
 *
 * @param page - The dashboard page.
 */
const openFilterDropdown = async (page: Page) => {
	await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
	await expect(page.getByRole('menu')).toBeVisible();
};

/**
 * Locates a timestamp's checkbox in an open filter dropdown.
 *
 * @param page - The dashboard page.
 * @param name - The timestamp to locate.
 */
const timestampOption = (page: Page, name: TimestampName) =>
	page.getByRole('menuitemcheckbox', {
		exact: true,
		name: TIMESTAMP[name].label,
	});

/**
 * Toggles timestamps from the filter dropdown, closing it afterwards.
 *
 * @param page - The dashboard page.
 * @param names - The timestamps to toggle.
 */
const toggleTimestamps = async (page: Page, ...names: TimestampName[]) => {
	await openFilterDropdown(page);
	for (const name of names) await timestampOption(page, name).click();
	await page.keyboard.press('Escape');
	await expect(page.getByRole('menu')).toHaveCount(0);
};

test.describe('Dashboard', () => {
	test('has the correct title', async ({ page }) => {
		await page.goto('/');
		await expect(page).toHaveTitle(/Argus/);
	});

	test('dashboard is visible', async ({ page }) => {
		await page.goto('/');
		await expect(
			page.getByRole('heading', { exact: true, name: EXAMPLE_SERVICE_ID }),
		).toBeVisible();
	});
});

test.describe('Dashboard card timestamps', () => {
	test.beforeEach(async ({ page }) => {
		// GIVEN: the grid view, with a loaded service card.
		await page.goto('/approvals');
		await expectServiceLoaded(page, EXAMPLE_SERVICE_ID);
	});

	test('only the queried timestamp is shown until changed', async ({
		page,
	}) => {
		// WHEN: nothing has been toggled.
		// THEN: the card only carries the 'queried' timestamp.
		await expectTimestamps(page, ['queried']);

		// AND: the dropdown reports the same.
		await openFilterDropdown(page);
		await expect(page.getByText('Timestamps:')).toBeVisible();
		await expect(timestampOption(page, 'deployed')).not.toBeChecked();
		await expect(timestampOption(page, 'found')).not.toBeChecked();
		await expect(timestampOption(page, 'queried')).toBeChecked();
	});

	test('each timestamp can be shown, in a fixed order', async ({ page }) => {
		// WHEN: the other two timestamps are enabled, deployed last.
		await toggleTimestamps(page, 'found', 'deployed');

		// THEN: all three show, in card order rather than click order.
		await expectTimestamps(page, CARD_ORDER);
	});

	test('a timestamp can be hidden again', async ({ page }) => {
		// WHEN: 'found' is enabled and the default 'queried' is disabled.
		await toggleTimestamps(page, 'found', 'queried');

		// THEN: only 'found' remains.
		await expectTimestamps(page, ['found']);
	});

	test('every timestamp can be hidden', async ({ page }) => {
		// WHEN: the only enabled (default) timestamp is disabled.
		await toggleTimestamps(page, 'queried');

		// THEN: the card shows no timestamps.
		await expectTimestamps(page, []);
	});

	test('toggling a timestamp leaves the dropdown open', async ({ page }) => {
		// WHEN: a timestamp is toggled with the dropdown open.
		await openFilterDropdown(page);
		await timestampOption(page, 'found').click();

		// THEN: the dropdown stays open, with the change applied, so more can
		// be toggled without reopening it.
		await expect(page.getByRole('menu')).toBeVisible();
		await expect(timestampOption(page, 'found')).toBeChecked();

		await timestampOption(page, 'deployed').click();
		await expect(page.getByRole('menu')).toBeVisible();
		await expect(timestampOption(page, 'deployed')).toBeChecked();

		await page.keyboard.press('Escape');
		await expectTimestamps(page, CARD_ORDER);
	});

	test('the choice survives a reload', async ({ page }) => {
		// WHEN: the defaults are swapped for 'found' alone, and the page reloads.
		await toggleTimestamps(page, 'found', 'queried');
		await expectTimestamps(page, ['found']);
		await page.reload();
		await expectServiceLoaded(page, EXAMPLE_SERVICE_ID);

		// THEN: the card, and the dropdown, keep that choice.
		await expectTimestamps(page, ['found']);
		await openFilterDropdown(page);
		await expect(timestampOption(page, 'deployed')).not.toBeChecked();
		await expect(timestampOption(page, 'found')).toBeChecked();
		await expect(timestampOption(page, 'queried')).not.toBeChecked();
	});

	test('hiding every timestamp survives a reload', async ({ page }) => {
		// WHEN: every timestamp is disabled, and the page reloads.
		await toggleTimestamps(page, 'queried');
		await page.reload();
		await expectServiceLoaded(page, EXAMPLE_SERVICE_ID);

		// THEN: none come back.
		await expectTimestamps(page, []);
	});
});
