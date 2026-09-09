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
	RUN_ID,
	signedInContext,
} from './fixtures/auth';

/**
 * Account settings e2e specs - run only in the `chromium-auth` project,
 * against the auth-enabled Argus instance.
 *
 * Scope: what a user may change about their own account at /account/profile -
 * their profile fields, and their password. Each test owns a user of its own.
 */

const PASSWORD = 'e2e-account-password';
const NEW_PASSWORD = 'e2e-account-password-2';

test.describe('Account settings', () => {
	let adminAPI: APIRequestContext;
	const cleanup = cleanupStack();

	/** Signs in through the login form. */
	const signIn = async (page: Page, username: string, password: string) => {
		await page.goto('/login');
		await page.getByLabel('Username').fill(username);
		await page.getByLabel('Password', { exact: true }).fill(password);
		await page.getByRole('button', { name: 'Sign in' }).click();
	};

	/** Creates a user, registering its removal, and returns its username. */
	const createUser = async (suffix: string) => {
		const username = `e2e-account-${suffix}-${RUN_ID}`;
		cleanupByName(cleanup, adminAPI, 'users', username);
		const created = await adminAPI.post('api/v1/users', {
			data: { password: PASSWORD, username: username },
		});
		expect(created.status(), `create ${username}`).toBe(201);
		return username;
	};

	/**
	 * A signed-in browser context for username, closed during teardown. Logs in
	 * through the API: only the password test exercises the login form itself.
	 */
	const sessionFor = async (
		browser: Parameters<typeof signedInContext>[0],
		username: string,
		password: string,
	): Promise<BrowserContext> => {
		const context = await signedInContext(browser, username, password);
		cleanup.add(() => context.close());
		return context;
	};

	test.beforeAll(async ({ browser }) => {
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
	});

	test('renames the display name and email', async ({ browser }) => {
		const username = await createUser('profile');
		const context = await sessionFor(browser, username, PASSWORD);
		const page = await context.newPage();

		await page.goto('/account/profile');

		// The username identifies the account, so it is shown but not editable.
		await expect(page.getByLabel('Username')).toBeDisabled();
		await expect(page.getByLabel('Username')).toHaveValue(username);

		// Saving needs the current password, whatever the change.
		await page.getByLabel('Display name').fill('Renamed User');
		await page.getByLabel('Email').fill('renamed@example.com');
		await expect(
			page.getByRole('button', { name: 'Save changes' }),
		).toBeDisabled();

		await page
			.getByRole('textbox', { exact: true, name: 'Current password' })
			.fill(PASSWORD);
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Account updated')).toBeVisible();

		// The header picks the new name up without a reload.
		await page.getByRole('button', { name: 'User menu' }).click();
		await expect(page.getByText('Renamed User')).toBeVisible();
		await page.keyboard.press('Escape');

		// AND: the change survives a reload, so it reached the API.
		await page.reload();
		await expect(page.getByLabel('Display name')).toHaveValue('Renamed User');
		await expect(page.getByLabel('Email')).toHaveValue('renamed@example.com');
	});

	test('form rejects a wrong current password', async ({ browser }) => {
		const username = await createUser('wrong-pw');
		const context = await sessionFor(browser, username, PASSWORD);
		const page = await context.newPage();

		await page.goto('/account/profile');
		await page.getByLabel('Display name').fill('Nice Try');
		await page
			.getByRole('textbox', { exact: true, name: 'Current password' })
			.fill('not-my-password');
		await page.getByRole('button', { name: 'Save changes' }).click();

		// The error belongs to the field, and the session survives it.
		await expect(page.locator('#account-current-password-error')).toHaveText(
			'Current password is incorrect.',
		);
		await expect(page).toHaveURL(/\/account\/profile/);

		// AND: nothing was saved.
		await page.reload();
		await expect(page.getByLabel('Display name')).toHaveValue('');
	});

	test('changes the password, and only the new one works afterwards', async ({
		browser,
	}) => {
		const username = await createUser('password');
		const context = await sessionFor(browser, username, PASSWORD);
		const page = await context.newPage();

		// WHEN: the user sets a new password.
		await page.goto('/account/profile');
		await page
			.getByRole('textbox', { exact: true, name: 'New password' })
			.fill(NEW_PASSWORD);
		await page
			.getByRole('textbox', { exact: true, name: 'Confirm new password' })
			.fill(NEW_PASSWORD);
		await page
			.getByRole('textbox', { exact: true, name: 'Current password' })
			.fill(PASSWORD);
		await page.getByRole('button', { name: 'Save changes' }).click();

		// THEN: it succeeds, and this session survives its own change.
		await expect(
			page.getByText('Account updated - other sessions signed out'),
		).toBeVisible();
		await page.goto('/account/profile');
		await expect(page.getByLabel('Username')).toHaveValue(username);

		// WHEN: they log out.
		await page.getByRole('button', { name: 'User menu' }).click();
		await page.getByRole('menuitem', { name: 'Log out' }).click();
		await expect(page).toHaveURL(/\/login/);

		// THEN: the old password no longer gets them in.
		await signIn(page, username, PASSWORD);
		await expect(
			page.getByText(/invalid credentials|too many login attempts/),
		).toBeVisible();
		await expect(page).toHaveURL(/\/login/);

		// AND: the new one does - landing back on the page they logged out of.
		await signIn(page, username, NEW_PASSWORD);
		await expect(page).not.toHaveURL(/\/login/);
		await expect(page.getByRole('button', { name: 'User menu' })).toBeVisible();
	});
});
