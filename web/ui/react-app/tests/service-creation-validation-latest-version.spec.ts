import { expect, test } from '@playwright/test';
import { SECRET_VALUE, screenshotsUnder } from './fixtures/service';
import {
	clickViaKeyboard,
	expectError,
	expectValid,
	openCreateServiceModal,
	openSection,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Latest Version', () => {
		test('type=github - repository is required and must match owner/repo', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: `latest_version.type` is "github" (default).
			const repoInput = section.getByRole('textbox', { name: /repository/i });

			// GIVEN: the "Repository" field is focused and blurred.
			// THEN: a "Required." error is shown.
			await expectError(
				repoInput,
				undefined,
				'Required.',
				shot,
				'01-github-repo-empty',
			);

			// WHEN: an invalid "owner/repo" value is entered.
			// THEN: an error is shown.
			await expectError(
				repoInput,
				'not-a-repo',
				'Invalid repository.',
				shot,
				'02-github-repo-invalid',
			);

			// WHEN: a valid "owner/repo" value is entered.
			// THEN: the error clears.
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

			// GIVEN: `latest_version.type` is set to "forgejo".
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			// THEN: a "Host" field is revealed.
			const hostInput = section.getByRole('textbox', {
				name: /^Value field for Host$/i,
			});
			await expect(hostInput).toBeVisible();

			// GIVEN: the "Host" field is focused and blurred.
			// THEN: a "Required." error is shown.
			await expectError(
				hostInput,
				undefined,
				'Required.',
				shot,
				'01-forgejo-host-empty',
			);

			// WHEN: a value that cannot address an instance is entered.
			// THEN: an error is shown.
			await expectError(
				hostInput,
				'ftp://codeberg.org',
				'Invalid host.',
				shot,
				'02-forgejo-host-invalid',
			);

			// WHEN: a host padded with whitespace is entered.
			// THEN: an error is shown.
			await expectError(
				hostInput,
				'  https://codeberg.org  ',
				'Invalid host.',
				shot,
				'03-forgejo-host-whitespace',
			);

			// WHEN: a host is entered without a scheme.
			// THEN: the error clears.
			await expectValid(
				hostInput,
				'codeberg.org',
				shot,
				'04-forgejo-host-bare-valid',
			);

			const repoInput = section.getByRole('textbox', { name: /repository/i });

			// WHEN: the repository is given as a full address rather than owner/repo.
			// THEN: an error is shown.
			await expectError(
				repoInput,
				'https://codeberg.org/forgejo/forgejo',
				'Invalid repository.',
				shot,
				'05-forgejo-repo-invalid',
			);

			// WHEN: a valid "owner/repo" value is entered.
			// THEN: the error clears.
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

			// GIVEN: a forgejo service with a repository entered.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = section.getByRole('textbox', {
				name: /^Value field for Host$/i,
			});
			const repoInput = section.getByRole('textbox', { name: /repository/i });
			const repoLink = section.getByRole('link', { name: /open repository/i });

			await repoInput.fill('owner/repo');
			await repoInput.blur();

			// THEN: no link is offered without a host.
			await expect(repoLink).toHaveCount(0);

			// WHEN: a host is entered without a scheme.
			// THEN: the link defaults it to HTTPS.
			await hostInput.fill('codeberg.org');
			await hostInput.blur();
			await expect(repoLink).toHaveAttribute(
				'href',
				'https://codeberg.org/owner/repo',
			);

			// WHEN: a host is entered with a sub-path and a trailing slash.
			// THEN: the trailing slash is dropped, and the repository follows the sub-path.
			await hostInput.fill('https://example.com/git/');
			await hostInput.blur();
			await expect(repoLink).toHaveAttribute(
				'href',
				'https://example.com/git/owner/repo',
			);
			await shot('01-forgejo-repo-link', repoInput);
		});

		test('type=forgejo - a token typed here survives the host being set', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: a forgejo lookup in the create form.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = section.getByRole('textbox', {
				name: /^Value field for Host$/i,
			});
			const tokenInput = section.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			});

			// AND: an access token is entered before any host.
			await tokenInput.fill('dummy-e2e-token');
			await tokenInput.blur();
			await shot('01-forgejo-token-entered', tokenInput);

			// WHEN: the host it was meant for is set afterwards.
			await hostInput.fill('codeberg.org');
			await hostInput.blur();

			// THEN: the token is kept.
			await expect(tokenInput).toHaveValue('dummy-e2e-token');
			await shot('02-forgejo-token-kept', tokenInput);

			// WHEN: the host is changed to a different instance.
			await hostInput.fill('git.example.com');
			await hostInput.blur();

			// THEN: it is still kept.
			await expect(tokenInput).toHaveValue('dummy-e2e-token');
			await shot('03-forgejo-token-still-kept', tokenInput);
		});

		test('type=forgejo - a configured host offers its defaults, matched canonically', async ({
			page,
		}, testInfo) => {
			const shot = screenshotsUnder(
				page,
				testInfo.project.name,
				'service-creation-validation/latest-version',
			);
			const dialog = await openCreateServiceModal(page);
			const section = await openSection(dialog, 'Latest Version');

			// GIVEN: a forgejo lookup, with host defaults configured for two instances.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^forgejo$/i }).click();

			const hostInput = section.getByRole('textbox', {
				name: /^Value field for Host$/i,
			});
			const tokenInput = section.getByRole('textbox', {
				name: /^Value field for Access Token$/i,
			});
			// The inherited allow_invalid_certs boolean shows only as the Default option's
			// icon colour.
			const certsDefaultIcon = section
				.locator('[aria-labelledby="latest_version.allow_invalid_certs-label"]')
				.getByRole('radio', { name: /^Default:?$/ })
				.locator('svg');

			// WHEN: a host the defaults spell differently is entered.
			await hostInput.fill('forge.example.com');
			await hostInput.blur();

			// THEN: that entry is still matched, offering its token as the placeholder.
			await expect(tokenInput).toHaveAttribute('placeholder', SECRET_VALUE);
			// AND: its allow_invalid_certs=false shows as the default.
			await expect(certsDefaultIcon).toHaveClass(/text-destructive/);
			await shot('04-forgejo-host-defaults-inherited', section);

			// WHEN: the host addresses the other configured instance.
			await hostInput.fill('git.internal.example');
			await hostInput.blur();

			// THEN: that entry's defaults apply instead - allow_invalid_certs=true.
			await expect(tokenInput).toHaveAttribute('placeholder', SECRET_VALUE);
			await expect(certsDefaultIcon).toHaveClass(/text-success/);
			await shot('05-forgejo-host-defaults-swapped', section);

			// WHEN: the host has no defaults entry.
			await hostInput.fill('git.example.com');
			await hostInput.blur();

			// THEN: nothing is inherited, so no default is offered.
			await expect(tokenInput).not.toHaveAttribute('placeholder');
			await shot('06-forgejo-host-defaults-absent', section);
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

			// GIVEN: `latest_version.type` is "url".
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();

			const urlInput = section.getByRole('textbox', {
				name: /^Value field for URL$/i,
			});

			// GIVEN: the "URL" field is focused and blurred.
			// THEN: an error is shown.
			await expectError(urlInput, undefined, 'Required.', shot, '01-url-empty');

			// WHEN: a valid URL is entered.
			// THEN: no error is present.
			await expectValid(urlInput, 'https://example.com', shot, '02-url-valid');

			// WHEN: an invalid URL (missing http(s)://) is entered.
			// THEN: an error is shown.
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

			// GIVEN: `latest_version.type` is "url" and we've added a url command.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await clickViaKeyboard(
				section.getByRole('button', { name: /add new url command/i }),
			);

			const regexInput = section.locator(
				'input[name="latest_version.url_commands.0.regex"]',
			);
			await expect(regexInput).toBeVisible();

			// WHEN: an unbalanced (invalid) regular expression is entered.
			// THEN: an error is shown.
			await expectError(
				regexInput,
				'(',
				'Invalid regular expression.',
				shot,
				'01-url-command-regex-invalid',
			);

			// WHEN: the regex is corrected to a valid expression.
			// THEN: the error clears.
			await expectValid(
				regexInput,
				'v([0-9.]+)',
				shot,
				'02-url-command-regex-valid',
			);

			// WHEN: the regex field is emptied.
			// THEN: an error is shown.
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

			// GIVEN: `latest_version.type` is "url" with a url command added.
			await section.locator('#latest_version\\.type').click();
			await dialog.getByRole('option', { name: /^url$/i }).click();
			await clickViaKeyboard(
				section.getByRole('button', { name: /add new url command/i }),
			);

			// WHEN: the command type is switched to "Replace".
			// THEN: its "Replace" (`old`) field is required.
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

			// WHEN: the command type is switched to "Split".
			// THEN: both its "Text" and "Index" fields are required.
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

			// GIVEN: the "Require:" sub-accordion is expanded.
			await section.getByRole('button', { name: /^require:?$/i }).click();

			// Both regex fields are optional (empty is fine) but must compile to a
			// valid RegExp when set.
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

			// GIVEN: the "Require:" sub-accordion is expanded.
			await section.getByRole('button', { name: /^require:?$/i }).click();

			const imageInput = section.locator(
				'input[name="latest_version.require.docker.image"]',
			);
			const tagInput = section.locator(
				'input[name="latest_version.require.docker.tag"]',
			);

			// GIVEN: both image and tag are set (a satisfied pairing - no error).
			await imageInput.fill('release-argus/argus');
			await imageInput.blur();
			await tagInput.fill('latest');
			await tagInput.blur();

			// WHEN: the image is cleared while the tag remains.
			// THEN: the image becomes required - docker needs image and tag together
			// (or neither).
			await expectError(
				imageInput,
				'',
				'Required.',
				shot,
				'01-require-docker-image-required',
				section,
			);

			// WHEN: the image is restored.
			// THEN: the pairing is satisfied, the error clears.
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
