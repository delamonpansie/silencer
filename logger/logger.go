package logger

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"go.uber.org/zap"
)

var Log zap.Logger = *zap.NewNop()

type StdLogger struct {
	*zap.SugaredLogger
}

func (sl *StdLogger) Fatalln(v ...interface{}) {
	sl.Fatal(v...)
}

func (sl *StdLogger) Panicln(v ...interface{}) {
	sl.Panic(v...)
}

func (sl *StdLogger) Print(v ...interface{}) {
	sl.Info(v...)
}

func (sl *StdLogger) Printf(fmt string, v ...interface{}) {
	sl.Infof(fmt, v...)
}

func (sl *StdLogger) Println(v ...interface{}) {
	sl.Info(v...)
}

func reopenStdio(logFile string) error {
	var err error
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	os.Stdin, err = os.OpenFile("/dev/null", os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	if os.Stdin.Fd() != 0 {
		return fmt.Errorf("assert stdin: fd=%q", os.Stdin.Fd())
	}

	if err := os.Stdout.Close(); err != nil {
		return err
	}
	os.Stdout, err = os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0)
	if err != nil {
		return err
	}
	if os.Stdout.Fd() != 1 {
		return fmt.Errorf("assert stdout: fd=%q", os.Stdin.Fd())
	}

	if err := os.Stderr.Close(); err != nil {
		return err
	}
	if err := syscall.Dup2(1, 2); err != nil {
		return err
	}
	os.Stderr = os.NewFile(2, logFile)

	return nil
}

var (
	verbose = flag.Bool("verbose", false, "produce verbose output")
	LogFile = flag.String("log", "", "log destination (default stderr)")
)

func Init() {
	var err error

	stderr := os.Stderr
	logFatal := func(err error) {
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}

	fd, err := syscall.Dup(2)
	if err != nil {
		logFatal(err)
	}
	stderr = os.NewFile(uintptr(fd), "stderr")
	defer stderr.Close()

	if *LogFile != "" && *LogFile != "-" {
		if err := reopenStdio(*LogFile); err != nil {
			logFatal(err)
		}
	}

	var logp *zap.Logger
	if *verbose {
		logp, err = zap.NewDevelopment()
	} else {
		logp, err = zap.NewProduction()
	}
	if err != nil {
		logFatal(err)
	}
	Log = *logp
}
