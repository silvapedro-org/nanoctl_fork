# Conventions

Go style: see `AGENTS.md`. What follows is fork-specific.

## Changes should be upstreamable
Written so they could be offered to upstream: no NanoCluster hardcoding, no
LPi3H assumptions in code. Board specifics (PWM channel, polarity, frequency,
sensor path) are **config**, defaulted so existing users are unaffected — e.g.
`min_duty: 0` disables the stall floor entirely.

## Safety is the acceptance criterion
Every change to the control loop is judged by its failure mode, not its happy
path. The rule: **a controller that cannot measure must never conclude that
cooling is unnecessary.** Concretely — unreadable temperature drives the fan to
`failure_duty`, exiting leaves the fan running, and a crash restarts forever.

## Build and CI
- Build with `make build-arm64`, never a bare `go build`: the Makefile injects
  Version/Commit/BuildDate via ldflags.
- CI checks out with `fetch-depth: 0` — the Makefile reads `git describe --tags`.
- `go vet ./...` and `go test ./...` gate every build. Keep vet green: it was
  red for a long time on one bug, which made it useless as a gate.
- Never execute the built binary in CI (cross-compiled arm64, x86 runner).

## Testing
Upstream had no tests. Ours cover the pure logic that is cheap to get wrong and
expensive to debug on hardware: `applyDutyFloor` (the stall band) and
`FallbackSource` (degrade + recover). Hardware behaviour is verified by watching
a temperature curve on a real board, and the result recorded in gotchas.

## Commits
Conventional prefixes (`fix(fan):`, `ci:`, `docs:`). The body says what the
failure looked like, not just what changed — most of these bugs are invisible
until a specific hardware state occurs.
