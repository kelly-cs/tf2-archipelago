"""Private HTTP adapter around Archipelago's official seed generator.

The Compose admin cannot safely control Docker and must not need a host
Archipelago installation. This process lives only on the Compose network,
accepts one player YAML, and returns the archive made by the pinned Generate.py.
"""

from __future__ import annotations

import http.server
import os
import pathlib
import subprocess
import sys
import tempfile
import threading

MAX_PLAYER_FILE = 1 << 20
GENERATE_TIMEOUT = 180
generate_lock = threading.Lock()


class Handler(http.server.BaseHTTPRequestHandler):
    server_version = "tf2ap-generator"

    def do_GET(self) -> None:  # noqa: N802 - BaseHTTPRequestHandler API
        if self.path != "/healthz":
            self.fail(http.HTTPStatus.NOT_FOUND, "not found")
            return
        self.send_response(http.HTTPStatus.OK)
        self.end_headers()
        self.wfile.write(b"ready\n")

    def do_POST(self) -> None:  # noqa: N802 - BaseHTTPRequestHandler API
        if self.path != "/generate":
            self.fail(http.HTTPStatus.NOT_FOUND, "not found")
            return
        try:
            length = int(self.headers.get("Content-Length", ""))
        except ValueError:
            length = -1
        if length <= 0 or length > MAX_PLAYER_FILE:
            self.fail(http.HTTPStatus.BAD_REQUEST, "invalid player file size")
            return
        player_file = self.rfile.read(length)
        if len(player_file) != length:
            self.fail(http.HTTPStatus.BAD_REQUEST, "incomplete player file")
            return

        try:
            with generate_lock:
                archive_name, archive = generate(player_file)
        except subprocess.TimeoutExpired:
            self.fail(http.HTTPStatus.GATEWAY_TIMEOUT, "seed generation timed out")
            return
        except subprocess.CalledProcessError as error:
            detail = error.stdout.decode("utf-8", errors="replace")[-4000:]
            print(detail, file=sys.stderr, flush=True)
            self.fail(
                http.HTTPStatus.UNPROCESSABLE_ENTITY,
                detail or "seed generation failed",
            )
            return
        except (OSError, RuntimeError) as error:
            print(f"seed generation failed: {error}", file=sys.stderr, flush=True)
            self.fail(http.HTTPStatus.INTERNAL_SERVER_ERROR, "seed generation failed")
            return

        self.send_response(http.HTTPStatus.OK)
        self.send_header("Content-Type", "application/zip")
        self.send_header("Content-Disposition", f'attachment; filename="{archive_name}"')
        self.send_header("Content-Length", str(len(archive)))
        self.end_headers()
        self.wfile.write(archive)

    def log_message(self, message: str, *args: object) -> None:
        print(f"generator: {message % args}", flush=True)

    def fail(self, status: http.HTTPStatus, message: str) -> None:
        body = (message.strip() + "\n").encode()
        self.send_response(status)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def generate(player_file: bytes) -> tuple[str, bytes]:
    with tempfile.TemporaryDirectory(prefix="tf2ap-generate-") as temporary:
        work = pathlib.Path(temporary)
        players = work / "players"
        output = work / "output"
        players.mkdir()
        output.mkdir()
        (players / "tf2.yaml").write_bytes(player_file)
        subprocess.run(
            [
                sys.executable,
                "/ap/Generate.py",
                "--player_files_path",
                str(players),
                "--outputpath",
                str(output),
            ],
            cwd="/ap",
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            check=True,
            timeout=GENERATE_TIMEOUT,
        )
        archives = list(output.glob("AP_*.zip"))
        if len(archives) != 1:
            raise RuntimeError(f"generation produced {len(archives)} archives")
        return archives[0].name, archives[0].read_bytes()


def main() -> None:
    port = int(os.environ.get("AP_GENERATOR_PORT", "38282"))
    server = http.server.ThreadingHTTPServer(("0.0.0.0", port), Handler)
    print(f"seed generator listening on {port}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
