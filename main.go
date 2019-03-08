package main

import (
	"github.com/StewartThomson/ChartoGopher/chart_writer"
	"os"
)

func main() {
	songDir := "./toec/"
	filename := songDir + "gp.mid"

	info, err := GetMidiNotes(filename, 1)
	if err != nil {
		panic(err)
	}

	patternDenoted, err := PatternizeNotes(info.beats)
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

	if _, err := os.Stat(songDir + "beats.chart"); !os.IsNotExist(err) {
		err = os.Remove(songDir + "beats.chart")
		if err != nil {
			panic(err)
		}
	}
	f, err := os.OpenFile(songDir+"beats.chart", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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
