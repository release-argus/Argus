#!python3
from http.server import BaseHTTPRequestHandler, HTTPServer

hostName = "0.0.0.0"
serverPort = 9000


class MyServer(BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()
        print(self.headers.as_bytes())
        for k in self.headers.items():
            print(k)
        data = self.rfile.read(int(self.headers["Content-Length"]))
        print(data)


if __name__ == "__main__":
    webServer = HTTPServer((hostName, serverPort), MyServer)
    print("Server started http://%s:%s" % (hostName, serverPort))

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()
    print("Server stopped.")
