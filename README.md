# URFS [![CircleCI](https://circleci.com/gh/bbengfort/urfs.svg?style=svg)](https://circleci.com/gh/bbengfort/urfs)

**Uniform random sample of files in a directory**

## Usage

To install the utility, use `go get`:

    $ go get github.com/bbengfort/urfs/...

If you have added `$GOPATH/bin` to your `$PATH` then you will have the `urfs` utility installed and available:

    $ urfs --help

You can then sample a directory with a likelihood of 0.25 (approximately a quarter of the documents), copying the files to the destination directory as follows:

    $ urfs -s 0.25 src/path dst/path

For very large directories this may take a while, but should be faster than many other utilities.
