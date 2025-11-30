declare global {
	interface ImportMetaEnv {
		readonly PROD: boolean;
		readonly DEV: boolean;
		readonly VITE_WS_PORT?: string;
	}

	interface ImportMeta {
		readonly env: ImportMetaEnv;
	}
}

export const PROTOCOL = globalThis.document.location.protocol;
export const HOST = globalThis.document.location.host.replace(/:.*/, '');
export const PORT = import.meta.env?.PROD ? globalThis.location.port : 8080;

export const API_ADDRESS = `${PROTOCOL}//${HOST}:${PORT}`;

export const WS_PROTO = PROTOCOL === 'http:' ? 'ws' : 'wss';
export const WS_ADDRESS = `${WS_PROTO}://${HOST}:${PORT}`;
