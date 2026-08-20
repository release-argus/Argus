import {
	type APIRequestContext,
	type APIResponse,
	type Browser,
	type BrowserContext,
	expect,
} from '@playwright/test';

/**
 * Shared fixtures for the auth specs, which run only in the `chromium-auth`
 * project against the auth-enabled Argus instance (AUTH_BASE_URL, bootstrap
 * admin created by `make playwright-tests-setup`).
 */

/** Base URL of the auth-enabled Argus instance. */
export const AUTH_BASE = process.env.AUTH_BASE_URL ?? 'http://localhost:8081';

/** The bootstrap admin created by `make playwright-tests-setup`. */
export const ADMIN_USER = 'admin';
export const ADMIN_PASSWORD =
	process.env.PLAYWRIGHT_AUTH_PASSWORD ?? 'playwright-admin-pw';

/** The service the auth instance is configured with (fixtures/auth-config.yml). */
export const SERVICE = 'release-argus/Argus';

/**
 * The second fixture service, which always has a release pending, so the
 * approve/skip affordances have something to render.
 */
export const PENDING_SERVICE = {
	deployedVersion: '0.0.1',
	id: 'e2e-pending-release',
	latestVersion: '1.2.3',
};

/** Unique-enough suffix so re-runs never collide with leftovers. */
export const RUN_ID = Date.now().toString(36);

/**
 * Opens a fresh browser context on the auth instance and logs username in via
 * the API. Returns the context and the login response, so the caller can assert
 * the outcome (200, or a rejection) and drive its own pages.
 */
export const newLoginContext = async (
	browser: Browser,
	username: string,
	password: string,
): Promise<{ context: BrowserContext; login: APIResponse }> => {
	const context = await browser.newContext({ baseURL: AUTH_BASE });
	const login = await context.request.post('api/v1/auth/login', {
		data: { password, username },
	});
	return { context, login };
};

/** As [newLoginContext], for the callers that require the login to succeed. */
export const signedInContext = async (
	browser: Browser,
	username: string,
	password: string,
): Promise<BrowserContext> => {
	const { context, login } = await newLoginContext(browser, username, password);
	expect(login.status(), `login as ${username}`).toBe(200);
	return context;
};

export type CleanupStack = {
	add: (undo: () => Promise<unknown>) => void;
	run: () => Promise<void>;
};

/**
 * A LIFO stack of undo steps for `afterAll`. Steps run in reverse registration
 * order, so each undoes before whatever it was registered through. A failing
 * step is reported and skipped rather than thrown: teardown must never mask the
 * test failure that caused it.
 */
export const cleanupStack = (): CleanupStack => {
	const steps: Array<() => Promise<unknown>> = [];
	return {
		add: (undo) => {
			steps.push(undo);
		},
		run: async () => {
			for (const undo of steps.splice(0).reverse()) {
				await undo().catch((err: unknown) => {
					console.warn('cleanup step failed:', err);
				});
			}
		},
	};
};

/**
 * Registers an undo that deletes the named resource if it is still there, for
 * entities created through the UI (whose IDs the test never sees). Idempotent,
 * so a test that deletes its own entity leaves nothing for this to do.
 */
export const cleanupByName = (
	cleanup: CleanupStack,
	admin: APIRequestContext,
	resource: 'groups' | 'tokens' | 'users',
	name: string,
) => {
	cleanup.add(async () => {
		const list = await admin.get(`api/v1/${resource}`);
		if (!list.ok()) return;
		const items = (await list.json()) as Array<{
			id: string;
			name?: string;
			username?: string;
		}>;
		const match = items.find((item) => (item.username ?? item.name) === name);
		if (match) await admin.delete(`api/v1/${resource}/${match.id}`);
	});
};
