import { expect, test } from '@playwright/test';
import { SECRET_VALUE, screenshotsUnder } from './fixtures/service';
import {
	clickViaKeyboard,
	expectError,
	expectValid,
	openCreateServiceModal,
	openSection,
	selectInputFor,
	selectOrCreate,
} from './fixtures/validation';

// The first instance in `tests/fixtures/noauth-config.yml`.
const FIRST_FORGEJO_HOST = 'Forge';
const FIRST_FORGEJO_URL_CANONICAL = 'https://forge.example.com';
const SCHEMALESS_UNIQUE_FORGEJO_HOST = 'forge.other.example.com';
const PROJECT_FORGEJO_URL = 'https://valid.release-argus.io/forgejo';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Latest Version', () => {
		test('type=github - repository is required and must match owner/repo format', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// The type is left as "github", the default.
			const repoInput = section.getByRole('textbox', { name: /repository/i });

			await expectError(
				repoInput,
				undefined,
				'Required.',
				shot,
				'01-github-repo-empty',
			);

			await expectError(
				repoInput,
				'not-a-repo',
				'Invalid repository.',
				shot,
				'02-github-repo-invalid',
			);

			await expectValid(
				repoInput,
				'release-argus/Argus',
				shot,
				'03-github-repo-valid',
			);
		});

		test('type=forgejo - reveals a host field, and both it and the repository are validated', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = selectInputFor(section, 'Host');
			await expect(hostInput).toBeVisible();
			await expect(
				section.locator('input[name="latest_version.host"]'),
			).toHaveValue(FIRST_FORGEJO_HOST);
			await shot('01-forgejo-host-prefilled', section);

			await selectOrCreate(hostInput, 'ftp://codeberg.org');
			await expectError(
				hostInput,
				undefined,
				'Invalid host.',
				shot,
				'02-forgejo-host-invalid',
			);

			await selectOrCreate(hostInput, '  https://codeberg.org  ');
			await expectError(
				hostInput,
				undefined,
				'Invalid host.',
				shot,
				'03-forgejo-host-whitespace',
			);

			await selectOrCreate(hostInput, 'codeberg.org');
			await expectValid(
				hostInput,
				undefined,
				shot,
				'04-forgejo-host-bare-valid',
			);

			const repoInput = section.getByRole('textbox', { name: /repository/i });

			await expectError(
				repoInput,
				'https://codeberg.org/forgejo/forgejo',
				'Invalid repository.',
				shot,
				'05-forgejo-repo-invalid',
			);

			await expectValid(
				repoInput,
				'forgejo/forgejo',
				shot,
				'06-forgejo-repo-valid',
			);
		});

		test('type=forgejo - the repository link is built from the host, normalised', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = selectInputFor(section, 'Host');
			const repoInput = section.getByRole('textbox', { name: /repository/i });
			const repoLink = section.getByRole('link', { name: /open repository/i });

			await repoInput.fill('owner/repo');
			await repoInput.blur();

			await expect(repoLink, 'pre-filled host').toHaveAttribute(
				'href',
				'https://FORGE.example.com:443/owner/repo',
			);

			await selectOrCreate(hostInput, 'codeberg.org');
			await expect(repoLink, 'after scheme-less host').toHaveAttribute(
				'href',
				'https://codeberg.org/owner/repo',
			);

			await selectOrCreate(hostInput, 'https://example.com/git/');
			// The host's trailing slash is dropped, and the repository follows the sub-path.
			await expect(repoLink, 'after sub-path host').toHaveAttribute(
				'href',
				'https://example.com/git/owner/repo',
			);
			await shot('01-forgejo-repo-link', repoInput);

			await selectOrCreate(hostInput, 'ftp://codeberg.org');
			// The host no longer addresses an instance, so there is nowhere to point the link.
			await expect(repoLink).toBeHidden();
			await shot('02-forgejo-repo-link-absent', repoInput);

			await selectOrCreate(hostInput, 'codeberg.org');
			await expect(repoLink, 'after host restored').toHaveAttribute(
				'href',
				'https://codeberg.org/owner/repo',
			);
		});

		test('type=forgejo - a typed token survives the host being changed', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = selectInputFor(section, 'Host');
			const tokenInput = section.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			});

			await tokenInput.fill('dummy-e2e-token');
			await tokenInput.blur();
			await shot('01-forgejo-token-entered', tokenInput);

			await selectOrCreate(hostInput, 'codeberg.org');

			await expect(tokenInput, 'after first host change').toHaveValue(
				'dummy-e2e-token',
			);
			await shot('02-forgejo-token-kept', tokenInput);

			await selectOrCreate(hostInput, 'git.example.com');

			await expect(tokenInput, 'after second host change').toHaveValue(
				'dummy-e2e-token',
			);
			await shot('03-forgejo-token-still-kept', tokenInput);
		});

		test('type=forgejo - a configured host offers its defaults, matched by name', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = selectInputFor(section, 'Host');
			const tokenInput = section.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			});
			const hostField = section.locator('input[name="latest_version.host"]');
			// The Default option carries the inherited allow_invalid_certs boolean.
			const certsDefaultOption = section
				.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
				.getByRole('radio', { name: /^Default:?$/ });

			await hostInput.click();
			// The instance whose name is a URL is offered only once.
			await expect(dialog.getByRole('option')).toHaveText([
				`${FIRST_FORGEJO_HOST} (${FIRST_FORGEJO_URL_CANONICAL})`,
				SCHEMALESS_UNIQUE_FORGEJO_HOST,
				'Internal (https://git.internal.example)',
				`Project (${PROJECT_FORGEJO_URL})`,
			]);
			await hostInput.blur();

			await expect(hostField, 'pre-filled host').toHaveValue(
				FIRST_FORGEJO_HOST,
			);
			// The pre-filled instance offers its token as the placeholder.
			await expect(tokenInput, 'pre-filled host').toHaveAttribute(
				'placeholder',
				SECRET_VALUE,
			);
			await expect(certsDefaultOption, 'pre-filled host').toHaveAttribute(
				'data-default-value',
				'false',
			);
			await shot('04-forgejo-host-defaults-prefilled', section);

			await selectOrCreate(hostInput, 'https://git.internal.example');

			await expect(hostField, 'after Internal host').toHaveValue('Internal');
			await expect(tokenInput, 'after Internal host').toHaveAttribute(
				'placeholder',
				SECRET_VALUE,
			);
			await expect(certsDefaultOption, 'after Internal host').toHaveAttribute(
				'data-default-value',
				'true',
			);
			await shot('05-forgejo-host-defaults-swapped', section);

			await selectOrCreate(hostInput, 'forge.example.com');

			// The URL an option's label shows picks that option, so its entry is matched.
			await expect(hostField, 'after Forge by URL').toHaveValue(
				FIRST_FORGEJO_HOST,
			);
			await expect(tokenInput, 'after Forge by URL').toHaveAttribute(
				'placeholder',
				SECRET_VALUE,
			);
			await expect(certsDefaultOption, 'after Forge by URL').toHaveAttribute(
				'data-default-value',
				'false',
			);
			await shot('06-forgejo-host-defaults-inherited', section);

			await selectOrCreate(hostInput, 'https://git.internal.example');
			await expect(hostField, 'after Internal host again').toHaveValue(
				'Internal',
			);
			await selectOrCreate(hostInput, SCHEMALESS_UNIQUE_FORGEJO_HOST);

			await expect(hostField, 'after URL-named host').toHaveValue(
				SCHEMALESS_UNIQUE_FORGEJO_HOST,
			);
			await expect(tokenInput, 'after URL-named host').toHaveAttribute(
				'placeholder',
				SECRET_VALUE,
			);
			await expect(certsDefaultOption, 'after URL-named host').toHaveAttribute(
				'data-default-value',
				'false',
			);
			await shot('07-forgejo-host-defaults-url-named', section);

			await selectOrCreate(hostInput, 'git.example.com');

			// No defaults entry names this instance, so nothing is inherited.
			await expect(tokenInput).not.toHaveAttribute('placeholder');
			await shot('08-forgejo-host-defaults-absent', section);
		});

		test('type=url - URL is required and must start with http(s)://', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const urlInput = section.getByRole('textbox', {
				name: /^Value field for URL$/i,
			});

			await expectError(urlInput, undefined, 'Required.', shot, '01-url-empty');

			await expectValid(urlInput, 'https://example.com', shot, '02-url-valid');

			await expectError(
				urlInput,
				'example.com',
				"Invalid URL (Must start with 'http://' or 'https://').",
				shot,
				'03-url-invalid',
			);
		});

		test('url_commands - regex command must be a valid regular expression', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await clickViaKeyboard(
				section.getByRole('button', { name: /add new url command/i }),
			);

			const regexInput = section.locator(
				'input[name="latest_version.url_commands.0.regex"]',
			);
			await expect(regexInput).toBeVisible();

			await expectError(
				regexInput,
				'(',
				'Invalid regular expression.',
				shot,
				'01-url-command-regex-invalid',
			);

			await expectValid(
				regexInput,
				'v([0-9.]+)',
				shot,
				'02-url-command-regex-valid',
			);

			await expectError(
				regexInput,
				'',
				'Required.',
				shot,
				'03-url-command-regex-required',
			);
		});

		test('url_commands - replace and split commands require their fields', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await clickViaKeyboard(
				section.getByRole('button', { name: /add new url command/i }),
			);

			await section
				.locator('#latest_version\\.url_commands\\.0\\.type')
				.click();
			await dialog.getByRole('option', { name: /^replace$/i }).click();
			const oldInput = section.locator(
				'input[name="latest_version.url_commands.0.old"]',
			);
			await expectError(
				oldInput,
				undefined,
				'Required.',
				shot,
				'01-url-command-replace-old-empty',
				section,
			);
			await expectValid(
				oldInput,
				'beta',
				shot,
				'02-url-command-replace-old-valid',
				section,
			);

			await section
				.locator('#latest_version\\.url_commands\\.0\\.type')
				.click();
			await dialog.getByRole('option', { name: /^split$/i }).click();
			const textInput = section.locator(
				'input[name="latest_version.url_commands.0.text"]',
			);
			await expectError(
				textInput,
				undefined,
				'Required.',
				shot,
				'03-url-command-split-text-empty',
				section,
			);
			await expectValid(
				textInput,
				'-',
				shot,
				'04-url-command-split-text-valid',
				section,
			);
			const indexInput = section.locator(
				'input[name="latest_version.url_commands.0.index"]',
			);
			await expectError(
				indexInput,
				undefined,
				'Required.',
				shot,
				'05-url-command-split-index-empty',
				section,
			);
			await expectValid(
				indexInput,
				'0',
				shot,
				'06-url-command-split-index-valid',
				section,
			);
		});

		test('require - regex_content and regex_version must be valid regular expressions', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.getByRole('button', { name: /^require:?$/i }).click();

			// Both regex fields are optional, so empty is valid.
			const contentInput = section.locator(
				'input[name="latest_version.require.regex_content"]',
			);
			await expectValid(
				contentInput,
				undefined,
				shot,
				'01-require-regex-content-empty',
				section,
			);
			await expectError(
				contentInput,
				'(',
				'Invalid regular expression.',
				shot,
				'02-require-regex-content-invalid',
				section,
			);
			await expectValid(
				contentInput,
				'v([0-9.]+)',
				shot,
				'03-require-regex-content-valid',
				section,
			);

			const versionInput = section.locator(
				'input[name="latest_version.require.regex_version"]',
			);
			await expectValid(
				versionInput,
				undefined,
				shot,
				'04-require-regex-version-empty',
				section,
			);
			await expectError(
				versionInput,
				'[',
				'Invalid regular expression.',
				shot,
				'05-require-regex-version-invalid',
				section,
			);
			await expectValid(
				versionInput,
				'^[0-9.]+$',
				shot,
				'06-require-regex-version-valid',
				section,
			);
		});

		test('require.docker - image and tag must be set together', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			await section.getByRole('button', { name: /^require:?$/i }).click();

			const imageInput = section.locator(
				'input[name="latest_version.require.docker.image"]',
			);
			const tagInput = section.locator(
				'input[name="latest_version.require.docker.tag"]',
			);

			await imageInput.fill('release-argus/argus');
			await imageInput.blur();
			await tagInput.fill('latest');
			await tagInput.blur();

			await expectError(
				imageInput,
				'',
				'Required.',
				shot,
				'01-require-docker-image-required',
				section,
			);

			await expectValid(
				imageInput,
				'release-argus/argus',
				shot,
				'02-require-docker-paired-valid',
				section,
			);
		});
	});
});
