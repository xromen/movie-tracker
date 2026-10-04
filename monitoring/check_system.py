"""Check provisioned system queries: python monitoring/check_system.py [promtool]."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile


root = Path(__file__).resolve().parent
dashboard = json.loads((root / "grafana/dashboards/system.json").read_text(encoding="utf-8"))
groups = json.loads((root / "grafana/provisioning/alerting/system.json").read_text(encoding="utf-8"))["groups"]
rules = groups[0]["rules"]
instance = "host.docker.internal:9100"
promtool = sys.argv[1] if len(sys.argv) > 1 else "promtool"

assert len({panel["id"] for panel in dashboard["panels"]}) == len(dashboard["panels"])
assert len({rule["uid"] for rule in rules}) == len(rules)
for rule in rules:
    assert rule["noDataState"] == "OK"  # Exporter availability is monitored separately.
    assert rule["notification_settings"]["receiver"] == "Movie Tracker Telegram"
    assert rule["annotations"]["__dashboardUid__"] == dashboard["uid"]
    assert int(rule["annotations"]["__panelId__"]) in {panel["id"] for panel in dashboard["panels"]}


def fixtures(memory, available, free_inodes, busy=False):
    series = []

    def add(name, values, **labels):
        labels = {"job": "node", "instance": instance, **labels}
        selector = ",".join(f'{key}="{value}"' for key, value in labels.items())
        series.append({"series": f"{name}{{{selector}}}", "values": values})

    for name, value in [("MemTotal", 1000), ("MemAvailable", memory), ("MemFree", 1)]:
        add(f"node_memory_{name}_bytes", f"{value}+0x30")
    for cpu in ("0", "1"):
        add("node_cpu_seconds_total", "0+3x30" if busy else "0+54x30", cpu=cpu, mode="idle")
        add("node_cpu_seconds_total", "0+18x30" if busy else "0+0x30", cpu=cpu, mode="iowait")
    add("node_load15", "4+0x30" if busy else "1+0x30")
    add("node_vmstat_oom_kill", "0+0x19 1+0x10" if busy else "0+0x30")
    # Full temporary/overlay and read-only mounts must never produce low-space alerts.
    for mount, fstype, readonly in [("/", "ext4", 0), ("/run", "tmpfs", 0),
                                     ("/var/lib/docker/test", "overlay", 0), ("/snapshot", "ext4", 1),
                                     ("/zero", "ext4", 0)]:
        labels = {"device": "/dev/vda3", "mountpoint": mount, "fstype": fstype}
        for name, value in [("size_bytes", 0 if mount == "/zero" else 1000),
                            ("avail_bytes", available if mount == "/" else 1),
                            ("files", 0 if mount == "/zero" else 1000),
                            ("files_free", free_inodes if mount == "/" else 1),
                            ("readonly", readonly)]:
            add("node_filesystem_" + name, f"{value}+0x30", **labels)
    return series


def expected(rule, enabled):
    if not enabled:
        return []
    labels = {"instance": instance, "severity": rule["labels"]["severity"]}
    suffix = rule["uid"].removeprefix("movie_tracker_system_")
    if suffix.startswith(("memory_", "disk_", "inodes_")) or suffix == "oom":
        labels["job"] = "node"
    if suffix.startswith(("disk_", "inodes_")):
        labels.update({"device": "/dev/vda3", "mountpoint": "/", "fstype": "ext4"})
    return [{"exp_labels": labels, "exp_annotations": {}}]


prom_rules = []
for rule in rules:
    threshold = rule["data"][1]["model"]["conditions"][0]["evaluator"]
    operator = {"lt": "<", "gt": ">"}[threshold["type"]]
    prom_rules.append({"alert": rule["uid"],
                       "expr": f'({rule["data"][0]["model"]["expr"]}) {operator} {threshold["params"][0]}',
                       "for": rule["for"], "labels": rule["labels"]})
for panel in dashboard["panels"]:
    for target in panel["targets"]:
        prom_rules.append({"record": f'system_panel_{panel["id"]}_{target["refId"]}', "expr": target["expr"]})

tests = []
for name, series, enabled in [
    ("healthy with low MemFree and full excluded mounts", fixtures(200, 300, 500), lambda rule: False),
    ("warning thresholds only", fixtures(100, 150, 50), lambda rule: rule["uid"].endswith("_warning")),
    ("critical pressure and OOM", fixtures(40, 50, 20, True), lambda rule: True),
    ("exact warning boundaries stay healthy", fixtures(150, 200, 100), lambda rule: False),
    ("exact critical boundaries stay warning", fixtures(50, 100, 30), lambda rule: rule["uid"].endswith("_warning")),
]:
    tests.append({"name": name, "interval": "1m", "input_series": series,
                  "alert_rule_test": [{"eval_time": "20m", "alertname": rule["uid"],
                                       "exp_alerts": expected(rule, enabled(rule))} for rule in rules]})

# Check that short spikes have not passed the resource rules' pending periods.
tests.append({"name": "resource alerts remain pending for short spikes", "interval": "1m",
              "input_series": fixtures(40, 50, 20), "alert_rule_test": [
                  {"eval_time": "1m", "alertname": rule["uid"], "exp_alerts": []} for rule in rules]})

with tempfile.TemporaryDirectory(prefix="movie-tracker-system-") as directory:
    rules_path = Path(directory) / "rules.json"
    tests_path = Path(directory) / "tests.json"
    # JSON is also valid YAML, so promtool can read these without a YAML dependency.
    rules_path.write_text(json.dumps({"groups": [{"name": "system", "rules": prom_rules}]}), encoding="utf-8")
    tests_path.write_text(json.dumps({"rule_files": [rules_path.as_posix()], "evaluation_interval": "1m",
                                     "tests": tests}), encoding="utf-8")
    subprocess.run([promtool, "check", "rules", str(rules_path)], check=True)
    subprocess.run([promtool, "test", "rules", str(tests_path)], check=True)

print(f"System dashboard queries and {len(rules)} alerts: OK")
