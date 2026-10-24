"""Scores must expose missing waves, startup failures and stale game builds."""
import importlib.util
import json
import os
import subprocess
from pathlib import Path
import sys
import tempfile
import textwrap
from types import SimpleNamespace
import unittest
from unittest import mock
import zipfile
import zlib


def module(name):
    spec = importlib.util.spec_from_file_location(name.replace("-", "_"), Path(__file__).with_name(name + ".py"))
    value = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(value)
    return value


report = module("stability-report")
runner = module("stability-run")
host = module("stability-host")
publisher = module("stability-publish")
partitioner = module("waveprobe_partition")
local = module("local-waveprobe")


class StabilityTest(unittest.TestCase):
    def test_final_health_waits_through_transition_and_incomplete_configs(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "health.txt"
            attempt = 0
            def query(command):
                nonlocal attempt
                if command == "version":
                    attempt += 1
                    if attempt == 1:
                        raise subprocess.CalledProcessError(1, ["rcon"])
                return "TF2 MvM Archipelago\nTF2 MvM wave probe\nconfigsready=" + ("1" if attempt >= 3 else "0")
            with mock.patch.object(runner.time, "monotonic", side_effect=[0, 1, 2]), mock.patch.object(runner.time, "sleep"):
                result = runner.wait_for_health(query, output, 10)
            self.assertIn("configsready=1", result)
            self.assertEqual(output.read_text(), result)

    def test_final_health_timeout_retains_unhealthy_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "health.txt"
            with mock.patch.object(runner.time, "monotonic", side_effect=[0, 11]), mock.patch.object(runner.time, "sleep"):
                with self.assertRaisesRegex(RuntimeError, "final mission"):
                    runner.wait_for_health(lambda _: "Unknown command", output, 10)
            self.assertIn("Unknown command", output.read_text())

    def setUp(self):
        self.meta = {"commit": "abc", "startup": True, "fresh": True,
                     "completed": True, "robot_limiter": True, "crash_free": True, "speed": 50, "installed_version": "123"}
        self.summary = {"eligible_cases": 100, "passed_cases": 50, "tested_cases": 50}

    def test_interrupted_run_cannot_score_100_percent(self):
        self.meta["completed"] = False
        result = report.evaluate(self.meta, self.summary, "abc", "123")
        self.assertEqual(result["score"], 50)
        self.assertEqual(result["status"], "incomplete")

    def test_unverified_inputs_have_no_numeric_score(self):
        for change in ({"startup": False}, {"fresh": False}, {"commit": "other"}, {"robot_limiter": False}):
            with self.subTest(change=change):
                result = report.evaluate({**self.meta, **change}, self.summary, "abc", "123")
                self.assertIsNone(result["score"])
        self.assertIsNone(report.evaluate(self.meta, self.summary, "abc", "124")["score"])

    def test_crash_retains_percentage_without_a_passing_verdict(self):
        result = report.evaluate({**self.meta, "crash_free": False}, self.summary, "abc", "123")
        self.assertEqual(result["score"], 50)
        self.assertEqual(result["status"], "engine crash recorded")

    def test_all_waves_passing_after_crash_still_fails_report_and_keeps_date(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source, output = root/"source", root/"output"
            for environment in report.ENVIRONMENTS:
                folder = source/environment
                folder.mkdir(parents=True)
                (folder/"environment.json").write_text(json.dumps({**self.meta, "crash_free": False}))
                (folder/"SUMMARY.json").write_text(json.dumps({**self.summary, "passed_cases": 100, "tested_cases": 100}))
            self.assertFalse(report.main(source, output, "abc", "https://example.com/run", measured_utc="2026-10-24T14:10:00+00:00", latest="123"))
            badge = json.loads((output/"ubuntu24.json").read_text())
            self.assertIn("100.00%", badge["message"])
            self.assertIn("engine crash recorded", badge["message"])
            self.assertEqual(badge["color"], "red")
            self.assertEqual(json.loads((output/"stability.json").read_text())["measured_utc"], "2026-10-24T14:10:00+00:00")

    def test_bad_or_empty_plan_cannot_pass(self):
        for summary in ({}, {"eligible_cases": 1, "passed_cases": 2, "tested_cases": 1}):
            self.assertIsNone(report.evaluate(self.meta, summary, "abc")["score"])

    def test_balanced_missions_stay_whole_and_are_order_independent(self):
        plan = [{"mission": name, "state": "planned", "wave": wave, "mode": "normal"}
                for name, count in (("a", 12), ("b", 10), ("c", 8), ("d", 6), ("e", 4))
                for wave in range(count)]
        manifest = partitioner.partition(plan, 2)
        self.assertEqual(manifest, partitioner.partition(list(reversed(plan)), 2))
        self.assertEqual(sum(manifest["planned_cases"]), 40)
        self.assertLessEqual(max(manifest["planned_cases"])-min(manifest["planned_cases"]), 4)
        self.assertEqual(set(manifest["missions"]), {"a", "b", "c", "d", "e"})
        with self.assertRaises(ValueError):
            partitioner.assignment(manifest, "missing", 2)
        with self.assertRaises(ValueError):
            partitioner.assignment(manifest, "a", 3)

    def test_local_worker_limit_respects_memory_disk_and_host_reserve(self):
        self.assertEqual(local.capacity(32, 58, 100, 16), 12)
        self.assertEqual(local.capacity(32, 58, 100, 16, reserve_cores=0), 16)
        self.assertEqual(local.capacity(32, 14, 100, 16), 2)
        self.assertEqual(local.capacity(32, 58, 34, 16), 2)
        self.assertEqual(local.capacity(32, 7, 100, 16), 0)
        self.assertEqual(local.capacity(32, 58, 29, 16), 0)

    def test_native_rcon_uses_local_interfaces_and_docker_uses_only_published_loopback(self):
        with mock.patch.object(runner.socket, "getaddrinfo", return_value=[(None,None,None,None,("172.20.144.1",0))]), \
                mock.patch.object(runner.subprocess, "run", return_value=SimpleNamespace(stdout="172.20.144.1 10.0.0.2 ::1")) as call:
            self.assertEqual(runner.rcon_hosts("ubuntu24"), ["127.0.0.1", "172.20.144.1", "10.0.0.2"])
            call.reset_mock()
            self.assertEqual(runner.rcon_hosts("docker"), ["127.0.0.1"])
            call.assert_not_called()
        with mock.patch.object(runner.os, "name", "nt"), \
                mock.patch.object(runner.socket, "getaddrinfo", return_value=[]), \
                mock.patch.object(runner.subprocess, "run", return_value=SimpleNamespace(stdout='["127.0.0.1","172.20.144.1"]')):
            self.assertEqual(runner.rcon_hosts("windows"), ["127.0.0.1", "172.20.144.1"])

    def test_windows_interface_timeout_retains_readiness_candidates(self):
        with mock.patch.object(runner.os, "name", "nt"), \
                mock.patch.object(runner.socket, "getaddrinfo", return_value=[(None,None,None,None,("172.27.48.1",0))]), \
                mock.patch.object(runner.subprocess, "run", side_effect=runner.subprocess.TimeoutExpired("powershell", 60)), \
                mock.patch("builtins.print"):
            self.assertEqual(runner.rcon_hosts("windows"), ["127.0.0.1", "172.27.48.1"])

    def test_failed_native_discovery_retains_loopback(self):
        with mock.patch.object(runner.socket, "getaddrinfo", side_effect=runner.socket.gaierror("unavailable")), \
                mock.patch.object(runner.subprocess, "run", side_effect=OSError("unavailable")), \
                mock.patch("builtins.print"):
            self.assertEqual(runner.rcon_hosts("ubuntu24"), ["127.0.0.1"])

    def test_host_cleanup_refuses_a_personal_machine(self):
        with mock.patch.dict(os.environ, {"STABILITY_RUNNER_ENVIRONMENT": "self-hosted"}):
            with self.assertRaisesRegex(RuntimeError, "only on an ephemeral"):
                host.main()

    @unittest.skipIf(os.name == "nt", "Linux package setup fixture")
    def test_host_with_sufficient_space_keeps_sdks_and_installs_runtime_libraries(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp)
            with mock.patch.dict(os.environ,{"STABILITY_RUNNER_ENVIRONMENT":"github-hosted",
                    "STABILITY_ENV":"docker","RUNNER_TEMP":str(root/"game"),"GITHUB_ENV":str(root/"env")}), \
                    mock.patch.object(host.shutil,"disk_usage",return_value=SimpleNamespace(free=100*1024**3)), \
                    mock.patch.object(host.subprocess,"run") as run:
                host.main()
            self.assertFalse(any("rm" in call.args[0] for call in run.call_args_list))
            self.assertTrue(any("libcurl4t64:i386" in call.args[0] for call in run.call_args_list))
            self.assertIn("STABILITY_GAME_PARENT=",(root/"env").read_text())

    @unittest.skipIf(os.name == "nt", "Linux package setup fixture")
    def test_failed_dependency_install_cannot_export_a_ready_game_root(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            def install_fails(command, **kwargs):
                if "install" in command:
                    raise host.subprocess.CalledProcessError(124, command)
            with mock.patch.dict(os.environ, {"STABILITY_RUNNER_ENVIRONMENT": "github-hosted",
                    "STABILITY_ENV": "ubuntu24", "RUNNER_TEMP": str(root/"game"), "GITHUB_ENV": str(root/"env")}), \
                    mock.patch.object(host.shutil, "disk_usage", return_value=SimpleNamespace(free=100*1024**3)), \
                    mock.patch.object(host.subprocess, "run", side_effect=install_fails):
                with self.assertRaises(host.subprocess.CalledProcessError):
                    host.main()
            self.assertFalse((root/"env").exists())

    def test_badge_update_is_one_commit_with_all_environments(self):
        with tempfile.TemporaryDirectory() as temp:
            directory = Path(temp)
            for name in report.ENVIRONMENTS:
                (directory / f"{name}.json").write_text(json.dumps({"schemaVersion": 1, "message": "unverified"}))
            replies = [{"object": {"sha": "old"}}, {"sha": "tree"}, {"sha": "new"}, {}]
            with mock.patch.object(publisher, "api", side_effect=replies) as api:
                publisher.publish(directory, "kelly-cs/tf2-archipelago")
            tree = api.call_args_list[1].args[3]["tree"]
            self.assertEqual({file["path"] for file in tree}, {"ubuntu24.json", "docker.json", "windows.json"})
            self.assertEqual(api.call_args_list[2].args[3]["parents"], ["old"])
            self.assertEqual(api.call_args_list[3].args[3], {"sha": "new", "force": False})

    def test_missing_badge_does_not_partially_publish(self):
        with tempfile.TemporaryDirectory() as temp, mock.patch.object(publisher, "api") as api:
            with self.assertRaises(FileNotFoundError):
                publisher.publish(Path(temp), "kelly-cs/tf2-archipelago")
            api.assert_not_called()

    def test_missing_hosted_shard_stays_in_full_plan(self):
        with tempfile.TemporaryDirectory() as temp:
            base = Path(temp)
            source, plan = base / "evidence", base / "plan.jsonl"
            missions = {index: next(f"mission{n}" for n in range(100)
                                    if zlib.crc32(f"mission{n}".encode()) % 2 == index)
                        for index in range(2)}
            rows = [{"mission": mission, "mode": "normal", "wave": 1,
                     "map": "mvm_decoy", "state": "planned"} for mission in missions.values()]
            plan.write_text("".join(json.dumps(row) + "\n" for row in rows))
            for env in report.ENVIRONMENTS:
                for shard in range(2):
                    if env == "windows" and shard == 1:
                        continue
                    path = source / env / f"part-{shard}"
                    path.mkdir(parents=True)
                    row = next(row for row in rows if row["mission"] == missions[shard])
                    (path / "environment.json").write_text(json.dumps({
                        **self.meta, "shard": shard, "shards": 2, "running_version": "123"}))
                    (path / "plan.jsonl").write_text(json.dumps(row) + "\n")
                    (path / "shard-0.jsonl").write_text(json.dumps({**row, "state": "passed"}) + "\n")
            report.combine_shards(source, plan, 2, "abc")
            complete = json.loads((source / "ubuntu24/SUMMARY.json").read_text())
            partial = json.loads((source / "windows/SUMMARY.json").read_text())
            self.assertEqual(complete["passed_cases"], complete["eligible_cases"])
            self.assertEqual(partial["passed_cases"], 1)
            self.assertEqual(partial["eligible_cases"], 2)
            self.assertEqual(partial["outcomes"]["not run"], 1)
            meta = json.loads((source / "windows/environment.json").read_text())
            self.assertFalse(meta["completed"])
            self.assertIn("shard 1: missing result", meta["error"])
            self.assertIsNone(report.evaluate(meta, partial, "abc", "123")["score"])

    def test_missing_partition_never_falls_back_to_hash_or_stale_passes(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            plan = root / "plan.jsonl"
            plan.write_text(json.dumps({"mission":"fixture", "wave":1, "mode":"normal", "map":"mvm_decoy", "state":"planned"}) + "\n")
            report.combine_shards(root/"evidence", plan, 2, "abc", root/"missing.json")
            meta = json.loads((root/"evidence/docker/environment.json").read_text())
            self.assertFalse(meta["fresh"])
            self.assertIn("invalid canonical partition", meta["error"])
            summary = json.loads((root/"evidence/docker/SUMMARY.json").read_text())
            self.assertEqual(summary["eligible_cases"], 1)
            self.assertEqual(summary["passed_cases"], 0)

    def test_single_shard_smoke_aggregates_without_claiming_full_catalog(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            plan = root/"plan.jsonl"
            row = {"mission":"fixture","wave":1,"mode":"normal","map":"mvm_decoy","state":"planned"}
            plan.write_text(json.dumps(row)+"\n")
            manifest=root/"partition.json"
            manifest.write_text(json.dumps(partitioner.partition([row],1)))
            for environment in report.ENVIRONMENTS:
                out=root/"input"/environment/"part-0"
                out.mkdir(parents=True)
                (out/"plan.jsonl").write_text(plan.read_text())
                (out/"environment.json").write_text(json.dumps({**self.meta,"shard":0,"shards":1,"running_version":"123"}))
                (out/"shard-0.jsonl").write_text(json.dumps({**row,"state":"passed"})+"\n")
            report.combine_shards(root/"input",plan,1,"abc",manifest)
            self.assertTrue(report.main(root/"input",root/"report","abc","https://example.com",scope="smoke"))
            data=json.loads((root/"report/stability.json").read_text())
            self.assertEqual(data["scope"],"smoke")
            self.assertEqual(data["environments"]["docker"]["passed"],1)
            self.assertIn("not full-catalog stability",(root/"report/REPORT.md").read_text())

    def test_no_results_are_rendered_and_archived_as_failures(self):
        with tempfile.TemporaryDirectory() as temp:
            source, output = Path(temp) / "input", Path(temp) / "output"
            source.mkdir()
            self.assertFalse(report.main(source, output, "abc", "https://example.com/run"))
            result = json.loads((output / "stability.json").read_text())
            self.assertEqual(len(result["environments"]), 3)
            self.assertTrue(all(row["status"] == "no result" for row in result["environments"].values()))
            self.assertIn("unverified", (output / "REPORT.md").read_text())
            with zipfile.ZipFile(output / "stability-evidence.zip") as archive:
                self.assertIn("REPORT.md", archive.namelist())

    def test_full_results_preserve_reverse_exclusion_and_evidence(self):
        with tempfile.TemporaryDirectory() as temp:
            source, output = Path(temp) / "input", Path(temp) / "output"
            for key in report.ENVIRONMENTS:
                path = source / key
                path.mkdir(parents=True)
                (path / "environment.json").write_text(json.dumps(self.meta))
                (path / "SUMMARY.json").write_text(json.dumps({
                    "eligible_cases": 10, "passed_cases": 10, "tested_cases": 10,
                    "outcomes": {"reverse objective": 2}}))
                (path / "REPORT.md").write_text("Detailed wave evidence")
            self.assertTrue(report.main(source, output, "abc", "https://example.com/run"))
            result = json.loads((output / "stability.json").read_text())
            self.assertEqual(result["environments"]["windows"]["unsupported_reverse"], 2)
            badge = json.loads((output / "windows.json").read_text())
            self.assertEqual(badge["color"], "brightgreen")
            with zipfile.ZipFile(output / "stability-evidence.zip") as archive:
                self.assertEqual(archive.read("windows/REPORT.md"), b"Detailed wave evidence")

    def test_publication_rechecks_build_when_an_earlier_environment_finished(self):
        with tempfile.TemporaryDirectory() as temp:
            source, output = Path(temp) / "input", Path(temp) / "output"
            source.mkdir()
            for key in report.ENVIRONMENTS:
                path = source / key
                path.mkdir()
                (path / "environment.json").write_text(json.dumps(self.meta))
                (path / "SUMMARY.json").write_text(json.dumps({
                    "eligible_cases": 1, "passed_cases": 1, "tested_cases": 1}))
            fake_runner = mock.Mock()
            fake_runner.live_version.return_value = "124"
            with mock.patch.object(report.importlib.util, "module_from_spec", return_value=fake_runner), \
                    mock.patch.object(report.importlib.util, "spec_from_file_location") as spec:
                self.assertFalse(report.main(source, output, "abc", "https://example.com", check_live=True))
                spec.return_value.loader.exec_module.assert_called_once_with(fake_runner)
            data = json.loads((output / "stability.json").read_text())
            self.assertTrue(all(row["status"] == "latest build unverified"
                                for row in data["environments"].values()))

    def test_mod_load_failure_is_not_server_readiness(self):
        good = "TF2 MvM Archipelago\nTF2 MvM wave probe\nprobe status"
        self.assertTrue(runner.startup_ready(good))
        for error in ("<Failed>", "<Error>", "<Bad Load>", "failed to load", "Unknown command"):
            self.assertFalse(runner.startup_ready(good + "\n" + error))
        self.assertFalse(runner.startup_ready("TF2 MvM wave probe"))

    def test_crash_followed_by_successful_restart_is_still_a_failure(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "server.log").write_text("Segmentation fault (core dumped)\nServer restart in 10 seconds\nready\n")
            (root / "console.log").write_text("healthy probe\n")
            self.assertEqual(len(runner.crash_evidence(root)), 1)
        complete = {"eligible_cases": 1, "passed_cases": 1, "tested_cases": 1}
        result = report.evaluate({**self.meta, "crash_free": False}, complete, "abc")
        self.assertEqual(result["score"], 100)
        self.assertEqual(result["status"], "engine crash recorded")

    def test_limiter_must_be_enabled_and_readable(self):
        self.assertTrue(runner.robot_limiter_enabled('"sig_mvm_robot_limit" = "1" ( def. "0" )'))
        for reply in ('"sig_mvm_robot_limit" = "0" ( def. "1" )',
                      'Unknown command "sig_mvm_robot_limit"',
                      '"tf_mvm_max_invaders" = "1"', ''):
            self.assertFalse(runner.robot_limiter_enabled(reply))
        with tempfile.TemporaryDirectory() as temp:
            evidence = Path(temp) / "limiter.txt"
            with self.assertRaisesRegex(RuntimeError, "limiter is disabled"):
                runner.check_robot_limiter(lambda _: '"sig_mvm_robot_limit" = "0"', evidence)
            self.assertIn('"0"', evidence.read_text())
        self.meta.pop("robot_limiter")
        self.assertIsNone(report.evaluate(self.meta, self.summary, "abc")["score"])

    def test_reads_server_version_instead_of_steam_depot_buildid(self):
        with tempfile.TemporaryDirectory() as temp:
            game = Path(temp)
            (game / "tf").mkdir()
            (game / "tf/steam.inf").write_bytes(b"PatchVersion=123\r\nServerVersion=999\r\n")
            self.assertEqual(runner.installed_version(game), "999")

    def test_live_rcon_server_version_does_not_require_a_breakpad_console_line(self):
        reply = "Network PatchVersion: 111\nProtocol version: 24\nServer version:       999\n"
        self.assertEqual(runner.running_version(Path("/no-game-needed"), reply), "999")

    @unittest.skipIf(os.name == "nt", "POSIX executable fixtures; real Windows uses .exe beds")
    def test_runner_executes_plan_install_start_probe_retry_and_report(self):
        # Exercise process orchestration and the real report/retry scripts with
        # an inert fixture in place of SteamCMD, SRCDS and Docker.
        with tempfile.TemporaryDirectory() as temp:
            base = Path(temp)
            bed = base / "bed"
            bed.mkdir()
            script = "#!" + sys.executable + "\n" + textwrap.dedent('''\
                import json, pathlib, sys, time
                name = pathlib.Path(sys.argv[0]).name
                args = sys.argv[1:]
                if name == 'winbed':
                    root = pathlib.Path(args[args.index('-root') + 1])
                    (root / 'tf-dedicated/tf').mkdir(parents=True)
                    (root / 'tf-dedicated/tf/steam.inf').write_text('ServerVersion=123\\n')
                    (root / 'tf-dedicated/tf/console.log').write_text('Using Breakpad minidump system. Version: 123\\n')
                elif name == 'rcon' and args == ['sig_mvm_robot_limit']:
                    print('\"sig_mvm_robot_limit\" = \"1\" ( def. \"0\" )')
                elif name == 'rcon':
                    print('TF2 MvM Archipelago\\nTF2 MvM wave probe\\nstatus: idle configsready=1')
                elif name == 'docker':
                    print(json.dumps(args), flush=True)
                    if args[0] == 'run' and '--entrypoint' not in args:
                        time.sleep(300)
                else:
                    with pathlib.Path(__file__).with_name('probe-args.jsonl').open('a') as f:
                        f.write(json.dumps(args) + '\\n')
                    for mode in ('normal', 'surge'):
                        print(json.dumps({'mission': 'fixture', 'map': 'mvm_decoy',
                            'wave': 1, 'mode': mode,
                            'state': 'planned' if '-plan' in args else 'passed'}), flush=True)
            ''')
            for name in ("winbed", "waveprobe", "rcon", "docker"):
                (bed / name).write_text(script)
                (bed / name).chmod(0o755)
            (bed / "tf2_waveprobe.smx").touch()
            output = base / "results"
            opt = SimpleNamespace(output=output, bed=bed, root=base / "game", hours=0.02,
                                  environment="docker", image="fixture")
            with mock.patch.dict(os.environ, {"PATH": str(bed) + os.pathsep + os.environ["PATH"],
                                              "WAVEPROBE_RCONPW": "fixture", "GITHUB_SHA": "abc"}), \
                    mock.patch.object(runner, "live_version", return_value="123"):
                self.assertEqual(runner.main(opt), 0)
            meta = json.loads((output / "environment.json").read_text())
            summary = json.loads((output / "SUMMARY.json").read_text())
            self.assertTrue(meta["startup"] and meta["fresh"] and meta["completed"])
            self.assertEqual(summary["passed_cases"], 2)
            self.assertEqual(summary["tested_cases"], summary["eligible_cases"])
            self.assertIn("Retesting 0", (output / "retest.log").read_text())
            self.assertIn("SRCDS_MOD_LOADING=always", (output / "server.log").read_text())
            self.assertIn("speed=50", (output / "config.txt").read_text())
            self.assertIn("reverse_bombs=1", (output / "config.txt").read_text())
            calls = [json.loads(line) for line in (bed / "probe-args.jsonl").read_text().splitlines()]
            self.assertTrue(all("-reverse-objectives" in args for args in calls))
            self.assertTrue((output / "startup.txt").exists())
            self.assertTrue((output / "final-health.txt").exists())


if __name__ == "__main__":
    unittest.main()
