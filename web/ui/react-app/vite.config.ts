import path from 'node:path';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react-swc';
import { defineConfig } from 'vite';
import babel from 'vite-plugin-babel';
import viteTsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
	base: '',
	plugins: [
		react(),
		tailwindcss(),
		viteTsconfigPaths(),
		babel({
			babelConfig: {
				plugins: [['babel-plugin-react-compiler']],
				presets: ['@babel/preset-typescript'],
			},
			filter: /\.tsx?$/,
		}),
	],
	resolve: {
		alias: {
			'@': path.resolve(__dirname, 'src'),
		},
	},
	server: {
		open: true,
		port: 3000,
		proxy: {
			'/api': {
				changeOrigin: true,
				configure: (proxy) => {
					proxy.on('error', (err) => {
						console.log('Proxy error', err);
					});
					proxy.on('proxyReq', (_proxyReq, req) => {
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
				secure: false,
				target: 'http://localhost:8080',
				ws: true,
			},
		},
	},
});
