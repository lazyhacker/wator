This is an implementation of the  Wa-Tor (http://en.wikipedia.org/wiki/Wa-Tor)
simulation.

[!wator.gif](wator.gif)

I started this project as an exercise for me to learn the
[Go Programming Language](http://golang.org) so there will many newbie mistakes
and non-idiomatic patterns.  This project is probably not much use to anyone
else except maybe to help people realize that there is an even less
knowledgeable gopher out there.

## Install

`go get lazyhacker.dev/wator`

##

To run wator just run the `wator` binary that is built.

```
Usage:
  -fbreed int
    	# of cycles for fish to reproduce (default: 500) (default 500)
  -fish int
    	Initial # of fish. (default 1000)
  -height int
    	Height of the world (North-South). (default: 240) (default 240)
  -sbreed int
    	# of cycles for shark to reproduce (default: 100) (default 100)
  -sharks int
    	Initial # of sharks. (default 500)
  -starve int
    	# of cycles shark can go with feeding before dying (default: 100) (default 100)
  -width int
    	Width of the world (East - West). (default: 320) (default 320)
```
