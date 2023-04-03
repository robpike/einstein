# Einstein

To install: `go get robpike.io/cmd/einstein`.


This is a simple program to print an STL representation of the "einstein hat"
tile described in [An aperiodic monotile](https://arxiv.org/pdf/2303.10798.pdf) by
David Smith, Joseph Samuel Myers, Craig S. Kaplan, and Chaim Goodman-Strauss.

The output is simple textual unoptimized STL.
For the reflected tile, run with the -r flag.

The output is included in the repo, if you just want the result.

The Quad and Box code could be pulled out to create a library, but it's
so small it's probably not worth it.
