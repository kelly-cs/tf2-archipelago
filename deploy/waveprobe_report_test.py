import importlib.util
import pathlib
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
