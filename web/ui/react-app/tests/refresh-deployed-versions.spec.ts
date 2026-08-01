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
	let createdIDs: string[] = [];
	test.afterEach(async ({ page }) => {
		if (createdIDs.length) await cleanupServices(page, createdIDs);
		createdIDs = [];
	});

	test('filter dropdown item re-queries the deployed_version of 2+ services', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'refresh-deployed-versions',
		);
		const ids = ['REFRESH-ALL, DV, 1', 'REFRESH-ALL, DV, 2'].map((baseID) =>
			withProject(baseID, testInfo.project.name),
		);
		createdIDs = ids;

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();

		// GIVEN: two services with a manual deployed_version (no network lookup
		// needed for the refresh itself).
		for (const id of ids) {
			await createService(page, id, {
				deployedVersion: { type: 'manual', version: '0.0.1' },
				latestVersion: LOOKUP_LATEST_VERSION_JSON,
			});
			await expect(page.getByRole('heading', { name: id })).toBeVisible();
		}
		await shot('01-before-refresh');

		// WHEN: "Refresh all deployed versions" is picked from the filter
		// dropdown. The backend is shared across specs, so other services may
		// also be refreshed - only assert on the responses for the services this
		// test created, matched by their `service_id` query param (not string
		// containment, since URLSearchParams encodes differently to
		// encodeURIComponent for values like these IDs' spaces/commas).
		await page.getByRole('button', { name: 'Filter shown services' }).click();
		const refreshMenuItem = page.getByRole('menuitem', {
			name: 'Refresh all deployed versions',
		});
		await expect(refreshMenuItem).toBeVisible();
		const [responses] = await Promise.all([
			Promise.all(
				ids.map((id) =>
					page.waitForResponse((res) => {
						if (!res.url().includes('/api/v1/deployed_version/refresh'))
							return false;
						if (res.request().method() !== 'GET') return false;
						return new URL(res.url()).searchParams.get('service_id') === id;
					}),
				),
			),
			refreshMenuItem.click(),
		]);

		// THEN: both refreshes succeed and a completion toast is shown.
		for (const response of responses) {
			expect(response.ok()).toBeTruthy();
		}
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
