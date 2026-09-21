import { expect, type Locator, type Page, test } from '@playwright/test';
import {
	DASHBOARD_SHOW_ALL,
	openDashboardInEditMode,
	serviceCard,
} from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	SECRET_VALUE,
	screenshot,
	setBooleanWithDefault,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import {
	LOOKUP_BASIC_AUTH,
	LOOKUP_WITH_HEADER_AUTH,
	NOTIFY_GOTIFY,
	WEBHOOK_GITHUB,
} from './fixtures/test-endpoints';
import {
	openSection,
	selectInputFor,
	selectOrCreate,
} from './fixtures/validation';

// Instances in `tests/fixtures/noauth-config.yml`, and the private repository they
// reach. `argus/private-releases` answers 404 without a token, so a resolved version
// proves the credential survived the save.
const FORGEJO_HOST_WITH_TOKEN = 'Project';
const FORGEJO_HOST_UNMATCHED = 'https://valid.release-argus.io:443/forgejo';
const FORGEJO_PRIVATE_REPO = 'argus/private-releases';
const FORGEJO_PRIVATE_VERSION = '0.2.0';

/**
 * Opens the edit modal for an existing service (edit mode must already be on).
 *
 * @param page - The dashboard page.
 * @param id - The ID of the service to edit.
 * @returns The open edit dialog.
 */
const openEditModal = async (page: Page, id: string) => {
	const card = serviceCard(page, id);
	await expect(card).toBeVisible();
	await card.getByRole('button', { name: /edit/i }).click();
	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	return dialog;
};

/**
 * Submits the edit modal via "Confirm" and waits for it to close. The button
 * only enables once the form is dirty, so asserting it's enabled also confirms
 * the edit registered. The save re-verifies server-side.
 *
 * @param dialog - The open edit dialog.
 */
const saveEdit = async (dialog: Locator) => {
	const confirm = dialog.locator('#modal-action');
	await expect(confirm).toBeEnabled();
	await confirm.click();
	await expect(dialog).not.toBeVisible({ timeout: 30_000 });
};

/**
 * Picks a `latest_version.type`.
 *
 * @param section - The open 'Latest Version' section.
 * @param dialog - The open edit dialog, which hosts the option list.
 * @param type - The type to select.
 */
const selectType = async (section: Locator, dialog: Locator, type: string) => {
	await section.locator('#latest_version\\.type').click();
	await dialog
		.getByRole('option', { name: new RegExp(`^${type}$`, 'i') })
		.click();
};

