package main

import (
  "github.com/fatih/color"
)

var cyan = color.New(color.FgCyan).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var italic = color.New(color.Italic).SprintFunc()

func Cyan(args ...interface{}) string {
  return cyan(args...)
}

func Red(args ...interface{}) string {
  return red(args...)
}

func Yellow(args ...interface{}) string {
  return yellow(args...)
}

func Green(args ...interface{}) string {
  return green(args...)
}

func Blue(args ...interface{}) string {
  return blue(args...)
}

func Italic(args ...interface{}) string {
  return italic(args...)
}
