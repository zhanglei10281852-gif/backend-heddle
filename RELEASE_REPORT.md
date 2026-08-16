# Heddle release report

This report records what was checked for this release. Every result below was produced by
running the command shown, on `linux/amd64` for the container checks and on the development host
for the Go toolchain checks.

## Size

Effective lines are counted after removing blank lines and lines that are only comments.

| part                              | effective lines |
| --------------------------------- | --------------- |
| production Go (`cmd`, `internal`) | 2635            |
| tests (`*_test.go`)               | 2563            |

Per file:

| file                          | production | tests |
| ----------------------------- | ---------- | ----- |
| `cmd/heddle/main.go`          | 8          | –     |
| `internal/cli/cli.go`         | 755        | 368   |
| `internal/cloth/cloth.go`     | 314        | 443   |
| `internal/draft/catalogue.go` | 74         | –     |
| `internal/draft/draft.go`     | 292        | 410   |
| `internal/draft/parse.go`     | 241        | 257   |
| `internal/pattern/pattern.go` | 397        | 455   |
| `internal/quality/quality.go` | 198        | 210   |
| `internal/report/report.go`   | 137        | 111   |
| `internal/shafts/shafts.go`   | 219        | 309   |

## Toolchain checks

| check                    | result                                                    |
| ------------------------ | --------------------------------------------------------- |
| `gofmt -l internal cmd`  | no output                                                 |
| `go vet ./...`           | no findings                                               |
| `go test ./... -count=1` | ok for all seven packages, `cmd/heddle` has no test files |

What the packages cover:

- `shafts`: building a set from shafts in any order with five rejections; adding and removing a
  shaft, including one already there, one that is not, and one off the loom; the shafts of a set
  coming back in order as a slice the caller owns, with the highest and the empty case; union,
  intersection, difference and complement, with the checks that a set and its complement cover
  the loom and share nothing and that the complement of the complement is the set again, and
  three rejections; ordering by count and then by shafts, and sorting a list; printing as
  separated numbers, as a run of digits, and as a grid, over a set that reaches past nine and
  over the empty set; validation against a loom over a shaft off the loom, the empty set and
  three bad shaft counts; reading a set as digits and as separated numbers with either order and
  surrounding space, and the check that what is read prints and reads back to itself; twelve
  rejections in `Parse`; reading a set as a grid with five rejections; and the rendered
  description.
- `draft`: straight and point draws against their ends and their repeats, with five and two
  rejections; threading validation over a shaft off the loom, a shaft of zero, a negative shaft,
  no ends and two bad shaft counts; the ends on each shaft, the shafts with no end on them, and
  the check that the counts add up to the ends; clone independence and reversing without
  touching the threading it was read from, including reversing twice and a threading that reads
  the same either way; the repeat over seven sequences including one that does not repeat, and
  the same for a treadling; treadling counts, reversal, cloning, printing and five rejections;
  the tie-up read forwards and backwards with the check that the two agree on how many shafts
  are lifted, and clone independence; the refusal of a treadle that lifts every shaft, one that
  lifts none, a shaft off the loom, a tie-up short of a treadle and two bad shaft counts; draft
  validation over four faults; the shafts lifted by a pick and whether an end is raised, with
  five rejections; tromping as writ including the refusal when there are fewer treadles than
  shafts; mirroring including twice over; the catalogue being ordered, every entry validating,
  weaving right through and leaving no shaft empty, with two lookup cases; the file format with
  comments, blank lines, keys in either case and a missing name; nineteen rejections of bad
  files; and every draft in the catalogue written out and read back to itself.
