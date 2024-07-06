package main

import (
	"github.com/fatih/color"
)

var gray = color.New(90).SprintFunc()
var orange = color.New(208).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var italic = color.New(color.Italic).SprintFunc()

func Gray(args ...interface{}) string {
	return gray(args...)
}

func Orange(args ...interface{}) string {
	return orange(args...)
}

func Cyan(args ...interface{}) string {
	return cyan(args...)
}

func Red(args ...interface{}) string {
	return red(args...)
}

func Yellow(args ...interface{}) string {
	return yellow(args...)
}

func Magenta(args ...interface{}) string {
	return magenta(args...)
}

func Green(args ...interface{}) string {
	return green(args...)
}

func Blue(args ...interface{}) string {
	return blue(args...)
}

func Bold(args ...interface{}) string {
	return bold(args...)
}

func Italic(args ...interface{}) string {
	return italic(args...)
}
