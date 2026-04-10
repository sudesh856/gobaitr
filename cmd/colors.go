package cmd

import "github.com/fatih/color"

var (
	colorSuccess  = color.New(color.FgGreen, color.Bold)
	colorError    = color.New(color.FgRed)
	colorAlert    = color.New(color.FgRed, color.Bold)
	colorInfo     = color.New(color.FgCyan)
	// colorWarning  = color.New(color.FgYellow)
	// colorWarnBold = color.New(color.FgYellow, color.Bold)
)
