package main

import (
	"log"
	"os"
)

type Logger struct {
	Error *log.Logger
	Info  *log.Logger
}

var logger = &Logger{
	Error: log.New(os.Stderr, "ERROR|", log.LstdFlags|log.Lshortfile),
	Info:  log.New(os.Stdout, "INFO|", log.LstdFlags|log.Lshortfile),
}