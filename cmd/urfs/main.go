// CLI Command for the UFS App
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bbengfort/urfs"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var fs *urfs.FSWalker

//===========================================================================
// Main Method
//===========================================================================

func main() {

	// Load the .env file if it exists
	godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "urfs"
	app.Version = "0.3"
	app.Usage = "perform computations on files in a large directory"
	app.Before = initWalker

	// Define the global flags for the application
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "t, timeout",
			Value: "",
			Usage: "specify a parsable duration to limit sampling",
		},
		cli.IntFlag{
			Name:  "w, workers",
			Value: urfs.DefaultWorkers,
			Usage: "specify size of workers pool for system threads",
		},
		cli.BoolFlag{
			Name:  "D, no-skip-dirs",
			Usage: "do not skip directories",
		},
		cli.BoolFlag{
			Name:  "H, no-skip-hidden",
			Usage: "do not skip hidden files and directories",
		},
		cli.StringFlag{
			Name:  "m, match",
			Value: "*",
			Usage: "specify a pattern to match files on",
		},
	}

	// Define the commands for the application
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "sample",
			Usage:     "uniform random sample of files in a directory",
			ArgsUsage: "src dst",
			Action:    sample,
			Flags: []cli.Flag{
				cli.Float64Flag{
					Name:  "s, sample",
					Value: 0.1,
					Usage: "approximate fractional size of sample",
				},
			},
		},
		cli.Command{
			Name:      "count",
			Usage:     "compute number of files and bytes per directory",
			ArgsUsage: "dir [dir ...]",
			Action:    count,
		},
	}

	// Run the application
	app.Run(os.Args)
}

//===========================================================================
// Initialize Walker
//===========================================================================

func initWalker(c *cli.Context) (err error) {
	// Parse the timeout duration
	var timeout time.Duration
	if c.String("timeout") != "" {
		if timeout, err = time.ParseDuration(c.String("timeout")); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	}

	// Create the context for the walk function
	ctx := context.Background()
	if timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, timeout)
	}

	// Initialize the walker
	fs = new(urfs.FSWalker)
	fs.Init(ctx)

	// Set other defaults from the command line
	fs.Workers = c.Int("workers")
	fs.SkipDirs = !c.Bool("no-skip-dirs")
	fs.SkipHidden = !c.Bool("no-skip-hidden")
	fs.Match = c.String("match")

	return nil
}

//===========================================================================
// Sample Command
//===========================================================================

func sample(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.NewExitError("specify the src and dst directories", 1)
	}

	args := c.Args()
	result, err := fs.Sample(args.Get(0), args.Get(1), c.Float64("sample"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Println(result)
	return nil
}

//===========================================================================
// Count Command
//===========================================================================

func count(c *cli.Context) error {
	_, err := fs.Count(true, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
