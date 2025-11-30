export type BuildInfo = {
	/* The version of the app. */
	version: string;
	/* The build date of the app. */
	build_date: string;
	/* The Go version of the app. */
	go_version: string;
};

export type RuntimeInfo = {
	/* The start time of the app. */
	start_time: number;
	/* The current working directory of the app. */
	cwd: string;
	/* The number of goroutines. */
	goroutines: number;
	/* The maximum number of system threads that can execute simultaneously. */
	GOMAXPROCS: number;
	/* The GOGC environment variable. */
	GOGC: string;
	/* The GODEBUG environment variable. */
	GODEBUG: string;
};
