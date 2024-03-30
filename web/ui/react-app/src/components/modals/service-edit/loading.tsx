import {
  Accordion,
  Container,
  FormControl,
  FormGroup,
  Stack,
} from "react-bootstrap";

import { FC } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { FormLabel } from "components/generic/form";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";

interface Props {
  name: string;
}

/**
 * The disabled edit form for a service whilst loading
 *
 * @param name - The name of the service
 * @returns The edit form for the service, but disabled whilst loading
 */
export const Loading: FC<Props> = ({ name }) => {
  const delayedRender = useDelayedRender(500);
  const accordionHeaders = [
    "Options:",
    "Latest Version:",
    "Deployed Version:",
    "Commands:",
    "WebHooks:",
    "Notify:",
    "Dashboard:",
  ];
  const formControlClassName = "pt-1 pb-1 col-form col-sm-12 col-12";

  return (
    <Stack gap={3}>
      <FormGroup className="mb-2">
        <FormGroup className={formControlClassName}>
          <FormLabel text="Name" required />
          <FormControl
            autoFocus={false}
            defaultValue={name}
            disabled
            className="bg-transparent"
          />
        </FormGroup>
        <FormGroup className={formControlClassName}>
          <FormLabel text="Comment" />
          <FormControl autoFocus={false} disabled className="bg-transparent" />
        </FormGroup>
        {delayedRender(() => (
          <Container className="empty">
            <FontAwesomeIcon icon={faCircleNotch} className={"fa-spin"} />
            <span style={{ paddingLeft: "0.5rem" }}>Loading...</span>
          </Container>
        ))}
      </FormGroup>
      {accordionHeaders.map((title) => {
        return (
          <Accordion key={title}>
            <Accordion.Header>{title}</Accordion.Header>
          </Accordion>
        );
      })}
    </Stack>
  );
};
