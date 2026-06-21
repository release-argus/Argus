import { expect, type Page, test } from '@playwright/test';

/**
 * Opens the dashboard and reveals a top-nav dropdown by hovering its button.
 *
 * @param page - The page to navigate.
 * @param menu - The accessible name of the nav button to hover (e.g. 'Status').
 */
const openNav = async (page: Page, menu: string) => {
	await page.goto('/approvals');
	await page.getByRole('button', { name: menu }).hover();
};

test.describe('Navigation', () => {
	test('root should redirect to /approvals', async ({ page }) => {
		await page.goto('/');
		await expect(page).toHaveURL(/.*\/approvals/);
	});

	test.describe('Status dropdown', () => {
		const links: {
			name: string;
			url: RegExp;
			check: (page: Page) => Promise<void>;
		}[] = [
			{
				check: (page) => expect(page.getByRole('table')).toHaveCount(2),
				name: 'Runtime & Build Information',
				url: /.*\/status/,
			},
			{
				check: (page) => expect(page.getByRole('table')).toBeVisible(),
				name: 'Command-Line Flags',
				url: /.*\/flags/,
			},
			{
				check: (page) => expect(page.locator('pre')).toBeVisible(),
				name: 'Configuration',
				url: /.*\/config/,
			},
		];

		for (const { name, url, check } of links) {
			test(name, async ({ page }) => {
				await openNav(page, 'Status');
				await page.getByRole('link', { name }).click();
				await expect(page).toHaveURL(url);
				await check(page);
			});
		}
	});

	test.describe('Help dropdown', () => {
		const links = [
			{
				name: 'GitHub (source)',
				url: 'https://github.com/release-argus/Argus',
			},
			{
				name: 'Report an issue/feature request',
				url: 'https://github.com/release-argus/Argus/issues',
			},
			{ name: 'Docs', url: 'https://release-argus.io/docs' },
		];

		for (const { name, url } of links) {
			test(name, async ({ page }) => {
				await openNav(page, 'Help');
				const link = page.getByRole('link', { name });
				await expect(link).toHaveAttribute('href', url);
				await expect(link).toHaveAttribute('target', '_blank');
			});
		}
	});
});
