import { expect, type Locator, type Page, test } from '@playwright/test';
import { bareEndpoint } from './test-endpoints';

export type KeyVal = { key: string; value: string };

/**
 * Appends the browser project's name to a base ID so browser tests never
 * collide on the same literal name against the shared backend.
 *
 * @param id - The base service ID.
 * @param projectName - The browser project name (`testInfo.project.name`).
 * @returns The project-suffixed ID.
 */
export const withProject = (id: string, projectName: string) =>
	`${id}-${projectName}`;

export type LatestVersionOptions = {
	/* `latest_version.type` - defaults to 'github'. */
	type?: 'github' | 'url';
	/* `latest_version.url` (the GitHub "Repository" or generic "URL" field). */
	url: string;
	/* `latest_version.allow_invalid_certs` - only applicable to type 'url'. */
	allowInvalidCerts?: boolean;
	/* `latest_version.headers` - only applicable to type 'url'. */
	headers?: KeyVal[];
	/* `latest_version.url_commands` (only the first command, type 'regex') -
	 * only applicable to type 'url'. Used to extract a version from a
	 * response body that isn't a bare semver string. */
	urlCommandRegex?: { regex: string; index?: string };
};

export type DeployedVersionOptions =
	| {
			/* `deployed_version.type` = 'manual'. */
			type: 'manual';
			/* `deployed_version.version`. */
			version: string;
	  }
	| {
			/* `deployed_version.type` = 'url'. */
			type: 'url';
			/* `deployed_version.url`. */
			url: string;
			/* `deployed_version.method` - defaults to 'GET'. */
			method?: 'GET' | 'POST';
			/* `deployed_version.allow_invalid_certs`. */
			allowInvalidCerts?: boolean;
			/* `deployed_version.basic_auth`. */
			basicAuth?: { username: string; password: string };
			/* `deployed_version.headers`. */
			headers?: KeyVal[];
			/* `deployed_version.body` - only sent when `method` is 'POST'. */
			body?: string;
			/* `deployed_version.target_header`. */
			targetHeader?: string;
			/* `deployed_version.json`. */
			json?: string;
			/* `deployed_version.regex`. */
			regex?: string;
	  };

export type WebHookOptions = {
	/* `webhook.X.name`. */
	name: string;
	/* `webhook.X.url`. */
	url: string;
	/* `webhook.X.secret`. */
	secret: string;
	/* `webhook.X.max_tries`. */
	maxTries?: string;
};

/* A `gotify` notifier. */
export type NotifyOptions = {
	/* `notify.X.type` - only 'gotify' is supported here. */
	type: 'gotify';
	/* `notify.X.name`. */
	name: string;
	/* `notify.X.url_fields.host`. */
	host: string;
	/* `notify.X.url_fields.path`. */
	path?: string;
	/* `notify.X.url_fields.token` (the masked secret). */
	token: string;
	/* `notify.X.params.title`. */
	title?: string;
};

export type CreateServiceOptions = {
	latestVersion?: LatestVersionOptions;
	deployedVersion?: DeployedVersionOptions;
	/* `options.semantic_versioning` - set to `false` for lookups whose raw
	 * response isn't a parseable MAJOR.MINOR.PATCH version. */
	semanticVersioning?: boolean;
	notifiers?: NotifyOptions[];
	webhooks?: WebHookOptions[];
};

// A `url` latest-version lookup against the JSON fixture, extracting its
// version via a regex url_command.
export const LOOKUP_LATEST_VERSION_JSON: LatestVersionOptions = {
	allowInvalidCerts: false,
	type: 'url',
	url: bareEndpoint('{"version":"1.2.3"}'),
	urlCommandRegex: {
		index: '0',
		regex: '"version":\\s*"([\\d.]+)"',
	},
};

/**
 * Screenshots `page` to `test-results/screenshots/<dir>/<browser>/<file>.png`.
 *
 * @param page - The page to capture.
 * @param name - Path-like `<dir>/<file>` (no browser segment or extension).
 * @param projectName - Browser project; becomes the `<browser>` segment (its
 *   `-mutating` suffix stripped) so projects don't collide.
 * @param centerOn - If set, scroll this field to centre and capture only the
 *   viewport rather than the full page.
 * @returns The screenshot buffer.
 */
