import { Approvals } from "./approvals";
import { Config } from "./status/configuration";
import { Flags } from "./status/cli_flags";
import { Status } from "./status/runtime_and_build_info";
// import { withStartingIndicator } from 'components/withStartingIndicator';

// const ApprovalsPage = withStartingIndicator(Agent);
const ApprovalsPage = Approvals;
// const StatusPage = withStartingIndicator(Status);
const StatusPage = Status;
const FlagsPage = Flags;
const ConfigPage = Config;

// prettier-ignore
export {
	ApprovalsPage,
	StatusPage,
	FlagsPage,
	ConfigPage
};
