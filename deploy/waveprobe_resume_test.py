import json
import os
import pathlib
import subprocess
import tempfile
import unittest


class ResumeSweepTest(unittest.TestCase):
    def test_resume_writes_final_report_from_screened_pass(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)
            bin_dir = root / "bin"
            bin_dir.mkdir()
            (root / "plan.jsonl").write_text(json.dumps({
                "mission": "mvm_decoy", "map": "mvm_decoy", "mode": "normal",
                "wave": 1, "state": "planned"}) + "\n")
            (root / "config.txt").write_text(
                "shards=1\nfirst_pass_timeout=3s\nwave_timeout=3s\n")
            (root / "projects.txt").write_text("probe-test\n")
            runner = root / "waveprobe"
            runner.write_text("#!/bin/sh\n"
                              "echo '{\"mission\":\"mvm_decoy\",\"mode\":\"normal\",\"wave\":1,\"state\":\"passed\",\"outcome\":\"passed\",\"bot_spawns\":1}'\n")
            runner.chmod(0o755)
            docker = bin_dir / "docker"
            docker.write_text("#!/bin/sh\n"
                              "if [ \"$1\" = inspect ]; then echo SRCDS_RCONPW=testpw; fi\n")
            docker.chmod(0o755)
            env = dict(os.environ, PATH=str(bin_dir) + os.pathsep + os.environ["PATH"])
            env.pop("WAVEPROBE_RCONPW", None)
            script = pathlib.Path(__file__).with_name("resume-waveprobe.sh")
            result = subprocess.run(["bash", str(script), str(root)], env=env,
                                    capture_output=True, text=True, timeout=30)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue((root / "REPORT.md").exists())
            summary = json.loads((root / "SUMMARY.json").read_text())
            self.assertEqual(summary["passed_cases"], 1)


if __name__ == "__main__":
    unittest.main()
