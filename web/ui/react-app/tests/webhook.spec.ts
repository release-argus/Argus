import { expect, type Page, type Route, test } from '@playwright/test';
import { openDashboardInEditMode, serviceCard } from './fixtures/dashboard';
import {
	createService,
	LOOKUP_LATEST_VERSION_JSON,
	screenshot,
	trackCreatedServices,
	withProject,
} from './fixtures/service';
import { WEBHOOK_GITHUB } from './fixtures/test-endpoints';

/* The send endpoint, which a read-only instance refuses with a 403. */
const ACTIONS_ROUTE = '**/api/v1/service/actions*';

/**
 * Refuses sends the way the demo guard does, leaving reads alone.
 *
 * @param route - The intercepted route.
 */
const refuseSends = async (route: Route) => {
	if (route.request().method() !== 'POST') return route.fallback();

	await route.fulfill({
		body: JSON.stringify({ message: 'Read-only demo instance.' }),
		contentType: 'application/json; charset=utf-8',
		status: 403,
	});
};

/**
 * @param page - The page to find the toast on.
 * @param text - Text the toast must contain.
 * @returns The error toast containing `text`.
 */
const errorToast = (page: Page, text: string) =>
	page.locator('[data-sonner-toast][data-type="error"]').filter({
		hasText: text,
	});

test.describe('Webhook actions', () => {
	// Scoped to this test's ID so its afterEach can't race the other test's
	// service. IDs are project-suffixed to avoid cross-browser collisions.
	const createdIDs = trackCreatedServices();

	test('Webhook send succeeds', async ({ page }, testInfo) => {
		const baseID = 'WEBHOOK=PASS';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// The deployed version differs from the latest, so an update is available.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
			webhooks: [
				{
					name: id,
					secret: WEBHOOK_GITHUB.secretPass,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await screenshot(
			page,
			`webhook/${baseID}/01-after-create`,
			testInfo.project.name,
		);

		const card = serviceCard(page, id);

		// Labelled "Resend actions", though its visible text is "Approve".
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/02-action-modal`,
			testInfo.project.name,
		);

		const sendButton = dialog.getByRole('button', {
			exact: true,
			name: 'Send',
		});
		await expect(sendButton).toBeVisible();
		await sendButton.click();

		// The call to valid.release-argus.io is real, so allow extra time.
		await expect(dialog.locator('[aria-label="Successful"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/03-send-successful`,
			testInfo.project.name,
		);

		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		// Retry stays disabled until the Webhook's next_runnable time.
		await expect(retryButton).toBeDisabled();

		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/04-send-blocked`,
			testInfo.project.name,
		);

		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		await expect(dialog).not.toBeVisible();
	});

	test('Webhook send fails', async ({ page }, testInfo) => {
		const baseID = 'WEBHOOK=FAIL';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// The deployed version differs from the latest, so an update is available.
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
			webhooks: [
				{
					// One try, so the send fails fast.
					maxTries: '1',
					name: id,
					secret: WEBHOOK_GITHUB.secretFail,
					url: WEBHOOK_GITHUB.urlValid,
				},
			],
		});
		await screenshot(
			page,
			`webhook/${baseID}/01-after-create`,
			testInfo.project.name,
		);

		const card = serviceCard(page, id);

		// Labelled "Resend actions", though its visible text is "Approve".
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog, 'after open').toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/02-action-modal`,
			testInfo.project.name,
		);

		const sendButton = dialog.getByRole('button', {
			exact: true,
			name: 'Send',
		});
		await expect(sendButton).toBeVisible();

		await page.route(ACTIONS_ROUTE, refuseSends);
		await dialog.locator('#modal-action').click();

		await expect(dialog, 'after send all').not.toBeVisible();
		const sendAllToast = errorToast(page, 'Failed to send');
		await expect(sendAllToast.first()).toBeVisible();
		await expect(sendAllToast.first()).toContainText(
			'Read-only demo instance.',
		);

		await card.getByRole('button', { name: /approve|resend/i }).click();
		await expect(dialog, 'after reopen').toBeVisible();
		// The refusal leaves nothing stuck sending.
		await expect(sendButton, 'after reopen').toBeEnabled();
		const spinner = dialog.locator('.animate-spin');
		await expect(spinner, 'after reopen').toHaveCount(0);

		await sendButton.click();

		const sendOneToast = errorToast(page, 'Failed to send Webhook');
		await expect(sendOneToast).toBeVisible();
		await expect(sendOneToast).toContainText('Read-only demo instance.');
		await expect(spinner, 'after send one').toHaveCount(0);
		await expect(sendButton, 'after send one').toBeEnabled();
		await screenshot(
			page,
			`webhook/${baseID}/03-send-refused`,
			testInfo.project.name,
		);

		await page.unroute(ACTIONS_ROUTE, refuseSends);
		await sendButton.click();

		// The real receiver rejects the wrong secret.
		await expect(dialog.locator('[aria-label="Failed"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/04-send-failed`,
			testInfo.project.name,
		);

		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		// Retry stays disabled until the Webhook's next_runnable time.
		await expect(retryButton).toBeDisabled();

		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/05-send-blocked`,
			testInfo.project.name,
		);

		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		await expect(dialog, 'after done').not.toBeVisible();
	});
});