- `cloth`: weaving plain weave and a twill out cell by cell against their drawn rows, their
  balance and the count of ends each pick raises; two rejections in `Weave`; building a cloth
  from cells the caller keeps, with four rejections and two validation cases; reading a cell, a
  pick and an end with the rows and columns belonging to the caller and six rejections; the face
  of a warp faced, a weft faced and a balanced cloth, including the two satins of the catalogue;
  equality over size and cells; turning the cloth over and the check that what was raised is
  then crossed; floats running round the repeat over a float of two and a float of three that
  cross the edge; the float of a line that never changes, in both directions; the floats of plain
  weave being one long and as many as there are cells, with the histogram and the over-the-limit
  filter; the check over ten drafts that the floats cover every cell of the cloth exactly once
  in each direction; the ordering of the floats over the limit; interlacements and firmness for
  plain weave at one, a twill and a satin in descending order, and a solid cloth at zero; and
  both rendered descriptions.
- `pattern`: plain weave recognised for two drafts and refused for five, with the one pick case
  and a rejection; twills read for five drafts against their up, down, step, repeat and
  direction, with the check that the counts add up to the repeat, and four cases that are not
  twills; a twill climbing the other way read as a left hand twill with the same slope; the
  satin steps of four prime shaft counts, with the checks that each step it offers the rule
  accepts and that they come back in order, and six rejections; the refusal of a step of one, of
  one less than the shafts, of four steps outside the repeat and of four shaft counts too small
  or too large, with the acceptance of every step in between on a prime number of shafts; satin
  tie-ups over three prime shaft counts and every step they allow, checked so that each shaft is
  lifted exactly once and the two faces of the tie-up share no shaft, with two rejections;
  building a whole satin draft over two faces with four rejections; satins read back from the
  cloth for both faces and refused for five other drafts; repeats over three hand built cloths
  and the check that the repeat of every catalogue draft divides its size, with a rejection;
  threading symmetry over six threadings; the classifier over eight drafts and a rejection; the
  whole analysis of a twill; and the step list including the check that it leaves its input
  alone.
- `quality`: plain weave read as sound with floats of one, a firmness of one, nothing unbound
  and no problems; a 2/2 twill sound at the default limit and unsound at a limit of one, with
  the floats worst first; the floating end draft read for the two ends that never interlace,
  their float over the whole repeat, and the problem naming them before the floats; a hand built
  cloth read for an unbound pick and two unbound ends; an eight shaft satin read for floats of
  seven, for being sleazy at a firmness under the threshold, and for the firmness not depending
  on the float limit; interlacements counted for plain weave at twice per cell with the three
  firmnesses in order; four rejections; and both rendered descriptions.
- `report`: table rendering, right alignment and trailing padding, short rows, rune counting for
  multi byte headers, the fixed digit float and scientific formats, the boolean, integer and
  integer list formats, and the bar cell.
- `cli`: the usage screen with no arguments, `help` in three spellings, `version` in three
  spellings, an unknown command, per command help, an unknown flag, `draft` over a straight draw
  and a point draw and mirrored with two rejections, `cloth` over three drafts and mirrored and
  tromped, `floats` over a twill, over plain weave with nothing over the limit, over one repeat
  of a twill whose floats cross the edge, and with a bad limit, `pattern` over plain weave, a
  satin and a basket, `satin` on five and on seven shafts with five rejections, `tromp` with a
  rejection, `drafts`, `report` over a sound draft and a faulty one, and reading a draft from a
  file with four rejections.

## Container checks

| check                                                                                     | result                                                                                                 |
| ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `docker build --platform linux/amd64 -t heddle:verify .`                                  | succeeded; the builder stage ran `go vet ./...` and `go test ./...` with `GOPROXY=off` and both passed |
| `docker run --rm --network none heddle:verify version`                                    | version line plus the study tool notice, exit 0                                                        |
| `docker run --rm --network none heddle:verify drafts`                                     | the same twelve rows as the host run                                                                   |
| `docker run --rm --network none heddle:verify cloth -file /examples/herringbone.draft`    | 12 ends by 8 picks, repeat 6 by 4, firmness 0.5, no plain weave, twill or satin                        |
| `docker run --rm --network none heddle:verify cloth -file /examples/opens-no-shed.draft`  | refused the tie-up, naming the treadle that lifts all four shafts, exit 1                              |
| `docker run --rm --network none heddle:verify satin -shafts 6`                            | every step shares a factor with six or sits at an end, so there is no six shaft satin, exit 2          |
| `docker run --rm --network none heddle:verify quality -file /examples/floating-end.draft` | ends 1 and 5 never interlace, longest warp float 8, sound no                                           |

