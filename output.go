package main

import (
	"fmt"
	"os"
)

type Output interface {
	Print(args ...interface{})
	Println(args ...interface{})
}

type StandardOutput struct{}

func (s StandardOutput) Print(args ...interface{}) {
  fmt.Fprint(os.Stdout, args...)
}

func (s StandardOutput) Println(args ...interface{}) {
	fmt.Fprintln(os.Stdout, args...)
}

type ErrorOutput struct{}

func (e ErrorOutput) Print(args ...interface{}) {
  fmt.Fprint(os.Stderr, args...)
}

func (e ErrorOutput) Println(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

type DebugOutput struct{}

func (d DebugOutput) Print(args ...interface{}) {
  debug := os.Getenv("DEBUG")
  if debug == "true" {
    fmt.Fprint(os.Stderr, append([]interface{}{Cyan("[DEBUG]")}, args...)...)
  }
}

func (d DebugOutput) Println(args ...interface{}) {
	debug := os.Getenv("DEBUG")
	if debug == "true" {
		fmt.Fprintln(os.Stderr, append([]interface{}{Cyan("[DEBUG]")}, args...)...)
	}
}

type NoOutput struct{}

func (n NoOutput) Print(args ...interface{}) {
}

func (n NoOutput) Println(args ...interface{}) {
}

type OutputType string

const (
	StdOut OutputType = "STANDARD"
	StdErr OutputType = "ERROR"
	Debug  OutputType = "DEBUG"
	None   OutputType = "NONE"
)

func (outputType OutputType) Out() Output {
	switch outputType {
	case StdOut:
		return StandardOutput{}
	case StdErr:
		return ErrorOutput{}
	case Debug:
		return DebugOutput{}
	case None:
		return NoOutput{}
	default:
		return StandardOutput{}
	}
}

func Out(outputType OutputType) Output {
  return outputType.Out()
}
