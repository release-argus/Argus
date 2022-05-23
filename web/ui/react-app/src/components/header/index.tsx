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
            <NavDropdown title="Help" id="basic-nav-dropdown">
              <NavDropdown.Item
                href="https://github.com/release-argus/Argus"
                target="_blank"
                rel="noreferrer noopener"
              >
                GitHub (source)
              </NavDropdown.Item>
              <NavDropdown.Item
                href="https://github.com/release-argus/Argus/issues"
                target="_blank"
                rel="noreferrer noopener"
              >
                Report an issue/feature request
              </NavDropdown.Item>
              <NavDropdown.Item
                href="https://release-argus.io/docs"
                target="_blank"
                rel="noreferrer noopener"
              >
                Docs
              </NavDropdown.Item>
            </NavDropdown>
          </Nav>
        </Navbar.Collapse>
        <ThemeToggle />
      </Container>
    </Navbar>
  );
};

export default Header;
