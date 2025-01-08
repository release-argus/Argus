import {
	Button,
	Dropdown,
	DropdownButton,
	Form,
	FormControl,
	InputGroup,
} from 'react-bootstrap';
import { FC, memo, useContext, useEffect, useMemo, useRef } from 'react';
import { ModalType, ServiceSummaryType } from 'types/summary';
import {
	faEye,
	faPen,
	faPlus,
	faTimes,
} from '@fortawesome/free-solid-svg-icons';

import { ApprovalsToolbarOptions } from 'types/util';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ModalContext } from 'contexts/modal';

type TypeMappingItem = string | boolean | number | number[];
type Props = {
	values: ApprovalsToolbarOptions;
	setValues: React.Dispatch<React.SetStateAction<ApprovalsToolbarOptions>>;
};

/**
 * The toolbar for the approvals page, including a search bar, hide options, and edit mode toggle.
 * - Hide options - Select box with filters to hide services that are up-to-date, updatable, skipped, or inactive.
 * - Edit mode - Toggles the ability to add/edit services.
 * - Search bar - Filter services by name.
 *
 * @param values - The values of the toolbar.
 * @param setValues - The function to set the values of the toolbar.
 * @returns A component that displays the toolbar for the approvals page.
 */
const ApprovalsToolbar: FC<Props> = ({ values, setValues }) => {
	const setValue = (param: keyof typeof values, value: TypeMappingItem) => {
		setValues((prevState) => ({
			...prevState,
			[param]: value as (typeof values)[typeof param],
		}));
	};

	const optionsMap = useMemo(
		() => ({
			upToDate: () => setValue('hide', toggleHideValue(0)),
			updatable: () => setValue('hide', toggleHideValue(1)),
			skipped: () => setValue('hide', toggleHideValue(2)),
			inactive: () => setValue('hide', toggleHideValue(3)),
			reset: () => setValue('hide', [3]),
			flipAllHideOptions: () => setValue('hide', toggleAllHideValues()),
		}),
		[values.hide],
	);

	const toggleHideValue = (value: number) =>
		values.hide.includes(value)
			? values.hide.filter((v) => v !== value)
			: [...values.hide, value];

	const toggleAllHideValues = () =>
		[0, 1, 2, 3].filter((n) => !(n !== 3 && values.hide.includes(n)));

	const handleOption = (option: string) => {
		const hideUpdatable = values.hide.includes(0);
		const hideUpToDate = values.hide.includes(1);
		const hideSkipped = values.hide.includes(2);
		switch (option) {
			case 'upToDate': // 0
				hideUpToDate && hideSkipped // 1 && 2
					? optionsMap.flipAllHideOptions()
					: optionsMap.upToDate();
				break;
			case 'updatable': // 1
				hideUpdatable && hideSkipped // 0 && 2
					? optionsMap.flipAllHideOptions()
					: optionsMap.updatable();
				break;
			case 'skipped': // 2
				hideUpdatable && hideUpToDate // 0 && 1
					? optionsMap.flipAllHideOptions()
					: optionsMap.skipped();
				break;
			case 'inactive': // 3
				optionsMap.inactive();
				break;
			case 'reset':
				optionsMap.reset();
				break;
		}
	};

	const toggleEditMode = () => {
		setValue('editMode', !values.editMode);
	};

	const { handleModal } = useContext(ModalContext);
	const showModal = useMemo(
		() => (type: ModalType, service: ServiceSummaryType) => {
			handleModal(type, service);
		},
		[],
	);

	const searchInputRef = useRef<HTMLInputElement | null>(null);
	useEffect(() => {
		const handleKeyPress = (event: KeyboardEvent) => {
			// Ignore when in an input/textarea
			if (
				event.target instanceof HTMLInputElement ||
				event.target instanceof HTMLTextAreaElement
			) {
				return;
			}

			if (event.key === '/') {
				// Focus on the search box.
				event.preventDefault();
				searchInputRef.current?.focus();
			}
		};

		const handleEscape = (event: KeyboardEvent) => {
			// Escape pressed and we are in the search box.
			if (
				event.key === 'Escape' &&
				searchInputRef.current &&
				document.activeElement === searchInputRef.current
			) {
				searchInputRef.current.blur();
				setValue('search', ''); // Clear search on escape
			}
		};

		// Add event listener
		window.addEventListener('keydown', handleKeyPress);
		window.addEventListener('keydown', handleEscape);

		// Clean up the event listener on unmount
		return () => {
			window.removeEventListener('keydown', handleKeyPress);
			window.removeEventListener('keydown', handleEscape);
		};
	}, []);

	return (
		<Form className="mb-3" style={{ display: 'flex' }}>
			<InputGroup className="me-3">
				<FormControl
					type="text"
					ref={searchInputRef}
					placeholder="Type '/' to filter services"
					value={values.search}
					onChange={(e) => setValue('search', e.target.value)}
					aria-label="Filter services"
				/>
				{values.search.length > 0 && (
					<Button
						variant="secondary"
						onClick={() => setValue('search', '')}
						aria-label="Clear search"
					>
						<FontAwesomeIcon icon={faTimes} />
					</Button>
				)}
			</InputGroup>
			<DropdownButton
				className="me-2"
				variant="secondary"
				title={<FontAwesomeIcon icon={faEye} />}
			>
				<Dropdown.Item
					eventKey="upToDate"
					active={values.hide.includes(0)}
					onClick={() => handleOption('upToDate')}
				>
					Hide up-to-date
				</Dropdown.Item>
				<Dropdown.Item
					eventKey="updatable"
					active={values.hide.includes(1)}
					onClick={() => handleOption('updatable')}
				>
					Hide updatable
				</Dropdown.Item>
				<Dropdown.Item
					eventKey="skipped"
					active={values.hide.includes(2)}
					onClick={() => handleOption('skipped')}
				>
					Hide skipped
				</Dropdown.Item>
				<Dropdown.Item
					eventKey="inactive"
					active={values.hide.includes(3)}
					onClick={() => handleOption('inactive')}
				>
					Hide inactive
				</Dropdown.Item>
				<Dropdown.Divider />
				<Dropdown.Item eventKey="reset" onClick={() => handleOption('reset')}>
					Reset
				</Dropdown.Item>
			</DropdownButton>
			{values.editMode && (
				<Button
					variant="secondary"
					onClick={() => showModal('EDIT', { id: '', loading: false })}
					className="me-2"
				>
					<FontAwesomeIcon icon={faPlus} />
				</Button>
			)}
			<Button variant="secondary" onClick={toggleEditMode}>
				<FontAwesomeIcon icon={faPen} />
			</Button>
		</Form>
	);
};

export default memo(ApprovalsToolbar);
