import { expect, type Page, test } from '@playwright/test';
import {
	expectServiceLoaded,
	expectUpdateState,
	openDashboardInEditMode,
	recordDeployedVersionRefreshes,
	serviceCard,
	waitForDeployedVersionRefresh,
} from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshotsUnder,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import { bareEndpoint } from './fixtures/test-endpoints';

const SEARCH_INPUT_NAME = 'Search and filter services by name';
const FILTER_BUTTON_NAME = 'Filter shown services';
const REFRESH_MENU_ITEM_NAME = 'Refresh visible deployed versions';

// The version LOOKUP_LATEST_VERSION_JSON resolves to.
const LATEST_VERSION = '1.2.3';

/**
 * Creates a service whose deployed_version is a url lookup, so the refresh
 * action can re-query it.
 *
 * @param page - The dashboard page (edit mode must already be on).
 * @param id - The service ID.
 * @param version - The version the lookup resolves to (Default='0.0.1').
 */
const createTrackedService = (page: Page, id: string, version = '0.0.1') =>
	createService(page, id, {
		deployedVersion: {
			json: 'version',
			type: 'url',
			url: bareEndpoint(`{"version":"${version}"}`),
		},
		latestVersion: LOOKUP_LATEST_VERSION_JSON,
	});

const refreshMenuItem = (page: Page) =>
	page.getByRole('menuitem', { exact: true, name: REFRESH_MENU_ITEM_NAME });

/**
 * Opens the filter dropdown and picks the refresh action.
 *
 * @param page - The dashboard page.
 */
const pickRefreshFromFilterDropdown = async (page: Page) => {
	await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
	const menuItem = refreshMenuItem(page);
	await expect(menuItem).toBeVisible();
	await menuItem.click();
};

/**
 * Picks the refresh action and waits for each of `serviceIDs` to be re-queried.
 *
 * @param page - The dashboard page.
 * @param serviceIDs - The IDs expected to be refreshed.
 */
const refreshAndAwait = async (page: Page, serviceIDs: string[]) => {
	const [responses] = await Promise.all([
		Promise.all(
			serviceIDs.map((id) => waitForDeployedVersionRefresh(page, id)),
		),
		pickRefreshFromFilterDropdown(page),
	]);
	for (const response of responses) {
		expect(response.ok()).toBeTruthy();
	}
	const toast = page
		.locator('[data-sonner-toast][data-type="success"]')
		.filter({ hasText: 'Refreshed deployed versions' })
		.first();
	await expect(toast).toBeVisible();
	await expect(toast).toContainText(`Succeeded: ${serviceIDs.length}`);
};

/**
 * Asserts the refresh action isn't offered for the currently visible services.
 *
 * @param page - The dashboard page.
 */
const expectRefreshActionHidden = async (page: Page) => {
	await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
	// Wait for the menu to render, so the absence below isn't just "not yet".
	await expect(page.getByRole('menu')).toBeVisible();
	await expect(refreshMenuItem(page)).toHaveCount(0);
};

/**
 * Narrows the dashboard to the services matching `term`. The refresh action
 * targets every visible service, so this keeps it off other specs' services.
 *
 * @param page - The dashboard page.
 * @param term - The search term to filter by.
 */
const filterServices = (page: Page, term: string) =>
	page.getByRole('textbox', { name: SEARCH_INPUT_NAME }).fill(term);

/**
 * Flips one of the filter dropdown's 'hide' toggles.
 *
 * @param page - The dashboard page.
 * @param label - The toggle's label, e.g. 'Hide inactive'.
 */
const toggleHideOption = async (page: Page, label: string) => {
	await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
	await page.getByRole('menuitemcheckbox', { name: label }).click();
	await page.keyboard.press('Escape');
	await expect(page.getByRole('menu')).toHaveCount(0);
};

