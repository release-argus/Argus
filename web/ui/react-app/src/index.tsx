import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import '@/app/globals.css';

const rootElement = document.getElementById('root');
if (!rootElement) {
	throw new Error('Could not find the root element to mount the application.');
}

createRoot(rootElement).render(
	<StrictMode>
		<App />
	</StrictMode>,
);
