import { expect, type Page, test } from '@playwright/test';
import {
	type CreateServiceOptions,
	cleanupServices,
	createService,
	deleteService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshot,
	withProject,
} from './fixtures/service';
import {
	bareEndpoint,
	LOOKUP_BASIC_AUTH,
	LOOKUP_WITH_HEADER_AUTH,
} from './fixtures/test-endpoints';

// Creating a service makes a real server-side network call that can exceed the
// default timeout under load - so triple it.
test.beforeEach(() => {
	test.slow();
});

/**
 * Runs a create -> verify -> refresh -> delete cycle, screenshotting each stage.
 *
 * @param page - The dashboard page.
 * @param id - The (project-suffixed) ID used for backend operations.
 * @param baseID - The raw name used for screenshot paths.
 * @param projectName - The browser project name.
 * @param options - Options passed through to `createService`.
 */
const runCreateServiceTest = async (
	page: Page,
	id: string,
	baseID: string,
	projectName: string,
	options?: CreateServiceOptions,
) => {
	await page.goto('/');
	await page.getByRole('button', { name: /toggle edit mode/i }).click();
	await screenshot(
		page,
		`service-create/${baseID}/01-before-create`,
		projectName,
	);

	await createService(page, id, options);
	// `exact: true` - some IDs here are prefixes of others, and name matching is
	// substring-based by default (would match multiple headings).
	await expect(
		page.getByRole('heading', { exact: true, name: id }),
	).toBeVisible();
	await screenshot(
		page,
		`service-create/${baseID}/02-after-create`,
		projectName,
	);

	// `reload` (unlike `goto`) keeps edit mode active - no need to re-toggle.
	await page.reload();
	await expect(
		page.getByRole('heading', { exact: true, name: id }),
	).toBeVisible();
	await screenshot(
		page,
		`service-create/${baseID}/03-after-refresh`,
		projectName,
	);

	await deleteService(page, id);
	await screenshot(
		page,
		`service-create/${baseID}/04-after-delete`,
		projectName,
	);
};

test.describe('Service creation', () => {
	// Safety net for a test that fails before its own `deleteService` step.
	let createdID: string | undefined;
	test.afterEach(async ({ page }) => {
		if (createdID) await cleanupServices(page, [createdID]);
		createdID = undefined;
	});

	test('latest-version=github', async ({ page }, testInfo) => {
		const baseID = 'LATEST_VERSION=GITHUB';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			latestVersion: {
				type: 'github',
				url: 'release-argus/Argus',
			},
		});
	});

	test('latest-version=url', async ({ page }, testInfo) => {
		const baseID = 'LATEST_VERSION=URL';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			latestVersion: {
				type: 'url',
				url: bareEndpoint('1.2.3'),
			},
		});
	});

	test('deployed-version=manual', async ({ page }, testInfo) => {
		const baseID = 'DEPLOYED_VERSION=MANUAL';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			deployedVersion: {
				type: 'manual',
				version: '1.2.3',
			},
		});
	});

	test('deployed-version=url (JSON)', async ({ page }, testInfo) => {
		const baseID = 'DEPLOYED_VERSION=URL';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			deployedVersion: {
				json: 'version',
				type: 'url',
				url: bareEndpoint('{"version":"1.2.3"}'),
			},
		});
	});

	test('deployed-version=url (basic auth)', async ({ page }, testInfo) => {
		const baseID = 'DEPLOYED_VERSION=URL BASIC-AUTH';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			deployedVersion: {
				basicAuth: {
					password: LOOKUP_BASIC_AUTH.password,
					username: LOOKUP_BASIC_AUTH.username,
				},
				type: 'url',
				url: LOOKUP_BASIC_AUTH.urlValid,
			},
			// /basic-auth returns a sentence, not a bare semver - disable semantic versioning.
			semanticVersioning: false,
		});
	});

	test('deployed-version=url (header auth)', async ({ page }, testInfo) => {
		const baseID = 'DEPLOYED_VERSION=URL HEADER-AUTH';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;
		await runCreateServiceTest(page, id, baseID, testInfo.project.name, {
			deployedVersion: {
				headers: [
					{
						key: LOOKUP_WITH_HEADER_AUTH.headerKey,
						value: LOOKUP_WITH_HEADER_AUTH.headerValuePass,
					},
				],
				type: 'url',
				url: LOOKUP_WITH_HEADER_AUTH.urlValid,
			},
		});
	});
});