export const screenshot = async (
	page: Page,
	name: string,
	projectName: string,
	centerOn?: Locator,
) => {
	const lastSlash = name.lastIndexOf('/');
	const dir = name.slice(0, lastSlash + 1);
	const file = name.slice(lastSlash + 1);
	const browser = projectName.replace(/-mutating$/, '');
	if (centerOn) {
		await centerOn.evaluate((el) =>
			el.scrollIntoView({ behavior: 'instant', block: 'center' }),
		);
	}
	return page.screenshot({
		// Fast-forward in-flight CSS transitions so a field whose validation
		// just cleared isn't captured mid-fade.
		animations: 'disabled',
		fullPage: !centerOn,
		path: `test-results/screenshots/${dir}${browser}/${file}.png`,
	});
};

// A filesystem-safe slug of a test's title - used as a per-test folder name.
const titleSlug = (title: string) =>
	title.replace(/[^\w=.]+/g, '-').replace(/^-+|-+$/g, '');

/**
 * Binds `screenshot` to a fixed `base` dir and project. Each call's screenshot
 * lands under `base/<test-title>/`, so every test in a describe gets its own
 * folder (the test title comes from `test.info()`).
 *
 * @param page - The page to capture.
 * @param projectName - The browser project name.
 * @param base - The shared screenshot directory prefix.
 * @returns A bound screenshot taker.
 */
export const screenshotsUnder =
	(page: Page, projectName: string, base: string) =>
	(file: string, centerOn?: Locator) =>
		screenshot(
			page,
			`${base}/${titleSlug(test.info().title)}/${file}`,
			projectName,
			centerOn,
		);

/**
 * Fills the headers map for a lookup `section`, adding a row per entry.
 *
 * @param section - The accordion section containing the headers map.
 * @param prefix - The lookup's field-name prefix (`latest_version` or
 *   `deployed_version`); the rows live under `<prefix>.headers`.
 * @param entries - The key/value pairs to add.
 */
const fillHeaders = async (
	section: Locator,
	prefix: 'latest_version' | 'deployed_version',
	entries: KeyVal[],
) => {
	for (let index = 0; index < entries.length; index++) {
		await section.getByRole('button', { name: /add new headers/i }).click();
		await section
			.locator(`input[name="${prefix}.headers.${index}.key"]`)
			.fill(entries[index].key);
		await section
			.locator(`input[name="${prefix}.headers.${index}.value"]`)
			.fill(entries[index].value);
	}
};

/**
 * Sets a Yes/No/Default toggle for `fieldName`.
 *
 * @param section - The accordion section containing the toggle.
 * @param fieldName - The form field name.
 * @param value - `true` selects Yes, `false` No; omit to select Default.
 */
export const setBooleanWithDefault = async (
	section: Locator,
	fieldName: string,
	value?: boolean,
) => {
	const name =
		value === undefined ? /^Default:?$/i : value ? /^Yes$/i : /^No$/i;
	await section
		.locator(`[aria-labelledby="${fieldName}-label"]`)
		.getByRole('radio', { name })
		.click();
};

/**
 * Fills in the `latest_version` accordion section.
 *
 * @param dialog - The modal dialog.
 * @param options - The latest-version lookup options.
 */
const fillLatestVersion = async (
	dialog: Locator,
	options: LatestVersionOptions,
) => {
	const type = options.type ?? 'github';

	await dialog.getByRole('button', { name: /^Latest Version:?$/i }).click();
	const section = dialog
		.locator('[data-slot="accordion-item"]', { hasText: 'Latest Version' })
		.first();

	await section.locator('#latest_version\\.type').click();
	await dialog.getByRole('option', { name: type }).click();

	if (type === 'github') {
		await section
			.getByRole('textbox', { name: /repository/i })
			.fill(options.url);
		return;
	}

	// type === 'url'
	await section
		.getByRole('textbox', { name: /^Value field for URL$/i })
		.fill(options.url);

	if (options.allowInvalidCerts !== undefined) {
		await setBooleanWithDefault(
			section,
			'latest_version.allow_invalid_certs',
			options.allowInvalidCerts,
		);
	}

	if (options.headers?.length) {
		await fillHeaders(section, 'latest_version', options.headers);
	}

	if (options.urlCommandRegex) {
		// Add a regex url_command to extract the version from the response body.
		await section.getByRole('button', { name: /add new url command/i }).click();
		await section
			.locator('input[name="latest_version.url_commands.0.regex"]')
			.fill(options.urlCommandRegex.regex);
		if (options.urlCommandRegex.index !== undefined) {
			await section
				.locator('input[name="latest_version.url_commands.0.index"]')
				.fill(options.urlCommandRegex.index);
		}
	}
};

