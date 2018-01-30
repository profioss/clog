//
// Copyright 2017 Radovan Vrždiak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clog exports logging primitives that log to stderr and also to io.Writer (file, network, etc).
// Usage is optimized for command line utilities and simple services.
//
package clog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// TODO
//   tests

// Logger is the interface for logging messages.
type Logger interface {
	// Debug writes a debug message to the log.
	Debug(msg ...interface{})

	// Debugf writes a formated debug message to the log.
	Debugf(fmt string, msg ...interface{})

	// Info writes an info message to the log.
	Info(msg ...interface{})

	// Infof writes a formated info message to the log.
	Infof(fmt string, msg ...interface{})

	// Warn writes a warning message to the log.
	Warn(msg ...interface{})

	// Warnf writes a formated warning message to the log.
	Warnf(fmt string, msg ...interface{})

	// Error writes an error message to the log.
	Error(msg ...interface{})

	// Errorf writes a formated error message to the log.
	Errorf(fmt string, msg ...interface{})

	// Error writes an error message to the log and aborts using os.Exit(1).
	Fatal(msg ...interface{})

	// Error writes a formated error message to the log and aborts using os.Exit(1).
	Fatalf(fmt string, msg ...interface{})
}

// Level represents the level of logging.
type Level int

// Different levels of logging.
const (
	DisabledLevel Level = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

type logger struct {
	level   Level
	w       io.Writer
	verbose bool
	// loggers for each log level
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	fatal *log.Logger
}

// New creates new Logger.
func New(w io.Writer, level string, verbose bool) (Logger, error) {
	l := &logger{w: w, verbose: verbose}

	lv, err := levelOfStr(level)
	if err != nil {
		return l, err
	}
	l.level = lv

	flags := log.Ldate | log.Ltime
	/*
		if l.level == DebugLevel {
			flags = log.Ldate | log.Ltime | log.Lshortfile
		}
	*/

	storage := w
	if storage == nil {
		storage = ioutil.Discard
	}
	multiOut := io.MultiWriter(storage, os.Stdout)
	multiErr := io.MultiWriter(storage, os.Stderr)

	l.fatal = log.New(multiErr, "FATAL: ", flags)

	if l.level == DisabledLevel {
		return l, nil // leave debug, info, ... to be nil
	}

	l.error = log.New(multiErr, "ERROR: ", flags)
	if l.level == ErrorLevel {
		return l, nil // leave debug, info, ... to be nil
	}

	l.warn = log.New(multiErr, "WARN:  ", flags)
	if l.level == WarnLevel {
		return l, nil // leave debug, info to be nil
	}

	l.info = log.New(storage, "INFO:  ", flags)
	if l.verbose {
		l.info = log.New(multiOut, "INFO:  ", flags)
	}
	if l.level == InfoLevel {
		return l, nil // leave debug to be nil
	}

	l.debug = log.New(storage, "DEBUG: ", flags)
	if l.verbose {
		l.debug = log.New(multiOut, "DEBUG: ", flags)
	}

	return l, nil
}

// Fatal is for fatal error messages.
func (l *logger) Fatal(msg ...interface{}) {
	if l.fatal == nil {
		return // Don't log at disabled level.
	}
	l.fatal.Fatal(l.compose(msg...))
}

// Fatalf is for formatted fatal error messages.
func (l *logger) Fatalf(fmt string, msg ...interface{}) {
	if l.fatal == nil {
		return // Don't log at disabled level.
	}
	l.fatal.Fatal(l.composef(fmt, msg...))
}

// Error is for error messages.
func (l *logger) Error(msg ...interface{}) {
	if l.level < ErrorLevel || l.error == nil {
		return // Don't log at lower levels.
	}
	l.error.Print(l.compose(msg...))
}

// Errorf is for formatted error messages.
func (l *logger) Errorf(fmt string, msg ...interface{}) {
	if l.level < ErrorLevel || l.error == nil {
		return // Don't log at lower levels.
	}
	l.error.Print(l.composef(fmt, msg...))
}

// Warn is for warning messages.
func (l *logger) Warn(msg ...interface{}) {
	if l.level < WarnLevel || l.warn == nil {
		return // Don't log at lower levels.
	}
	l.warn.Print(l.compose(msg...))
}

// Warnf is for formatted warning messages.
func (l *logger) Warnf(fmt string, msg ...interface{}) {
	if l.level < WarnLevel || l.warn == nil {
		return // Don't log at lower levels.
	}
	l.warn.Print(l.composef(fmt, msg...))
}

// Info is for info messages.
func (l *logger) Info(msg ...interface{}) {
	if l.level < InfoLevel || l.info == nil {
		return // Don't log at lower levels.
	}
	// l.info.Println(msg...)
	l.info.Print(l.compose(msg...))
}

// Infof is for formatted info messages.
func (l *logger) Infof(fmt string, msg ...interface{}) {
	if l.level < InfoLevel || l.info == nil {
		return // Don't log at lower levels.
	}
	l.info.Print(l.composef(fmt, msg...))
}

// Debug is for debug messages.
func (l *logger) Debug(msg ...interface{}) {
	if l.level < DebugLevel || l.debug == nil {
		return // Don't log at lower levels.
	}
	l.debug.Print(l.compose(msg...))
}

// Debugf is for formatted debug messages.
func (l *logger) Debugf(fmt string, msg ...interface{}) {
	if l.level < DebugLevel || l.debug == nil {
		return // Don't log at lower levels.
	}
	l.debug.Print(l.composef(fmt, msg...))
}

// caller adds inforation about source code file and line.
// Runtime information is expensive so it is used only in DebugLevel.
func (l *logger) caller() string {
	c := ""
	if l.level == DebugLevel {
		// see log/log.go of standard library
		_, file, line, ok := runtime.Caller(3) // 3 - show file of code where logger is used
		if !ok {
			file = "???"
			line = 0
		}
		c = fmt.Sprintf("%s:%d ", file, line)
	}

	return c
}

// compose prepares full log message. This time it ads caller info & PID if appropriate.
func (l *logger) compose(msg ...interface{}) string {
	output := []interface{}{l.pid(), l.caller()}
	output = append(output, msg...)

	return fmt.Sprint(output...)
}

// composef prepares formatted full log message. This time it ads caller info & PID if appropriate.
func (l *logger) composef(format string, msg ...interface{}) string {
	return fmt.Sprint(l.pid(), l.caller(), fmt.Sprintf(format, msg...))
}

func (l *logger) pid() string {
	p := ""
	if l.level == DebugLevel {
		p = fmt.Sprintf("[%d] ", os.Getpid())
	}

	return p
}

// OpenFile helper function opens log file with options suitable for logging i.e. O_APPEND, etc.
func OpenFile(fname string) (fd *os.File, err error) {
	err = os.MkdirAll(filepath.Dir(fname), os.ModePerm)
	if err != nil {
		return
	}

	fd, err = os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return
}

// levelOfString returns log level from provided string.
// Valid string parameters are: "disabled" | "error" | "warning" | "info" | "debug"
func levelOfStr(s string) (Level, error) {
	level := InfoLevel
	switch s {
	case "disabled":
		level = DisabledLevel
	case "error":
		level = ErrorLevel
	case "warning":
		level = WarnLevel
	case "info":
		level = InfoLevel
	case "debug":
		level = DebugLevel
	default:
		return level,
			fmt.Errorf("invalid log level: %s; use one of: disabled | error | warning | info | debug ", s)
	}

	return level, nil
}
