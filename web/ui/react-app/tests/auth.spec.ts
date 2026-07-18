import {
	type APIRequestContext,
	type BrowserContext,
	expect,
	type Page,
	test,
} from '@playwright/test';
import {
	ADMIN_PASSWORD,
	ADMIN_USER,
	cleanupByName,
	cleanupStack,
	newLoginContext,
	RUN_ID,
	SERVICE,
	signedInContext,
} from './fixtures/auth';

/**
 * Session and administration e2e specs - run only in the `chromium-auth`
 * project, against the auth-enabled Argus instance (AUTH_BASE_URL, bootstrap
 * admin created by `make playwright-tests-setup`).
 *
 * Scope: getting in and managing things - the session lifecycle (login, deep
 * link, expiry, logout) and the user/group/token administration flows through
 * the UI.
 *
 * Serial: the tests share one logged-in browser context, so a full run spends
 * only 2 of the login rate limit's budget (5 failed attempts per 5 minutes,
 * keyed on client IP. CI retries replay the whole serial group against the
 * same long-lived instance, so 3 attempts can exceed that budget;
 * the two tests that spend it therefore accept a 429 as a valid rejection.
 */

const E2E_USER = `e2e-user-${RUN_ID}`;
const E2E_VIEWER = `e2e-viewer-${RUN_ID}`;
const E2E_GROUP = `e2e-group-${RUN_ID}`;
const E2E_TOKEN = `e2e-token-${RUN_ID}`;

test.describe.configure({ mode: 'serial' });

