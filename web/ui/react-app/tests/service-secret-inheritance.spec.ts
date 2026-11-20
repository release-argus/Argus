import { expect, type Locator, type Page, test } from '@playwright/test';
import {
	DASHBOARD_SHOW_ALL,
	openDashboardInEditMode,
	serviceCard,
} from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	openEditModal,
	openEditModalFromBackend,
	retryCreate,
	SECRET_VALUE,
	saveEdit,
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

/**
 * Creates a service through the API, with the `LOOKUP_LATEST_VERSION_JSON` lookup and `notify`.
 *
 * @param page - The dashboard page.
 * @param id - The service ID.
 * @param notify - The notifiers, as the API takes them.
 */
const createServiceWithNotify = async (
	page: Page,
	id: string,
	notify: Record<string, unknown>[],
) => {
	const { url, urlCommandRegex } = LOOKUP_LATEST_VERSION_JSON;
	await retryCreate(() =>
		page.request.put('api/v1/service/new', {
			data: {
				id,
				latest_version: {
					type: 'url',
					url,
					url_commands: [
						{ index: 0, regex: urlCommandRegex?.regex, type: 'regex' },
					],
				},
				notify,
			},
		}),
	);
};

/**
 * Saves the edit modal, returning the first notifier's params as the UI sent them.
 *
 * @param page - The dashboard page.
 * @param dialog - The open edit dialog.
 */
