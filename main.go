// f5-auto-uploader is a service that watches directories and automatically
// updates or creates iFiles for the LTM module of an F5 BigIP instance.
//
// For usage information, please see:
//    f5-auto-uploader -h
//    f5-auto-uploader -help
//
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
)

const (
	major  = "1"
	minor  = "0"
	bugfix = "0"
)

// Print usage and exit with status 1.
func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
	os.Exit(1)
}

// Print version number and exit with status 0.
func version() {
	fmt.Printf("%s - %s.%s.%s\n", filepath.Base(os.Args[0]), major, minor, bugfix)
	os.Exit(0)
}

// Hide stderr and stout behind an io.Writer in order to ease testing.
var (
	stderr io.Writer = os.Stderr
	stdout io.Writer = os.Stdout
)

// Wrap os.Exit call into a variable function in order to ease testing.
var exit = func(status int) {
	os.Exit(status)
}

// fatal prints to standard error output (stderr) and then exit the program with
// a status 1. The prefix "fatal:" is prepended to the message. Arguments are
// handled in the manner of fmt.Print.
func fatal(v ...interface{}) {
	fmt.Fprintln(stderr, "fatal:", fmt.Sprint(v...))
	exit(1)
}

// verbose prints to standard output (stdout) only when the verboseMode is
// enabled. Otherwise it does nothing. The prefix "verbose:" is prepended to the
// message. Arguments are handled in the manner of fmt.Print.
func verbose(v ...interface{}) {
	if *verboseMode {
		fmt.Fprintln(stdout, "verbose:", fmt.Sprint(v...))
	}
}

// info prints to standard output (stdout). The prefix "info:" is prepended to
// the message. Arguments are handled in the manner of fmt.Print.
func info(v ...interface{}) {
	fmt.Fprintln(stdout, "info:", fmt.Sprint(v...))
}

// initF5Client initializes a new f5.Client with the provided configuration.
func initF5Client(cfg f5Config) (*f5.Client, error) {
	var (
		f5Client *f5.Client
		err      error
	)
	switch authMethod := cfg.AuthMethod; authMethod {
	case "basic":
		f5Client, err = f5.NewBasicClient(cfg.URL, cfg.User, cfg.Password)
	case "token":
		f5Client, err = f5.NewTokenClient(
			cfg.URL,
			cfg.User,
			cfg.Password,
			cfg.LoginProviderName,
			cfg.SSLCheck,
		)
	default:
		err = errors.New("unsupported auth method \"" + authMethod + "\"")
	}
	if err != nil {
		return nil, err
	}
	if !cfg.SSLCheck {
		f5Client.DisableCertCheck()
	}
	return f5Client, nil
}

var (
	configPath   = flag.String("config", "config.toml", "path to configuration file")
	verboseMode  = flag.Bool("verbose", false, "enable verbose mode")
	printVersion = flag.Bool("version", false, "print current version and exit")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *printVersion {
		version()
	}

	cfg, err := readConfig(*configPath)
	if err != nil {
		fatal(err)
	}

	f5Client, err := initF5Client(cfg.F5)
	if err != nil {
		fatal(err)
	}

	l := newLogger(os.Stderr)

	var routines []*watchRoutine
	defer func() {
		for i, r := range routines {
			l.Noticef("stopping routine %d", i)
			if err := r.stop(); err != nil {
				l.Errorf("cannot stop routine %d: %v", i, err)
			}
		}
	}()
	for _, watchCfg := range cfg.Watch {
		if err := scanDir(watchCfg.Dir, watchCfg.Exclude, f5Client); err != nil {
			l.Errorf("cannot scan directory %q: %v", watchCfg.Dir, err)
			return
		}
		routine, err := watchDir(f5Client, l, watchCfg)
		if err != nil {
			l.Error(err)
			return
		}
		routines = append(routines, routine)
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Kill, os.Interrupt)

	<-sig

	info("bye.")
}
