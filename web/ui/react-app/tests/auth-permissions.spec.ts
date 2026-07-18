import {
	type APIRequestContext,
	type BrowserContext,
	expect,
	test,
} from '@playwright/test';
import type { Action, Grant, Resource } from '../src/types/auth';
import {
	ADMIN_PASSWORD,
	ADMIN_USER,
	AUTH_BASE,
	cleanupStack,
	PENDING_SERVICE,
	RUN_ID,
	SERVICE,
	signedInContext,
} from './fixtures/auth';

/**
 * Permission-gating e2e specs - run only in the `chromium-auth` project,
 * against the auth-enabled Argus instance (AUTH_BASE_URL, bootstrap admin
 * created by `make playwright-tests-setup`).
 *
 * Scope: what each grant permits, at the API and in the UI.
 *
 * Three route sets, each probed against every persona:
 *   - PROBES - grant-guarded routes. Probed as the personas whose grants
 *     satisfy the guard (must not be refused) and as every persona whose
 *     grants do not (must be 403). The personas hold one permission each, so
 *     the "denied" half of each probe covers the whole rest of the catalogue
 *     rather than just a grantless user.
 *   - ADMIN_ONLY - routes behind requireAdmin, which no catalogue grant reaches.
 *   - AUTHENTICATED - routes any session reaches, whatever its grants.
 *
 * Probes that would mutate config or reach the network target a service ID
 * that does not exist: a global grant still admits the request, so the guard is
 * exercised, and the handler then fails on the missing service without side
 * effects. Write probes likewise carry a body that fails validation after the
 * guard, so no probe can leave an entity behind. The routes that must be shown
 * to genuinely *work* (create, update, delete) get their own round-trip test
 * against a throwaway service.
 *
 * Every login here succeeds.
 */

/** A service ID (and user/group ID) that does not exist. */
const ABSENT = `absent-${RUN_ID}`;

const PASSWORD = 'e2e-permission-password';

/** Statuses the auth layer returns; a probe that yields neither got through. */
const REFUSALS = [401, 403];

const grant = (
	resource: Resource,
	action: Action,
	scope: Grant['scope'] = { type: 'global' },
): Grant => ({
	action,
	resource,
	scope,
});

/**
 * One persona per catalogue permission, plus a service-scoped reader, the
 * pairings that `guardReadable` needs (its routes require service:read on top
 * of their own permission) and a user with no grants at all.
 */
const PERSONAS = [
	{ grants: [], key: 'grantless' },
	{ grants: [grant('config', 'read')], key: 'config-read' },
	{ grants: [grant('service', 'read')], key: 'service-read' },
	{
		grants: [grant('service', 'read', { ref: SERVICE, type: 'service' })],
		key: 'service-read-scoped',
	},
	{ grants: [grant('service', 'create')], key: 'service-create' },
	{ grants: [grant('service', 'update')], key: 'service-update' },
	{ grants: [grant('service', 'delete')], key: 'service-delete' },
	{ grants: [grant('service_order', 'update')], key: 'order-update' },
	{ grants: [grant('service_action', 'execute')], key: 'action-execute' },
	{ grants: [grant('version_refresh', 'execute')], key: 'refresh-execute' },
	{ grants: [grant('notify', 'execute')], key: 'notify-execute' },
	{
		grants: [grant('service_action', 'execute'), grant('service', 'read')],
		key: 'action-execute+read',
	},
	{
		grants: [grant('version_refresh', 'execute'), grant('service', 'read')],
		key: 'refresh-execute+read',
	},
] as const;

type PersonaKey = (typeof PERSONAS)[number]['key'];

/** The username a persona's fixture user is created under. */
const personaUsername = (key: PersonaKey) => `e2e-perm-${key}-${RUN_ID}`;

/**
 * Personas holding service:read, for the routes that require it. Every such
 * route targets SERVICE or accepts any scope, so the scoped reader qualifies
 * alongside the global ones.
 */
const READERS: PersonaKey[] = [
	'service-read',
	'service-read-scoped',
	'action-execute+read',
	'refresh-execute+read',
];

type Route = {
	name: string;
	method: 'DELETE' | 'GET' | 'PATCH' | 'POST' | 'PUT';
	path: string;
	data?: unknown;
	/**
	 * Exact status an entitled caller must get. Set only where the route is
	 * side-effect free, so it can be driven to a real success.
	 */
	ok?: number;
};

