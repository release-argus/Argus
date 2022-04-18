export const PROTOCOL = window.document.location.protocol;
export const HOST = window.document.location.host.replace(/:.*/, "");
export const WS_PORT =
  process.env.NODE_ENV === "development" ? 8080 : window.location.port;
export const API_BASE = `/api/v1`;
export const WS_PROTO = PROTOCOL === "http:" ? "ws" : "wss";
export const WS_ADDRESS = `${WS_PROTO}://${HOST}:${WS_PORT}`;
