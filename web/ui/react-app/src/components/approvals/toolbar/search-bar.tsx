import { Button, FormControl, InputGroup } from 'react-bootstrap';
import { FC, useEffect, useRef, useState } from 'react';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTimes } from '@fortawesome/free-solid-svg-icons';

type Props = {
	search: string;
	setSearch: (value: string) => void;
};

const SearchBar: FC<Props> = ({ search, setSearch }) => {
	const searchInputRef = useRef<HTMLInputElement | null>(null);
	const [placeholder, setPlaceholder] = useState('');

	// Update placeholder based on screen size
	const updatePlaceholder = () => {
		const newPlaceholder =
			window.innerWidth <= 576
				? 'Filter services'
				: 'Type "/" to filter services';

		// Only update if the placeholder text is different
		if (newPlaceholder !== placeholder) {
			setPlaceholder(newPlaceholder);
		}
	};

	useEffect(() => {
		updatePlaceholder();
		window.addEventListener('resize', updatePlaceholder);

		// Cleanup listener on unmount
		return () => window.removeEventListener('resize', updatePlaceholder);
	}, []);

	useEffect(() => {
		const handleKeyPress = (event: KeyboardEvent) => {
			if (
				event.target instanceof HTMLInputElement ||
				event.target instanceof HTMLTextAreaElement
			) {
				return;
			}
			if (event.key === '/') {
				event.preventDefault();
				searchInputRef.current?.focus();
			}
		};

		const handleEscape = (event: KeyboardEvent) => {
			if (
				event.key === 'Escape' &&
				searchInputRef.current &&
				document.activeElement === searchInputRef.current
			) {
				searchInputRef.current.blur();
				setSearch('');
			}
		};

		window.addEventListener('keydown', handleKeyPress);
		window.addEventListener('keydown', handleEscape);

		return () => {
			window.removeEventListener('keydown', handleKeyPress);
			window.removeEventListener('keydown', handleEscape);
		};
	}, [setSearch]);

	return (
		<InputGroup>
			<FormControl
				className="search-input"
				ref={searchInputRef}
				type="text"
				placeholder={placeholder}
				value={search}
				onChange={(e) => setSearch(e.target.value)}
				aria-label="Search and filter services by name"
			/>
			{search && (
				<Button
					variant="outline-secondary"
					onClick={() => setSearch('')}
					aria-label="Clear search"
				>
					<FontAwesomeIcon icon={faTimes} />
				</Button>
			)}
		</InputGroup>
	);
};

export default SearchBar;
