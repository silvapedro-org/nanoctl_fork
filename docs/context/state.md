# Current state (volatile — update when it changes)

Last updated: 2026-08-02

## Why this fork exists

The NanoCluster's slot 1 runs an LPi3H whose fan wiring inverts the PWM sense,
so the kernel's own thermal governor drives the fan **backwards** — it raises
duty as the board heats, which switches the fan off, which heats it further. The
board hit the 110 C critical trip and hard-shut-down twice before that was
understood. nanoctl replaces the kernel's loop entirely: it drives
`/sys/class/pwm` directly and has an `inverted` flag, which nothing else did.

## Changes on `my_main` (all committed, none merged upstream)

- `feat(fan)`: `pwm.min_duty` + `pwm.off_below` — a floor for fans that stall.
- `fix(fan)`: fail safe when the temperature cannot be read — after
  `failure_after` failures the fan goes to `failure_duty` (default 100%).
  Previously a failed read only logged, freezing the fan at its last duty.
- `fix(fan)`: leave the fan running when the daemon exits. `defer
  controller.Stop()` wrote `enable=0`, so `systemctl stop` stopped the **fan**.
- `fix(service)`: `StartLimitIntervalSec=0` in the generated unit, in `[Unit]`.
- `fix(temperature)`: `FallbackSource` — the source choice used to be made once
  at startup, so a Prometheus that was merely slow to start condemned the daemon
  to the local sensor forever. Now retries every 30s.
- `fix(info)`: `c.Lines` -> `c.Lines()`; this was the only thing keeping
  `go vet ./...` red.
- `ci`: build via the Makefile, `fetch-depth: 0`, vet+test gates, and the
  workflows split — `release.yml` (v* tag) and `prerelease.yml` (manual,
  publishes a prerelease so it can be fetched by URL like a real release).

All of the fan fixes are upstream-worthy and none are NanoCluster-specific.
Offering them is a good moment to ask for a LICENSE file.

## Deployed

`prerelease-1` (commit d19b9ff) runs on the NanoCluster orchestrator, installed
by nanocluster's `cluster fan-setup --prerelease`. Holding 55 C at ~30% duty,
survives reboots. Config: `mode: hardware`, `channel: 2`, `inverted: true`,
`frequency_khz: 20`, sensor `thermal_zone2`, `min_duty: 0`.

## Known gaps

- **Config is read once at startup** — no reload, no API, so every tuning change
  needs a service restart. The values you actually tune (target, gains,
  min_duty, off_below) can all be swapped into the running loop; only
  mode/chip/channel/frequency/inverted need the controller rebuilt. A SIGHUP
  reload is the obvious next change.
- **No slot-1 sensor in the Prometheus path.** The orchestrator is outside the
  k3s cluster, so `max(node_hwmon_temp_celsius)` covers only the CM5s unless
  node-exporter runs standalone on slot 1 and is scraped as an external target —
  and slot 1 is often the hottest board.
- **Nothing here helps if the board loses power.** With slot 1 unpowered the fan
  stops (measured), so the cluster has no airflow. That is a hardware fix — a
  pull-up on the fan's PWM line — not a software one.