const saveEditNotifyParams = async (page: Page, dialog: Locator) => {
	const [request] = await Promise.all([
		page.waitForRequest(
			(req) =>
				req.url().includes('/api/v1/service/config') && req.method() === 'PUT',
		),
		saveEdit(dialog),
	]);
	return (
		request.postDataJSON() as { notify: { params: Record<string, string> }[] }
	).notify[0].params;
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

		// The lookup only returns a version (1.2.3) when the header is correct.
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

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const headerValue = section.locator(
			'input[name="latest_version.headers.0.value"]',
		);

		await expect(headerValue).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await setBooleanWithDefault(
			section,
			'latest_version.allow_invalid_certs',
			true,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

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

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const hostInput = selectInputFor(section, 'Host');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});
		const certToggle = section
			.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
			.getByRole('radio', { checked: true });

		await expect(tokenInput, 'on open').toHaveValue(SECRET_VALUE);
		await expect(certToggle, 'on open').toHaveText(/^Yes$/);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await selectOrCreate(hostInput, 'https://git.example.com');

		await expect(tokenInput, 'after host changed').toHaveValue('');
		await expect(certToggle, 'after host changed').toHaveText(/^Default:?$/);
		await screenshot(page, `${shotDir}/02-host-changed`, testInfo.project.name);

		await selectOrCreate(hostInput, 'CODEBERG.org:443/');

		// The host still addresses the original, but its spelling, not its URL, names
		// the instance, so two instances sharing a URL can hold different tokens.
		await expect(tokenInput, 'after host respelled').toHaveValue('');
		await expect(certToggle, 'after host respelled').toHaveText(/^Default:?$/);
		await screenshot(
			page,
			`${shotDir}/03-host-respelled`,
			testInfo.project.name,
		);

		await selectOrCreate(hostInput, 'https://codeberg.org');

		await expect(tokenInput, 'after host restored').toHaveValue(SECRET_VALUE);
		await expect(certToggle, 'after host restored').toHaveText(/^Yes$/);
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

		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: 'dummy-e2e-token',
				type: 'github',
				url: 'release-argus/Argus',
			},
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});

		await expect(tokenInput, 'on open').toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await selectType(section, dialog, 'forgejo');

		// Forgejo shares the Access Token field, but is not offered the GitHub credential.
		await expect(tokenInput, 'after forgejo').toHaveValue('');
		await screenshot(
			page,
			`${shotDir}/02-forgejo-cleared`,
			testInfo.project.name,
		);

		await selectType(section, dialog, 'github');

		await expect(tokenInput, 'after github restored').toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/03-github-restored`,
			testInfo.project.name,
		);

		await section
			.getByRole('textbox', { name: /repository/i })
			.fill('release-argus/argus');
		await saveEdit(dialog);

		const reopened = await openEditModal(page, id);
		const reopenedSection = await openSection(reopened, 'Latest Version');
		await expect(
			reopenedSection.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			}),
			'after save',
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

		// The host is the 'Forge' defaults entry, saved by name.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: 'dummy-e2e-token',
				host: 'Forge',
				type: 'forgejo',
				url: 'forgejo/forgejo',
			},
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const hostInput = selectInputFor(section, 'Host');
		const tokenInput = section.getByRole('textbox', {
			name: /^Value field for Access Token$/i,
		});

		await expect(tokenInput, 'on open').toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await selectOrCreate(hostInput, 'https://git.example.com');

		await expect(tokenInput, 'after host changed').toHaveValue('');
		await screenshot(page, `${shotDir}/02-host-changed`, testInfo.project.name);

		// This is the original instance's URL, spelled with an explicit port.
		await selectOrCreate(hostInput, 'forge.example.com:443');

		await expect(tokenInput, 'after host respelled').toHaveValue('');
		await screenshot(
			page,
			`${shotDir}/03-host-respelled`,
			testInfo.project.name,
		);

		await selectOrCreate(hostInput, 'Forge');

		await expect(tokenInput, 'after host restored').toHaveValue(SECRET_VALUE);
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

		await createService(page, id, {
			active: false,
			latestVersion: {
				...LOOKUP_LATEST_VERSION_JSON,
				allowInvalidCerts: true,
			},
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		const certToggle = section
			.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
			.getByRole('radio', { checked: true });

		await expect(certToggle, 'on open').toHaveText(/^Yes$/);
		await screenshot(page, `${shotDir}/01-certs-stored`, testInfo.project.name);

		await selectType(section, dialog, 'forgejo');
		await selectType(section, dialog, 'url');

		// Forgejo shares the field, yet the stored setting returns rather than the default.
		await expect(certToggle, 'after type round-trip').toHaveText(/^Yes$/);
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

		// The lookup only returns a version (1.2.3) when the header is correct.
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

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Deployed Version');
		const headerValue = section.locator(
			'input[name="deployed_version.headers.0.value"]',
		);

		await expect(headerValue).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await setBooleanWithDefault(
			section,
			'deployed_version.allow_invalid_certs',
			true,
		);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

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

		// An update is available, and the receiver only accepts this Webhook secret.
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

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Webhook');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const secretInput = section.getByRole('textbox', {
			name: 'Value field for Secret',
		});

		await expect(secretInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await section
			.getByRole('textbox', { name: 'Value field for Max tries' })
			.fill('2');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

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

		// The Gotify endpoint is real and only accepts the configured token.
		await createService(page, id, {
			notifiers: [
				{
					name: 'gotify-test',
					text: {
						Host: NOTIFY_GOTIFY.host,
						Path: NOTIFY_GOTIFY.path,
						Title: 'Original Title',
						Token: NOTIFY_GOTIFY.tokenPass,
					},
					typeLabel: 'Gotify',
				},
			],
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const tokenInput = section.getByRole('textbox', {
			name: 'Value field for Token',
		});

		await expect(tokenInput).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await section
			.getByRole('textbox', { name: 'Value field for Title' })
			.fill('Changed Title');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

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

	test('notify: a masked generic header secret survives a key rename', async ({
		page,
	}, testInfo) => {
		const baseID = 'SECRET_INHERIT=NOTIFY_HEADER_RENAME';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `secret-inheritance/${baseID}`;
		const endpoint = new URL(LOOKUP_BASIC_AUTH.urlValid);
		const credentials = `${LOOKUP_BASIC_AUTH.username}:${LOOKUP_BASIC_AUTH.password}`;

		await openDashboardInEditMode(page);

		// The endpoint answers 401 unless the Authorization header holds its credentials.
		await createService(page, id, {
			notifiers: [
				{
					headers: [
						{
							key: 'authorization',
							value: `Basic ${Buffer.from(credentials).toString('base64')}`,
						},
					],
					name: 'generic-test',
					text: { Host: endpoint.host, Path: endpoint.pathname.slice(1) },
					typeLabel: 'Generic Webhook',
				},
			],
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const headerKey = section.locator(
			'input[name="notify.0.url_fields.headers.0.key"]',
		);
		const headerValue = section.locator(
			'input[name="notify.0.url_fields.headers.0.value"]',
		);

		await expect(headerValue).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await headerKey.fill('Authorization');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

		const dialog2 = await openEditModal(page, id);
		const section2 = await openSection(dialog2, 'Notify');
		await section2
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await expect(
			section2.locator('input[name="notify.0.url_fields.headers.0.key"]'),
		).toHaveValue('Authorization');
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

	test('notify: a masked ntfy action extra is sent with its old key when renamed', async ({
		page,
	}, testInfo) => {
		const id = withProject(
			'SECRET_INHERIT=NOTIFY_NTFY_RENAME',
			testInfo.project.name,
		);
		createdIDs.push(id);
		await openDashboardInEditMode(page);
		await createServiceWithNotify(page, id, [
			{
				name: 'ntfy-test',
				params: {
					actions: JSON.stringify([
						{ action: 'broadcast', extras: { a: 'x' }, label: 'B' },
					]),
				},
				type: 'ntfy',
				url_fields: { topic: 'argus' },
			},
		]);

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const row = 'notify.0.params.actions.0.extras.0';
		await expect(section.locator(`input[name="${row}.value"]`)).toHaveValue(
			SECRET_VALUE,
		);
		await section.locator(`input[name="${row}.key"]`).fill('b');

		// Sends the old key with the mask, so the server can restore the secret.
		const params = await saveEditNotifyParams(page, dialog);
		expect(JSON.parse(params.actions)[0].extras).toEqual({
			b: { old_key: 'a', value: SECRET_VALUE },
		});
	});

	test('notify: a masked gotify custom extra is sent with its old namespace when renamed', async ({
		page,
	}, testInfo) => {
		const id = withProject(
			'SECRET_INHERIT=NOTIFY_GOTIFY_RENAME',
			testInfo.project.name,
		);
		createdIDs.push(id);
		await openDashboardInEditMode(page);
		await createServiceWithNotify(page, id, [
			{
				name: 'gotify-test',
				params: { extras: JSON.stringify({ 'my::ns': 'x' }) },
				type: 'gotify',
				url_fields: {
					host: NOTIFY_GOTIFY.host,
					path: NOTIFY_GOTIFY.path,
					token: NOTIFY_GOTIFY.tokenPass,
				},
			},
		]);

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		const row = 'notify.0.params.extras.0';
		await expect(section.locator(`input[name="${row}.value"]`)).toHaveValue(
			SECRET_VALUE,
		);
		await section.locator(`input[name="${row}._namespace"]`).fill('your::ns');

		// Sends the old key with the mask, so the server can restore the secret.
		const params = await saveEditNotifyParams(page, dialog);
		expect(JSON.parse(params.extras)).toEqual({
			'your::ns': { old_key: 'my::ns', value: SECRET_VALUE },
		});
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

		await createService(page, id, {
			active: false,
			latestVersion: {
				host: FORGEJO_HOST_WITH_TOKEN,
				type: 'forgejo',
				url: FORGEJO_PRIVATE_REPO,
			},
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		// The service never holds a token of its own.
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

		// The host's defaults hold no token, so the service's own is the only credential.
		await createService(page, id, {
			active: false,
			latestVersion: {
				accessToken: token,
				host: FORGEJO_HOST_UNMATCHED,
				type: 'forgejo',
				url: FORGEJO_PRIVATE_REPO,
			},
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Latest Version');
		await expect(
			section.getByRole('textbox', { name: /^Value field for Access Token$/i }),
		).toHaveValue(SECRET_VALUE);
		await screenshot(
			page,
			`${shotDir}/01-secret-masked`,
			testInfo.project.name,
		);

		await setBooleanWithDefault(
			section,
			'latest_version.use_prerelease',
			false,
		);
		// The form submits the mask rather than the real token.
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/02-after-save`, testInfo.project.name);

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

		// The lookup only succeeds with the correct basic-auth password.
		await createService(page, id, {
			deployedVersion: {
				basicAuth: {
					password: LOOKUP_BASIC_AUTH.password,
					username: LOOKUP_BASIC_AUTH.username,
				},
				type: 'url',
				url: LOOKUP_BASIC_AUTH.urlValid,
			},
			// /basic-auth returns a sentence, not a bare semver, so semVer is off.
			semanticVersioning: false,
		});

		const dialog = await openEditModalFromBackend(page, id);
		const idInput = dialog.locator('input[name="id"]');
		await expect(idInput).toHaveValue(id);
		await idInput.fill(renamedId);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

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

		// The Gotify endpoint is real and only accepts the configured token.
		await createService(page, id, {
			notifiers: [
				{
					name: 'gotify-test',
					text: {
						Host: NOTIFY_GOTIFY.host,
						Path: NOTIFY_GOTIFY.path,
						Title: 'Original Title',
						Token: NOTIFY_GOTIFY.tokenPass,
					},
					typeLabel: 'Gotify',
				},
			],
		});

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Notify');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section
			.getByRole('textbox', { name: 'Value field for Name' })
			.fill('gotify-renamed');
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

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

		// An update is available, and the receiver only accepts this Webhook secret.
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

		const dialog = await openEditModalFromBackend(page, id);
		const section = await openSection(dialog, 'Webhook');
		await section
			.locator('[data-slot="accordion-trigger"]', { hasText: /^0:/ })
			.click();
		await section
			.getByRole('textbox', { name: 'Value field for Name' })
			.fill(`${id}-renamed`);
		await saveEdit(dialog);
		await screenshot(page, `${shotDir}/01-after-rename`, testInfo.project.name);

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
