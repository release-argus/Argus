import { Container, Nav, NavDropdown, Navbar } from "react-bootstrap";

import { Link } from "react-router-dom";
import { ThemeToggle } from "theme";

const Header = () => {
  return (
    <Navbar className="navbar" variant="dark" expand="sm">
      <Container fluid>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Brand as={Link} to="/approvals">
          <img src="favicon.svg" className="d-inline-block" alt="Argus logo" />
          Argus
        </Navbar.Brand>
        <Navbar.Collapse id="navbar-nav">
          <Nav className="me-auto">
            <Nav.Link as={Link} to="/approvals">
              Approvals
            </Nav.Link>
            <NavDropdown title="Status" id="basic-nav-dropdown">
              <NavDropdown.Item as={Link} to="/status">
                Runtime & Build Information
              </NavDropdown.Item>
              <NavDropdown.Item as={Link} to="/flags">
                Command-Line Flags
              </NavDropdown.Item>
              <NavDropdown.Item as={Link} to="/config">
                Configuration
              </NavDropdown.Item>
            </NavDropdown>
            <Nav.Link href="https://release-argus.io/docs">Help</Nav.Link>
          </Nav>
        </Navbar.Collapse>
        <ThemeToggle />
      </Container>
    </Navbar>
  );
};

export default Header;
