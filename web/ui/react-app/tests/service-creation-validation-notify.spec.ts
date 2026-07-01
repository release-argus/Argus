import { expect, test } from '@playwright/test';
import { screenshotsUnder } from './fixtures/service';
import {
	addNotify,
	clickViaKeyboard,
	expectError,
	expectValid,
	fieldOf,
	MUST_BE_UNIQUE,
	numeric,
	openCreateServiceModal,
	openSection,
	optional,
	REQUIRED,
	required,
	runValidations,
	waitForAccordionAnimations,
} from './fixtures/validation';

test.describe('Service creation modal - field validation', () => {
	// These tests never submit, so they're safe to run fully parallel.
	test.describe.configure({ mode: 'parallel' });

	test.describe('Notify', () => {
		test.describe('Bark', () => {
			test('host and device key are required; port and badge must be valid numbers', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/bark',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Bark" notifier is added.
				await addNotify(section, dialog, 'Bark');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({ good: 'bark.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '8080', input: 'Port', slug: 'port' }),
					required({
						good: 'abc123devicekey',
						input: 'Device Key',
						slug: 'devicekey',
					}),
					numeric({ good: '1', input: 'Badge', slug: 'badge' }),
				]);
			});
		});

		test.describe('Discord', () => {
			test('name, WebHook ID, and token are each required', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/discord',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Discord" notifier is added.
				await addNotify(section, dialog, 'Discord');

				// WHEN/THEN: each field is blurred empty then valid.
				await runValidations(section, shot, [
					required({ good: 'my-discord', input: 'Name', slug: 'name' }),
					required({ good: '123456', input: 'WebHook ID', slug: 'webhookid' }),
					required({ good: 'abcdef', input: 'Token', slug: 'token' }),
				]);
			});
		});

		test.describe('Google Chat', () => {
			test('raw is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/googlechat',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Google Chat" notifier is added.
				await addNotify(section, dialog, 'Google Chat');

				// WHEN/THEN: Raw is blurred empty then valid.
				await runValidations(section, shot, [
					required({
						good: 'chat.googleapis.com/v1/spaces/foo/messages?key=bar&token=baz',
						input: 'Raw',
						inputType: 'textbox',
						slug: 'raw',
					}),
				]);
			});
		});

		test.describe('Gotify', () => {
			test('host and token are required; port and priority must be valid numbers', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/gotify',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Gotify" notifier is added.
				await addNotify(section, dialog, 'Gotify');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({ good: 'gotify.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '443', input: 'Port', slug: 'port' }),
					required({ good: 'abc123token', input: 'Token', slug: 'token' }),
					numeric({ good: '5', input: 'Priority', slug: 'priority' }),
				]);
			});

			test('extras (namespace "other") requires name and value', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/gotify',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Gotify" notifier is added.
				await addNotify(section, dialog, 'Gotify');

				// WHEN: an "extra" is added and its namespace switched to "other".
				await section.getByRole('button', { name: /add new extras/i }).click();
				await section
					.locator('#notify\\.0\\.params\\.extras\\.0\\.namespace')
					.click();
				await dialog
					.getByRole('option', { exact: true, name: /^other$/i })
					.click();

				// Scoped by `name` attribute - the extras "Name" shares its
				// "Value field for Name" aria-label with the notifier "Name".
				const nameInput = section.locator(
					'input[name="notify.0.params.extras.0._namespace"]',
				);
				const nameField = fieldOf(nameInput);
				const valueInput = section.locator(
					'input[name="notify.0.params.extras.0.value"]',
				);
				const valueField = fieldOf(valueInput);

				// WHEN: "Name" and "Value" are left empty and blurred.
				await nameInput.fill('');
				await nameInput.blur();
				await valueInput.fill('');
				await valueInput.blur();

				// THEN: an error is shown for both.
				await expect(nameInput).toHaveAttribute('aria-invalid', 'true');
				await expect(nameField.getByRole('alert')).toHaveText(REQUIRED);
				await expect(valueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(valueField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-extras-empty', valueField);

				// WHEN: values are entered.
				await nameInput.fill('my-namespace');
				await nameInput.blur();
				await valueInput.fill('{"key": "value"}');
				await valueInput.blur();

				// THEN: both errors clear.
				await expect(nameInput).toHaveAttribute('aria-invalid', 'false');
				await expect(nameField.getByRole('alert')).not.toBeVisible();
				await expect(valueInput).toHaveAttribute('aria-invalid', 'false');
				await expect(valueField.getByRole('alert')).not.toBeVisible();
				await shot('02-extras-valid', valueField);
			});
		});

		test.describe('Generic', () => {
			test('host is required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/generic',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Generic WebHook" notifier is added.
				await addNotify(section, dialog, 'Generic WebHook');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({
						good: 'generic.example.com',
						input: 'Host',
						slug: 'host',
					}),
					numeric({ good: '8080', input: 'Port', slug: 'port' }),
				]);
			});

			test('headers, JSON payload vars, and query vars require a key and value, and keys must be unique', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/generic',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Generic WebHook" notifier is added.
				await addNotify(section, dialog, 'Generic WebHook');

				// WHEN: a header row is added with an empty key and value.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new headers/i }),
				);
				const headerKeyInput = section.locator(
					'input[name="notify.0.url_fields.headers.0.key"]',
				);
				const headerValueInput = section.locator(
					'input[name="notify.0.url_fields.headers.0.value"]',
				);
				await headerKeyInput.focus();
				await headerKeyInput.blur();
				await headerValueInput.focus();
				await headerValueInput.blur();

				// THEN: an error is shown for both.
				await expect(headerKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(headerValueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerValueInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-headers-empty');

				// WHEN: a second header row is added with the same key as the first.
				await headerKeyInput.fill('X-Header');
				await headerValueInput.fill('one');
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new headers/i }),
				);
				const headerKeyInput2 = section.locator(
					'input[name="notify.0.url_fields.headers.1.key"]',
				);
				const headerValueInput2 = section.locator(
					'input[name="notify.0.url_fields.headers.1.value"]',
				);
				await headerKeyInput2.fill('X-Header');
				await headerValueInput2.fill('two');
				await headerValueInput2.blur();
				// Re-touch the first key so its duplicate error re-evaluates.
				await headerKeyInput.focus();
				await headerKeyInput.blur();

				// THEN:an error is shown on both duplicate keys.
				await expect(headerKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(headerKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('02-headers-duplicate-key');

				// WHEN: the second key is changed to be unique.
				await headerKeyInput2.fill('X-Header-2');
				await headerKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await headerKeyInput.focus();
				await headerKeyInput.blur();

				// THEN: both errors clear.
				await expect(headerKeyInput).toHaveAttribute('aria-invalid', 'false');
				await expect(
					fieldOf(headerKeyInput).getByRole('alert'),
				).not.toBeVisible();
				await expect(headerKeyInput2).toHaveAttribute('aria-invalid', 'false');
				await expect(
					fieldOf(headerKeyInput2).getByRole('alert'),
				).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('03-headers-valid');

				// JSON Payload vars only render once "Template" has a value.
				const templateInput = section.getByRole('textbox', {
					name: 'Value field for Template',
				});
				await templateInput.fill('{}');
				await templateInput.blur();

				// WHEN: a JSON payload var row is added with an empty key and value.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new json payload vars/i }),
				);
				const payloadVarKeyInput = section.locator(
					'input[name="notify.0.url_fields.json_payload_vars.0.key"]',
				);
				const payloadVarValueInput = section.locator(
					'input[name="notify.0.url_fields.json_payload_vars.0.value"]',
				);
				await payloadVarKeyInput.focus();
				await payloadVarKeyInput.blur();
				await payloadVarValueInput.focus();
				await payloadVarValueInput.blur();

				// THEN: an error is shown for both.
				await expect(payloadVarKeyInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(payloadVarKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(payloadVarValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(
					fieldOf(payloadVarValueInput).getByRole('alert'),
				).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('04-json-payload-vars-empty');

				// WHEN: a second JSON payload var row is added with the same key.
				await payloadVarKeyInput.fill('titleKey');
				await payloadVarValueInput.fill('title');
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new json payload vars/i }),
				);
				const payloadVarKeyInput2 = section.locator(
					'input[name="notify.0.url_fields.json_payload_vars.1.key"]',
				);
				const payloadVarValueInput2 = section.locator(
					'input[name="notify.0.url_fields.json_payload_vars.1.value"]',
				);
				await payloadVarKeyInput2.fill('titleKey');
				await payloadVarValueInput2.fill('message');
				await payloadVarValueInput2.blur();
				// Re-touch the first key so its duplicate error re-evaluates.
				await payloadVarKeyInput.focus();
				await payloadVarKeyInput.blur();

				// THEN: an error is shown on both duplicate keys.
				await expect(payloadVarKeyInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(payloadVarKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(payloadVarKeyInput2).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(
					fieldOf(payloadVarKeyInput2).getByRole('alert'),
				).toHaveText(MUST_BE_UNIQUE);
				await waitForAccordionAnimations(section);
				await shot('05-json-payload-vars-duplicate-key');

				// WHEN: the second key is changed to be unique.
				await payloadVarKeyInput2.fill('messageKey');
				await payloadVarKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await payloadVarKeyInput.focus();
				await payloadVarKeyInput.blur();

				// THEN: both errors clear.
				await expect(payloadVarKeyInput).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(
					fieldOf(payloadVarKeyInput).getByRole('alert'),
				).not.toBeVisible();
				await expect(payloadVarKeyInput2).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(
					fieldOf(payloadVarKeyInput2).getByRole('alert'),
				).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('06-json-payload-vars-valid');

				// WHEN: a query var row is added with an empty key and value.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new query vars/i }),
				);
				const queryVarKeyInput = section.locator(
					'input[name="notify.0.url_fields.query_vars.0.key"]',
				);
				const queryVarValueInput = section.locator(
					'input[name="notify.0.url_fields.query_vars.0.value"]',
				);
				await queryVarKeyInput.focus();
				await queryVarKeyInput.blur();
				await queryVarValueInput.focus();
				await queryVarValueInput.blur();

				// THEN: an error is shown for both.
				await expect(queryVarKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(queryVarKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(queryVarValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(queryVarValueInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('07-query-vars-empty');

				// WHEN: a second query var row is added with the same key as the first.
				await queryVarKeyInput.fill('foo');
				await queryVarValueInput.fill('one');
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new query vars/i }),
				);
				const queryVarKeyInput2 = section.locator(
					'input[name="notify.0.url_fields.query_vars.1.key"]',
				);
				const queryVarValueInput2 = section.locator(
					'input[name="notify.0.url_fields.query_vars.1.value"]',
				);
				await queryVarKeyInput2.fill('foo');
				await queryVarValueInput2.fill('two');
				await queryVarValueInput2.blur();
				// Re-touch the first key so its duplicate error re-evaluates.
				await queryVarKeyInput.focus();
				await queryVarKeyInput.blur();

				// THEN:an error is shown on both duplicate keys.
				await expect(queryVarKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(queryVarKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(queryVarKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(queryVarKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('08-query-vars-duplicate-key');

				// WHEN: the second key is changed to be unique.
				await queryVarKeyInput2.fill('bar');
				await queryVarKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await queryVarKeyInput.focus();
				await queryVarKeyInput.blur();

				// THEN: both errors clear.
				await expect(queryVarKeyInput).toHaveAttribute('aria-invalid', 'false');
				await expect(
					fieldOf(queryVarKeyInput).getByRole('alert'),
				).not.toBeVisible();
				await expect(queryVarKeyInput2).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(
					fieldOf(queryVarKeyInput2).getByRole('alert'),
				).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('09-query-vars-valid');
			});
		});

		test.describe('IFTTT', () => {
			test('WebHook ID and events are required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/ifttt',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "IFTTT" notifier is added.
				await addNotify(section, dialog, 'IFTTT');

				// WHEN/THEN: each required field is blurred empty then valid.
				await runValidations(section, shot, [
					required({ good: 'abc123', input: 'WebHook ID', slug: 'webhookid' }),
					required({ good: 'event1,event2', input: 'Events', slug: 'events' }),
				]);

				// Message Value / Title Value are selects preselected with a hard-default.
				const messageValueField = fieldOf(
					section.getByRole('combobox', { name: 'Message Value' }),
				);
				const titleValueField = fieldOf(
					section.getByRole('combobox', { name: 'Title Value' }),
				);
				await expect(messageValueField).toContainText(' (default)');
				await expect(titleValueField).toContainText(' (default)');
				await shot('05-message-and-title-value-defaults');
			});
		});

		test.describe('Join', () => {
			test('API key and devices are required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/join',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Join" notifier is added.
				await addNotify(section, dialog, 'Join');

				// WHEN/THEN: each field is blurred empty then valid.
				await runValidations(section, shot, [
					required({ good: 'abc123apikey', input: 'API Key', slug: 'apikey' }),
					required({
						good: 'device1,device2',
						input: 'Devices',
						slug: 'devices',
					}),
				]);
			});
		});

		test.describe('MatterMost', () => {
			test('host and token are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/mattermost',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "MatterMost" notifier is added.
				await addNotify(section, dialog, 'MatterMost');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({
						good: 'mattermost.example.com',
						input: 'Host',
						slug: 'host',
					}),
					numeric({ good: '8065', input: 'Port', slug: 'port' }),
					required({ good: 'abc123token', input: 'Token', slug: 'token' }),
				]);
			});
		});

		test.describe('Matrix', () => {
			test('host and password are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/matrix',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Matrix" notifier is added.
				await addNotify(section, dialog, 'Matrix');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({ good: 'matrix.example.com', input: 'Host', slug: 'host' }),
					required({ good: 's3cret', input: 'Password', slug: 'password' }),
					numeric({ good: '8448', input: 'Port', slug: 'port' }),
				]);
			});
		});

		test.describe('Notifiarr', () => {
			test('API key is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/notifiarr',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Notifiarr" notifier is added.
				await addNotify(section, dialog, 'Notifiarr');

				// WHEN/THEN: the API key is blurred empty then valid.
				await runValidations(section, shot, [
					required({ good: 'abc123apikey', input: 'API Key', slug: 'apikey' }),
				]);
			});
		});

		test.describe('Ntfy', () => {
			test('host is optional; topic is required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/ntfy',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Ntfy" notifier is added.
				await addNotify(section, dialog, 'Ntfy');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					optional({ good: 'ntfy.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '443', input: 'Port', slug: 'port' }),
					required({ good: 'my-topic', input: 'Topic', slug: 'topic' }),
				]);
			});

			test('a "Broadcast" action requires a label, and a "View" action requires a URL', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/ntfy',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Ntfy" notifier is added.
				await addNotify(section, dialog, 'Ntfy');

				// WHEN: an action is added (defaults to type "Broadcast", which
				// requires `label`; "View"/"HTTP").
				await section.getByRole('button', { name: /add new actions/i }).click();

				const labelInput = section.getByRole('textbox', {
					name: 'Value field for Label',
				});

				// WHEN: "Label" is left empty and blurred.
				// THEN: an error is shown.
				await expectError(
					labelInput,
					undefined,
					REQUIRED,
					shot,
					'01-action-label-empty',
					section,
				);

				// WHEN: a value is entered.
				// THEN: the error clears.
				await expectValid(
					labelInput,
					'Open page',
					shot,
					'02-action-label-valid',
				);

				// WHEN: its type is switched to "View".
				await section
					.locator('#notify\\.0\\.params\\.actions\\.0\\.action')
					.click();
				await dialog
					.getByRole('option', { exact: true, name: /^View$/i })
					.click();

				const urlInput = section.getByRole('textbox', {
					name: 'Value field for URL',
				});

				// WHEN: "URL" is left empty and blurred.
				// THEN: an error is shown.
				await expectError(
					urlInput,
					undefined,
					REQUIRED,
					shot,
					'03-action-url-empty',
					section,
				);

				// WHEN: a value is entered.
				// THEN: the error clears.
				await expectValid(
					urlInput,
					'https://example.com',
					shot,
					'04-action-url-valid',
				);
			});
		});

		test.describe('OpsGenie', () => {
			test('API key is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/opsgenie',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "OpsGenie" notifier is added.
				await addNotify(section, dialog, 'OpsGenie');

				// WHEN/THEN: apikey is blurred empty then valid.
				await runValidations(section, shot, [
					required({
						good: 'abc123apikey',
						input: 'API Key',
						slug: 'apikey',
					}),
				]);
			});

			test('an action requires a value', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/opsgenie',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "OpsGenie" notifier is added.
				await addNotify(section, dialog, 'OpsGenie');

				// WHEN: an action is added and left empty.
				await section.getByRole('button', { name: /add new actions/i }).click();
				const actionInput = section.locator(
					'input[name="notify.0.params.actions.0.arg"]',
				);
				const actionField = fieldOf(actionInput);
				await actionInput.fill('');
				await actionInput.blur();

				// THEN: an error is shown.
				await expect(actionInput).toHaveAttribute('aria-invalid', 'true');
				await expect(actionField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-action-empty', actionField);

				// WHEN: a value is entered.
				await actionInput.fill('Acknowledge');
				await actionInput.blur();

				// THEN: the error clears.
				await expect(actionInput).toHaveAttribute('aria-invalid', 'false');
				await expect(actionField.getByRole('alert')).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('02-action-valid', actionField);
			});

			test('a details entry requires a key and value, and keys must be unique', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/opsgenie',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "OpsGenie" notifier is added.
				await addNotify(section, dialog, 'OpsGenie');

				// WHEN: a details row is added with an empty key and value.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new details/i }),
				);
				const detailKeyInput = section.locator(
					'input[name="notify.0.params.details.0.key"]',
				);
				const detailValueInput = section.locator(
					'input[name="notify.0.params.details.0.value"]',
				);
				await detailKeyInput.focus();
				await detailKeyInput.blur();
				await detailValueInput.focus();
				await detailValueInput.blur();

				// THEN: an error is shown for both.
				await expect(detailKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(detailValueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailValueInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-details-empty');

				// WHEN: a second details row is added with the same key as the first.
				await detailKeyInput.fill('priority');
				await detailValueInput.fill('1');
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new details/i }),
				);
				const detailKeyInput2 = section.locator(
					'input[name="notify.0.params.details.1.key"]',
				);
				const detailValueInput2 = section.locator(
					'input[name="notify.0.params.details.1.value"]',
				);
				await detailKeyInput2.fill('priority');
				await detailValueInput2.fill('2');
				await detailValueInput2.blur();
				// Re-touch the first key so its duplicate error re-evaluates.
				await detailKeyInput.focus();
				await detailKeyInput.blur();

				// THEN:an error is shown on both duplicate keys.
				await expect(detailKeyInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(detailKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('02-details-duplicate-key');

				// WHEN: the second key is changed to be unique.
				await detailKeyInput2.fill('severity');
				await detailKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await detailKeyInput.focus();
				await detailKeyInput.blur();

				// THEN: both errors clear.
				await expect(detailKeyInput).toHaveAttribute('aria-invalid', 'false');
				await expect(
					fieldOf(detailKeyInput).getByRole('alert'),
				).not.toBeVisible();
				await expect(detailKeyInput2).toHaveAttribute('aria-invalid', 'false');
				await expect(
					fieldOf(detailKeyInput2).getByRole('alert'),
				).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('03-details-valid');
			});

			test('a responder/visible-to target requires a value', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/opsgenie',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "OpsGenie" notifier is added.
				await addNotify(section, dialog, 'OpsGenie');

				// WHEN: a responder is added and left empty (defaults to type "team").
				await section
					.getByRole('button', { name: /add new responders/i })
					.click();
				const responderValueInput = section.locator(
					'input[name="notify.0.params.responders.0.value"]',
				);
				const responderValueField = fieldOf(responderValueInput);
				await responderValueInput.focus();
				await responderValueInput.blur();

				// THEN: an error is shown.
				await expect(responderValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(responderValueField.getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-responders-empty');

				// WHEN: a value is entered.
				await responderValueInput.fill('my-team');
				await responderValueInput.blur();

				// THEN: the error clears.
				await expect(responderValueInput).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(responderValueField.getByRole('alert')).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('02-responders-valid');

				// WHEN: a "Visible To" target is added and left empty.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new visible to/i }),
				);
				const visibleToValueInput = section.locator(
					'input[name="notify.0.params.visibleto.0.value"]',
				);
				await visibleToValueInput.focus();
				await visibleToValueInput.blur();

				// THEN: an error is shown.
				await expect(visibleToValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(
					fieldOf(visibleToValueInput).getByRole('alert'),
				).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('03-visibleto-empty');

				// WHEN: a value is entered.
				await visibleToValueInput.fill('my-user');
				await visibleToValueInput.blur();

				// THEN: the error clears.
				await expect(visibleToValueInput).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(
					fieldOf(visibleToValueInput).getByRole('alert'),
				).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('04-visibleto-valid');
			});
		});

		test.describe('PushBullet', () => {
			test('access token and targets are required', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/pushbullet',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "PushBullet" notifier is added.
				await addNotify(section, dialog, 'PushBullet');

				// WHEN/THEN: each field is blurred empty then valid.
				await runValidations(section, shot, [
					required({
						good: 'abc123token',
						input: 'Access Token',
						slug: 'accesstoken',
					}),
					required({
						good: 'DEVICE1,DEVICE2',
						input: 'Targets',
						slug: 'targets',
					}),
				]);
			});
		});

		test.describe('PushOver', () => {
			test('API token/key and user key are required; priority must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/pushover',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "PushOver" notifier is added.
				await addNotify(section, dialog, 'PushOver');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({
						good: 'abc123token',
						input: 'API Token/Key',
						slug: 'token',
					}),
					required({
						good: 'abc123userkey',
						input: 'User Key',
						slug: 'userkey',
					}),
					numeric({ good: '1', input: 'Priority', slug: 'priority' }),
				]);
			});
		});

		test.describe('Rocket.Chat', () => {
			test('channel, host, token A, and token B are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/rocketchat',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Rocket.Chat" notifier is added.
				await addNotify(section, dialog, 'Rocket.Chat');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({ good: 'general', input: 'Channel', slug: 'channel' }),
					required({
						good: 'rocketchat.example.io',
						input: 'Host',
						slug: 'host',
					}),
					numeric({ good: '443', input: 'Port', slug: 'port' }),
					required({ good: 'tokenA123', input: 'Token A', slug: 'tokena' }),
					required({ good: 'tokenB123', input: 'Token B', slug: 'tokenb' }),
				]);
			});
		});

		test.describe('Shoutrrr', () => {
			test('raw is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/shoutrrr',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Shoutrrr" notifier is added.
				await addNotify(section, dialog, 'Shoutrrr');

				// WHEN/THEN: Raw is blurred empty then valid.
				await runValidations(section, shot, [
					required({
						good: 'slack://xoxb:token@channel',
						input: 'Raw',
						slug: 'raw',
					}),
				]);
			});
		});

		test.describe('Slack', () => {
			test('token and channel are required; color must be a valid hex string', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/slack',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Slack" notifier is added.
				await addNotify(section, dialog, 'Slack');

				// WHEN/THEN: token/channel are blurred empty then valid.
				// color rejects a non-hex value, then accepts a valid one.
				await runValidations(section, shot, [
					required({ good: 'xoxb:abc123', input: 'Token', slug: 'token' }),
					required({ good: 'general', input: 'Channel', slug: 'channel' }),
					{
						bad: 'zzzzzz',
						badSlug: 'color-invalid-hex',
						error: 'Invalid hexadecimal. Must be 6 characters, 0-9 and A-F.',
						good: 'ff0000',
						goodSlug: 'color-valid',
						input: 'Color',
						inputType: 'textbox',
					},
				]);
			});
		});

		test.describe('Email (SMTP)', () => {
			test('host, from address, and to address(es) are required; port must be a valid number and timeout a valid duration', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/smtp',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Email (SMTP)" notifier is added.
				await addNotify(section, dialog, 'Email (SMTP)');

				// WHEN/THEN: each field is blurred bad then valid;
				// timeout rejects a non-duration value then accepts a valid one.
				await runValidations(section, shot, [
					required({ good: 'smtp.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '587', input: 'Port', slug: 'port' }),
					required({
						good: 'from@example.com',
						input: 'From Address',
						slug: 'fromaddress',
					}),
					required({
						good: 'to@example.com',
						input: 'To Address(es)',
						slug: 'toaddresses',
					}),
					{
						bad: 'abc',
						badSlug: 'timeout-invalid-duration',
						error: "Invalid duration. Use 'AhBmCs' duration format.",
						good: '10s',
						goodSlug: 'timeout-valid',
						input: 'Timeout',
					},
				]);
			});
		});

		test.describe('Teams', () => {
			test('host is required; title and color have no field-level validation', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/teams',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Teams" notifier is added.
				await addNotify(section, dialog, 'Teams');

				// --- Host: empty=Required -> valid ---
				const hostInput = section.getByRole('textbox', {
					name: 'Value field for Host',
				});
				const hostField = fieldOf(hostInput);

				// WHEN: "Host" is cleared and blurred.
				await hostInput.fill('');
				await hostInput.blur();

				// THEN: an error is shown.
				await expect(hostInput).toHaveAttribute('aria-invalid', 'true');
				await expect(hostField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-host-empty', hostField);

				// WHEN: a value is entered.
				await hostInput.fill(
					'prod-00.westus.logic.azure.com:443/workflows/abc123',
				);
				await hostInput.blur();

				// THEN: the error clears.
				await expect(hostInput).toHaveAttribute('aria-invalid', 'false');
				await expect(hostField.getByRole('alert')).not.toBeVisible();
				await shot('02-host-valid', hostField);

				// --- Title and Color: arbitrary values -> no validation errors ---
				const titleInput = section.getByRole('textbox', {
					name: 'Value field for Title',
				});
				const colorInput = section.getByRole('textbox', { name: 'Color' });

				// WHEN: arbitrary values ("zzzzzz" isn't valid hex) are entered into
				// "Title" and "Color" - neither has a validator.
				await titleInput.fill('My Notification');
				await titleInput.blur();
				await colorInput.fill('zzzzzz');
				await colorInput.blur();
				await shot('03-title-and-color-arbitrary');

				// THEN: no validation errors are rendered for either field (a field
				// that was never invalid omits `aria-invalid`, so assert not "true").
				await expect(titleInput).not.toHaveAttribute('aria-invalid', 'true');
				await expect(colorInput).not.toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(titleInput).getByRole('alert')).toHaveCount(0);
				await expect(fieldOf(colorInput).getByRole('alert')).toHaveCount(0);
			});
		});

		test.describe('Telegram', () => {
			test('token is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/telegram',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Telegram" notifier is added.
				await addNotify(section, dialog, 'Telegram');

				// WHEN/THEN: the token is blurred empty then valid.
				await runValidations(section, shot, [
					required({ good: '123456:ABC-DEF', input: 'Token', slug: 'token' }),
				]);
			});
		});

		test.describe('Zulip Chat', () => {
			test('bot mail, bot key, and host are required; port must be a valid number; message type and "to" have no field-level validation', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/zulip',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				// GIVEN: a new "Zulip Chat" notifier is added.
				await addNotify(section, dialog, 'Zulip Chat');

				// WHEN/THEN: each field is blurred bad then valid.
				await runValidations(section, shot, [
					required({
						good: 'bot@example.com',
						input: 'Bot Mail',
						slug: 'botmail',
					}),
					required({ good: 'abc123key', input: 'Bot Key', slug: 'botkey' }),
					required({ good: 'zulip.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '443', input: 'Port', slug: 'port' }),
				]);

				// --- Message Type and To: arbitrary values -> no validation errors ---
				const messageTypeSelect = section.getByRole('combobox', {
					name: 'Message Type',
				});
				const toInput = section.getByRole('textbox', { name: 'To (DM)' });

				// WHEN: "Message Type" is changed and "To" is filled and blurred.
				await messageTypeSelect.click();
				await dialog.getByRole('option', { name: 'Direct' }).click();
				await toInput.fill('user@example.com');
				await toInput.blur();

				// THEN: no validation errors are rendered for either field.
				await expect(toInput).not.toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(toInput).getByRole('alert')).toHaveCount(0);
				await shot('09-messagetype-and-to-arbitrary');
			});
		});
	});
});
