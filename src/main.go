package main

import (
	"os"
	"fmt"
	"flag"
	herr "hike/error"
	spc "hike/spec"
	knw "hike/known"
	rdr "hike/reader"
	prs "hike/parser"
	abs "hike/abstract"
)

import _ "hike/concrete"
import _ "hike/generic"

const DEFAULT_HIKEFILE = "hikefile"
const DEFAULT_GOAL = "build"

func die(err herr.BuildError) {
	nerr := err.PrintBuildError(0)
	if nerr != nil {
		fmt.Fprintln(os.Stderr, "Yikes! Failed to print the true error:", nerr.Error())
		fmt.Fprintln(os.Stderr, "I guess that means you probably can't see this...")
	} else {
		fmt.Fprintln(os.Stderr)
	}
	os.Exit(1)
}

func main() {
	var hikefileName string
	const hikefileUsage = "Filename of hikefile to read for root project."
	flag.StringVar(&hikefileName, "hikefile", DEFAULT_HIKEFILE, hikefileUsage)
	flag.StringVar(&hikefileName, "f", DEFAULT_HIKEFILE, hikefileUsage)
	flag.Parse()
	cwd, nerr := os.Getwd()
	if nerr != nil {
		fmt.Fprintln(os.Stderr, "Oyyyy, couldn't determine current working directory (say whaaaaat):", nerr.Error())
		os.Exit(1)
	}
	config := &spc.Config {
		ProjectName: "this",
		TopDir: cwd,
	}
	rootState := spc.NewState(config)
	//TODO: search upward
	knownStructures := prs.NewKnownStructures()
	knw.RegisterAllKnownStructures(knownStructures)
	err := rdr.ReadFile(hikefileName, knownStructures, rootState)
	if err != nil {
		die(err)
	}
	err = rootState.Compile()
	if err != nil {
		die(err)
	}
	goalNames := flag.Args()
	if len(goalNames) == 0 {
		goalNames = []string{DEFAULT_GOAL}
	}
	var goals []*abs.Goal
	for _, goalName := range goalNames {
		goal := rootState.Goal(goalName)
		if goal == nil {
			fmt.Fprintln(os.Stderr, "No such goal:", goalName)
			os.Exit(1)
		}
		goals = append(goals, goal)
	}
	plan := abs.NewPlan()
	for _, goal := range goals {
		for _, action := range goal.Actions() {
			err = action.Perform(plan)
			if err != nil {
				die(err)
			}
		}
	}
	//TODO
}
