export interface Info {
  build?: BuildInfo;
  runtime?: RuntimeInfo;
}
export interface BuildInfo {
  version: string;
  build_date: string;
  go_version: string;
}

export interface RuntimeInfo {
  start_time: number;
  cwd: string;
  goroutines: number;
  GOMAXPROCS: number;
  GOGC: string;
  GODEBUG: string;
}
