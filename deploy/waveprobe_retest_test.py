import importlib.util
import json
import os
import pathlib
import tempfile
import unittest
from unittest.mock import patch

module_path = pathlib.Path(__file__).with_name("waveprobe-retest.py")
spec = importlib.util.spec_from_file_location("waveprobe_retest", module_path)
retest = importlib.util.module_from_spec(spec)
spec.loader.exec_module(retest)


class RetestRecoveryTest(unittest.TestCase):
    def test_rcon_failure_restarts_server_and_retries_case(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)
            (root / "config.txt").write_text("shards=1\nbase_port=27035\nwave_timeout=1s\n")
            case = {"mission": "mvm_decoy", "map": "mvm_decoy", "mode": "normal", "wave": 1}
            (root / "plan.jsonl").write_text(json.dumps({**case, "state": "planned"}) + "\n")
            (root / "shard-0.jsonl").write_text(json.dumps({**case, "state": "failed"}) + "\n")
            (root / "projects.txt").write_text("probe-test\n")
            bin_dir = root / "bin"
            bin_dir.mkdir()
            docker = bin_dir / "docker"
            docker.write_text("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$TEST_DOCKER_LOG\"\n")
            docker.chmod(0o755)
            runner = bin_dir / "probe"
            runner.write_text("#!/bin/sh\n"
                              "if [ ! -e \"$TEST_FIRST_RUN\" ]; then\n"
                              "  touch \"$TEST_FIRST_RUN\"\n"
                              "  echo '{\"wave\":0,\"error\":\"connection reset by peer\"}'\n"
                              "else\n"
                              "  echo '{\"mission\":\"mvm_decoy\",\"mode\":\"normal\",\"wave\":1,\"state\":\"passed\",\"outcome\":\"passed\"}'\n"
                              "fi\n")
            runner.chmod(0o755)
            env = {"PATH": str(bin_dir) + os.pathsep + os.environ["PATH"],
                   "TEST_DOCKER_LOG": str(root / "docker.log"),
                   "TEST_FIRST_RUN": str(root / "first-run")}
            with patch.dict(os.environ, env):
                retest.main(root, runner, [0])
            result = json.loads((root / "retest-0.jsonl").read_text())
            self.assertEqual(result["outcome"], "passed")
            self.assertEqual((root / "docker.log").read_text().splitlines(),
                             ["restart probe-test-srcds-1", "restart probe-test-srcds-1"])


if __name__ == "__main__":
    unittest.main()
