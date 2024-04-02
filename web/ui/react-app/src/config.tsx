export const PROTOCOL = window.document.location.protocol;
export const HOST = window.document.location.host.replace(/:.*/, "");
export const WS_PORT = import.meta.env.PROD ? window.location.port : 8080;
export const WS_PROTO = PROTOCOL === "http:" ? "ws" : "wss";
export const WS_ADDRESS = `${WS_PROTO}://${HOST}:${WS_PORT}`;
