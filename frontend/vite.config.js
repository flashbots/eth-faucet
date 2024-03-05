import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import basicSsl from '@vitejs/plugin-basic-ssl';

export default defineConfig({
	server: {
		https: true,
	},

	plugins: [
		sveltekit(),

		basicSsl({
			certDir: '.tls',
		}),
	]
});