type Probe = Route & {
	/** Personas whose grants satisfy this route's guard. */
	allow: PersonaKey[];
};

const PROBES: Probe[] = [
	// config:read.
	{
		allow: ['config-read'],
		method: 'GET',
		name: 'config',
		ok: 200,
		path: '/api/v1/config',
	},
	{
		allow: ['config-read'],
		method: 'GET',
		name: 'runtime info',
		ok: 200,
		path: '/api/v1/status/runtime',
	},
	{
		allow: ['config-read'],
		method: 'GET',
		name: 'build info',
		ok: 200,
		path: '/api/v1/version',
	},
	{
		allow: ['config-read'],
		method: 'GET',
		name: 'flags',
		ok: 200,
		path: '/api/v1/flags',
	},
	{
		allow: ['config-read'],
		method: 'GET',
		name: 'counts',
		ok: 200,
		path: '/api/v1/counts',
	},

	// service:read, per-service target.
	{
		allow: READERS,
		method: 'GET',
		name: 'service summary',
		ok: 200,
		path: `/api/v1/service/summary?service_id=${encodeURIComponent(SERVICE)}`,
	},
	{
		allow: READERS,
		method: 'GET',
		name: 'service actions',
		ok: 200,
		path: `/api/v1/service/actions?service_id=${encodeURIComponent(SERVICE)}`,
	},
	{
		allow: READERS,
		method: 'GET',
		name: 'service config',
		ok: 200,
		path: `/api/v1/service/config?service_id=${encodeURIComponent(SERVICE)}`,
	},
	// service outside the scoped reader's grant refuses them.
	{
		allow: ['service-read', 'action-execute+read', 'refresh-execute+read'],
		method: 'GET',
		name: 'service summary outside the scope',
		path: `/api/v1/service/summary?service_id=${ABSENT}`,
	},
	// service:read under any scope (guardAnyScope).
	{
		allow: READERS,
		method: 'GET',
		name: 'service defaults',
		ok: 200,
		path: '/api/v1/service/defaults',
	},
	{
		allow: READERS,
		method: 'GET',
		name: 'template parse',
		ok: 200,
		path: `/api/v1/template?service_id=${encodeURIComponent(SERVICE)}&template=${encodeURIComponent('{{ service_id }}')}`,
	},

	// service_order:update.
	{
		allow: ['order-update'],
		data: { order: [SERVICE] },
		method: 'PUT',
		name: 'set service order',
		ok: 200,
		path: '/api/v1/service/order',
	},

	// guardReadable - the primary grant alone is not enough without service:read.
	{
		allow: ['action-execute+read'],
		data: { target: 'ARGUS_ALL' },
		method: 'POST',
		name: 'run service actions',
		path: `/api/v1/service/actions?service_id=${ABSENT}`,
	},
	{
		allow: ['refresh-execute+read'],
		method: 'GET',
		name: 'latest version refresh',
		path: `/api/v1/latest_version/refresh?service_id=${ABSENT}`,
	},
	{
		allow: ['refresh-execute+read'],
		method: 'GET',
		name: 'deployed version refresh',
		path: `/api/v1/deployed_version/refresh?service_id=${ABSENT}`,
	},

	// service:create.
	{
		allow: ['service-create'],
		method: 'GET',
		name: 'refresh uncreated latest version',
		path: '/api/v1/latest_version/refresh_uncreated',
	},
	{
		allow: ['service-create'],
		method: 'GET',
		name: 'refresh uncreated deployed version',
		path: '/api/v1/deployed_version/refresh_uncreated',
	},
	{
		allow: ['service-create'],
		data: { id: ABSENT },
		method: 'PUT',
		name: 'create service',
		path: '/api/v1/service/new',
	},

	// service:update / service:delete.
	{
		allow: ['service-update'],
		data: {},
		method: 'PUT',
		name: 'update service',
		path: `/api/v1/service/config?service_id=${ABSENT}`,
	},
	{
		allow: ['service-delete'],
		method: 'DELETE',
		name: 'delete service',
		path: `/api/v1/service/delete?service_id=${ABSENT}`,
	},

	// notify:execute.
	{
		allow: ['notify-execute'],
		data: {},
		method: 'POST',
		name: 'test notify',
		path: '/api/v1/notify/test',
	},
];