test.describe('refresh visible deployed versions', () => {
	const createdIDs = trackCreatedServices();

	test('filter dropdown item re-queries the deployed_version of 2+ visible services', async ({
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
		createdIDs.push(...ids);

		await openDashboardInEditMode(page);

		for (const id of ids) {
			await createTrackedService(page, id);
		}

		await filterServices(page, 'REFRESH-ALL');
		await shot('01-before-refresh');

		await refreshAndAwait(page, ids);
		await shot('02-after-refresh');
	});

	test('only re-queries deployed_version for services visible under the current search filter', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'refresh-deployed-versions',
		);
		const visibleIDs = ['REFRESH-VISIBLE, DV, 1', 'REFRESH-VISIBLE, DV, 2'].map(
			(baseID) => withProject(baseID, testInfo.project.name),
		);
		const hiddenID = withProject(
			'REFRESH-HIDDEN, DV, 1',
			testInfo.project.name,
		);
		const ids = [...visibleIDs, hiddenID];
		createdIDs.push(...ids);

		await openDashboardInEditMode(page);

		for (const id of ids) {
			await createTrackedService(page, id);
		}
		await shot('01-before-filter');

		await filterServices(page, 'REFRESH-VISIBLE');
		await expect(page.getByRole('heading', { name: hiddenID })).toHaveCount(0);
		for (const id of visibleIDs) {
			await expect(page.getByRole('heading', { name: id })).toBeVisible();
		}
		await shot('02-filtered');

		// Record every deployed_version refresh fired while the filter is
		// applied, so we can assert the hidden service's id never appears.
		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		await refreshAndAwait(page, visibleIDs);
		await shot('03-after-refresh');

		await page.getByRole('button', { name: 'Clear search' }).click();
		await expect(page.getByRole('heading', { name: hiddenID })).toBeVisible();
		await shot('04-filters-cleared');

		expect(refreshedServiceIDs).toEqual(expect.arrayContaining(visibleIDs));
		expect(refreshedServiceIDs).not.toContain(hiddenID);
	});

	test('only re-queries deployed_version for services visible under the current hide filter', async ({
		page,
	}, testInfo) => {
		const upToDateID = withProject(
			'REFRESH-HIDE, DV, up-to-date',
			testInfo.project.name,
		);
		const updatableID = withProject(
			'REFRESH-HIDE, DV, updatable',
			testInfo.project.name,
		);
		createdIDs.push(upToDateID, updatableID);

		await openDashboardInEditMode(page);

		await createTrackedService(page, upToDateID, LATEST_VERSION);
		await createTrackedService(page, updatableID, '0.0.1');

		await filterServices(page, 'REFRESH-HIDE');
		await expectUpdateState(page, upToDateID, 'UP_TO_DATE');
		await expectUpdateState(page, updatableID, 'AVAILABLE');

		await toggleHideOption(page, 'Hide up to date');
		await expect(page.getByRole('heading', { name: upToDateID })).toHaveCount(
			0,
		);
		await expect(
			page.getByRole('heading', { name: updatableID }),
		).toBeVisible();

		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		await refreshAndAwait(page, [updatableID]);

		expect(refreshedServiceIDs).toContain(updatableID);
		expect(refreshedServiceIDs).not.toContain(upToDateID);
	});

	test('only re-queries deployed_version for services visible under the current skipped filter', async ({
		page,
	}, testInfo) => {
		const skippedID = withProject(
			'REFRESH-SKIP, DV, skipped',
			testInfo.project.name,
		);
		const updatableID = withProject(
			'REFRESH-SKIP, DV, updatable',
			testInfo.project.name,
		);
		createdIDs.push(skippedID, updatableID);

		await openDashboardInEditMode(page);

		await createTrackedService(page, skippedID);
		await createTrackedService(page, updatableID);

		await filterServices(page, 'REFRESH-SKIP');
		await expectUpdateState(page, skippedID, 'AVAILABLE');
		await expectUpdateState(page, updatableID, 'AVAILABLE');

		await serviceCard(page, skippedID)
			.getByRole('button', { name: /reject release/i })
			.click();
		const dialog = page.getByRole('dialog', { name: /skip this release\?/i });
		await dialog.locator('#modal-action').click();
		await expect(dialog).not.toBeVisible();
		await expectUpdateState(page, skippedID, 'SKIPPED');

		await toggleHideOption(page, 'Hide skipped');
		await expect(page.getByRole('heading', { name: skippedID })).toHaveCount(0);
		await expect(
			page.getByRole('heading', { name: updatableID }),
		).toBeVisible();

		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		await refreshAndAwait(page, [updatableID]);

		// SKIPPED maps to its own hide toggle, so the up-to-date case does not cover it.
		expect(refreshedServiceIDs).toContain(updatableID);
		expect(refreshedServiceIDs).not.toContain(skippedID);
	});

	test('re-queries a service whose deployed_version already matches latest_version', async ({
		page,
	}, testInfo) => {
		const id = withProject('REFRESH-UP-TO-DATE, DV, 1', testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createTrackedService(page, id, LATEST_VERSION);
		await filterServices(page, 'REFRESH-UP-TO-DATE');
		await expectUpdateState(page, id, 'UP_TO_DATE');

		// Its deployed_version may change, so being up to date doesn't exclude it.
		await refreshAndAwait(page, [id]);
	});

	test('excludes manual deployed_version services from the refresh action', async ({
		page,
	}, testInfo) => {
		const urlID = withProject(
			'REFRESH-EXCLUDE-MANUAL, DV, url',
			testInfo.project.name,
		);
		const manualID = withProject(
			'REFRESH-EXCLUDE-MANUAL, DV, manual',
			testInfo.project.name,
		);
		createdIDs.push(urlID, manualID);

		await openDashboardInEditMode(page);

		await createTrackedService(page, urlID);
		await createService(page, manualID, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		await filterServices(page, 'REFRESH-EXCLUDE-MANUAL');
		await expectServiceLoaded(page, urlID);
		await expectServiceLoaded(page, manualID);

		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		await refreshAndAwait(page, [urlID]);

		expect(refreshedServiceIDs).toContain(urlID);
		// A manual deployed_version has no source to re-query.
		expect(refreshedServiceIDs).not.toContain(manualID);
	});

	test('excludes services with no deployed_version lookup from the refresh action', async ({
		page,
	}, testInfo) => {
		const trackedID = withProject(
			'REFRESH-EXCLUDE-NONE, DV, url',
			testInfo.project.name,
		);
		const untrackedID = withProject(
			'REFRESH-EXCLUDE-NONE, DV, none',
			testInfo.project.name,
		);
		createdIDs.push(trackedID, untrackedID);

		await openDashboardInEditMode(page);

		await createTrackedService(page, trackedID);
		await createService(page, untrackedID, {
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		await filterServices(page, 'REFRESH-EXCLUDE-NONE');
		await expectServiceLoaded(page, trackedID);
		await expectServiceLoaded(page, untrackedID);

		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		await refreshAndAwait(page, [trackedID]);

		expect(refreshedServiceIDs).toContain(trackedID);
		expect(refreshedServiceIDs).not.toContain(untrackedID);
	});

	test('only re-queries deployed_version for services visible under the current inactive filter', async ({
		page,
	}, testInfo) => {
		const activeID = withProject(
			'REFRESH-INACTIVE, DV, active',
			testInfo.project.name,
		);
		const inactiveID = withProject(
			'REFRESH-INACTIVE, DV, inactive',
			testInfo.project.name,
		);
		createdIDs.push(activeID, inactiveID);

		await openDashboardInEditMode(page);

		// Hide inactive is turned off, so an inactive service renders once created.
		await toggleHideOption(page, 'Hide inactive');

		await createTrackedService(page, activeID);
		await createService(page, inactiveID, {
			active: false,
			deployedVersion: {
				json: 'version',
				type: 'url',
				url: bareEndpoint('{"version":"0.0.1"}'),
			},
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		await filterServices(page, 'REFRESH-INACTIVE');
		await expectServiceLoaded(page, activeID);
		await expect(
			serviceCard(page, inactiveID).locator('[data-version-type="deployed"]'),
		).toBeVisible();

		const refreshedServiceIDs = recordDeployedVersionRefreshes(page);

		// Being inactive doesn't exclude a visible service.
		await refreshAndAwait(page, [activeID, inactiveID]);

		refreshedServiceIDs.length = 0;
		await toggleHideOption(page, 'Hide inactive');
		await expect(page.getByRole('heading', { name: inactiveID })).toHaveCount(
			0,
		);

		await refreshAndAwait(page, [activeID]);

		expect(refreshedServiceIDs).toContain(activeID);
		expect(refreshedServiceIDs).not.toContain(inactiveID);
	});

	test('is hidden when the only visible service has a manual deployed_version', async ({
		page,
	}, testInfo) => {
		const id = withProject('REFRESH-ONLY-MANUAL, DV, 1', testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		await filterServices(page, id);
		await expectServiceLoaded(page, id);

		// A manual deployed_version has no source to re-query.
		await expectRefreshActionHidden(page);
	});

	test('is hidden when the current filters leave no visible service with a deployed version', async ({
		page,
	}, testInfo) => {
		const id = withProject(
			'REFRESH-NONE-VISIBLE, DV, 1',
			testInfo.project.name,
		);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createTrackedService(page, id);
		await filterServices(page, 'REFRESH-NONE-VISIBLE');
		await expectServiceLoaded(page, id);
		await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
		// A refreshable service is visible, so the action is on offer.
		await expect(refreshMenuItem(page)).toBeVisible();
		await page.keyboard.press('Escape');

		await filterServices(page, 'z-no-service-matches-this-filter-z');
		await expect(page.getByRole('heading', { name: id })).toHaveCount(0);

		await expectRefreshActionHidden(page);
	});

	test('is reachable the same way regardless of viewport width', async ({
		page,
	}, testInfo) => {
		const id = withProject('REFRESH-VIEWPORT, DV, 1', testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		await createTrackedService(page, id);

		await page.setViewportSize({ height: 844, width: 390 });
		await filterServices(page, 'REFRESH-VIEWPORT');
		await expectServiceLoaded(page, id);

		// The action is only ever reachable through the filter dropdown.
		await expect(refreshMenuItem(page)).toHaveCount(0);

		await page.getByRole('button', { name: FILTER_BUTTON_NAME }).click();
		await expect(refreshMenuItem(page)).toBeVisible();
	});
});
