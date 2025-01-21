import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import viteTsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
	base: '',
	plugins: [react(), viteTsconfigPaths()],
	server: {
		open: true,
		port: 3000,
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false,
				ws: true,
				configure: (proxy) => {
					proxy.on('error', (err) => {
						console.log('Proxy error', err);
					});
					proxy.on('proxyReq', (proxyReq, req) => {
						console.log('Sending Request to the Target:', req.method, req.url);
					});
					proxy.on('proxyRes', (proxyRes, req) => {
						console.log(
							'Received Response from the Target:',
							proxyRes.statusCode,
							req.url,
						);
					});
				},
			},
		},
	},
});
