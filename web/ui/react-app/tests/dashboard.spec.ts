import { expect, test } from '@playwright/test';

test.describe('Dashboard', () => {
	test('has the correct title', async ({ page }) => {
		await page.goto('/');
		await expect(page).toHaveTitle(/Argus/);
	});

	test('dashboard is visible', async ({ page }) => {
		await page.goto('/');
		await expect(
			page.getByRole('heading', { exact: true, name: 'release-argus/Argus' }),
		).toBeVisible();
	});
});
