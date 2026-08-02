# CLAUDE.md

Fork of [AlejandroPerez92/nanoctl](https://github.com/AlejandroPerez92/nanoctl),
maintained for the Sipeed NanoCluster in
[silvapedro-org/nanocluster](https://github.com/silvapedro-org/nanocluster).

`AGENTS.md` (upstream's) is the Go style guide — imports, naming, error
handling, Cobra patterns. Follow it; do not duplicate it here.

These files carry what AGENTS.md cannot: why this fork exists, the hardware
behaviour that is not guessable from the code, and what has changed.

@docs/context/architecture.md
@docs/context/conventions.md
@docs/context/gotchas.md
@docs/context/state.md

## Working rules

- **Upstream is MIT** (declared in its README; there is no LICENSE file).
  Changes here should be written so they can be offered upstream: no
  nanocluster-specific hardcoding, board specifics belong in config.
- **Never paste in GPL-3.0 code** (e.g. from meteyou/sipeed-nanocluster-server).
  Reimplementing behaviour is fine — ideas are not copyrightable — but pasting
  would force GPL-3 on this MIT work.
- **This controls cooling.** Every change is judged on what happens when it
  fails: a controller that cannot measure must never conclude that cooling is
  unnecessary. See the fail-safe entries in gotchas.
- **Test on hardware before claiming done.** Every bug worth finding here was
  found by watching a temperature curve, not by reading code.
