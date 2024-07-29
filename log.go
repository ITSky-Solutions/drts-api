package main

import (
	"log"
)

var Log *log.Logger = log.New(log.Writer(), "[LOG] ", log.Ldate|log.Ltime|log.Lshortfile)

