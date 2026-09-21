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
import { openSection } from './fixtures/validation';

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

	test('latest_version=forgejo: changing the host clears the access token, reverting restores it', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=LATEST_VERSION_FORGEJO_HOST';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;

		// 'Hide inactive' off, so the inactive service below renders once created.
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
		const hostInput = section.getByRole('textbox', {
			name: /^Value field for Host$/i,
		});
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
		await hostInput.fill('https://git.example.com');
		await hostInput.blur();

		// THEN: the credential is cleared immediately, and so is the trust
		// relaxation.
		await expect(tokenInput).toHaveValue('');
		await expect(certToggle).toHaveText(/^Default:?$/);
		await screenshot(page, `${shotDir}/02-host-changed`, testInfo.project.name);

		// WHEN: the host is spelled differently, but still addresses the original.
		await hostInput.fill('CODEBERG.org:443/');
		await hostInput.blur();

		// THEN: both are restored.
		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await expect(certToggle).toHaveText(/^Yes$/);
		await screenshot(
			page,
			`${shotDir}/03-host-reverted`,
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
