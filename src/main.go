package main

import (
	"os"
	"fmt"
	"flag"
	"time"
	"path/filepath"
	herr "hike/error"
	spc "hike/spec"
	knw "hike/known"
	rdr "hike/reader"
	prs "hike/parser"
	abs "hike/abstract"
)

import _ "hike/concrete"
import _ "hike/generic"
import _ "hike/hilevel"

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

func numWidth(n int) (exp int) {
	for n > 0 {
		n /= 10
		exp++
	}
	if exp == 0 {
		exp = 1
	}
	return
}

func fileExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	switch {
		case err == nil:
			exists = true
		case os.IsNotExist(err):
			err = nil
	}
	return
}

func main() {
	var hikefileName string
	const hikefileUsage = "Filename of hikefile to read for root project."
	flag.StringVar(&hikefileName, "hikefile", DEFAULT_HIKEFILE, hikefileUsage)
	flag.StringVar(&hikefileName, "f", DEFAULT_HIKEFILE, hikefileUsage)
	var pretend bool
	const pretendUsage = "Print the plan, but do not execute it."
	flag.BoolVar(&pretend, "pretend", false, pretendUsage)
	flag.BoolVar(&pretend, "p", false, pretendUsage)
	var dumpStruct bool
	const dumpStructUsage = "Dump artifact/transform structure (and quit if no goal given)."
	flag.BoolVar(&dumpStruct, "dump", false, dumpStructUsage)
	flag.Parse()
	noDefaultBuild := dumpStruct
	// find hikefile
	cwd, nerr := os.Getwd()
	if nerr != nil {
		fmt.Fprintln(os.Stderr, "Oyyyy, couldn't determine current working directory (say whaaaaat):", nerr.Error())
		os.Exit(1)
	}
	if hikefileName == "" {
		hikefileName = DEFAULT_HIKEFILE
	}
	hikefileName = filepath.FromSlash(hikefileName)
	var topDir, hikefilePath string
	if filepath.IsAbs(hikefileName) {
		topDir = cwd
		hikefilePath = filepath.Clean(hikefileName)
	} else {
		topDir = cwd
		for {
			hikefilePath = filepath.Join(topDir, hikefileName)
			exists, xerr := fileExists(hikefilePath)
			if xerr != nil {
				fmt.Fprintln(os.Stderr, "Failed to stat '%s': %s\n", hikefileName, xerr.Error())
				os.Exit(1)
			}
			if exists {
				break
			}
			nextParent := filepath.Dir(topDir)
			if nextParent == topDir || len(nextParent) == 0 {
				fmt.Fprintf(os.Stderr, "No '%s' found in '%s' nor any ancestor\n", hikefileName, cwd)
				os.Exit(1)
			}
			topDir = nextParent
		}
	}
	// init state
	config := &spc.Config {
		ProjectName: "this",
		TopDir: topDir,
		CurrentHikefile: hikefilePath,
	}
	rootState := spc.NewState(config)
	knownStructures := prs.NewKnownStructures()
	knw.RegisterAllKnownStructures(knownStructures)
	// compile hikefile
	fullStartTime := time.Now()
	err := rdr.ReadFile(hikefileName, knownStructures, rootState)
	if err != nil {
		die(err)
	}
	err = rootState.Compile()
	if err != nil {
		die(err)
	}
	// perform dumps
	if dumpStruct {
		for _, artifact := range rootState.KnownArtifacts() {
			nerr = artifact.DumpArtifact(0)
			if nerr != nil {
				fmt.Fprintln(os.Stderr, "Failed to dump artifacts:", nerr.Error())
				os.Exit(1)
			}
			_, nerr = fmt.Println()
			if nerr != nil {
				fmt.Fprintln(os.Stderr, "Failed to dump artifacts:", nerr.Error())
				os.Exit(1)
			}
		}
	}
	// retrieve goals
	goalNames := flag.Args()
	if len(goalNames) == 0 {
		if noDefaultBuild {
			os.Exit(0)
		}
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
	// build plan
	plan := abs.NewPlan()
	for _, goal := range goals {
		for _, action := range goal.Actions() {
			err = action.Perform(plan)
			if err != nil {
				die(err)
			}
		}
	}
	// execute plan
	stepCount := plan.StepCount()
	stepIndexWidth := numWidth(stepCount)
	planDuration := time.Since(fullStartTime)
	startTime := time.Now()
	for stepIndex, step := range plan.Steps() {
		fmt.Printf("%*d/%d %s\n", stepIndexWidth, stepIndex + 1, stepCount, step.SimpleDescr())
		if !pretend {
			err = step.Perform()
			if err != nil {
				die(err)
			}
		}
	}
	duration := time.Since(startTime)
	switch {
		case stepCount == 0:
			fmt.Println("Nandemonai yo.")
		case !pretend:
			fmt.Printf("Success after %s (+ %s for setup).\n", duration.String(), planDuration.String())
	}
}
