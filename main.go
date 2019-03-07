package main

import (
	"github.com/StewartThomson/ChartoGopher/chart_writer"
	"os"
)

type Note struct {
	value    int
	duration int
	position int
}

type Notes struct {
	noteSet      []Note
	globalRange  int
	numRealNotes int
	setLength    int
}

func main() {
	songDir := "./toec/"
	filename := songDir + "gp.mid"

	info, err := GetMidiNotes(filename, 0)
	if err != nil {
		panic(err)
	}

	patternDenoted, err := PatternizeNotes(info.notes)
	if err != nil {
		panic(err)
	}

	patterns, err := GetPatternsFromNoteSet(patternDenoted)
	if err != nil {
		panic(err)
	}

	chart, err := CreateChartFromMidi(patterns, info)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(songDir + "notes.chart"); !os.IsNotExist(err) {
		err = os.Remove(songDir + "notes.chart")
		if err != nil {
			panic(err)
		}
	}
	f, err := os.OpenFile(songDir+"notes.chart", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := chart_writer.New(f)

	_, err = chart.Write(writer)
	if err != nil {
		panic(err)
	}
}
