# Architecture

Go CLI + daemon. Single static binary, no runtime deps. Module is still
`github.com/AlejandroPerez92/nanoctl` (upstream's path — don't rename, it would
break the ldflags targets in the Makefile).

```
cmd/                cobra commands: fan, poweron/poweroff/reset, info, service,
                    check_prometheus, version
pkg/fan/            the control loop
  monitor.go        RunMonitor: read temp -> PID -> duty. Owns the fail-safes.
  pwm_controller.go controller interface + newPWMController + applyDutyFloor
  hardware_pwm.go   sysfs /sys/class/pwm/<chip>/pwm<n>  (the one we use)
  pwm.go            software bit-banged GPIO PWM (unusable at real fan
                    frequencies — see gotchas)
pkg/temperature/    Source interface: GetTemperature() (float64, error)
  file.go           reads a thermal zone file (millidegrees)
  prometheus.go     queries a Prometheus server — the "cluster aware" feature
  fallback.go       FallbackSource: primary, degrading to secondary, RETRYING
pkg/config/         FanConfig + defaults + validation; default.yaml is embedded
pkg/gpio/           slot power control (poweron/poweroff/reset)
pkg/metrics/        OTLP push (temperature + duty cycle)
```

## Control flow (`nanoctl fan`)

`cmd/fan.go` loads `/etc/nanoctl/fan.yaml`, builds a `temperature.Source`, and
calls `fan.RunMonitor`. Each tick: read temperature, `pid.Update`, clamp through
`applyDutyFloor`, `SetDutyCycle`. On read failure the fan is driven to
`failure_duty`; on exit it is left running, not stopped.

Two things deliberately live outside the binary and belong to the consumer
(nanocluster's `cluster fan-setup`): taking the PWM channel away from the kernel
driver, and the systemd unit. nanoctl assumes the channel is already free.

## How it ships

`release.yml` on a `v*` tag, `prerelease.yml` on manual dispatch. Both build via
`make build-arm64` (never a bare `go build` — see gotchas) and gate on
`go vet ./...` + `go test ./...`. The consumer downloads
`nanoctl-linux-arm64` from a release by tag; it never builds locally.