The final image is `scratch` and contains only the binary and `examples/`.

## Command smoke tests

The exit code column is the observed code.

| command                                    | observed result                                                                                | exit |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------- | ---- |
| `version`                                  | version line plus the study tool notice                                                        | 0    |
| `help`                                     | nine commands, the exit codes, the twelve drafts, the drawing convention                       | 0    |
| `drafts`                                   | twelve drafts, each woven out and read for structure, floats, firmness and faults              | 0    |
| `draft -draft twill-2-2`                   | threading `1 2 3 4 1 2 3 4` repeating over 4, tie-up grids, two ends per shaft, not symmetric  | 0    |
| `draft -draft herringbone`                 | 12 ends, threading repeating over 6, the other selvedge reading `2 3 4 3 2 1 2 3 4 3 2 1`      | 0    |
| `draft -draft twill-2-2 -mirror`           | threading `4 3 2 1 4 3 2 1`, the other selvedge reading `1 2 3 4 1 2 3 4`                      | 0    |
| `cloth -draft plain-weave`                 | `x.x.x.x.` and `.x.x.x.x`, 128 interlacements, firmness 1, plain weave                         | 0    |
| `cloth -draft twill-2-2`                   | the diagonal drawn out, 32 of 64 cells warp up, repeat 4 by 4, 2/2 twill stepping 1 right      | 0    |
| `cloth -draft twill-3-1`                   | 48 of 64 cells warp up, warp faced, 3/1 twill stepping 1 to the right                          | 0    |
| `cloth -draft basket-2-2`                  | each shed woven twice, repeat 4 by 4, no plain weave, twill or satin                           | 0    |
| `cloth -draft twill-2-2 -mirror`           | the same twill climbing 1 to the left                                                          | 0    |
| `cloth -file examples/sateen-5.draft`      | 10 by 10, one end raised in five, weft faced satin stepping 2                                  | 0    |
| `floats -draft twill-3-1-repeat -over 2`   | 8 warp floats and 8 weft floats, longest 3 in both directions                                  | 0    |
| `floats -draft satin-8 -over 5`            | 64 floats each way, all of them 1 or 7, the 64 over the limit listed worst first               | 0    |
| `floats -draft plain-weave`                | every float one long, so nothing is over the limit of 3                                        | 2    |
| `pattern -draft plain-weave`               | repeat 2 by 2, plain weave, and also a 1/1 twill, which is why plain weave is looked for first | 0    |
| `pattern -draft satin-5`                   | satin yes, step 2, 5 shafts, warp faced                                                        | 0    |
| `pattern -draft sateen-5`                  | satin yes, step 2, 5 shafts, weft faced                                                        | 0    |
| `pattern -draft herringbone`               | repeat 6 by 4, no plain weave, twill or satin                                                  | 0    |
| `quality -draft plain-weave`               | floats of 1, firmness 1, nothing unbound, sound                                                | 0    |
| `quality -draft satin-8`                   | longest float 7 both ways, 64 over the limit, firmness 0.25, sleazy, sound no                  | 0    |
| `quality -draft twill-2-2 -limit 1`        | the floats of 2 reported against a limit of 1                                                  | 0    |
| `quality -draft floating-end`              | ends 1 and 5 never interlace, longest warp float 8, the worst float drawn out, sound no        | 0    |
| `satin -shafts 5`                          | steps 2 and 3 usable, 1 and 4 refused as twills                                                | 0    |
| `satin -shafts 7`                          | steps 2, 3, 4 and 5 usable, seven being prime                                                  | 0    |
| `satin -shafts 9`                          | steps 2, 4, 5 and 7 usable; 3 and 6 share the factor 3                                         | 0    |
| `satin -shafts 6`                          | no step works, so there is no six shaft satin                                                  | 2    |
| `satin -shafts 5 -counter 2`               | the tie-up `1 3 5 2 4`, every shaft carrying ends, repeat 5, longest float 4                   | 0    |
| `satin -shafts 7 -counter 2 -warp`         | the warp faced satin of the same stepping, drawn out                                           | 0    |
| `tromp -draft twill-2-2`                   | the treadling already reads the threading, so the cloth is the same                            | 0    |
| `tromp -draft twill-1-3`                   | the same, over the weft faced tie-up                                                           | 0    |
| `tromp -draft herringbone`                 | 8 picks become 12, balance 0.5 becomes 0.527778, and the cloth is not the same                 | 0    |
| `report -draft twill-2-2`                  | ten rows over five packages, longest float 2, the cloth sound                                  | 0    |
| `report -draft satin-5`                    | firmness 0.4, longest float 4, the cloth not sound                                             | 0    |
| `report -draft floating-end`               | longest float 8, the cloth not sound                                                           | 0    |
| `cloth -h`                                 | the flag list on standard output                                                               | 0    |
| `cloth -nope`                              | `flag provided but not defined: -nope` and the flag list                                       | 1    |
| `sley`                                     | `unknown command "sley"` and the usage screen                                                  | 1    |
| `draft -draft crackle`                     | the draft is not in the catalogue, and the twelve that are were listed                         | 1    |
| `draft -draft plain-weave-4 -tromp`        | four shafts and two treadles cannot be treadled as threaded                                    | 1    |
| `floats -draft plain-weave -over 0`        | a limit of 0 allows no float at all                                                            | 1    |
| `satin -shafts 4`                          | a satin needs at least five shafts                                                             | 1    |
| `satin -shafts 5 -counter 1`               | a step of 1 puts every interlacement beside the one before it                                  | 1    |
| `cloth -file examples/opens-no-shed.draft` | named the file, the line and the treadle that lifts every shaft                                | 1    |

