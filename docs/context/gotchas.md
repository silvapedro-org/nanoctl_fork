# Gotchas (measured on hardware — trust them)

Target hardware: Sipeed NanoCluster, slot 1 = LPi3H/LM3H (Allwinner), fan is a
**Noctua NF-A6x25 5V PWM retrofit**, not the stock 2-pin fan.

## Fan / PWM

- **The fan's control sense is INVERTED, and it is not a bug in this code.**
  Measured on the exported channel: `duty_cycle 0` = fan **FULL**,
  `duty_cycle = period` = fan **OFF**. So `inverted: true` is correct for this
  wiring, and `SetDutyCycle(0)` genuinely means off.
  The cause is the retrofit: yellow to header positive, black to a GPIO ground,
  and **blue (the fan's active-high PWM input) to the header's negative**, so the
  header's low-side switch drives the PWM input. More duty = less fan. With the
  stock 2-pin fan the same board behaves normally — so **never "fix" this by
  patching the device tree polarity**: that encodes a wiring mod as a board fact.
- **`cooling_device0/cur_state` is DERIVED from `pwm1`.** Reading it to compute a
  target duty and writing that back is a feedback oscillator — live-verified,
  the duty flapped 255 -> 1 -> 100 -> 1 between polls. Drive from **temperature**,
  never from the cooling device's own state.
- **The kernel's `pwm-fan` driver owns the channel.** While bound, sysfs
  `export` fails EBUSY. It must be unbound
  (`echo pwm-fan > /sys/bus/platform/drivers/pwm-fan/unbind`) before nanoctl can
  take the channel. Unbinding drops the 60/70 C passive trips but **leaves the
  110 C critical trip armed** (verified), so the board still protects itself.
  nanoctl does not do this — the consumer's unit does it in `ExecStartPre`.
- **Set the channel's `polarity` explicitly.** It is inherited from whatever
  configured the channel last; relying on the default would silently invert the
  fan. `polarity` is only writable while the channel is disabled.
- **`thermal_zone0` is board-specific.** On the LPi3H it is the **GPU**; the CPU
  is `thermal_zone2`. The shipped default points at zone0, which on this board
  regulates against the wrong sensor while looking perfectly healthy.
- **Software PWM mode cannot work at its own default frequency.** It bit-bangs a
  GPIO from Go with `time.Sleep`, and the default config pairs `mode: software`
  with `frequency_khz: 25`. Userspace Go cannot produce a 25 kHz waveform. Use
  `mode: hardware`.
- **Not every 4-pin fan stalls at low duty.** The assumption behind `min_duty`
  is that fans stop below ~30%. This Noctua still moves air at 10% (measured:
  +6.7 C in 40s at 10% versus +10.7 C with the fan off, same start). Set
  `min_duty` from measurement, not folklore.
- **PID gains are negated at runtime.** Config validation demands positive
  `kp/ki/kd`, then the loop calls `pid.SetPID(-Kp, -Ki, -Kd)` so rising
  temperature raises duty. Nothing in the config says so. (`pidctrl.SetPID` only
  assigns the constants and does not reset accumulated state, so calling it every
  tick is redundant but harmless.)

## Systemd / packaging

- **`StartLimitIntervalSec` is only honoured in `[Unit]`.** Under `[Service]`
  systemd **silently ignores it** and the unit file still looks correct — check
  the effective value with `systemctl show <unit> -p StartLimitIntervalUSec`
  (note: the property is `...USec`, not `...Sec`). Without it, five quick
  failures make systemd give up permanently and leave the machine with no fan
  control and nothing watching.

## CI

- **Never execute the built binary on the runner.** It is an arm64
  cross-build and the runner is x86_64: `cannot execute binary file: Exec format
  error`, exit 126, whole workflow dead. Use `go version -m <binary>` to read the
  embedded build info instead — it also proves the ldflags landed, which running
  it does not.
- **Build via `make`, and check out with `fetch-depth: 0`.** A bare `go build`
  drops the ldflags and a shallow clone has no tags for `git describe`. Either
  mistake produces a binary that reports itself as `dev`, silently.

## Licence

- **MIT, declared only in upstream's README** — there is no LICENSE file, so
  GitHub does not detect it and there is no copyright notice to carry.
  Redistributing builds is fine. The GPL direction is one-way: MIT may go into a
  GPL-3 work, but pasting GPL-3 code in here would force GPL-3 on the result.
