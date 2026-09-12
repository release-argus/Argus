import {
	type APIRequestContext,
	type BrowserContext,
	expect,
	test,
} from '@playwright/test';
import {
	ADMIN_PASSWORD,
	ADMIN_USER,
	cleanupByName,
	cleanupStack,
	RUN_ID,
	SERVICE,
	signedInContext,
} from './fixtures/auth';
import { expectServiceLoaded, serviceCard } from './fixtures/dashboard';

/**
 * Dashboard display defaults e2e specs - run only in the `chromium-auth`
 * project, against the auth-enabled Argus instance.
 *
 * Scope: what a user saves at /account/defaults, and how the approvals
 * dashboard applies it.
 */

const PASSWORD = 'e2e-defaults-password';

test.describe('Dashboard defaults', () => {
	let adminAPI: APIRequestContext;
	const cleanup = cleanupStack();

	/**
	 * Creates a user, registering its removal, and returns its username.
	 * In the viewer group, so the dashboard has services to lay out.
	 */
	const createUser = async (suffix: string) => {
		const username = `e2e-defaults-${suffix}-${RUN_ID}`;
		cleanupByName(cleanup, adminAPI, 'users', username);
		const created = await adminAPI.post('api/v1/users', {
			data: { groups: ['viewer'], password: PASSWORD, username: username },
		});
		expect(created.status(), `create ${username}`).toBe(201);
		return username;
	};

	/** A signed-in browser context for username. */
	const sessionFor = async (
		browser: Parameters<typeof signedInContext>[0],
		username: string,
	): Promise<BrowserContext> => {
		const context = await signedInContext(browser, username, PASSWORD);
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

	test('saved defaults apply to a clean dashboard URL', async ({ browser }) => {
		// GIVEN: a user on the defaults page, with nothing saved.
		const username = await createUser('apply');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');

		// WHEN: the table layout and an extra filter are saved.
		await page.getByRole('radio', { name: 'Table' }).click();
		await page.getByRole('checkbox', { name: 'Hide up to date' }).click();
		const save = page.getByRole('button', { name: 'Save changes' });
		await save.click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// AND: the button keeps focus.
		await expect(save).toBeFocused();
		await expect(save).toHaveAttribute('aria-disabled', 'true');

		// THEN: a clean /approvals opens in the table layout.
		await page.goto('/approvals');
		await expect(page.getByRole('table')).toBeVisible();

		// AND: the URL stays clean - the defaults are not echoed into it.
		expect(new URL(page.url()).search).toBe('');

		// AND: the saved filter is the one the dropdown reports.
		await page.getByRole('button', { name: 'Filter shown services' }).click();
		await expect(
			page.getByRole('menuitemcheckbox', { name: 'Hide up to date' }),
		).toBeChecked();
		await page.keyboard.press('Escape');

		// AND: they survive a reload, so they came from the account.
		await page.reload();
		await expect(page.getByRole('table')).toBeVisible();
	});

	test('a URL param still overrides the saved default', async ({ browser }) => {
		// GIVEN: a user who defaults to the table layout.
		const username = await createUser('override');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');
		await page.getByRole('radio', { name: 'Table' }).click();
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// WHEN: the dashboard is opened with an explicit layout.
		await page.goto('/approvals?view=grid');

		// THEN: the link view wins.
		await expect(
			page.getByRole('heading', { exact: true, name: SERVICE }),
		).toBeVisible();
		await expect(page.getByRole('table')).toHaveCount(0);
	});

	test('the toolbar Reset returns to the saved filters', async ({
		browser,
	}) => {
		// GIVEN: a user who defaults to hiding up-to-date services.
		const username = await createUser('toolbar-reset');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');
		await page.getByRole('checkbox', { name: 'Hide up to date' }).click();
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// WHEN: that filter is turned off on the dashboard, then Reset is used.
		await page.goto('/approvals');
		await page.getByRole('button', { name: 'Filter shown services' }).click();
		await page
			.getByRole('menuitemcheckbox', { name: 'Hide up to date' })
			.click();
		await expect(
			page.getByRole('menuitemcheckbox', { name: 'Hide up to date' }),
		).not.toBeChecked();
		await page.getByRole('menuitem', { name: 'Reset filters' }).click();

		// THEN: it comes back.
		await expect(
			page.getByRole('menuitemcheckbox', { name: 'Hide up to date' }),
		).toBeChecked();
	});

	test('each control group carries its heading as its name', async ({
		browser,
	}) => {
		// GIVEN: a user on the defaults page.
		const username = await createUser('a11y');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');

		// THEN: every group of controls is announced with what it configures.
		await expect(
			page.getByRole('radiogroup', { name: 'Layout' }),
		).toBeVisible();
		await expect(page.getByRole('group', { name: 'Filters' })).toBeVisible();
		await expect(page.getByRole('group', { name: 'Timestamps' })).toBeVisible();
	});

	test('a saved timestamp choice applies to the cards', async ({ browser }) => {
		// GIVEN: a user on the defaults page.
		const username = await createUser('timestamps');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');

		// WHEN: only the 'found' timestamp is left on, and saved.
		await page.getByRole('checkbox', { name: 'Show queried' }).click();
		await page.getByRole('checkbox', { name: 'Show found' }).click();
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// THEN: the card shows that timestamp and no other.
		await page.goto('/approvals');
		await expectServiceLoaded(page, SERVICE);
		const lines = serviceCard(page, SERVICE).locator('[data-timestamp]');
		await expect(lines).toHaveCount(1);
		await expect(lines.first()).toHaveAttribute('data-timestamp', 'found');
	});

	test('a saved empty filter set hides nothing', async ({ browser }) => {
		// GIVEN: a user on the defaults page.
		const username = await createUser('nofilters');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');

		// WHEN: every filter is off - an explicit none, not "inherit".
		await page.getByRole('checkbox', { name: 'Hide inactive' }).click();
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// THEN: a clean /approvals applies the empty set rather than the built-in
		// default, so nothing is filtered out.
		await page.goto('/approvals');
		await page.getByRole('button', { name: 'Filter shown services' }).click();
		for (const label of [
			'Hide up to date',
			'Hide updatable',
			'Hide skipped',
			'Hide inactive',
		])
			await expect(
				page.getByRole('menuitemcheckbox', { exact: true, name: label }),
			).not.toBeChecked();
	});

	test('reset discards the saved defaults', async ({ browser }) => {
		// GIVEN: a user who has saved the table layout as default.
		const username = await createUser('reset');
		const page = await (await sessionFor(browser, username)).newPage();
		await page.goto('/account/defaults');
		await page.getByRole('radio', { name: 'Table' }).click();
		await page.getByRole('button', { name: 'Save changes' }).click();
		await expect(page.getByText('Dashboard defaults saved')).toBeVisible();

		// WHEN: they reset the defaults.
		await page.getByRole('button', { name: 'Reset to Argus defaults' }).click();
		await expect(page.getByText('Dashboard defaults reset')).toBeVisible();

		// AND: the button keeps focus.
		const reset = page.getByRole('button', { name: 'Reset to Argus defaults' });
		await expect(reset).toHaveAttribute('aria-disabled', 'true');
		await expect(reset).toBeFocused();

		// AND: the dashboard is back to the grid layout.
		await page.goto('/approvals');
		await expect(
			page.getByRole('heading', { exact: true, name: SERVICE }),
		).toBeVisible();
		await expect(page.getByRole('table')).toHaveCount(0);
	});
});