/**
 * Admin-only routes: no catalogue grant reaches them. The writes carry an empty
 * body (a missing username/name is rejected after the guard) and the rest an
 * absent ID, so an admin probe cannot create or destroy anything.
 */
const ADMIN_ONLY: Route[] = [
	{ method: 'GET', name: 'list users', ok: 200, path: '/api/v1/users' },
	{ data: {}, method: 'POST', name: 'create user', path: '/api/v1/users' },
	{ method: 'GET', name: 'get user', path: `/api/v1/users/${ABSENT}` },
	{
		data: {},
		method: 'PATCH',
		name: 'update user',
		path: `/api/v1/users/${ABSENT}`,
	},
	{ method: 'DELETE', name: 'delete user', path: `/api/v1/users/${ABSENT}` },
	{ method: 'GET', name: 'list groups', ok: 200, path: '/api/v1/groups' },
	{ data: {}, method: 'POST', name: 'create group', path: '/api/v1/groups' },
	{ method: 'GET', name: 'get group', path: `/api/v1/groups/${ABSENT}` },
	{
		data: {},
		method: 'PATCH',
		name: 'update group',
		path: `/api/v1/groups/${ABSENT}`,
	},
	{ method: 'DELETE', name: 'delete group', path: `/api/v1/groups/${ABSENT}` },
	{
		method: 'GET',
		name: 'permission catalogue',
		ok: 200,
		path: '/api/v1/permissions',
	},
];

/**
 * Routes any session reaches, whatever its grants: the caller's own account and
 * tokens, and the per-user-filtered service order.
 */
const AUTHENTICATED: Route[] = [
	{ method: 'GET', name: 'own account', ok: 200, path: '/api/v1/auth/me' },
	{ method: 'GET', name: 'own tokens', ok: 200, path: '/api/v1/tokens' },
	{ data: {}, method: 'POST', name: 'create token', path: '/api/v1/tokens' },
	{ method: 'DELETE', name: 'revoke token', path: `/api/v1/tokens/${ABSENT}` },
	{ method: 'GET', name: 'websocket token', path: '/api/v1/ws-token' },
	{
		method: 'GET',
		name: 'service order',
		ok: 200,
		path: '/api/v1/service/order',
	},
];

/** send issues a probe through ctx. */
const send = (ctx: APIRequestContext, route: Route) =>
	ctx.fetch(route.path, {
		...(route.data === undefined ? {} : { data: route.data }),
		method: route.method,
	});

/*
 * Not serial: the probes are independent, and the project already caps itself
 * to one worker. A failing probe must not mask the rest of the matrix.
 */