/**
 * Fills in the `deployed_version` accordion section.
 *
 * @param dialog - The modal dialog.
 * @param options - The deployed-version lookup options.
 */
const fillDeployedVersion = async (
	dialog: Locator,
	options: DeployedVersionOptions,
) => {
	await dialog.getByRole('button', { name: /^Deployed Version:?$/i }).click();
	const section = dialog
		.locator('[data-slot="accordion-item"]', { hasText: 'Deployed Version' })
		.first();

	await section.locator('#deployed_version\\.type').click();
	await dialog.getByRole('option', { name: options.type }).click();

	if (options.type === 'manual') {
		await section
			.locator('input[name="deployed_version.version"]')
			.fill(options.version);
		return;
	}

	// type === 'url'
	await section
		.getByRole('textbox', { name: /^Value field for URL$/i })
		.fill(options.url);

	if (options.method) {
		await section.locator('#deployed_version\\.method').click();
		await dialog
			.getByRole('option', { exact: true, name: options.method })
			.click();
	}

	if (options.allowInvalidCerts !== undefined) {
		await setBooleanWithDefault(
			section,
			'deployed_version.allow_invalid_certs',
			options.allowInvalidCerts,
		);
	}

	if (options.basicAuth) {
		await section
			.locator('input[name="deployed_version.basic_auth.username"]')
			.fill(options.basicAuth.username);
		await section
			.locator('input[name="deployed_version.basic_auth.password"]')
			.fill(options.basicAuth.password);
	}

	if (options.headers?.length) {
		await fillHeaders(section, 'deployed_version', options.headers);
	}

	if (options.method === 'POST' && options.body !== undefined) {
		await section.getByRole('textbox', { name: /^Body$/i }).fill(options.body);
	}

	if (options.targetHeader !== undefined) {
		await section
			.getByRole('textbox', { name: /target header/i })
			.fill(options.targetHeader);
	}

	if (options.json !== undefined) {
		await section.getByRole('textbox', { name: /json/i }).fill(options.json);
	}

	if (options.regex !== undefined) {
		await section
			.getByRole('textbox', { name: 'Value field for RegEx' })
			.fill(options.regex);
	}
};

/**
 * Fills in the `webhook` accordion section, adding one item per entry.
 *
 * @param dialog - The modal dialog.
 * @param webhooks - The webhooks to add.
 */
const fillWebHooks = async (dialog: Locator, webhooks: WebHookOptions[]) => {
	await dialog.getByRole('button', { name: /^WebHook:?$/i }).click();
	const section = dialog
		.locator('[data-slot="accordion-item"]', { hasText: 'WebHook' })
		.first();

	for (let index = 0; index < webhooks.length; index++) {
		await section.getByRole('button', { name: /add webhook/i }).click();

		const item = section.locator('[data-slot="accordion-item"]').nth(index);
		await item.locator('[data-slot="accordion-trigger"]').click();

		await item
			.getByRole('textbox', { name: /^Value field for Name$/i })
			.fill(webhooks[index].name);
		await item
			.getByRole('textbox', { name: /^Value field for Target URL$/i })
			.fill(webhooks[index].url);
		await item
			.getByRole('textbox', { name: /^Value field for Secret$/i })
			.fill(webhooks[index].secret);

		const { maxTries } = webhooks[index];
		if (maxTries !== undefined) {
			await item
				.getByRole('textbox', { name: /^Value field for Max tries$/i })
				.fill(maxTries);
		}
	}
};

/**
 * Fills in the `notify` accordion section, adding one item per entry.
 *
 * @param dialog - The modal dialog.
 * @param notifiers - The notifiers to add.
 */