test.describe('Service update status', () => {
	// Safety net for a test that fails before its own `deleteService` step.
	let createdID: string | undefined;
	test.afterEach(async ({ page }) => {
		if (createdID) await cleanupServices(page, [createdID]);
		createdID = undefined;
	});

	test('latest_version === deployed_version (up to date)', async ({
		page,
	}, testInfo) => {
		const baseID = 'UPDATE_STATUS=UP_TO_DATE';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();

		// GIVEN: a service whose `latest_version` and `deployed_version` lookups
		// both resolve to the same version (1.2.3).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '1.2.3' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});
		await expect(page.getByRole('heading', { name: id })).toBeVisible();

		const serviceCard = page.locator(`[data-service-id="${id}"]`);

		// THEN: the "Deployed" version is shown as 1.2.3.
		await expect(serviceCard.getByText('1.2.3')).toBeVisible();

		// AND: there is no separate "latest version" indicator for a different
		// version - only the single (deployed) version value is rendered.
		await expect(serviceCard.getByText(/^\d+\.\d+\.\d+$/)).toHaveCount(1);

		// AND: the card is not flagged as having an update available.
		await expect(serviceCard).toHaveAttribute('data-update-available', 'false');

		// AND: there is no "Skip" button (no update to skip).
		await expect(
			serviceCard.getByRole('button', { name: /reject release/i }),
		).not.toBeVisible();

		await screenshot(
			page,
			`service-update-status/${baseID}/01-up-to-date`,
			testInfo.project.name,
		);
	});

	test('latest_version !== deployed_version (update available)', async ({
		page,
	}, testInfo) => {
		const baseID = 'UPDATE_STATUS=AVAILABLE';
		const id = withProject(baseID, testInfo.project.name);
		createdID = id;

		await page.goto('/');
		await page.getByRole('button', { name: /toggle edit mode/i }).click();

		// GIVEN: a service whose `latest_version` (1.2.3) and `deployed_version`
		// (0.0.1) lookups resolve to different versions.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});
		await expect(page.getByRole('heading', { name: id })).toBeVisible();

		const serviceCard = page.locator(`[data-service-id="${id}"]`);
		const skipButton = serviceCard.getByRole('button', {
			name: /reject release/i,
		});

		// THEN: both the "Latest" (1.2.3) and "Deployed" (0.0.1) versions are
		// displayed.
		await expect(serviceCard.getByText('1.2.3')).toBeVisible();
		await expect(serviceCard.getByText('0.0.1')).toBeVisible();

		// AND: the card is flagged as having an update available.
		await expect(serviceCard).toHaveAttribute('data-update-available', 'true');

		// AND: a "Skip" button is visible.
		await expect(skipButton).toBeVisible();
		await screenshot(
			page,
			`service-update-status/${baseID}/01-update-available`,
			testInfo.project.name,
		);

		// WHEN: the user clicks "Skip" and confirms in the resulting modal.
		await skipButton.click();
		const dialog = page.getByRole('dialog', { name: /skip this release\?/i });
		await expect(dialog).toBeVisible();
		await expect(dialog.getByText('Stay on: 0.0.1')).toBeVisible();
		await expect(dialog.getByText('Skip: 1.2.3')).toBeVisible();
		await screenshot(
			page,
			`service-update-status/${baseID}/02-skip-modal`,
			testInfo.project.name,
		);

		await dialog.locator('#modal-action').click();

		// THEN: the modal closes.
		await expect(dialog).not.toBeVisible();

		// AND: the card is no longer flagged as having an update available.
		await expect(serviceCard).toHaveAttribute('data-update-available', 'false');

		// AND: the "Skip" button is no longer shown.
		await expect(skipButton).not.toBeVisible();
		await screenshot(
			page,
			`service-update-status/${baseID}/03-after-skip`,
			testInfo.project.name,
		);
	});
});
