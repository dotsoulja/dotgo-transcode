from http.server import HTTPServer, SimpleHTTPRequestHandler
import time

# Record server start time
start_time = time.time()

class CORSHandler(SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        super().end_headers()

    def do_GET(self):
        current_time = time.time()
        elapsed = current_time - start_time

        # Delay .ts segment responses for the first 4 minutes (240 seconds)
        if self.path.endswith(".ts") and elapsed < 120:
            print(f"[DELAY] Delaying segment: {self.path} by 4s")
            time.sleep(4)

        super().do_GET()

# Start the server
print("Starting adaptive test server on port 8000...")
HTTPServer(("0.0.0.0", 8000), CORSHandler).serve_forever()
