import { expect, type Page, type Response } from '@playwright/test';

const DEPLOYED_VERSION_REFRESH_PATH = '/api/v1/deployed_version/refresh';

/**
 * Loads the dashboard and turns edit mode on.
 *
 * @param page - The dashboard page.
 */
export const openDashboardInEditMode = async (page: Page) => {
	await page.goto('/');
	await page.getByRole('button', { name: /toggle edit mode/i }).click();
	await expect(
		page.getByRole('button', { name: 'Create a service' }),
	).toBeVisible();
};

/**
 * Locates a service's card/row on the dashboard.
 *
 * @param page - The dashboard page.
 * @param serviceID - The service ID.
 * @returns The service's `[data-service-id]` container.
 */
export const serviceCard = (page: Page, serviceID: string) =>
	page.locator(`[data-service-id="${serviceID}"]`);

/* The update states a loaded card reports via `data-update-state`. */
export type ServiceUpdateStateName = 'AVAILABLE' | 'SKIPPED' | 'UP_TO_DATE';

/**
 * Locates one of a service's two version items.
 *
 * @param page - The dashboard page.
 * @param serviceID - The service ID.
 * @param versionType - Which version to locate.
 */
const versionItem = (
	page: Page,
	serviceID: string,
	versionType: 'deployed' | 'latest',
) =>
	serviceCard(page, serviceID).locator(`[data-version-type="${versionType}"]`);

export type ExpectedVersions = {
	/* Expected deployed version, or null if that item must not render. */
	deployed: string | null;
	/* Expected latest version, or null if that item must not render. */
	latest: string | null;
};

/**
 * Asserts both of a service's version items.
 *
 * - `{deployed: X, latest: null}`  - a deployed_version lookup, already up to date.
 * - `{deployed: X, latest: Y}`     - a deployed_version lookup, behind the latest.
 * - `{deployed: null, latest: Y}`  - no deployed_version lookup.
 *
 * @param page - The dashboard page.
 * @param serviceID - The service ID.
 * @param versions - The expected version items.
 */
export const expectVersions = async (
	page: Page,
	serviceID: string,
	versions: ExpectedVersions,
) => {
	for (const versionType of ['deployed', 'latest'] as const) {
		const item = versionItem(page, serviceID, versionType);
		const expected = versions[versionType];
		await (expected === null
			? expect(item).toHaveCount(0)
			: expect(item).toHaveText(expected));
	}
};

/**
 * Waits for a service's summary to arrive.
 *
 * @param page - The dashboard page.
 * @param serviceID - The service ID.
 */
export const expectServiceLoaded = (page: Page, serviceID: string) =>
	expect(serviceCard(page, serviceID)).toHaveAttribute(
		'data-update-state',
		/AVAILABLE|SKIPPED|UP_TO_DATE/,
	);

/**
 * Asserts a service's update state.
 *
 * @param page - The dashboard page.
 * @param serviceID - The service ID.
 * @param state - The expected `data-update-state`.
 */
export const expectUpdateState = (
	page: Page,
	serviceID: string,
	state: ServiceUpdateStateName,
) =>
	expect(serviceCard(page, serviceID)).toHaveAttribute(
		'data-update-state',
		state,
	);

/**
 * The `service_id` a `deployed_version/refresh` response was for, or null for
 * any other response.
 *
 * @param response - The response to inspect.
 */
const refreshedServiceID = (response: Response) => {
	if (!response.url().includes(DEPLOYED_VERSION_REFRESH_PATH)) return null;
	if (response.request().method() !== 'GET') return null;
	return new URL(response.url()).searchParams.get('service_id');
};

/**
 * Waits for a `deployed_version/refresh` response.
 *
 * @param page - The dashboard page.
 * @param serviceID - Restrict to the refresh of this service; omit to match the
 *   next refresh of any service.
 */
export const waitForDeployedVersionRefresh = (page: Page, serviceID?: string) =>
	page.waitForResponse((response) => {
		const refreshedID = refreshedServiceID(response);
		if (refreshedID === null) return false;
		return serviceID === undefined || refreshedID === serviceID;
	});

/**
 * Starts recording the service IDs of every `deployed_version/refresh` response,
 * so a test can assert which services were (and weren't) targeted. Reset between
 * phases with `.length = 0`.
 *
 * @param page - The dashboard page.
 * @returns The recorded IDs, appended to as responses arrive.
 */
export const recordDeployedVersionRefreshes = (page: Page) => {
	const serviceIDs: string[] = [];
	page.on('response', (response) => {
		const serviceID = refreshedServiceID(response);
		if (serviceID) serviceIDs.push(serviceID);
	});
	return serviceIDs;
};