test.describe('Authentication', () => {
	let context: BrowserContext;
	let page: Page;
	/** A second admin session, for teardown: 'logs out' ends the shared one. */
	let adminAPI: APIRequestContext;
	const cleanup = cleanupStack();

	/** Signs in through the login form. */
	const signIn = async (username: string, password: string) => {
		await page.getByLabel('Username').fill(username);
		await page.getByLabel('Password', { exact: true }).fill(password);
		await page.getByRole('button', { name: 'Sign in' }).click();
	};

	test.beforeAll(async ({ browser }) => {
		context = await browser.newContext();
		page = await context.newPage();

		// Registered first, so LIFO closes it last.
		const adminContext = await signedInContext(
			browser,
			ADMIN_USER,
			ADMIN_PASSWORD,
		);
		cleanup.add(() => adminContext.close());
		adminAPI = adminContext.request;
	});

	test.afterAll(async () => {
		await cleanup.run();
		await context.close();
	});

	test('redirects unauthenticated visitors to /login', async () => {
		await page.goto('/approvals');
		await expect(page).toHaveURL(/\/login/);
		await expect(page.getByRole('form', { name: 'Login' })).toBeVisible();
	});

	test('rejects invalid credentials', async () => {
		await page.goto('/login');
		await signIn(ADMIN_USER, 'not-the-password');
		// On a CI retry the login limiter may already be exhausted, so accept
		// either rejection.
		await expect(
			page.getByText(/invalid credentials|too many login attempts/),
		).toBeVisible();
		await expect(page).toHaveURL(/\/login/);
	});

	test('logs in and returns to the page the user came from', async () => {
		// Deep link while logged out -> login -> back to the deep link.
		await page.goto('/account/tokens');
		await expect(page).toHaveURL(/\/login/);
		await signIn(ADMIN_USER, ADMIN_PASSWORD);
		await expect(page).toHaveURL(/\/account\/tokens/);
		await expect(page.getByRole('button', { name: 'User menu' })).toBeVisible();
	});

	test('shows the dashboard once authenticated', async () => {
		await page.goto('/approvals');
		await expect(page.getByText(SERVICE)).toBeVisible();
	});

	test('the bare /admin and /account paths redirect onto a page', async () => {
		for (const [parent, landing] of [
			['/admin', '/admin/users'],
			['/admin/', '/admin/users'],
			['/account', '/account/tokens'],
			['/account/', '/account/tokens'],
		]) {
			await page.goto(parent);
			await expect(page, `${parent} lands on ${landing}`).toHaveURL(
				new RegExp(`${landing}$`),
			);
		}
	});

	test('creates, edits, and deletes a user', async () => {
		await page.getByRole('button', { name: 'User menu' }).click();
		await page.getByRole('menuitem', { name: 'Users' }).click();
		await expect(page).toHaveURL(/\/admin\/users/);

		// Create.
		cleanupByName(cleanup, adminAPI, 'users', E2E_USER);
		await page.getByRole('button', { name: 'Add user' }).click();
		await page.getByLabel('Username').fill(E2E_USER);
		await page.getByLabel('Password').fill('e2e-password');
		await page.getByLabel('Display name').fill('E2E User');
		await page.getByLabel('viewer').check();
		await page.getByRole('button', { name: 'Save' }).click();
		const row = page.getByRole('row', { name: new RegExp(E2E_USER) });
		await expect(row).toBeVisible();
		await expect(row).toContainText('viewer');

		// Edit.
		await page.getByRole('button', { name: `Edit user ${E2E_USER}` }).click();
		await page.getByLabel('Display name').fill('Renamed E2E User');
		await page.getByRole('button', { name: 'Save' }).click();
		await expect(row).toContainText('Renamed E2E User');

		// Delete.
		await page.getByRole('button', { name: `Delete user ${E2E_USER}` }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete' })
			.click();
		await expect(row).toHaveCount(0);
	});

	test('creates and deletes a group with grants', async () => {
		await page.goto('/admin/groups');

		// Seeded system groups are listed.
		await expect(page.getByRole('row', { name: /^admin/ })).toBeVisible();

		// Create, with one grant row.
		cleanupByName(cleanup, adminAPI, 'groups', E2E_GROUP);
		await page.getByRole('button', { name: 'Add group' }).click();
		await page.getByLabel('Name').fill(E2E_GROUP);
		await page.getByLabel('Description').fill('E2E test group');
		await page.getByRole('button', { name: 'Add permission' }).click();
		await expect(page.getByLabel('Grant 0 resource')).toBeVisible();
		await page.getByRole('button', { name: 'Save' }).click();
		const row = page.getByRole('row', { name: new RegExp(E2E_GROUP) });
		await expect(row).toBeVisible();
		// Grants are summarised as a count; the detail is in a tooltip.
		await expect(row).toContainText('1 permission');
		await row.getByRole('button', { name: '1 permission' }).hover();
		await expect(page.getByRole('tooltip')).toContainText('service:read');

		// System groups have no delete button.
		await expect(
			page.getByRole('button', { name: 'Delete group admin' }),
		).toHaveCount(0);

		// Delete.
		await page
			.getByRole('button', { name: `Delete group ${E2E_GROUP}` })
			.click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete' })
			.click();
		await expect(row).toHaveCount(0);
	});

	test('creates, uses, and revokes an API token', async ({ request }) => {
		await page.goto('/account/tokens');

		// Create.
		cleanupByName(cleanup, adminAPI, 'tokens', E2E_TOKEN);
		await page.getByRole('button', { name: 'Create token' }).click();
		await page.getByLabel('Name').fill(E2E_TOKEN);
		await page.getByRole('button', { exact: true, name: 'Create' }).click();

		// The plaintext is revealed once.
		const revealed = page.getByLabel('New token');
		await expect(revealed).toBeVisible();
		const plaintext = await revealed.inputValue();
		expect(plaintext).toMatch(/^argus_/);
		await page.getByRole('button', { name: 'Done' }).click();

		// The token authenticates Bearer requests.
		const bearer = await request.get('/api/v1/flags', {
			headers: { Authorization: `Bearer ${plaintext}` },
		});
		expect(bearer.status()).toBe(200);

		// Revoke.
		const row = page.getByRole('row', { name: new RegExp(E2E_TOKEN) });
		await expect(row).toBeVisible();
		await page
			.getByRole('button', { name: `Revoke token ${E2E_TOKEN}` })
			.click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Revoke' })
			.click();
		await expect(row).toHaveCount(0);

		// The revoked token no longer authenticates.
		const revoked = await request.get('/api/v1/flags', {
			headers: { Authorization: `Bearer ${plaintext}` },
		});
		expect(revoked.status()).toBe(401);
	});

	test('the seeded viewer group carries its grants', async ({ browser }) => {
		const created = await context.request.post('api/v1/users', {
			data: {
				groups: ['viewer'],
				password: 'e2e-viewer-password',
				username: E2E_VIEWER,
			},
		});
		expect(created.status()).toBe(201);
		const viewerID = ((await created.json()) as { id: string }).id;
		cleanup.add(() => adminAPI.delete(`api/v1/users/${viewerID}`));

		const viewerContext = await signedInContext(
			browser,
			E2E_VIEWER,
			'e2e-viewer-password',
		);
		cleanup.add(() => viewerContext.close());
		const viewerPage = await viewerContext.newPage();
		await viewerPage.goto('/approvals');

		// service:read - the fixture service is on their dashboard.
		await expect(viewerPage.getByText(SERVICE)).toBeVisible();
		// config:read - the Status pages are visible.
		await expect(
			viewerPage.getByRole('button', { name: 'Status' }),
		).toBeVisible();
		// No write action is a read, so nothing that changes a service.
		await expect(
			viewerPage.getByRole('button', { name: 'Toggle edit mode' }),
		).toHaveCount(0);
	});

	test('builds a service-scoped grant through the grant editor', async () => {
		const scopedGroup = `e2e-uiscope-${RUN_ID}`;
		await page.goto('/admin/groups');

		// Add a grant, then narrow its default (service:read, global) scope down
		// to a single service through the scope select + ref input.
		cleanupByName(cleanup, adminAPI, 'groups', scopedGroup);
		await page.getByRole('button', { name: 'Add group' }).click();
		await page.getByLabel('Name').fill(scopedGroup);
		await page.getByRole('button', { name: 'Add permission' }).click();
		await page.getByLabel('Grant 0 scope').click();
		await page.getByRole('option', { exact: true, name: 'service' }).click();
		// The ref input mounts on the scope change; it is required for a
		// non-global scope.
		await page.getByLabel('Grant 0 scope ref').fill(SERVICE);
		await page.getByRole('button', { name: 'Save' }).click();

		// The saved grant reads service:read @ service/release-argus/Argus.
		const row = page.getByRole('row', { name: new RegExp(scopedGroup) });
		await expect(row).toBeVisible();
		await row.getByRole('button', { name: '1 permission' }).hover();
		await expect(page.getByRole('tooltip')).toContainText(
			`service:read @ service/${SERVICE}`,
		);

		// Clean up.
		await page
			.getByRole('button', { name: `Delete group ${scopedGroup}` })
			.click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete' })
			.click();
		await expect(row).toHaveCount(0);
	});

	test('creates a user and disables them', async ({ browser }) => {
		const disabledUser = `e2e-disabled-${RUN_ID}`;
		await page.goto('/admin/users');

		// Create (accounts are enabled by default).
		cleanupByName(cleanup, adminAPI, 'users', disabledUser);
		await page.getByRole('button', { name: 'Add user' }).click();
		await page.getByLabel('Username').fill(disabledUser);
		await page.getByLabel('Password').fill('e2e-password');
		await page.getByRole('button', { name: 'Save' }).click();
		const row = page.getByRole('row', { name: new RegExp(disabledUser) });
		await expect(row).toBeVisible();
		// Enabled column: index 3 (Username, Display name, Groups, Enabled).
		await expect(row.getByRole('cell').nth(3)).toHaveText('Yes');

		// The Enabled toggle only shows when editing; untick it and save.
		await page
			.getByRole('button', { name: `Edit user ${disabledUser}` })
			.click();
		await page.getByLabel('Enabled').uncheck();
		await page.getByRole('button', { name: 'Save' }).click();
		await expect(row.getByRole('cell').nth(3)).toHaveText('No');

		// The disabled account is refused at login.
		const { context: attempt, login } = await newLoginContext(
			browser,
			disabledUser,
			'e2e-password',
		);
		try {
			// 429 if a CI retry already spent the login-limiter budget.
			expect([401, 429]).toContain(login.status());
		} finally {
			await attempt.close();
		}

		// Clean up.
		await page
			.getByRole('button', { name: `Delete user ${disabledUser}` })
			.click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete' })
			.click();
		await expect(row).toHaveCount(0);
	});

	test('creates a token with a preset expiry', async () => {
		const expiringToken = `e2e-expiring-${RUN_ID}`;
		await page.goto('/account/tokens');

		cleanupByName(cleanup, adminAPI, 'tokens', expiringToken);
		await page.getByRole('button', { name: 'Create token' }).click();
		await page.getByLabel('Name').fill(expiringToken);
		await page.getByLabel('Token expiry').click();
		await page.getByRole('option', { name: '30 days' }).click();
		await page.getByRole('button', { exact: true, name: 'Create' }).click();

		// Dismiss the one-time reveal.
		await expect(page.getByLabel('New token')).toBeVisible();
		await page.getByRole('button', { name: 'Done' }).click();

		// The Expires column shows a date rather than "Never".
		const row = page.getByRole('row', { name: new RegExp(expiringToken) });
		await expect(row).toBeVisible();
		await expect(row).not.toContainText('Never');

		// Clean up.
		await page
			.getByRole('button', { name: `Revoke token ${expiringToken}` })
			.click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Revoke' })
			.click();
		await expect(row).toHaveCount(0);
	});

	test('an expired session drops back to /login and returns after re-login', async () => {
		// Simulate expiry: the cookie is gone/invalid, the SPA state is not.
		await context.clearCookies();

		// The next full load fails /auth/me -> login page.
		await page.goto('/approvals');
		await expect(page).toHaveURL(/\/login/);

		// Re-login returns to the page the user was heading to.
		await signIn(ADMIN_USER, ADMIN_PASSWORD);
		await expect(page).toHaveURL(/\/approvals/);
	});

	test('logs out', async () => {
		await page.getByRole('button', { name: 'User menu' }).click();
		await page.getByRole('menuitem', { name: 'Log out' }).click();
		await expect(page).toHaveURL(/\/login/);

		// Protected pages stay locked.
		await page.goto('/approvals');
		await expect(page).toHaveURL(/\/login/);
	});
});
