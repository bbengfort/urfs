# URFS [![CircleCI](https://circleci.com/gh/bbengfort/urfs.svg?style=svg)](https://circleci.com/gh/bbengfort/urfs)

**Perform computations on files in a large directory**

## Usage

To install the utility, use `go get`:

```bash
$ go get github.com/bbengfort/urfs/...
```

If you have added `$GOPATH/bin` to your `$PATH` then you will have the `urfs` utility installed and available:

```bash
$ urfs --help
```

The `urfs` utility works on all files under a directory except for hidden files that start with a "." or a "~". Use the `--no-skip-dir` and `--no-skip-hidden` to include directories and hidden files. You can also filter directories using a glob like syntax on the file names. For example:

```bash
$ urfs -m *.txt cmd dir
```

Will only match files with a .txt extension. You can also specify a timeout to stop directory processing.

```bash
$ urfs -t 1m cmd dir
```

Will limit the command to only 1 minute of processing. There are a number of commands available in the utility, listed as follows:

### Sample

You can sample a directory with a likelihood of 0.25 (approximately a quarter of the documents), copying the files to the destination directory as follows:

```bash
$ urfs sample -s 0.25 src/path dst/path
```

For very large directories this may take a while, but should be faster than many other utilities.

### Count

You can count the number of files and bytes in a directory as follows:

```bash
$ urfs count src/path/*
```

This will return the number of files, bytes and average number of bytes per file for each of the paths passed to the utility.

## Writing Commands

URFS stands for "uniform random file sample", which was the original purpose of the command, still implemented as the `sample` command. It has since been generalized. To develop a parallel file system utility, simply create a `WalkFunc` and pass it to the `FSWalker.Walk` method.

Create an `FSWalker` as follows:

```go
fs := new(FSWalker)
fs.Init(context.Background())
```

At this point you can modify any of the walker's variables such as `Workers` or `Match` to modify operation.

A function that operates concurrently on all paths is written as follows:

```go
func (fs *FSWalker) MyWalker(root string) error {

    fs.Walk(root, func(path string) (string, error) {
        // Do something with each path, note that this function
        // is executed concurrently with other functions, so be
        // sure to use appropriate synchronization mechanisms
        return path, nil
    })

    // Access global results of the walker
    fmt.Printf(
        "%d of %d paths executed in %d\n",
        fs.nResults, fs.nPaths, fs.Duration,
    )

}
```

Note that each call of the `WalkFunc` happens concurrently, and is limited by the number of workers set in `fs.Workers` (to prevent too many files open or max number of threads reached).

If the `WalkFunc` returns an error, then processing is canceled. If the `WalkFunc` returns an empty string `""` then the result is not counted. This allows you to correctly use the state of the walker on complete.

If you execute `fs.Walk` you'll need to reset the `FSWalker` in order to call `fs.Walk` a second time. Use `fs.Reset(nil)` to reset it with the original context.