test.describe('Permission gating', () => {
	let admin: APIRequestContext;
	/** One logged-in browser context per persona, reused by every test. */
	const sessions = new Map<PersonaKey, BrowserContext>();
	const cleanup = cleanupStack();

	/** The logged-in context for a persona; hard failure if setup missed it. */
	const sessionFor = (key: PersonaKey): BrowserContext => {
		const context = sessions.get(key);
		if (!context) throw new Error(`no session for ${key}`);
		return context;
	};

	test.beforeAll(async ({ browser }) => {
		const adminContext = await signedInContext(
			browser,
			ADMIN_USER,
			ADMIN_PASSWORD,
		);
		cleanup.add(() => adminContext.close());
		admin = adminContext.request;

		// The pending-release fixture's first lookup lands a beat after startup,
		// and the approve/skip affordances have nothing to render until it does.
		await expect(async () => {
			const summary = await admin.get(
				`/api/v1/service/summary?service_id=${encodeURIComponent(PENDING_SERVICE.id)}`,
			);
			expect(summary.status()).toBe(200);
			const { status } = (await summary.json()) as {
				status?: { deployed_version?: string; latest_version?: string };
			};
			expect(status?.latest_version, 'pending-release latest').toBe(
				PENDING_SERVICE.latestVersion,
			);
			expect(status?.deployed_version, 'pending-release deployed').toBe(
				PENDING_SERVICE.deployedVersion,
			);
		}).toPass({ timeout: 30_000 });

		for (const persona of PERSONAS) {
			const name = personaUsername(persona.key);
			const groups: string[] = [];

			if (persona.grants.length > 0) {
				const group = await admin.post('/api/v1/groups', {
					data: { name, permissions: persona.grants },
				});
				expect(group.status(), `create group for ${persona.key}`).toBe(201);
				const groupID = ((await group.json()) as { id: string }).id;
				groups.push(name);
				cleanup.add(() => admin.delete(`/api/v1/groups/${groupID}`));
			}

			const user = await admin.post('/api/v1/users', {
				data: { groups: groups, password: PASSWORD, username: name },
			});
			expect(user.status(), `create user for ${persona.key}`).toBe(201);
			const userID = ((await user.json()) as { id: string }).id;
			cleanup.add(() => admin.delete(`/api/v1/users/${userID}`));

			const context = await signedInContext(browser, name, PASSWORD);
			sessions.set(persona.key, context);
			cleanup.add(() => context.close());
		}
	});

	test.afterAll(async () => {
		await cleanup.run();
	});

	for (const probe of PROBES) {
		test(`${probe.method} ${probe.name}`, async () => {
			for (const persona of PERSONAS) {
				const status = (
					await send(sessionFor(persona.key).request, probe)
				).status();
				const as = `${probe.method} ${probe.name} as ${persona.key}`;

				if (!(probe.allow as readonly string[]).includes(persona.key)) {
					expect.soft(status, `${as} must be forbidden`).toBe(403);
					continue;
				}
				// Allowed: the guard must not be what stops them.
				expect
					.soft(REFUSALS, `${as} must not be refused`)
					.not.toContain(status);
				if (probe.ok !== undefined) {
					expect.soft(status, `${as} must succeed`).toBe(probe.ok);
				}
			}
		});
	}

	for (const route of ADMIN_ONLY) {
		test(`${route.method} ${route.name} is admin-only`, async () => {
			// No catalogue permission reaches the admin API.
			for (const persona of PERSONAS) {
				const status = (
					await send(sessionFor(persona.key).request, route)
				).status();
				expect
					.soft(status, `${route.name} as ${persona.key} must be forbidden`)
					.toBe(403);
			}

			// The admin group reaches it.
			const status = (await send(admin, route)).status();
			expect(
				REFUSALS,
				`${route.name} as admin must not be refused`,
			).not.toContain(status);
			if (route.ok !== undefined) {
				expect(status, `${route.name} as admin must succeed`).toBe(route.ok);
			}
		});
	}

	for (const route of AUTHENTICATED) {
		test(`${route.method} ${route.name} needs no grant`, async () => {
			for (const persona of PERSONAS) {
				const status = (
					await send(sessionFor(persona.key).request, route)
				).status();
				const as = `${route.name} as ${persona.key}`;
				expect
					.soft(REFUSALS, `${as} must not be refused`)
					.not.toContain(status);
				if (route.ok !== undefined) {
					expect.soft(status, `${as} must succeed`).toBe(route.ok);
				}
			}
		});
	}

	test('unauthenticated requests are refused, not forbidden', async ({
		browser,
	}) => {
		const anon = await browser.newContext({ baseURL: AUTH_BASE });
		try {
			for (const route of [...PROBES, ...ADMIN_ONLY, ...AUTHENTICATED]) {
				const status = (await send(anon.request, route)).status();
				expect.soft(status, `${route.name} while logged out`).toBe(401);
			}
		} finally {
			await anon.close();
		}
	});

	/**
	 * The affordances each permission unlocks in the UI, and the personas that
	 * should see them.
	 *
	 * The approve/skip pair renders on PENDING_SERVICE only, and needs
	 * service:read as well as service_action:execute: a persona who cannot read
	 * the service never gets a card to act on. That is what leaves
	 * `action-execute` (no read) out of their allow lists.
	 */
	const AFFORDANCES = [
		{ allow: ['config-read'], name: 'Status', role: 'button' as const },
		{
			allow: ['service-create', 'order-update', 'service-update'],
			name: 'Toggle edit mode',
			role: 'button' as const,
		},
		{
			allow: ['action-execute+read'],
			name: 'Approve release',
			role: 'button' as const,
		},
		{
			allow: ['action-execute+read'],
			name: 'Reject release',
			role: 'button' as const,
		},
	];

	for (const affordance of AFFORDANCES) {
		test(`the ${affordance.name} control follows its grant`, async () => {
			for (const persona of PERSONAS) {
				const page = await sessionFor(persona.key).newPage();
				try {
					await page.goto('/approvals');
					await expect(
						page.getByRole('button', { name: 'User menu' }),
					).toBeVisible();

					const control = page.getByRole(affordance.role, {
						name: affordance.name,
					});
					const as = `${affordance.name} for ${persona.key}`;
					if (affordance.allow.includes(persona.key)) {
						await expect.soft(control, `${as} should show`).toBeVisible();
					} else {
						await expect.soft(control, `${as} should be hidden`).toHaveCount(0);
					}
				} finally {
					await page.close();
				}
			}
		});
	}

	test('the dashboard lists only the services a persona may read', async () => {
		// Global read: both fixture services.
		const readerPage = await sessionFor('service-read').newPage();
		try {
			await readerPage.goto('/approvals');
			await expect(readerPage.getByText(SERVICE)).toBeVisible();
			await expect(readerPage.getByText(PENDING_SERVICE.id)).toBeVisible();
		} finally {
			await readerPage.close();
		}

		// Scoped to one of them: that one only, and nothing to act on it.
		const scopedPage = await sessionFor('service-read-scoped').newPage();
		try {
			await scopedPage.goto('/approvals');
			await expect(scopedPage.getByText(SERVICE)).toBeVisible();
			await expect(scopedPage.getByText(PENDING_SERVICE.id)).toHaveCount(0);
			await expect(
				scopedPage.getByRole('button', { name: 'Approve release' }),
			).toHaveCount(0);
		} finally {
			await scopedPage.close();
		}

		// No grants at all: a session, but an empty dashboard.
		const lonerPage = await sessionFor('grantless').newPage();
		try {
			await lonerPage.goto('/approvals');
			await expect(
				lonerPage.getByRole('button', { name: 'User menu' }),
			).toBeVisible();
			await expect(lonerPage.getByText(SERVICE)).toHaveCount(0);
			await expect(lonerPage.getByText(PENDING_SERVICE.id)).toHaveCount(0);
		} finally {
			await lonerPage.close();
		}
	});

	test('admin-only navigation is hidden from every catalogue permission', async () => {
		for (const persona of PERSONAS) {
			const page = await sessionFor(persona.key).newPage();
			try {
				await page.goto('/approvals');
				await page.getByRole('button', { name: 'User menu' }).click();
				// Tokens are per-user, so everyone keeps them.
				await expect(
					page.getByRole('menuitem', { name: 'API Tokens' }),
				).toBeVisible();
				for (const entry of ['Users', 'Groups']) {
					await expect
						.soft(
							page.getByRole('menuitem', { name: entry }),
							`${entry} for ${persona.key}`,
						)
						.toHaveCount(0);
				}

				// Routing them there directly bounces them back.
				await page.goto('/admin/groups');
				await expect
					.soft(page, `/admin/groups for ${persona.key}`)
					.toHaveURL(/\/approvals/);
			} finally {
				await page.close();
			}
		}
	});

	test('service create, update, and delete work for the granted personas', async () => {
		const serviceID = `e2e-perm-service-${RUN_ID}`;
		const creator = sessionFor('service-create').request;
		const updater = sessionFor('service-update').request;
		const deleter = sessionFor('service-delete').request;

		const body = {
			id: serviceID,
			latest_version: { type: 'github', url: SERVICE },
			// Inactive, so creating it never starts a real lookup.
			options: { active: false },
		};

		// Create.
		const created = await creator.put('/api/v1/service/new', { data: body });
		expect(created.status(), 'create as service-create').toBe(200);

		try {
			// The creator cannot then edit or delete it.
			expect(
				(
					await creator.put(
						`/api/v1/service/config?service_id=${encodeURIComponent(serviceID)}`,
						{ data: body },
					)
				).status(),
				'update as service-create',
			).toBe(403);

			// Update. The service stays inactive: activating it makes the edit
			// re-query GitHub, which fails on the unauthenticated rate limit.
			const updated = await updater.put(
				`/api/v1/service/config?service_id=${encodeURIComponent(serviceID)}`,
				{ data: { ...body, dashboard: { web_url: 'https://example.com' } } },
			);
			expect(updated.status(), 'update as service-update').toBe(200);
		} finally {
			// Delete.
			const deleted = await deleter.delete(
				`/api/v1/service/delete?service_id=${encodeURIComponent(serviceID)}`,
			);
			expect(deleted.status(), 'delete as service-delete').toBe(200);
		}
	});
});
