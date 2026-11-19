import { expect, type Locator, type Page, test } from '@playwright/test';
import { openDashboardInEditMode } from './fixtures/dashboard';
import {
	createService,
	type NotifyFieldsOptions,
	notifyItem,
	openEditModalFromBackend,
	openNotifyItem,
	SECRET_VALUE,
	saveEdit,
	screenshot,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import { openSection } from './fixtures/validation';

/**
 * Asserts a saved notifier shows what `notify` filled, with each of `secrets`
 * (text labels) masked.
 *
 * @param item - The notifier's accordion item.
 * @param index - The notifier's index.
 * @param notify - The options the notifier was created with.
 * @param secrets - Text labels the API masks.
 * @param step - Labels the assertions' failure messages.
 */
const expectNotifyFields = async (
	item: Locator,
	index: number,
	notify: NotifyFieldsOptions,
	secrets: string[] = [],
	step?: string,
) => {
	for (const [label, value] of Object.entries(notify.text ?? {})) {
		await expect(
			item.getByRole('textbox', {
				exact: true,
				name: `Value field for ${label}`,
			}),
			step,
		).toHaveValue(secrets.includes(label) ? SECRET_VALUE : value);
	}
	for (const [path, option] of Object.entries(notify.selects ?? {})) {
		await expect(
			item.locator(`[id="notify.${index}.${path}"]`),
			step,
		).toContainText(option);
	}
	for (const [path, value] of Object.entries(notify.booleans ?? {})) {
		await expect(
			item
				.locator(`[aria-labelledby="notify.${index}.${path}-label"]`)
				.getByRole('radio', { name: value ? /^Yes$/i : /^No$/i }),
			step,
		).toBeChecked();
	}
};

/**
 * Creates a service with `notifiers`, then reopens it from the backend and
 * asserts each notifier round-tripped.
 *
 * @param page - The dashboard page.
 * @param id - The service ID.
 * @param shotDir - The screenshot directory.
 * @param projectName - The browser project name.
 * @param notifiers - The notifiers, each with the text labels the API masks.
 * @param step - Labels the assertions' failure messages.
 * @returns The open edit dialog and its Notify section.
 */
const createAndReopen = async (
	page: Page,
	id: string,
	shotDir: string,
	projectName: string,
	notifiers: { notify: NotifyFieldsOptions; secrets?: string[] }[],
	step?: string,
) => {
	await openDashboardInEditMode(page);
	await createService(page, id, {
		notifiers: notifiers.map(({ notify }) => notify),
	});

	const dialog = await openEditModalFromBackend(page, id);
	const section = await openSection(dialog, 'Notify');
	for (let index = 0; index < notifiers.length; index++) {
		const { notify, secrets } = notifiers[index];
		const item = await openNotifyItem(section, index);
		await expectNotifyFields(item, index, notify, secrets, step);
		await screenshot(
			page,
			`${shotDir}/0${index + 1}-${notify.name}-reloaded`,
			projectName,
			item,
		);
	}
	return { dialog, section };
};

test.describe('Notify save', () => {
	const createdIDs = trackCreatedServices();

	test('signalgrid: secrets are masked, its select and toggle round-trip, and an unrelated edit saves', async ({
		page,
	}, testInfo) => {
		const baseID = 'NOTIFY=SIGNALGRID';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);
		const shotDir = `notify/${baseID}`;

		const signalgrid: NotifyFieldsOptions = {
			booleans: { 'params.critical': true },
			name: 'signalgrid',
			selects: { 'params.type': 'Critical' },
			text: { Channel: 'CHANNEL', 'Client Key': 'CLIENTKEY' },
			typeLabel: 'Signalgrid',
		};
		const { dialog, section } = await createAndReopen(
			page,
			id,
			shotDir,
			testInfo.project.name,
			[{ notify: signalgrid, secrets: ['Channel', 'Client Key'] }],
			'before edit',
		);

		await notifyItem(section, 0)
			.getByRole('textbox', { exact: true, name: 'Value field for Title' })
			.fill('Changed Title');
		await saveEdit(dialog);

		const dialog2 = await openEditModalFromBackend(page, id);
		const section2 = await openSection(dialog2, 'Notify');
		const item2 = await openNotifyItem(section2, 0);
		await expectNotifyFields(
			item2,
			0,
			{ ...signalgrid, text: { ...signalgrid.text, Title: 'Changed Title' } },
			['Channel', 'Client Key'],
			'after edit',
		);
		await screenshot(
			page,
			`${shotDir}/02-after-edit`,
			testInfo.project.name,
			item2,
		);
	});

	test('mqtt: select and toggle params round-trip', async ({
		page,
	}, testInfo) => {
		const baseID = 'NOTIFY=MQTT';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					booleans: { 'params.disabletls': true },
					name: 'mqtt',
					selects: { 'params.qos': '2 - Exactly once' },
					text: { Host: 'mqtt.example.com', Topic: 'argus/releases' },
					typeLabel: 'MQTT',
				},
			},
		]);
	});

	test('pagerduty: select params round-trip', async ({ page }, testInfo) => {
		const baseID = 'NOTIFY=PAGERDUTY';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					name: 'pagerduty',
					selects: {
						'params.action': 'Resolve',
						'params.severity': 'Warning',
					},
					// Shoutrrr rejects a key that isn't 32 hex characters.
					text: { 'Integration Key': '0123456789abcdef0123456789abcdef' },
					typeLabel: 'PagerDuty',
				},
				secrets: ['Integration Key'],
			},
		]);
	});

	test('signal: select params round-trip', async ({ page }, testInfo) => {
		const baseID = 'NOTIFY=SIGNAL';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					name: 'signal',
					selects: { 'params.textmode': 'Styled' },
					text: {
						Host: 'signal.example.com',
						Recipients: '+15557654321',
						Source: '+15551234567',
					},
					typeLabel: 'Signal',
				},
			},
		]);
	});

	test('xmpp: toggle param round-trip', async ({ page }, testInfo) => {
		const baseID = 'NOTIFY=XMPP';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					booleans: { 'params.skiptlsverify': true },
					name: 'xmpp',
					text: {
						Host: 'xmpp.example.com',
						Password: 'pass',
						To: 'user@example.com',
						User: 'argus',
					},
					typeLabel: 'XMPP',
				},
				secrets: ['Password'],
			},
		]);
	});

	test('bark: disable TLS round-trips', async ({ page }, testInfo) => {
		const baseID = 'NOTIFY=BARK';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					booleans: { 'params.disabletls': true },
					name: 'bark',
					text: { 'Device Key': 'DEVICEKEY', Host: 'bark.example.com' },
					typeLabel: 'Bark',
				},
				secrets: ['Device Key'],
			},
		]);
	});

	test('ntfy: disable TLS round-trips', async ({ page }, testInfo) => {
		const baseID = 'NOTIFY=NTFY';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await createAndReopen(page, id, `notify/${baseID}`, testInfo.project.name, [
			{
				notify: {
					booleans: { 'params.disabletls_': true },
					name: 'ntfy',
					text: { Topic: 'argus' },
					typeLabel: 'Ntfy',
				},
			},
		]);
	});
});
