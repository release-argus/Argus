// fetchJSON will GET the JSON data, rewritten for use in React-Query
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