const fillNotify = async (dialog: Locator, notifiers: NotifyOptions[]) => {
	await dialog.getByRole('button', { name: /^Notify:?$/i }).click();
	const section = dialog
		.locator('[data-slot="accordion-item"]', { hasText: 'Notify' })
		.first();

	for (let index = 0; index < notifiers.length; index++) {
		const notify = notifiers[index];
		await section.getByRole('button', { name: /add notify/i }).click();

		const item = section.locator('[data-slot="accordion-item"]').nth(index);
		await item.locator('[data-slot="accordion-trigger"]').click();

		await section.locator(`#notify\\.${index}\\.type`).click();
		await dialog.getByRole('option', { exact: true, name: 'Gotify' }).click();

		await item
			.getByRole('textbox', { name: /^Value field for Name$/i })
			.fill(notify.name);
		await item
			.getByRole('textbox', { name: /^Value field for Host$/i })
			.fill(notify.host);
		if (notify.path !== undefined) {
			await item
				.getByRole('textbox', { name: /^Value field for Path$/i })
				.fill(notify.path);
		}
		await item
			.getByRole('textbox', { name: /^Value field for Token$/i })
			.fill(notify.token);
		if (notify.title !== undefined) {
			await item
				.getByRole('textbox', { name: /^Value field for Title$/i })
				.fill(notify.title);
		}
	}
};

/**
 * Creates a service. Defaults to a `url` 'latest version' lookup against the
 * test server's `/bare/1.2.3` endpoint; pass `latestVersion`/`deployedVersion`
 * to exercise other lookup types.
 *
 * @param page - The dashboard page (edit mode must already be on).
 * @param id - The service ID.
 * @param options - Lookup/notify/webhook options for the service.
 */
export const createService = async (
	page: Page,
	id: string,
	options?: CreateServiceOptions,
) => {
	await page.getByRole('button', { name: /create a service/i }).click();
	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	await dialog.locator('input[name="id"]').fill(id);

	await fillLatestVersion(
		dialog,
		options?.latestVersion ?? LOOKUP_LATEST_VERSION_JSON,
	);

	if (options?.deployedVersion) {
		await fillDeployedVersion(dialog, options.deployedVersion);
	}

	if (options?.semanticVersioning !== undefined) {
		await dialog.getByRole('button', { name: /^Options:?$/i }).click();
		const optionsSection = dialog
			.locator('[data-slot="accordion-item"]', { hasText: 'Options' })
			.first();
		await setBooleanWithDefault(
			optionsSection,
			'options.semantic_versioning',
			options.semanticVersioning,
		);
	}

	if (options?.notifiers?.length) {
		await fillNotify(dialog, options.notifiers);
	}

	if (options?.webhooks?.length) {
		await fillWebHooks(dialog, options.webhooks);
	}

	await dialog.locator('#modal-action').click();
	// The server verifies the lookup via a real network call before
	// responding, so the modal can take longer than the default 5s to close.
	await expect(dialog).not.toBeVisible({ timeout: 30_000 });
};

/**
 * Deletes the service with the given ID via its edit modal.
 *
 * @param page - The dashboard page (edit mode must already be on).
 * @param serviceID - The ID of the service to delete.
 */
export const deleteService = async (page: Page, serviceID: string) => {
	const serviceCard = page.locator(`[data-service-id="${serviceID}"]`);
	await serviceCard.getByRole('button', { name: /edit/i }).click();

	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	await dialog.getByRole('button', { name: /^Delete$/i }).click();

	// Confirm in the resulting dialog.
	await page
		.getByRole('button', { name: /^Delete$/i })
		.last()
		.click();

	// The card disappears once the delete is broadcast over the WebSocket -
	// allow extra time under parallel load.
	await expect(
		page.locator(`[data-service-id="${serviceID}"]`),
	).not.toBeVisible({ timeout: 30_000 });
};

/**
 * Deletes each of `serviceIDs` that is currently present.
 *
 * @param page - The dashboard page.
 * @param serviceIDs - The IDs to remove if present.
 */
export const cleanupServices = async (page: Page, serviceIDs: string[]) => {
	await page.goto('/');
	await page.getByRole('button', { name: /toggle edit mode/i }).click();
	for (const id of serviceIDs) {
		if (await page.locator(`[data-service-id="${id}"]`).isVisible()) {
			await deleteService(page, id);
		}
	}
};
