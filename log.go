package main

import (
	"github.com/zbh255/bilog"
	"os"
)

var Logger = bilog.NewLogger(os.Stdout,bilog.PANIC,bilog.WithCaller())