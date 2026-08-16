# Heddle

Heddle reads a weaving draft and works out the cloth it makes: which warp ends are raised over
which picks, how long the floats run, what structure the result is, and whether anything in it
would fall apart on the loom. It is written in Go with the standard library only, builds
offline, and runs as a single static binary.

This is an independent study project. It is not affiliated with, sponsored by, or endorsed by
any company, guild, mill, or organisation, and it reads no file format anybody else uses. It
exists to make one point concrete: a draft determines its cloth completely, and almost every
wrong answer about a cloth comes from counting something along the piece that only makes sense
counted round the repeat.

Nothing here is a substitute for weaving software. The catalogue is twelve drafts, the float
limits are this tool's own, and the tool checks every draft it ships with rather than asking to
be trusted.

## Quick start

```
go build -o heddle ./cmd/heddle

./heddle help
./heddle drafts
./heddle cloth -draft twill-2-2
./heddle pattern -draft satin-5
./heddle quality -draft floating-end
./heddle satin -shafts 6
```

Or with Docker, which builds and tests offline and produces a `scratch` image:

```
docker build -t heddle .
docker run --rm --network none heddle drafts
docker run --rm --network none heddle cloth -file /examples/herringbone.draft
```

## What a draft is

Four things, of which the fourth follows from the other three.

| part          | what it says                                   |
| ------------- | ---------------------------------------------- |
| **threading** | which shaft each warp end passes through       |
| **tie-up**    | which shafts each treadle lifts                |
| **treadling** | which treadle is pressed for each pick of weft |
| **cloth**     | which ends are raised over which picks         |

An end is raised over a pick exactly when the treadle for that pick lifts the shaft that end is
threaded on. That single sentence is the whole of `cloth.Weave`, and everything else this tool
prints is read off the result.

## Commands

| command   | what it does                                       |
| --------- | -------------------------------------------------- |
| `draft`   | read a draft and report its four parts             |
| `cloth`   | weave a draft out and draw the cloth               |
| `floats`  | the floats of a cloth along the warp and the weft  |
| `pattern` | name the structure a cloth has                     |
| `quality` | read a cloth for the faults a weaver would find    |
| `satin`   | the steps a satin can use, and the satin they give |
| `tromp`   | treadle a draft the way it is threaded and compare |
| `drafts`  | the drafts this tool knows, with their checks      |
| `report`  | one figure per package for a draft                 |
| `version` | print the version                                  |
| `help`    | print the command surface                          |

Every command that needs a draft takes it with `-draft` for a catalogue entry or `-file` for a
draft written down, and can be asked for `-mirror` to thread it from the other selvedge or
`-tromp` to treadle it the way it is threaded.

The file format is one key and its value per line, with `#` for a comment. The shaft and
treadle counts have to come before the parts they size:

```
name 2/2 Twill
shafts 4
treadles 4
threading straight 8
tieup 12 23 34 14
treadling tromp
```

A threading or a treadling is a list of numbers, or `straight N`, or `point N`; a treadling may
also be `tromp`. A tie-up is one set of shafts per treadle, written as a run of digits (`12` is
shafts one and two) or separated by full stops when a shaft needs two digits (`1.10.12`).

A cloth is drawn with `x` where the warp is raised and `.` where the weft crosses over it, the
first pick at the top.

## Exit codes

| code | meaning                                     |
| ---- | ------------------------------------------- |
| `0`  | the command ran and produced an answer      |
| `1`  | the command could not run                   |
| `2`  | the command ran and there was nothing there |

The third code says something. `satin -shafts 6` reports that no step works, because there is
no six shaft satin; `floats -draft plain-weave` reports that no float is longer than the limit.
Both are answers a script has to be able to act on, and neither is a failure.

## The drafts

