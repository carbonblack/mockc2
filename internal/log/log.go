package log

import (
	"github.com/fatih/color"
)

var DebugEnabled = false

func Debug(format string, a ...interface{}) {
	if DebugEnabled {
		color.Red("[DEBUG] "+format, a...)
	}
}

func Info(format string, a ...interface{}) {
	color.Cyan("[i] "+format, a...)
}

func Note(format string, a ...interface{}) {
	color.Yellow("[-] "+format, a...)
}

func Success(format string, a ...interface{}) {
	color.Green("[+] "+format, a...)
}

func Warn(format string, a ...interface{}) {
	color.Red("[!] "+format, a...)
}