test.describe('Service secret inheritance', () => {
	// Safety net for a test that fails before its own cleanup.
	const createdIDs = trackCreatedServices();

	test('latest_version=url: a masked header secret survives an unrelated edit', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service whose latest_version is a header-authenticated URL
		// lookup - the header *value* is a secret, and the lookup only returns a
		// version (1.2.3) when it's correct.
		await createService(page, id, {
			latestVersion: {
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

		// AND: the page is reloaded so the edit modal loads the secret from the
		// backend (where it comes back masked) rather than from create-time state.
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const headerValue = section.locator(
			'input[name="latest_version.headers.0.value"]',
		);

		// THEN: the stored header value is shown masked.
		await expect(headerValue).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: an unrelated field (Allow Invalid Certs) is changed and saved,
		// without modifying the masked header secret.
		await setBooleanWithDefault(
			section,
			'latest_version.allow_invalid_certs',
			true,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		// THEN: reopening and refreshing the latest version still resolves to
		// 1.2.3 - only possible if the header secret was inherited on save (a lost
		// secret would fail the lookup and leave the version blank with an error).
		const dialog2 = await openEditModal(page, id);
		const section2 = await openSection(dialog2, 'Latest Version');
		await section2
			.getByRole('button', { name: /refresh the version/i })
			.click();
		await expect(section2.getByText('Failed to refresh:')).not.toBeVisible();
		await expect(section2.getByText('Latest version: 1.2.3')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`${shotDir}/03-refresh-succeeded`,
			testInfo.project.name,
		);
	});

	test('latest_version=forgejo: changing the host clears the access token, restoring the spelling restores it', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION_FORGEJO_HOST';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive forgejo service with an access token and relaxed
		// certificate trust.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: 'dummy-e2e-token',
				allowInvalidCerts: true,
				host: 'https://codeberg.org',
				type: 'forgejo',
				url: 'forgejo/forgejo',
			},
		});

		// AND: the page is reloaded so the edit modal loads the token from the
		// backend.
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const hostInput = selectInputFor(section, 'Host');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});
		const certToggle = section
			.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
			.getByRole('radio', { checked: true });

		// THEN: the stored token is shown masked, and the trust setting is as stored.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await expect(certToggle).toHaveText(/^Yes$/);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: the host is changed to a different instance.
		await selectOrCreate(hostInput, 'https://git.example.com');

		// THEN: the credential is cleared immediately, and so is the trust
		// relaxation.
		await expect(tokenInput).toHaveValue('');
		await expect(certToggle).toHaveText(/^Default:?$/);
		await screenshot(page, `${shotDir}/02-host-changed`, testInfo.project.name);

		// WHEN: the host is spelled differently, but still addresses the original.
		await selectOrCreate(hostInput, 'CODEBERG.org:443/');

		// THEN: neither returns - a spelling identifies the instance, as two
		// configured instances may share a URL while holding different tokens.
		await expect(tokenInput).toHaveValue('');
		await expect(certToggle).toHaveText(/^Default:?$/);
		await screenshot(
			page,
			`${shotDir}/03-host-respelled`,
			testInfo.project.name,
		);

		// WHEN: the host is spelled as it was stored.
		await selectOrCreate(hostInput, 'https://codeberg.org');

		// THEN: both are restored.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await expect(certToggle).toHaveText(/^Yes$/);
		await screenshot(
			page,
			`${shotDir}/04-host-reverted`,
			testInfo.project.name,
		);
	});

	test('latest_version: a github token survives a round-trip through type=forgejo', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION_TYPE_ROUNDTRIP';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive github service with an access token.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: 'dummy-e2e-token',
				type: 'github',
				url: 'release-argus/Argus',
			},
		});

		// AND: the page is reloaded so the edit modal loads the token from the
		// backend.
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});

		// THEN: the stored token is shown masked.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: the type is switched to forgejo, which shares the Access Token field.
		await selectType(section, dialog, 'forgejo');

		// THEN: the github credential is not offered to the Forgejo instance.
		await expect(tokenInput).toHaveValue('');
		await screenshot(
			page,
			`${shotDir}/02-forgejo-cleared`,
			testInfo.project.name,
		);

		// WHEN: the type is switched back to github.
		await selectType(section, dialog, 'github');

		// THEN: the stored token is offered again, masked rather than blanked.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/03-github-restored`,
			testInfo.project.name,
		);

		// WHEN: an unrelated field is edited and the service saved.
		await section
			.getByRole('textbox', { name: /repository/i })
			.fill('release-argus/argus');
		await saveEdit(dialog);

		// THEN: reopening still shows a stored token - the round-trip kept it.
		const reopened = await openEditModal(page, id);
		const reopenedSection = await openSection(reopened, 'Latest Version');
		await expect(
			reopenedSection.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			}),
		).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/04-token-persisted`,
			testInfo.project.name,
		);
	});

	test('latest_version=forgejo: a configured host name is not matched by URL', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION_FORGEJO_HOST_SPELLING';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive forgejo service saved against the 'Forge' defaults
		// entry by name, with its own access token.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: 'dummy-e2e-token',
				host: 'Forge',
				type: 'forgejo',
				url: 'forgejo/forgejo',
			},
		});
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const hostInput = selectInputFor(section, 'Host');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});

		// THEN: the stored token is shown masked.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: a genuinely different instance is picked.
		await selectOrCreate(hostInput, 'https://git.example.com');

		// THEN: the credential is cleared.
		await expect(tokenInput).toHaveValue('');
		await screenshot(page, `${shotDir}/02-host-changed`, testInfo.project.name);

		// WHEN: the original instance is typed with an explicit port - a new spelling.
		await selectOrCreate(hostInput, 'forge.example.com:443');

		// THEN: it is not recognised as the same instance, so the token remains absent.
		await expect(tokenInput).toHaveValue('');
		await screenshot(
			page,
			`${shotDir}/03-host-respelled`,
			testInfo.project.name,
		);

		// WHEN: the original instance is picked by name.
		await selectOrCreate(hostInput, 'Forge');

		// THEN: it is recognised as the same instance, so the token returns.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/04-host-restored`,
			testInfo.project.name,
		);
	});

	test('latest_version: allow_invalid_certs survives a round-trip through type=forgejo', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION_CERTS_ROUNDTRIP';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive url service that relaxes certificate trust.
		await createService(page, id, {
			active: false,
			latestVersion: {
				...LOOKUP_LATEST_VERSION_JSON,
				allowInvalidCerts: true,
			},
		});
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const certToggle = section
			.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
			.getByRole('radio', { checked: true });

		// THEN: the stored trust setting is shown.
		await expect(certToggle).toHaveText(/^Yes$/);
		await screenshot(page, `${shotDir}/01-certs-stored`, testInfo.project.name);

		// WHEN: the type is switched to forgejo, which shares the field, and back.
		await selectType(section, dialog, 'forgejo');
		await selectType(section, dialog, 'url');

		// THEN: the stored setting returns rather than falling back to the default.
		await expect(certToggle).toHaveText(/^Yes$/);
		await screenshot(
			page,
			`${shotDir}/02-certs-restored`,
			testInfo.project.name,
		);
	});

	test('deployed_version=url: a masked header secret survives an unrelated edit', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=DEPLOYED_VERSION';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service whose deployed_version is a header-authenticated URL
		// lookup - the header *value* is a secret, and the lookup only returns a
		// version (1.2.3) when it's correct.
		await createService(page, id, {
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

		// AND: the page is reloaded so the edit modal loads the secret from the
		// backend (where it comes back masked) rather than from create-time state.
		await page.reload();

		// WHEN: the service is reopened for editing.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Deployed Version');
		const headerValue = section.locator(
			'input[name="deployed_version.headers.0.value"]',
		);

		// THEN: the stored header value is shown masked.
		await expect(headerValue).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: an unrelated field (Allow Invalid Certs) is changed and saved,
		// without modifying the masked header secret.
		await setBooleanWithDefault(
			section,
			'deployed_version.allow_invalid_certs',
			true,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		// THEN: reopening and refreshing the deployed version still resolves to
		// 1.2.3 - only possible if the header secret was inherited on save (a lost
		// secret would fail the lookup and leave the version blank with an error).
		const dialog2 = await openEditModal(page, id);
		const section2 = await openSection(dialog2, 'Deployed Version');
		await section2
			.getByRole('button', { name: /refresh the version/i })
			.click();
		await expect(section2.getByText('Failed to refresh:')).not.toBeVisible();
		await expect(section2.getByText('Deployed version: 1.2.3')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`${shotDir}/03-refresh-succeeded`,
			testInfo.project.name,
		);
	});

	test('webhook: a masked secret survives an unrelated edit', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=WEBHOOK';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service with an update available and a Webhook whose secret is
		// the one the receiver requires.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: {
				...LOOKUP_LATEST_VERSION_JSON,
				allowInvalidCerts: true,
			},
			webhooks: [
				{
					name: id,
					secret: WEBHOOK_GITHUB.secretPass,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await page.reload();

		// WHEN: the service is reopened and its Webhook item expanded.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Webhook');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const secretInput = section.getByRole('textbox', {
			name: 'Value field for Secret',
		});

		// THEN: the stored secret is shown masked.
		await expect(secretInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: an unrelated field (Max tries) is changed and saved, without
		// modifying the masked secret.
		await section
			.getByRole('textbox', { name: 'Value field for Max tries' })
			.fill('2');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		// THEN: sending the Webhook succeeds - only possible if the secret was
		// inherited on save (a lost/corrupted secret is rejected by the receiver).
		const card = serviceCard(page, id);
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const actionDialog = page.getByRole('dialog');
		await expect(actionDialog).toBeVisible();
		await actionDialog
			.getByRole('button', { exact: true, name: 'Send' })
			.click();
		await expect(actionDialog.locator('[aria-label="Successful"]')).toBeVisible(
			{ timeout: 30_000 },
		);
		await screenshot(
			page,
			`${shotDir}/03-send-succeeded`,
			testInfo.project.name,
		);
	});

	test('notify: a masked gotify token survives an unrelated edit', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=NOTIFY';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service with a Gotify notifier pointed at a real endpoint that
		// only accepts the configured token.
		await createService(page, id, {
			notifiers: [
				{
					host: NOTIFY_GOTIFY.host,
					name: 'gotify-test',
					path: NOTIFY_GOTIFY.path,
					title: 'Original Title',
					token: NOTIFY_GOTIFY.tokenPass,
					type: 'gotify',
				},
			],
		});
		await page.reload();

		// WHEN: the service is reopened and its Notify item expanded.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const tokenInput = section.getByRole('textbox', {
			name: 'Value field for Token',
		});

		// THEN: the stored token is shown masked.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// WHEN: an unrelated field (Title) is changed and saved, without
		// modifying the masked token.
		await section
			.getByRole('textbox', { name: 'Value field for Title' })
			.fill('Changed Title');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		// THEN: a test message sends successfully - only possible if the token was
		// inherited on save (a lost/corrupted token gives "invalid gotify token").
		const dialog2 = await openEditModal(page, id);
		const section2 = await openSection(dialog2, 'Notify');
		await section2
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section2.getByRole('button', { name: /send test message/i }).click();
		await expect(dialog2.getByText('Success!')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`${shotDir}/03-test-succeeded`,
			testInfo.project.name,
		);
	});

	test('latest_version=forgejo: an inherited host token survives a save', async ({
		page,
	}, testInfo) => {
		test.skip(
			!process.env.ARGUS_TEST_FORGEJO_TOKEN,
			'ARGUS_TEST_FORGEJO_TOKEN is not set',
		);

		const baseID = 'SECRET_INHERIT=FORGEJO_HOST_TOKEN';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive forgejo service on a host whose defaults hold the token,
		// tracking a repository that answers 404 without it.
		await createService(page, id, {
			active: false,
			latestVersion: {
				host: FORGEJO_HOST_WITH_TOKEN,
				type: 'forgejo',
				url: FORGEJO_PRIVATE_REPO,
			},
		});

		// AND: the page is reloaded so the edit modal loads from the backend.
		await page.reload();

		// WHEN: an unrelated field is changed and the service saved, the service
		// never holding a token of its own.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		await expect(
			section.getByRole('textbox', { name: /^Value field for Access Token$/i }),
		).toHaveValue('');
		await setBooleanWithDefault(
			section,
			'latest_version.use_prerelease',
			false,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-save`, testInfo.project.name);

		// THEN: reopening and refreshing still resolves a version - only possible if
		// the host's token is still inherited by the saved service.
		const reopened = await openEditModal(page, id);
		const reopenedSection = await openSection(reopened, 'Latest Version');
		await reopenedSection
			.getByRole('button', { name: /refresh the version/i })
			.click();
		await expect(
			reopenedSection.getByText('Failed to refresh:'),
		).not.toBeVisible();
		await expect(
			reopenedSection.getByText(`Latest version: ${FORGEJO_PRIVATE_VERSION}`),
		).toBeVisible({ timeout: 30_000 });
		await screenshot(
			page,
			`${shotDir}/02-refresh-succeeded`,
			testInfo.project.name,
		);
	});

	test('latest_version=forgejo: a masked token is restored on save, not stored literally', async ({
		page,
	}, testInfo) => {
		const token = process.env.ARGUS_TEST_FORGEJO_TOKEN;
		test.skip(!token, 'ARGUS_TEST_FORGEJO_TOKEN is not set');

		const baseID = 'SECRET_INHERIT=FORGEJO_OWN_TOKEN';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		await openDashboardInEditMode(page, DASHBOARD_SHOW_ALL);

		// GIVEN: an inactive forgejo service carrying its own token, on a host whose
		// defaults hold none - so the service's token is the only credential.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: token,
				host: FORGEJO_HOST_UNMATCHED,
				type: 'forgejo',
				url: FORGEJO_PRIVATE_REPO,
			},
		});

		// AND: the page is reloaded so the edit modal loads the token from the
		// backend, where it comes back masked.
		await page.reload();

		// WHEN: the service is reopened, the stored token is shown masked.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Latest Version');
		await expect(
			section.getByRole('textbox', { name: /^Value field for Access Token$/i }),
		).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		// AND: an unrelated field is changed and the service saved, so the form
		// submits the mask rather than the real token.
		await setBooleanWithDefault(
			section,
			'latest_version.use_prerelease',
			false,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		// THEN: reopening and refreshing resolves a version - impossible if the mask
		// had been stored literally, since the repository 404s without a real token.
		const reopened = await openEditModal(page, id);
		const reopenedSection = await openSection(reopened, 'Latest Version');
		await reopenedSection
			.getByRole('button', { name: /refresh the version/i })
			.click();
		await expect(
			reopenedSection.getByText('Failed to refresh:'),
		).not.toBeVisible();
		await expect(
			reopenedSection.getByText(`Latest version: ${FORGEJO_PRIVATE_VERSION}`),
		).toBeVisible({ timeout: 30_000 });
		await screenshot(
			page,
			`${shotDir}/03-refresh-succeeded`,
			testInfo.project.name,
		);
	});
});

test.describe('Service secret inheritance on rename', () => {
	// A failed test may leave the service under either its original or its
	// renamed id, so clean up every id the test could have created.
	const createdIDs = trackCreatedServices();

	test('deployed_version=url basic-auth: lookup secret survives a service rename', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_RENAME=SERVICE';
		const id = withProject(baseID, testInfo.project.name);
		const renamedId = withProject(`${baseID}-RENAMED`, testInfo.project.name);
		createdIDs.push(id, renamedId);
		const shotDir = `secret-inheritance-rename/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service whose deployed_version is a basic-auth URL lookup - the
		// password is a secret, and the lookup only succeeds when it's correct.
		await createService(page, id, {
			deployedVersion: {
				basicAuth: {
					password: LOOKUP_BASIC_AUTH.password,
					username: LOOKUP_BASIC_AUTH.username,
				},
				type: 'url',
				url: LOOKUP_BASIC_AUTH.urlValid,
			},
			// /basic-auth returns a sentence, not a bare semver - disable semVer.
			semanticVersioning: false,
		});

		// AND: the page is reloaded so the edit modal loads the secret masked from
		// the backend rather than from create-time state.
		await page.reload();

		// WHEN: the service is reopened and renamed (its ID changed), leaving the
		// masked basic-auth password untouched, then saved.
		const dialog = await openEditModal(page, id);
		const idInput = dialog.locator('input[name="id"]');
		await expect(idInput).toHaveValue(id);
		await idInput.fill(renamedId);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

		// THEN: the renamed service refreshes its deployed version without error
		// - only possible if the password was inherited across the rename.
		const dialog2 = await openEditModal(page, renamedId);
		const section2 = await openSection(dialog2, 'Deployed Version');
		await section2
			.getByRole('button', { name: /refresh the version/i })
			.click();
		await expect(section2.getByText('Failed to refresh:')).not.toBeVisible();
		await expect(section2.getByText(/^Deployed version:/)).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`${shotDir}/02-refresh-succeeded`,
			testInfo.project.name,
		);
	});

	test('notify: a masked gotify token survives a notify rename', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_RENAME=NOTIFY';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance-rename/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service with a Gotify notifier pointed at a real endpoint that
		// only accepts the configured token.
		await createService(page, id, {
			notifiers: [
				{
					host: NOTIFY_GOTIFY.host,
					name: 'gotify-test',
					path: NOTIFY_GOTIFY.path,
					title: 'Original Title',
					token: NOTIFY_GOTIFY.tokenPass,
					type: 'gotify',
				},
			],
		});
		await page.reload();

		// WHEN: the service is reopened, the notifier's *name* changed (leaving the
		// masked token untouched), and saved.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section
			.getByRole('textbox', { name: 'Value field for Name' })
			.fill('gotify-renamed');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

		// THEN: a test message still sends successfully - only possible if the
		// token was inherited across the rename.
		const dialog2 = await openEditModal(page, id);
		const section2 = await openSection(dialog2, 'Notify');
		await section2
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section2.getByRole('button', { name: /send test message/i }).click();
		await expect(dialog2.getByText('Success!')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`${shotDir}/02-test-succeeded`,
			testInfo.project.name,
		);
	});

	test('webhook: a masked secret survives a webhook rename', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_RENAME=WEBHOOK';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance-rename/${baseID}`;

		await openDashboardInEditMode(page);

		// GIVEN: a service with an update available and a Webhook whose secret is
		// the one the receiver requires.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: {
				...LOOKUP_LATEST_VERSION_JSON,
				allowInvalidCerts: true,
			},
			webhooks: [
				{
					name: id,
					secret: WEBHOOK_GITHUB.secretPass,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await page.reload();

		// WHEN: the service is reopened, its Webhook item expanded and *renamed*
		// (leaving the masked secret untouched), then saved.
		const dialog = await openEditModal(page, id);
		const section = await openSection(dialog, 'Webhook');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section
			.getByRole('textbox', { name: 'Value field for Name' })
			.fill(`${id}-renamed`);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

		// THEN: sending the Webhook still succeeds - only possible if the secret
		// was inherited across the rename.
		const card = serviceCard(page, id);
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const actionDialog = page.getByRole('dialog');
		await expect(actionDialog).toBeVisible();
		await actionDialog
			.getByRole('button', { exact: true, name: 'Send' })
			.click();
		await expect(actionDialog.locator('[aria-label="Successful"]')).toBeVisible(
			{ timeout: 30_000 },
		);
		await screenshot(
			page,
			`${shotDir}/02-send-succeeded`,
			testInfo.project.name,
		);
	});
});
