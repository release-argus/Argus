import { expect, test } from '@playwright/test';
import {
	cleanupServices,
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshotsUnder,
	withProject,
} from './fixtures/service';

// createService makes a real network call to verify the lookups before the
// modal closes, so allow extra time.
test.beforeEach(() => {
	test.slow();
});

test.describe('refresh all deployed versions', () => {
	let createdID: string | undefined;
	test.afterEach(async ({ page }) => {
		if (createdID) await cleanupServices(page, [createdID]);
		createdID = undefined;
	});

	test('filter dropdown item re-queries the deployed_version of a service', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'refresh-deployed-versions',
		);
		const baseID = 'REFRESH-ALL, DV';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();

		// GIVEN: a service with a manual deployed_version (no network lookup
		// needed for the refresh itself).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});
		await expect(page.getByRole('heading', { name: id })).toBeVisible();
		await shot('01-before-refresh');

		// WHEN: "Refresh all deployed versions" is picked from the filter
		// dropdown. The backend is shared across specs, so other services may
		// also be refreshed - only assert on the response for the service this
		// test created, matched by its `service_id` query param (not string
		// containment, since URLSearchParams encodes differently to
		// encodeURIComponent for values like this ID's spaces/commas).
		await page.getByRole('button', { name: 'Filter shown services' }).click();
		const refreshMenuItem = page.getByRole('menuitem', {
			name: 'Refresh all deployed versions',
		});
		await expect(refreshMenuItem).toBeVisible();
		const [response] = await Promise.all([
			page.waitForResponse((res) => {
				if (!res.url().includes('/api/v1/deployed_version/refresh'))
					return false;
				if (res.request().method() !== 'GET') return false;
				return new URL(res.url()).searchParams.get('service_id') === id;
			}),
			refreshMenuItem.click(),
		]);

		// THEN: the refresh succeeds and a completion toast is shown.
		expect(response.ok()).toBeTruthy();
		await expect(
			page.getByText(/Refreshed deployed version/i).first(),
		).toBeVisible();
		await shot('02-after-refresh');
	});

	test('is reachable the same way regardless of viewport width', async ({
		page,
	}) => {
		await page.setViewportSize({ width: 390, height: 844 });
		await page.goto('/');

		await expect(
			page.getByRole('button', { name: 'Refresh all deployed versions' }),
		).toHaveCount(0);

		await page.getByRole('button', { name: 'Filter shown services' }).click();
		await expect(
			page.getByRole('menuitem', { name: 'Refresh all deployed versions' }),
		).toBeVisible();
	});
});
