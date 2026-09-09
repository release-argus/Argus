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

		// GIVEN: a service with an update available (latest_version !==
		// deployed_version) and a Webhook configured with valid credentials.
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

		// WHEN: the user opens the "Resend actions" (visible text "Approve")
		// modal for the service.
		await card.getByRole('button', { name: /approve|resend/i }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/02-action-modal`,
			testInfo.project.name,
		);

		// AND: clicks "Send" for the Webhook.
		const sendButton = dialog.getByRole('button', {
			exact: true,
			name: 'Send',
		});
		await expect(sendButton).toBeVisible();
		await sendButton.click();

		// THEN: the Webhook send succeeds (real network call to
		// valid.release-argus.io - allow extra time).
		await expect(dialog.locator('[aria-label="Successful"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/03-send-successful`,
			testInfo.project.name,
		);

		// AND: the send/retry button becomes disabled - sends are blocked until
		// the Webhook's `next_runnable` time.
		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		await expect(retryButton).toBeDisabled();

		// AND: an hourglass icon with a "Can resend ..." tooltip indicates when
		// the Webhook can next be sent.
		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/04-send-blocked`,
			testInfo.project.name,
		);

		// WHEN: the user closes the modal.
		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		// THEN: the modal closes.
		await expect(dialog).not.toBeVisible();
	});

	test('Webhook send fails', async ({ page }, testInfo) => {
		const baseID = 'WEBHOOK=FAIL';
		const id = withProject(baseID, testInfo.project.name);
		createdIDs.push(id);

		await openDashboardInEditMode(page);

		// GIVEN: a service with an update available and a Webhook configured
		// with an invalid secret and a single try (fail fast).
		await createService(page, id, {
			deployedVersion: { type: 'manual', version: '0.0.1' },
			latestVersion: LOOKUP_LATEST_VERSION_JSON,
			webhooks: [
				{
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

		// WHEN: the user opens the "Resend actions" (visible text "Approve")
		// modal for the service.
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

		// AND: the server refuses sends, as it does for a read-only demo user.
		await page.route(ACTIONS_ROUTE, refuseSends);

		// WHEN: the user confirms the send for every Webhook.
		await dialog.locator('#modal-action').click();

		// THEN: the modal closes and the refusal is surfaced as a toast.
		await expect(dialog).not.toBeVisible();
		const sendAllToast = errorToast(page, 'Failed to send');
		await expect(sendAllToast.first()).toBeVisible();
		await expect(sendAllToast.first()).toContainText(
			'Read-only demo instance.',
		);

		// AND: reopening the modal shows nothing stuck sending.
		await card.getByRole('button', { name: /approve|resend/i }).click();
		await expect(dialog).toBeVisible();
		await expect(sendButton).toBeEnabled();
		await expect(dialog.locator('.animate-spin')).toHaveCount(0);

		// WHEN: the user sends the single Webhook, still refused.
		await sendButton.click();

		// THEN: the refusal names the Webhook, and its spinner stops.
		const sendOneToast = errorToast(page, 'Failed to send Webhook');
		await expect(sendOneToast).toBeVisible();
		await expect(sendOneToast).toContainText('Read-only demo instance.');
		await expect(dialog.locator('.animate-spin')).toHaveCount(0);
		await expect(sendButton).toBeEnabled();
		await screenshot(
			page,
			`webhook/${baseID}/03-send-refused`,
			testInfo.project.name,
		);

		// WHEN: clicks "Send" for the Webhook, with the server accepting sends again.
		await page.unroute(ACTIONS_ROUTE, refuseSends);
		await sendButton.click();

		// THEN: the Webhook send fails (real network call - the receiver
		// responds non-2XX for an incorrect secret).
		await expect(dialog.locator('[aria-label="Failed"]')).toBeVisible({
			timeout: 30_000,
		});
		await screenshot(
			page,
			`webhook/${baseID}/04-send-failed`,
			testInfo.project.name,
		);

		// AND: the send/retry button becomes disabled - sends are blocked until
		// the Webhook's `next_runnable` time.
		const retryButton = dialog.getByRole('button', {
			exact: true,
			name: 'Retry',
		});
		await expect(retryButton).toBeVisible();
		await expect(retryButton).toBeDisabled();

		// AND: an hourglass icon with a "Can resend ..." tooltip indicates when
		// the Webhook can next be sent.
		await dialog.locator('[aria-label="Resend timer"]').hover();
		await expect(page.getByText(/can resend/i).first()).toBeVisible();
		await screenshot(
			page,
			`webhook/${baseID}/05-send-blocked`,
			testInfo.project.name,
		);

		// WHEN: the user closes the modal.
		const doneButton = dialog.locator('#modal-action');
		await expect(doneButton).toHaveText(/^Done$/i);
		await doneButton.click();

		// THEN: the modal closes.
		await expect(dialog).not.toBeVisible();
	});
});
