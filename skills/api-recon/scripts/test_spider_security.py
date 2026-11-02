import contextlib
import http.server
import os
import threading
import unittest
from unittest.mock import patch

import spider_mpa


@contextlib.contextmanager
def serve(routes, received):
    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            received.append((self.path, self.headers.get("Cookie")))
            status, headers, body = routes[self.path]
            self.send_response(status)
            for key, value in headers.items():
                self.send_header(key, value)
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, *args):
            pass

    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        yield "http://127.0.0.1:%d" % server.server_port
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


class SpiderSecurityTests(unittest.TestCase):
    def setUp(self):
        # Keep local integration fixtures independent of workstation proxies.
        self.enterContext(patch.dict(os.environ, {"NO_PROXY": "127.0.0.1,localhost", "no_proxy": "127.0.0.1,localhost"}))

    def test_cross_origin_redirect_never_contacts_destination(self):
        stolen, source = [], []
        with serve({"/": (200, {"Content-Type": "text/html"}, b"outside")}, stolen) as outside:
            with serve({"/": (302, {"Location": outside + "/"}, b"")}, source) as origin:
                spider_mpa.fetch(origin + "/", "session=test-secret")
        self.assertEqual(stolen, [])

    def test_same_origin_redirect_keeps_working(self):
        received = []
        routes = {
            "/": (302, {"Location": "/next"}, b""),
            "/next": (200, {"Content-Type": "text/html"}, b"safe page"),
        }
        with serve(routes, received) as origin:
            html, _ = spider_mpa.fetch(origin + "/", "session=test-secret")
        self.assertEqual(html, "safe page")
        self.assertEqual(received, [("/", "session=test-secret"), ("/next", "session=test-secret")])

    def test_origin_comparison_rejects_lookalikes_and_downgrades(self):
        for url in ("https://target.example.evil.test/", "https://target.example@evil.test/",
                    "http://target.example/", "https://target.example:444/", "file:///etc/passwd"):
            with self.subTest(url=url):
                self.assertFalse(spider_mpa.same_origin("https://target.example", url))
        self.assertTrue(spider_mpa.same_origin("https://target.example", "https://TARGET.example:443/path"))


if __name__ == "__main__":
    unittest.main()
