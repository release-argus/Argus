import {
  Button,
  Dropdown,
  DropdownButton,
  Form,
  InputGroup,
} from "react-bootstrap";
import { FC, memo, useContext, useMemo } from "react";
import { ModalType, ServiceSummaryType } from "types/summary";
import {
  faEye,
  faPen,
  faPlus,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { ApprovalsToolbarOptions } from "types/util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalContext } from "contexts/modal";

type TypeMappingItem = string | boolean | number | number[];
type Props = {
  values: ApprovalsToolbarOptions;
  setValues: React.Dispatch<React.SetStateAction<ApprovalsToolbarOptions>>;
};

const ApprovalsToolbar: FC<Props> = ({ values, setValues }) => {
  const setValue = (param: keyof typeof values, value: TypeMappingItem) => {
    setValues((prevState) => ({
      ...prevState,
      [param]: value as typeof values[typeof param],
    }));
  };

  const optionsMap = useMemo(
    () => ({
      upToDate: () => setValue("hide", toggleHideValue(0)),
      updatable: () => setValue("hide", toggleHideValue(1)),
      skipped: () => setValue("hide", toggleHideValue(2)),
      inactive: () => setValue("hide", toggleHideValue(3)),
      reset: () => setValue("hide", [3]),
      flipAllHideOptions: () => setValue("hide", toggleAllHideValues()),
    }),
    [values.hide]
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
      case "upToDate": // 0
        hideUpToDate && hideSkipped // 1 && 2
          ? optionsMap.flipAllHideOptions()
          : optionsMap.upToDate();
        break;
      case "updatable": // 1
        hideUpdatable && hideSkipped // 0 && 2
          ? optionsMap.flipAllHideOptions()
          : optionsMap.updatable();
        break;
      case "skipped": // 2
        hideUpdatable && hideUpToDate // 0 && 1
          ? optionsMap.flipAllHideOptions()
          : optionsMap.skipped();
        break;
      case "inactive": // 3
        optionsMap.inactive();
        break;
      case "reset":
        optionsMap.reset();
        break;
    }
  };

  const toggleEditMode = () => {
    setValue("editMode", !values.editMode);
  };

  const { handleModal } = useContext(ModalContext);
  const showModal = useMemo(
    () => (type: ModalType, service: ServiceSummaryType) => {
      handleModal(type, service);
    },
    []
  );

  return (
    <Form className="mb-3" style={{ display: "flex" }}>
      <InputGroup className="me-3">
        <Form.Control
          type="string"
          placeholder="Filter services"
          value={values.search}
          onChange={(e) => setValue("search", e.target.value)}
        />
        {values.search.length > 0 && (
          <Button variant="secondary" onClick={() => setValue("search", "")}>
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
          onClick={() => handleOption("upToDate")}
        >
          Hide up-to-date
        </Dropdown.Item>
        <Dropdown.Item
          eventKey="updatable"
          active={values.hide.includes(1)}
          onClick={() => handleOption("updatable")}
        >
          Hide updatable
        </Dropdown.Item>
        <Dropdown.Item
          eventKey="skipped"
          active={values.hide.includes(2)}
          onClick={() => handleOption("skipped")}
        >
          Hide skipped
        </Dropdown.Item>
        <Dropdown.Item
          eventKey="inactive"
          active={values.hide.includes(3)}
          onClick={() => handleOption("inactive")}
        >
          Hide inactive
        </Dropdown.Item>
        <Dropdown.Divider />
        <Dropdown.Item eventKey="reset" onClick={() => handleOption("reset")}>
          Reset
        </Dropdown.Item>
      </DropdownButton>
      {values.editMode && (
        <Button
          variant="secondary"
          onClick={() => showModal("EDIT", { id: "", loading: false })}
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
