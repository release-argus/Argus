/**
 * Returns the JSON response from the server
 *
 * @param url - The URL to fetch data from
 * @returns The JSON data returned from the request to the server
 */
const fetchJSON = async <T,>(url: string): Promise<T> => {
  const response = await Promise.race([
    fetch(url),
    new Promise<Response>((_, reject) =>
      setTimeout(() => reject(new Error("Timeout")), 10000)
    ),
  ]);
  const resp = await response.json();
  return resp;
};

export default fetchJSON;
