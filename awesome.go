package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const NginxDate = "02/Jan/2006:15:04:05"

func check(e error) {
	if e != nil {
		panic(e) // TODO: Gracefully handle error
	}
}

// Read returns the data read from a file while ensuring that the file being read is first passed through check.
// This allows us to reduce some code duplication and panic if the file is not found or there is some io error.
func Read(proArgs string) []byte {
	data, err := ioutil.ReadFile(proArgs)
	check(err)

	return data
}

// Retrieves fields from nginx logs passed as log and gets the ip address, date, and file columns
func extractNginxFields(log []string) []map[string]string {
	fields := make([]map[string]string, len(log))

	ipPattern := regexp.MustCompile("([0-9]\\.?)+")
	datePattern := regexp.MustCompile("[0-9]+\\/[A-Z][a-z]+\\/\\d+:\\d+:\\d+:\\d+ \\+")
	filePattern := regexp.MustCompile(" \\/.* HTTP")

	for e := range log {
		// Do some stuff to grab the columns from the nginx log.
		ipAddr := ipPattern.FindString(log[e])
		date := strings.TrimRight(datePattern.FindString(log[e]), " +")
		file := filePattern.FindString(log[e])
		file = file[:len(file)-5] // Remove the " HTTP" at the end of the string because I'm bad at RegEx. TODO: FIX

		fields[e] = make(map[string]string)
		fields[e]["ipAddr"] = ipAddr
		fields[e]["date"] = date
		fields[e]["file"] = file
	}

	return fields
}

func retrieve(fields []map[string]string, window int){
	for e := range fields {
		// retrieve the dates from the log files.
		ti, err := time.Parse(NginxDate, fields[e]["date"])
		elapsed := int(time.Since(ti).Minutes())
		if err == nil && elapsed <= window {
			fmt.Printf("%v\t%v\t%v\n", ti, fields[e]["ipAddr"], fields[e]["file"])
		} else if err != nil {
			fmt.Printf("Awful error: %v\n", err)
		}
	}
}

func main() {
	fmt.Println("Welcome to Awesome Monitor.")
	proArgs := os.Args[1]
	data := string(Read(proArgs))
	window, err := strconv.Atoi(os.Args[2])

	if err != nil{
		os.Exit(-1) // TODO: Clean this up with the panic in Read
	}

	// We want to break the logs up by newlines
	lines := strings.Split(data, "\n")
	fields := extractNginxFields(lines)
	retrieve(fields, window)
}