| draft              | shafts | tie-up                     | what it is                                        |
| ------------------ | ------ | -------------------------- | ------------------------------------------------- |
| `plain-weave`      | 2      | `1 2`                      | the firmest cloth there is                        |
| `plain-weave-4`    | 4      | `13 24`                    | the same cloth threaded over four shafts          |
| `basket-2-2`       | 2      | `1 2`                      | two ends and two picks together in each shed      |
| `twill-1-3`        | 4      | `1 2 3 4`                  | weft faced twill                                  |
| `twill-2-2`        | 4      | `12 23 34 14`              | balanced twill                                    |
| `twill-3-1`        | 4      | `123 234 134 124`          | warp faced twill                                  |
| `twill-3-1-repeat` | 4      | `123 234 134 124`          | one repeat of the same, four ends by four picks   |
| `herringbone`      | 4      | `12 23 34 14`              | the twill tie-up over a point draw                |
| `sateen-5`         | 5      | `1 3 5 2 4`                | five shaft sateen, stepped by two                 |
| `satin-5`          | 5      | `2345 1245 1234 1345 1235` | the same stepping, warp faced                     |
| `satin-8`          | 8      | `1 4 7 2 5 8 3 6`          | eight shaft satin, stepped by three               |
| `floating-end`     | 4      | `12 13 14 1`               | shaft one tied to every treadle, which is a fault |

Nothing in the structure or the float columns of `heddle drafts` is stored: every one of them
is read off the cloth after weaving the draft out. That is why `floating-end` is in the list.

## The ideas worth knowing before reading the code

**Floats run round the repeat, not along the piece.** A float is a run of picks over which one
end stays raised, and it is what decides whether the cloth holds together. A draft is woven
over and over, so the last pick of the repeat is followed by the first: an end raised over
picks 4 and 1 of a four pick repeat carries a float of two, not two floats of one. Cutting the
count off at the edge understates exactly the longest floats a weaver cares about.
`twill-3-1-repeat` is in the catalogue for this reason: it floats over three, and two of those
three cells sit on either side of the edge.

**A treadle that lifts everything opens no shed.** The tie-up says which shafts rise; the rest
stay down, and the shuttle passes through the gap between them. A treadle tied to every shaft,
or to none, leaves no gap, so it is not a treadle that weaves badly, it is a treadle that
cannot be trodden at all. `draft.TieUp.Validate` refuses both, and
`examples/opens-no-shed.draft` is kept in the repository so the refusal can be seen.

**A shaft the loom does not have is not a small mistake.** A tie-up naming shaft 9 on an eight
shaft loom will still lift, still print, and still weave: the ends it was meant to raise simply
stay down, and what comes out is a pick with no warp in it rather than an error. So
`shafts.Parse` checks every shaft against the loom it is being read for.

**Whether a satin exists at all is arithmetic.** A satin is a twill whose step is chosen so
that the single interlacement of each pick never lands beside the one before it. That needs the
step to share no factor with the number of shafts, or the interlacements fall on only some of
the shafts and the repeat comes out smaller than the loom; and it needs the step to be neither
one nor one less than the number of shafts, because either of those gives a twill. On six
shafts every step fails one of those tests, so there is no six shaft satin, and `satin -shafts
6` says so rather than offering something that is not one. On a prime number of shafts every
step in between works.

**An end that never interlaces is worse than a long float.** A float wears; an end that is
raised over every pick, or lowered under every one, is not woven in at all and will pull
straight out of the cloth. That is what `floating-end` has, and it is invisible in the balance,
the firmness and the structure, all three of which read as ordinary. `quality` looks along each
end and each pick for a line that never changes.

**Nothing hands out a slice the caller does not own.** A draft is read many times and altered
as often: mirrored, tromped, its rows and columns pulled out and measured. Reversing a
threading in place would leave the caller holding a threading it did not ask for, and the
mirrored draft it built would be wrong in a way that still validates, still weaves and still
prints.

## Tests

```
go vet ./...
go test ./...
```

Every package has its own tests, and `internal/cli` checks the exit code and the printed fields
of every command. The container build runs `go vet` and `go test` before it builds the binary,
with the module proxy off.

## Repository layout

```
cmd/heddle/         the entry point
internal/           the packages listed below
examples/           drafts written down, including one that is refused
Dockerfile          offline multi-stage build, scratch final image
RELEASE_REPORT.md   what was verified for this release
```

| package            | what lives there                                                   |
| ------------------ | ------------------------------------------------------------------ |
| `internal/shafts`  | sets of shafts, checked against the loom they belong to            |
| `internal/draft`   | threading, tie-up, treadling, the file format and the catalogue    |
| `internal/cloth`   | the drawdown, floats round the repeat, interlacements and firmness |
| `internal/pattern` | plain weave, twills, satins and the satin step rule                |
| `internal/quality` | long floats, ends and picks that never interlace, sleazy cloth     |
| `internal/report`  | aligned tables and fixed digit number formatting                   |
| `internal/cli`     | the command surface, the draft flags and the exit codes            |

No license is granted or implied by this repository.
