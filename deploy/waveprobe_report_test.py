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
    def test_outcome_does_not_depend_on_error_prose(self):
        for outcome in ("passed", "active at limit", "no enemies observed",
                        "wave lost", "inconclusive", "failed"):
            row = {"state": "failed", "outcome": outcome,
                   "error": "arbitrarily reworded error with Elapsed:900"}
            self.assertEqual(report.classify(row), outcome)

    def test_evidence_rank_uses_outcome(self):
        stalled = {"state": "failed", "outcome": "active at limit", "error": "new wording"}
        lost = {"state": "failed", "outcome": "wave lost", "error": "new wording"}
        self.assertGreater(report.evidence_rank(stalled), report.evidence_rank(lost))

    def test_missing_retest_wave_inherits_structured_load_failure(self):
        row = {"state": "inconclusive", "outcome": "inconclusive",
               "retest_no_wave_result": True,
               "error": "any text at all"}
        self.assertEqual(report.classify(row, prior_load_failure=True), "load blocked")
        self.assertEqual(report.classify(row, prior_load_failure=False), "inconclusive")

    def test_full_report_uses_codes_with_reworded_errors(self):
        with tempfile.TemporaryDirectory() as folder:
            root = pathlib.Path(folder)

            def write_rows(name, rows):
                (root / name).write_text("".join(json.dumps(row) + "\n" for row in rows))

            write_rows("plan.jsonl", [
                {"mission": "load_case", "map": "mvm_a", "mode": "normal", "wave": 1, "state": "planned"},
                {"mission": "lost_case", "map": "mvm_b", "mode": "surge", "wave": 1, "state": "planned"},
            ])
            write_rows("shard-0.jsonl", [
                {"mission": "load_case", "map": "mvm_a", "mode": "normal", "wave": 0,
                 "state": "load_failed", "error": "new load wording"},
                {"mission": "lost_case", "map": "mvm_b", "mode": "surge", "wave": 1,
                 "state": "failed", "outcome": "wave lost", "error": "new loss wording"},
            ])
            write_rows("retest-0.jsonl", [
                {"mission": "load_case", "map": "mvm_a", "mode": "normal", "wave": 1,
                 "state": "inconclusive", "outcome": "inconclusive",
                 "retest_no_wave_result": True, "error": "arbitrary message"},
            ])
            output = io.StringIO()
            with contextlib.redirect_stdout(output):
                report.main(root)
            self.assertIn("| load blocked | 1 |", output.getvalue())
            self.assertIn("| wave lost | 1 |", output.getvalue())
