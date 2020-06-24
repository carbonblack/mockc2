package log

import (
	"github.com/fatih/color"
)

// DebugEnabled controls where debug logging is shown or not.
var DebugEnabled = false

// Debug logs low level information.
func Debug(format string, a ...interface{}) {
	if DebugEnabled {
		color.Red("[DEBUG] "+format, a...)
	}
}

// Info logs general information for the user.
func Info(format string, a ...interface{}) {
	color.Cyan("[i] "+format, a...)
}

// Note logs general information for the user of higher importance.
func Note(format string, a ...interface{}) {
	color.Yellow("[-] "+format, a...)
}

// Success logs information about processes that completed successfully.
func Success(format string, a ...interface{}) {
	color.Green("[+] "+format, a...)
}

// Warn logs information about warnings or errors.
func Warn(format string, a ...interface{}) {
	color.Red("[!] "+format, a...)
}
