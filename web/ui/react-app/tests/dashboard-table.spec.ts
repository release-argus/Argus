import { expect, test } from '@playwright/test';

/**
 * The service seeded from fixtures/noauth-config.yml.
 * It has an icon, an icon link, and a web URL.
 */
const EXAMPLE_SERVICE = {
	/** `dashboard.icon`. */
	icon:
		'https://raw.githubusercontent.com/release-argus/Argus/master/web/ui/' +
		'react-app/public/favicon.svg',
	/** `dashboard.icon_link_to`. */
	iconLinkTo: 'https://release-argus.io',
	/** The service ID. */
	id: 'release-argus/Argus',
	/** `dashboard.web_url`. */
	webURL: 'https://github.com/release-argus/Argus/blob/master/CHANGELOG.md',
} as const;

test.describe('Dashboard table view', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/approvals?view=table');
	});

	test('the ID links to the service web URL', async ({ page }) => {
		const link = page.getByRole('link', {
			exact: true,
			name: EXAMPLE_SERVICE.id,
		});
		await expect(link).toHaveAttribute('href', EXAMPLE_SERVICE.webURL);
		await expect(link).toHaveAttribute('target', '_blank');
	});

	test('the icon column shows the service icon, linking to icon_link_to', async ({
		page,
	}) => {
		await expect(
			page.getByRole('columnheader', { name: 'Icon' }),
		).toBeVisible();

		// The icon is presentational (empty alt), so it has no accessible role.
		const row = page
			.getByRole('row')
			.filter({ hasText: EXAMPLE_SERVICE.id })
			.first();
		await expect(row.locator('img').first()).toHaveAttribute(
			'src',
			EXAMPLE_SERVICE.icon,
		);

		const iconLink = row.getByRole('link', {
			name: `${EXAMPLE_SERVICE.id} icon link`,
		});
		await expect(iconLink).toHaveAttribute('href', EXAMPLE_SERVICE.iconLinkTo);
		await expect(iconLink).toHaveAttribute('target', '_blank');
	});

	test('the icon column can be hidden, and stays hidden', async ({ page }) => {
		const iconHeader = page.getByRole('columnheader', { name: 'Icon' });
		await expect(iconHeader).toBeVisible();

		await page.getByRole('button', { name: 'Filter shown services' }).click();
		await page.getByRole('menuitemcheckbox', { name: 'Icon' }).click();
		await page.keyboard.press('Escape');
		await expect(iconHeader).toBeHidden();

		await page.reload();
		await expect(page.getByRole('table')).toBeVisible();
		await expect(iconHeader).toBeHidden();
	});
});
