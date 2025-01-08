import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css';

import App from './App';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

const container = document.getElementById('root') as HTMLElement;

const root = createRoot(container);

root.render(
	<StrictMode>
		<App />
	</StrictMode>,
);
