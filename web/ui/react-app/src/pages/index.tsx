import { Approvals } from './approvals';
import { Config } from './status/configuration';
import { Flags } from './status/cli_flags';
import { Status } from './status/runtime_and_build_info';

const ApprovalsPage = Approvals;
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