## Known limitations

- The catalogue is twelve drafts and the file format is this tool's own. There is no reader for
  any of the formats weaving software uses, and no way to look a draft up: every draft it knows
  is written into `draft.Catalogue` or read from a file.
- The float limit, the firmness threshold under which a cloth is called sleazy, and the very
  idea of scoring a cloth by its interlacements are this tool's own. They are set out in the
  package comments so they can be checked, and no number printed here is comparable with a
  number from anywhere else. A satin is reported as unsound at the default limit, which is
  correct arithmetic and normal weaving: a satin is made of long floats on purpose.
- The classifier knows plain weave, twills and satins. A basket, a waffle, a huck, a
  double weave and everything else come back as no plain weave, twill or satin, which is a
  statement about what this tool looks for rather than about the cloth.
- Plain weave satisfies the definition of a 1/1 twill, so `pattern` reports both and the
  classifier looks for plain weave first. That is a property of the definitions and not a
  special case in the code.
- Everything is one repeat. There are no selvedges, no take-up, no draw-in, no sett and no
  yarn: the cloth is a grid of cells and the floats run round it, which is what makes the
  arithmetic exact and also what makes it unlike a piece on a loom.
- The tie-up is a rising shed only. A loom that sinks shafts instead would need every set turned
  over, and `Cloth.Transpose` is the nearest this tool comes to that.
- Every draft is treadled one treadle at a time. Two treadles pressed together, which is how a
  countermarch loom is often used, cannot be written down here.
- The repeat is found by trying every divisor of the width and the height, and the twill step by
  trying every shift, so both are quadratic in the size of the cloth. The largest cloth in the
  catalogue is 16 ends by 16 picks and `cloth` draws nothing larger than 96 by 96.
- Ends and picks are capped at 512 and shafts at 32, because a set of shafts is held in a
  `uint32`.

## Provenance

Heddle was written from scratch for this repository. There is no upstream project and no
upstream fix commit behind any part of it, so any process that requires a real upstream repair
as evidence does not apply here and is left for human review.
