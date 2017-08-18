// CLI Command for the UFS App
package main

import (
	"os"
	"time"

	"github.com/bbengfort/urfs"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

//===========================================================================
// Main Method
//===========================================================================

func main() {

	// Load the .env file if it exists
	godotenv.Load()

	// Instantiate the command line application
	app := cli.NewApp()
	app.Name = "urfs"
	app.Version = "0.2.1"
	app.Usage = "uniform random sample of files in a directory"
	app.ArgsUsage = "src dst"

	// Define the flags for the application
	app.Flags = []cli.Flag{
		cli.Float64Flag{
			Name:  "s, sample",
			Value: 0.1,
			Usage: "approximate fractional size of sample",
		},
		cli.StringFlag{
			Name:  "t, timeout",
			Value: "",
			Usage: "specify a parsable duration to limit sampling",
		},
	}

	// Define the action for the application
	app.Action = sample

	// Run the application
	app.Run(os.Args)
}

func sample(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		return cli.NewExitError("specify the src and dst directories", 1)
	}

	// Parse the timeout duration
	var timeout time.Duration
	if c.String("timeout") != "" {
		if timeout, err = time.ParseDuration(c.String("timeout")); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	}

	// Create the context for the search function
	ctx := context.Background()
	if timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, timeout)
	}

	args := c.Args()
	if err = urfs.Sample(ctx, args.Get(0), args.Get(1), c.Float64("sample")); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
