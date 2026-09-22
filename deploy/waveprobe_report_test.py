import contextlib
import importlib.util
import io
import json
import pathlib
import tempfile
import unittest

module_path = pathlib.Path(__file__).with_name("waveprobe-report.py")
spec = importlib.util.spec_from_file_location("waveprobe_report", module_path)
report = importlib.util.module_from_spec(spec)
spec.loader.exec_module(report)


class ReportOutcomeTest(unittest.TestCase):
    def test_changelevel_failure_is_counted_and_shows_map_evidence(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)
            case = {"mission": "mission", "map": "mvm_deathpour_rc1", "mode": "normal", "wave": 5}
            (root / "plan.jsonl").write_text(json.dumps({**case, "state": "planned"}) + "\n")
            (root / "shard-0.jsonl").write_text(json.dumps({**case, "state": "failed",
                "outcome": "wave timed out", "bot_spawns": 82}) + "\n")
            evidence = {"requested_map": "mvm_deathpour_rc1", "before_map": "mvm_decoy",
                        "before_pop": "mvm_decoy_advanced", "after_map": "mvm_decoy",
                        "after_pop": "mvm_decoy_advanced", "after_state": "idle",
                        "command_reply": "rejected"}
            (root / "retest-0.jsonl").write_text(json.dumps({**case, "state": "load_failed",
                "outcome": "changelevel failure", "error": "map stayed on Decoy",
                "changelevel": evidence}) + "\n")
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                report.main(root)
            summary = json.loads((root / "SUMMARY.json").read_text())
            self.assertEqual(summary["outcomes"]["changelevel failure"], 1)
            self.assertIn("Changelevel evidence: requested `mvm_deathpour_rc1`", output.getvalue())
            self.assertIn("Reply: `rejected`", output.getvalue())

    def test_old_zero_kill_rows_do_not_claim_no_spawns(self):
        self.assertEqual(report.classify({"state": "failed", "outcome": "no enemies observed"}),
                         "inconclusive")

    def test_missing_retest_wave_inherits_load_failure(self):
        row = {"state": "inconclusive", "outcome": "inconclusive",
               "retest_no_wave_result": True}
        prior = {"state": "load_failed", "outcome": "wave 0"}
        self.assertEqual(report.classify(row, prior), "wave 0")

    def test_reset_cannot_erase_observed_spawns(self):
        row = {"state": "failed", "outcome": "no enemies spawned",
               "bot_spawns": 0, "tank_spawns": 0,
               "timeline": [{"spawned": 3}, {"spawned": 0}]}
        self.assertEqual(report.classify(row), "probe error")

    def test_resumed_screen_pass_counts_as_verified(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)
            case = {"mission": "mission", "map": "mvm_map", "mode": "normal", "wave": 1}
            (root / "plan.jsonl").write_text(json.dumps({**case, "state": "planned"}) + "\n")
            (root / "shard-0.jsonl").write_text(json.dumps({**case, "state": "failed", "outcome": "probe error"}) + "\n")
            (root / "screen-0.jsonl").write_text(json.dumps({**case, "state": "passed", "outcome": "passed",
                                                               "bot_spawns": 1}) + "\n")
            with contextlib.redirect_stdout(io.StringIO()):
                report.main(root)
            summary = json.loads((root / "SUMMARY.json").read_text())
            self.assertEqual(summary["passed_cases"], 1)

    def test_report_scores_modes_and_shows_timeline(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)

            def write_rows(name, rows):
                (root / name).write_text("".join(json.dumps(row) + "\n" for row in rows))

            plan = []
            for mode in ("normal", "surge"):
                for wave in (1, 2, 3):
                    plan.append({"mission": "mission", "map": "mvm_map", "mode": mode,
                                 "wave": wave, "state": "planned"})
            write_rows("plan.jsonl", plan)
            write_rows("shard-0.jsonl", [
                {"mission": "mission", "map": "mvm_map", "mode": "normal", "wave": 1,
                 "state": "passed", "outcome": "passed", "bot_spawns": 4, "bots": 4},
                {"mission": "mission", "map": "mvm_map", "mode": "surge", "wave": 1,
                 "state": "failed", "outcome": "wave failed", "bot_spawns": 4,
                 "bots": 3, "alive_at_end": 1, "progress_percent": 75,
                 "timeline": [{"wall_seconds": 10, "game_seconds": 200,
                               "spawned": 4, "killed": 3, "alive": 1,
                               "remaining": 1, "progress_percent": 75}]},
                {"mission": "mission", "map": "mvm_map", "mode": "normal", "wave": 2,
                 "state": "load_failed", "outcome": "wave 0"},
                {"mission": "mission", "map": "mvm_map", "mode": "surge", "wave": 2,
                 "state": "failed", "outcome": "no enemies spawned"},
                {"mission": "mission", "map": "mvm_map", "mode": "normal", "wave": 3,
                 "state": "failed", "outcome": "wave timed out", "bot_spawns": 3,
                 "bots": 2, "alive_at_end": 1, "progress_percent": 66,
                 "wall_seconds": 900, "game_seconds": 18000},
            ])
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                report.main(root)
            summary = json.loads((root / "SUMMARY.json").read_text())
            self.assertEqual(summary["eligible_cases"], 6)
            self.assertEqual(summary["tested_cases"], 5)
            self.assertEqual(summary["passed_cases"], 1)
            self.assertEqual(summary["stability_percent"], 20.0)
            self.assertEqual(summary["by_mode"]["normal"]["tested"], 3)
            self.assertEqual(summary["by_mode"]["surge"]["tested"], 2)
            text = output.getvalue()
            self.assertIn("Stability score: **20.0%**", text)
            self.assertIn("| Wave Failed | 0 | 1 |", text)
            self.assertIn("| Wave 0 | 1 | 0 |", text)
            self.assertIn("| No Enemies Spawned | 0 | 1 |", text)
            self.assertIn("| 10 | 200 | 4 | 3 | 1 | 1 | 75% |", text)


if __name__ == "__main__":
    unittest.main()
