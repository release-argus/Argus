import { expect, type Page, test } from '@playwright/test';
import {
	expectUpdateState,
	expectVersions,
	openDashboardInEditMode,
	serviceCard,
} from './fixtures/dashboard';
import {
	type CreateServiceOptions,
	createService,
	deleteService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshot,
	screenshotsUnder,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import {
	bareEndpoint,
	LOOKUP_BASIC_AUTH,
	LOOKUP_WITH_HEADER_AUTH,
} from './fixtures/test-endpoints';
import {
	expectError,
	expectValid,
	MUST_BE_UNIQUE,
} from './fixtures/validation';

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
	await openDashboardInEditMode(page);
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
	const createdIDs = trackCreatedServices();

	test('latest-version=github', async ({ page }, testInfo) => {
		const baseID = 'LATEST_VERSION=GITHUB';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
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
		createdIDs.push(id);
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
		createdIDs.push(id);
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
		createdIDs.push(id);
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
		createdIDs.push(id);
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
		createdIDs.push(id);
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
	const createdIDs = trackCreatedServices();

	test('latest_version === deployed_version (up to date)', async ({
		page,
	}, testInfo) => {
		const baseID = 'UPDATE_STATUS=UP_TO_DATE';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service whose `latest_version` and `deployed_version` lookups
		// both resolve to the same version (1.2.3).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '1.2.3' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);

		// THEN: the "Deployed" version is shown as 1.2.3, with no separate
		// "latest version" indicator - it collapses into the deployed item
		// when the two match.
		await expectVersions(page, id, { deployed: '1.2.3', latest: null });

		// AND: the card reports itself as up to date.
		await expectUpdateState(page, id, 'UP_TO_DATE');

		// AND: there is no "Skip" button (no update to skip).
		await expect(
			card.getByRole('button', { name: /reject release/i }),
		).not.toBeVisible();

		await screenshot(
			page,
			`service-update-status/${baseID}/01-up-to-date`,
			testInfo.project.name,
		);
	});

	test('no deployed_version lookup shows only the latest version', async ({
		page,
	}, testInfo) => {
		const baseID = 'UPDATE_STATUS=NO_DEPLOYED_VERSION';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service with a `latest_version` lookup, and no
		// `deployed_version` lookup to track what is actually running.
		await createService(page, id, {
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		// THEN: only the latest version is shown - with nothing tracking a
		// deployed version, there is no deployed version item to render.
		await expectVersions(page, id, { deployed: null, latest: '1.2.3' });

		// AND: the card still reports itself as up to date, as an untracked
		// deployed_version is taken to be the latest.
		await expectUpdateState(page, id, 'UP_TO_DATE');
	});

	test('latest_version !== deployed_version (update available)', async ({
		page,
	}, testInfo) => {
		const baseID = 'UPDATE_STATUS=AVAILABLE';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service whose `latest_version` (1.2.3) and `deployed_version`
		// (0.0.1) lookups resolve to different versions.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
		});

		const card = serviceCard(page, id);
		const skipButton = card.getByRole('button', {
			name: /reject release/i,
		});

		// THEN: both the "Latest" (1.2.3) and "Deployed" (0.0.1) versions are
		// displayed, each in its own item.
		await expectVersions(page, id, { deployed: '0.0.1', latest: '1.2.3' });

		// AND: the card is flagged as having an update available.
		await expectUpdateState(page, id, 'AVAILABLE');

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

		// AND: the card now reports the release as skipped, rather than merely
		// 'no update available'.
		await expectUpdateState(page, id, 'SKIPPED');

		// AND: the "Skip" button is no longer shown.
		await expect(skipButton).not.toBeVisible();
		await screenshot(
			page,
			`service-update-status/${baseID}/03-after-skip`,
			testInfo.project.name,
		);
	});
});

test.describe('Service name uniqueness', () => {
	const createdIDs = trackCreatedServices();

	test('a Name already used by another service is flagged as not unique', async ({
		page,
	}, testInfo) => {
		const shot = screenshotsUnder(
			page,
			testInfo.project.name,
			'service-name-uniqueness',
		);
		const baseID = 'NAME_UNIQUE=EXISTING';
		const id = withProject(baseID, testInfo.project.name);
		const name = withProject(`${baseID}-display`, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service whose Name differs from its ID.
		await createService(page, id, { name });

		// WHEN: a second service is given a unique ID, but that existing Name.
		await page.getByRole('button', { name: /create a service/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		const idInput = dialog.locator('input[name="id"]');
		await expect(idInput).toBeEditable();
		await idInput.fill(`${id}-second`);
		await dialog
			.getByRole('button', { name: /toggle to separate id/i })
			.click();
		const nameInput = dialog.locator('input[name="name"]');

		// THEN: the Name is flagged.
		await expectError(
			nameInput,
			name,
			MUST_BE_UNIQUE,
			shot,
			'01-name-duplicate',
		);

		// WHEN: a Name no other service uses is entered.
		// THEN: the error clears.
		await expectValid(nameInput, `${name}-second`, shot, '02-name-unique');

		await dialog.locator('#modal-cancel').click();
		await expect(dialog).not.toBeVisible();
	});
});
