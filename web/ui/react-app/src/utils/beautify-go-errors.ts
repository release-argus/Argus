/**
 * Beautify the error message.
 *
 * @param error - The error message.
 * @returns Replaces escape sequences in Go error messages with their actual characters.
 */
const beautifyGoErrors = (error: string) => {
	return error
		.replaceAll(/\\([ \t])/g, '\n$1') // \ + space/tab -> newline
		.replaceAll(`\\n`, '\n') // \n -> newline
		.replaceAll(`\\"`, `"`) // \" -> "
		.replaceAll(`\\\\`, `\\`) // \\ -> \
		.replaceAll(/\\$/g, '\n'); // \ + end of string -> newline
};

export default beautifyGoErrors;
