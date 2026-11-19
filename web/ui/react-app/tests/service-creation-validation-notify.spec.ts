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

const DURATION_INVALID = "Invalid duration. Use 'AhBmCs' duration format.";
const XMPP_RECIPIENT_REQUIRED = 'To or Rooms required.';

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

				await addNotify(section, dialog, 'Bark');

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
			test('name, Webhook ID, and token are each required', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/discord',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Discord');

				await runValidations(section, shot, [
					required({ good: 'my-discord', input: 'Name', slug: 'name' }),
					required({ good: '123456', input: 'Webhook ID', slug: 'webhookid' }),
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

				await addNotify(section, dialog, 'Google Chat');

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

				await addNotify(section, dialog, 'Gotify');

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

				await addNotify(section, dialog, 'Gotify');

				await clickViaKeyboard(
					section.getByRole('button', { name: /add new extras/i }),
				);
				await section
					.locator('#notify\\.0\\.params\\.extras\\.0\\.namespace')
					.click();
				await dialog
					.getByRole('option', { exact: true, name: /^other$/i })
					.click();

				// Scoped by name attribute, as the extras "Name" shares its
				// "Value field for Name" aria-label with the notifier "Name".
				const nameInput = section.locator(
					'input[name="notify.0.params.extras.0._namespace"]',
				);
				const nameField = fieldOf(nameInput);
				const valueInput = section.locator(
					'input[name="notify.0.params.extras.0.value"]',
				);
				const valueField = fieldOf(valueInput);

				await nameInput.fill('');
				await nameInput.blur();
				await valueInput.fill('');
				await valueInput.blur();

				await expect(nameInput).toHaveAttribute('aria-invalid', 'true');
				await expect(nameField.getByRole('alert')).toHaveText(REQUIRED);
				await expect(valueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(valueField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-extras-empty', valueField);

				await nameInput.fill('my-namespace');
				await nameInput.blur();
				await valueInput.fill('{"key": "value"}');
				await valueInput.blur();

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

				await addNotify(section, dialog, 'Generic Webhook');

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

				await addNotify(section, dialog, 'Generic Webhook');

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

				await expect(headerKeyInput, 'left empty').toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(headerKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(headerValueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerValueInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-headers-empty');

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

				await expect(headerKeyInput, 'duplicate key').toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(headerKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(headerKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(headerKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('02-headers-duplicate-key');

				await headerKeyInput2.fill('X-Header-2');
				await headerKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await headerKeyInput.focus();
				await headerKeyInput.blur();

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

				await expect(payloadVarKeyInput, 'left empty').toHaveAttribute(
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

				await expect(payloadVarKeyInput, 'duplicate key').toHaveAttribute(
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

				await payloadVarKeyInput2.fill('messageKey');
				await payloadVarKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await payloadVarKeyInput.focus();
				await payloadVarKeyInput.blur();

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

				await expect(queryVarKeyInput, 'left empty').toHaveAttribute(
					'aria-invalid',
					'true',
				);
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

				await expect(queryVarKeyInput, 'duplicate key').toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(queryVarKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(queryVarKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(queryVarKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('08-query-vars-duplicate-key');

				await queryVarKeyInput2.fill('bar');
				await queryVarKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await queryVarKeyInput.focus();
				await queryVarKeyInput.blur();

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

		test.describe('Home Assistant', () => {
			test('host and token are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/homeassistant',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Home Assistant');

				await runValidations(section, shot, [
					required({
						good: 'homeassistant.local',
						input: 'Host',
						slug: 'host',
					}),
					numeric({ good: '8123', input: 'Port', slug: 'port' }),
					required({ good: 'abc123token', input: 'Token', slug: 'token' }),
				]);
			});
		});

		test.describe('IFTTT', () => {
			test('Webhook ID and events are required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/ifttt',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'IFTTT');

				await runValidations(section, shot, [
					required({ good: 'abc123', input: 'Webhook ID', slug: 'webhookid' }),
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

				await addNotify(section, dialog, 'Join');

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

		test.describe('Lark', () => {
			test('token is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/lark',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Lark');

				await runValidations(section, shot, [
					required({ good: 'abc123token', input: 'Token', slug: 'token' }),
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

				await addNotify(section, dialog, 'MatterMost');

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

				await addNotify(section, dialog, 'Matrix');

				await runValidations(section, shot, [
					required({ good: 'matrix.example.com', input: 'Host', slug: 'host' }),
					required({ good: 's3cret', input: 'Password', slug: 'password' }),
					numeric({ good: '8448', input: 'Port', slug: 'port' }),
				]);
			});
		});

		test.describe('MQTT', () => {
			test('host and topic are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/mqtt',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'MQTT');

				await runValidations(section, shot, [
					required({ good: 'mqtt.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '8883', input: 'Port', slug: 'port' }),
					required({ good: 'argus/releases', input: 'Topic', slug: 'topic' }),
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

				await addNotify(section, dialog, 'Notifiarr');

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

				await addNotify(section, dialog, 'Ntfy');

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

				await addNotify(section, dialog, 'Ntfy');

				// A new action defaults to the "Broadcast" type, which needs a label.
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new actions/i }),
				);

				const labelInput = section.getByRole('textbox', {
					name: 'Value field for Label',
				});

				await expectError(
					labelInput,
					undefined,
					REQUIRED,
					shot,
					'01-action-label-empty',
					section,
				);

				await expectValid(
					labelInput,
					'Open page',
					shot,
					'02-action-label-valid',
				);

				await section
					.locator('#notify\\.0\\.params\\.actions\\.0\\.action')
					.click();
				await dialog
					.getByRole('option', { exact: true, name: /^View$/i })
					.click();

				const urlInput = section.getByRole('textbox', {
					name: 'Value field for URL',
				});

				await expectError(
					urlInput,
					undefined,
					REQUIRED,
					shot,
					'03-action-url-empty',
					section,
				);

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

				await addNotify(section, dialog, 'OpsGenie');

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

				await addNotify(section, dialog, 'OpsGenie');

				await clickViaKeyboard(
					section.getByRole('button', { name: /add new actions/i }),
				);
				const actionInput = section.locator(
					'input[name="notify.0.params.actions.0.arg"]',
				);
				const actionField = fieldOf(actionInput);
				await actionInput.fill('');
				await actionInput.blur();

				await expect(actionInput).toHaveAttribute('aria-invalid', 'true');
				await expect(actionField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-action-empty', actionField);

				await actionInput.fill('Acknowledge');
				await actionInput.blur();

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

				await addNotify(section, dialog, 'OpsGenie');

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

				await expect(detailKeyInput, 'left empty').toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(detailKeyInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await expect(detailValueInput).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailValueInput).getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-details-empty');

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

				await expect(detailKeyInput, 'duplicate key').toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(fieldOf(detailKeyInput).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await expect(detailKeyInput2).toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(detailKeyInput2).getByRole('alert')).toHaveText(
					MUST_BE_UNIQUE,
				);
				await waitForAccordionAnimations(section);
				await shot('02-details-duplicate-key');

				await detailKeyInput2.fill('severity');
				await detailKeyInput2.blur();
				// Re-touch the first key so its error clears.
				await detailKeyInput.focus();
				await detailKeyInput.blur();

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

				await addNotify(section, dialog, 'OpsGenie');

				// A new responder defaults to type "team".
				await clickViaKeyboard(
					section.getByRole('button', { name: /add new responders/i }),
				);
				const responderValueInput = section.locator(
					'input[name="notify.0.params.responders.0.value"]',
				);
				const responderValueField = fieldOf(responderValueInput);
				await responderValueInput.focus();
				await responderValueInput.blur();

				await expect(responderValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(responderValueField.getByRole('alert')).toHaveText(
					REQUIRED,
				);
				await waitForAccordionAnimations(section);
				await shot('01-responders-empty');

				await responderValueInput.fill('my-team');
				await responderValueInput.blur();

				await expect(responderValueInput).toHaveAttribute(
					'aria-invalid',
					'false',
				);
				await expect(responderValueField.getByRole('alert')).not.toBeVisible();
				await waitForAccordionAnimations(section);
				await shot('02-responders-valid');

				await clickViaKeyboard(
					section.getByRole('button', { name: /add new visible to/i }),
				);
				const visibleToValueInput = section.locator(
					'input[name="notify.0.params.visibleto.0.value"]',
				);
				await visibleToValueInput.focus();
				await visibleToValueInput.blur();

				await expect(visibleToValueInput).toHaveAttribute(
					'aria-invalid',
					'true',
				);
				await expect(
					fieldOf(visibleToValueInput).getByRole('alert'),
				).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('03-visibleto-empty');

				await visibleToValueInput.fill('my-user');
				await visibleToValueInput.blur();

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

		test.describe('PagerDuty', () => {
			test('integration key is required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/pagerduty',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'PagerDuty');

				await runValidations(section, shot, [
					numeric({ good: '443', input: 'Port', slug: 'port' }),
					required({
						good: '0123456789abcdef0123456789abcdef',
						input: 'Integration Key',
						slug: 'integrationkey',
					}),
				]);
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

				await addNotify(section, dialog, 'PushBullet');

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
			test('API token/key and user key are required; priority and timestamp must be valid numbers, retry and expire durations, and encryption key a 64-character hex string', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/pushover',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'PushOver');

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
					{
						bad: '30',
						badSlug: 'retry-invalid-duration',
						error: DURATION_INVALID,
						good: '30s',
						goodSlug: 'retry-valid',
						input: 'Retry',
					},
					{
						bad: '1',
						badSlug: 'expire-invalid-duration',
						error: DURATION_INVALID,
						good: '1h',
						goodSlug: 'expire-valid',
						input: 'Expire',
					},
					numeric({
						good: '1700000000',
						input: 'Timestamp',
						slug: 'timestamp',
					}),
					{
						bad: 'abc123',
						badSlug: 'encryptionkey-invalid-hex',
						error: 'Invalid key. Must be 64 characters, 0-9 and A-F.',
						good: '0123456789abcdef'.repeat(4),
						goodSlug: 'encryptionkey-valid',
						input: 'Encryption Key',
					},
				]);
			});
		});

		test.describe('Rocket.Chat', () => {
			test('host, token A, and token B are required; channel is optional; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/rocketchat',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Rocket.Chat');

				await runValidations(section, shot, [
					optional({ good: 'general', input: 'Channel', slug: 'channel' }),
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

				await addNotify(section, dialog, 'Shoutrrr');

				await runValidations(section, shot, [
					required({
						good: 'slack://xoxb:token@channel',
						input: 'Raw',
						slug: 'raw',
					}),
				]);
			});
		});

		test.describe('Signal', () => {
			test('host, source, and recipients are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/signal',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Signal');

				await runValidations(section, shot, [
					required({
						good: 'signal.example.com',
						input: 'Host',
						slug: 'host',
					}),
					numeric({ good: '8080', input: 'Port', slug: 'port' }),
					required({ good: '+15551234567', input: 'Source', slug: 'source' }),
					required({
						good: '+15557654321',
						input: 'Recipients',
						slug: 'recipients',
					}),
				]);
			});
		});

		test.describe('Signalgrid', () => {
			test('client key and channel are required', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/signalgrid',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Signalgrid');

				await runValidations(section, shot, [
					required({
						good: 'clientkey',
						input: 'Client Key',
						slug: 'clientkey',
					}),
					required({
						good: 'channel',
						input: 'Channel',
						slug: 'channel',
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

				await addNotify(section, dialog, 'Slack');

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

				await addNotify(section, dialog, 'Email (SMTP)');

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

				await addNotify(section, dialog, 'Teams');

				const hostInput = section.getByRole('textbox', {
					name: 'Value field for Host',
				});
				const hostField = fieldOf(hostInput);

				await hostInput.fill('');
				await hostInput.blur();

				await expect(hostInput).toHaveAttribute('aria-invalid', 'true');
				await expect(hostField.getByRole('alert')).toHaveText(REQUIRED);
				await waitForAccordionAnimations(section);
				await shot('01-host-empty', hostField);

				await hostInput.fill(
					'prod-00.westus.logic.azure.com:443/workflows/abc123',
				);
				await hostInput.blur();

				await expect(hostInput).toHaveAttribute('aria-invalid', 'false');
				await expect(hostField.getByRole('alert')).not.toBeVisible();
				await shot('02-host-valid', hostField);

				const titleInput = section.getByRole('textbox', {
					name: 'Value field for Title',
				});
				const colorInput = section.getByRole('textbox', { name: 'Color' });

				await titleInput.fill('My Notification');
				await titleInput.blur();
				// "zzzzzz" is not valid hex, yet Color has no validator.
				await colorInput.fill('zzzzzz');
				await colorInput.blur();
				await shot('03-title-and-color-arbitrary');

				// A field that was never invalid omits aria-invalid, so assert
				// not "true".
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

				await addNotify(section, dialog, 'Telegram');

				await runValidations(section, shot, [
					required({ good: '123456:ABC-DEF', input: 'Token', slug: 'token' }),
				]);
			});
		});

		test.describe('Twilio', () => {
			test('account SID, auth token, from, and to are required', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/twilio',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'Twilio (SMS)');

				await runValidations(section, shot, [
					required({
						// Split so secret scanners do not take it for a real account SID.
						good: `AC${'0123456789abcdef'.repeat(2)}`,
						input: 'Account SID',
						slug: 'accountsid',
					}),
					required({
						good: 'authtoken',
						input: 'Auth Token',
						slug: 'authtoken',
					}),
					required({ good: '+15551234567', input: 'From', slug: 'fromnumber' }),
					required({ good: '+15557654321', input: 'To', slug: 'tonumbers' }),
				]);
			});
		});

		test.describe('WeCom', () => {
			test('webhook key is required', async ({ page }, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/wecom',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'WeCom');

				await runValidations(section, shot, [
					required({ good: 'abc-123', input: 'Webhook Key', slug: 'key' }),
				]);
			});
		});

		test.describe('XMPP', () => {
			test('host, username, password, and one of to/rooms are required; port must be a valid number', async ({
				page,
			}, testInfo) => {
				const shot = screenshotsUnder(
					page,
					testInfo.project.name,
					'service-creation-validation/notify/xmpp',
				);
				const dialog = await openCreateServiceModal(page);
				const section = await openSection(dialog, 'Notify:');

				await addNotify(section, dialog, 'XMPP');

				await runValidations(section, shot, [
					required({ good: 'xmpp.example.com', input: 'Host', slug: 'host' }),
					numeric({ good: '5222', input: 'Port', slug: 'port' }),
					required({ good: 'argus', input: 'User', slug: 'user' }),
					required({ good: 'pass', input: 'Password', slug: 'password' }),
					{
						bad: undefined,
						badSlug: 'rooms-empty',
						error: XMPP_RECIPIENT_REQUIRED,
						input: 'Rooms',
					},
					{
						bad: undefined,
						badSlug: 'to-empty',
						error: XMPP_RECIPIENT_REQUIRED,
						good: 'user@example.com',
						goodSlug: 'to-valid',
						input: 'To',
					},
					{
						bad: undefined,
						badSlug: 'rooms-empty-to-set',
						error: null,
						input: 'Rooms',
					},
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

				await addNotify(section, dialog, 'Zulip Chat');

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

				const messageTypeSelect = section.getByRole('combobox', {
					name: 'Message Type',
				});
				const toInput = section.getByRole('textbox', { name: 'To (DM)' });

				await messageTypeSelect.click();
				await dialog.getByRole('option', { name: 'Direct' }).click();
				await toInput.fill('user@example.com');
				await toInput.blur();

				await expect(toInput).not.toHaveAttribute('aria-invalid', 'true');
				await expect(fieldOf(toInput).getByRole('alert')).toHaveCount(0);
				await shot('09-messagetype-and-to-arbitrary');
			});
		});
	});
});
