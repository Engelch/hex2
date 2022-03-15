//go:generate swagger generate spec

// TODO: shall it be changed to generic coding converter (iconv for codings)
// TODO: split between warnings (e.g. not found files for conversion) and errors so that computation can continue

package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	ce "github.com/engelch/go_libs/v2"
	cli "github.com/urfave/cli/v2"
)

const appVersion = "0.2.0"
const appName = "hex2"

// These CLI options are used more than once below. So let's use constants that we do not get
// misbehaviour by typoos.
const _debug = "debug"   // long (normal) name of CLI option
const _base64 = "base64" // dito
const _raw = "raw"       // dito

// =======================================================================================
func hex2(c *cli.Context, filename string) error {
	var data []byte
	if infile, err := ioutil.ReadFile(filename); err != nil {
		return errors.New(ce.CurrentFunctionName() + ":" + err.Error())
	} else {
		ce.CondDebugln(ce.CurrentFunctionName() + ":read input file " + filename + ":len " + fmt.Sprintf("%d", len(infile)))
		data, err = hex.DecodeString(strings.TrimSuffix(string(infile), "\n")) // strip line feeds added by some editors...
		if err != nil {
			return errors.New(ce.CurrentFunctionName() + ":" + err.Error())
		}
		ce.CondDebugln(ce.CurrentFunctionName() + ":hex decoding, len " + fmt.Sprintf("%d", len(infile)))
	}
	if c.Bool(_base64) {
		return processBase64Conversion(data)
	}
	return processRawConversion(data)
}

func processBase64Conversion(infile []byte) error {
	fmt.Print(base64.StdEncoding.EncodeToString(infile))
	return nil
}

func processRawConversion(infile []byte) error {
	n, err := os.Stdout.Write(infile) // flushing should be done implicitly by terminating the app
	if n != len(infile) {             // should not happen for such small files
		fmt.Fprint(os.Stderr, "Error writing raw data to stdout")
	}
	if err != nil { // should also never happen
		return errors.New(ce.CurrentFunctionName() + ":" + err.Error())
	}
	return nil
}

// =======================================================================================
// checkOptions checks the command line options if properly set or in range.
// POST: exactly one keyfile is not mt.
func checkOptions(c *cli.Context) error {
	if c.Bool(_debug) {
		ce.CondDebugSet(true)
	}
	ce.CondDebugln("Debug is enabled.")
	if !c.Bool(_raw) && !c.Bool(_base64) {
		return errors.New("At least one of the options -6 or -r must be set.")
	}
	if c.Bool(_raw) && c.Bool(_base64) {
		return errors.New("Only one of the options -6 or -r can be set.")
	}
	return nil
}

// commandLineOptions just separates the definition of command line options ==> creating a shorter main
func commandLineOptions() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    _debug,
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "OPTIONAL: enable debug",
		},
		&cli.BoolFlag{
			Name:    _base64,
			Aliases: []string{"6"},
			Value:   false,
			Usage:   "OPTIONAL: output base64 format",
		},
		&cli.BoolFlag{
			Name:    _raw,
			Aliases: []string{"r"},
			Value:   false,
			Usage:   "OPTIONAL: output raw format",
		},
	}
}

// main start routine
func main() {
	app := cli.NewApp() // global var, see discussion above
	app.Flags = commandLineOptions()
	app.Name = appName
	app.Version = appVersion
	app.Usage = "Convert hex into base64 or raw aka binary format.\n" +
		"\n                  hex2 [-d] -r <<file>> # creates raw format sent to stdout" +
		"\n                  hex2 [-d] -6 <<file>> # creates base64 sent to stdout"

	app.Action = func(c *cli.Context) error {
		err := checkOptions(c)
		ce.ExitIfError(err, 9, "checkOptions")
		for index, _ := range make([]int, c.NArg()) {
			err = hex2(c, c.Args().Get(index))
			ce.ExitIfError(err, 1, "Error for file "+c.Args().Get(index))
		}
		return nil
	}
	_ = app.Run(os.Args)
}

// eof
