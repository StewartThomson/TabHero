package main

import (
	"github.com/StewartThomson/ChartoGopher"
	"github.com/montanaflynn/stats"
)

const (
	DISABLE_OPENS = 1
)

// -1 denotes barrier in a pattern
func CreateChartFromMidi(patterns []Pattern, info MidiInfo) (chart ChartoGopher.Chart, err error) {
	chart, err = ChartoGopher.NewChart(ChartoGopher.SongInfo{
		SongName:    "Blah",
		Artist:      "Blah",
		MusicStream: "guitar.wav",
		Resolution:  int(info.tickRate),
	}, info.tempos[0].bpm, info.timeSigs[0].numerator, info.timeSigs[0].denominator)
	if err != nil {
		return
	}

	for _, tempo := range info.tempos {
		chart.AddTempoChange(tempo.bpm, tempo.position)
	}

	for _, timeSig := range info.timeSigs {
		err = chart.AddTimeSignatureChange(timeSig.numerator, timeSig.denominator, timeSig.position)
		if err != nil {
			return
		}
	}

	expertGuitar := ChartoGopher.NewTrack(ChartoGopher.DIFF_EXPERT, ChartoGopher.INSTR_GUITAR)
	chart.AddTrack(expertGuitar)

	//Building buckets based on note distribution
	percentiles := buildPercentiles(getNotesAsFloats(info.beats.unstructuredNotes), 5-DISABLE_OPENS)

	for _, pattern := range patterns {
		ProcessPattern(pattern, expertGuitar, percentiles)
	}

	return chart, nil
}

func buildPercentiles(notes []float64, numPercentiles int) (percentiles []float64) {
	div := 100.0 / float64(numPercentiles+1)

	for i := 1; i <= numPercentiles; i++ {
		percentile, err := stats.Percentile(notes, div*float64(i))
		if err == nil {
			percentiles = append(percentiles, percentile)
		}
	}
	return
}

func getNotesAsFloats(notes []Note) (fnotes []float64) {
	for _, n := range notes {
		if n.value != -1 {
			fnotes = append(fnotes, float64(n.value))
		}
	}
	return
}

func AssignColours(pattern Pattern, chartTrack *ChartoGopher.Track, startingPos, incr int) {
	colorMap := map[int8]int{}
	copyToSort := pattern.GetSortedCopy()
	i := startingPos
	for _, beat := range copyToSort.beats {
		if _, ok := colorMap[beat.Value()]; !ok {
			colorMap[beat.Value()] = i
			i += incr
		}
	}
	for _, beat := range pattern.beats {
		var noteColor ChartoGopher.Button
		switch colorMap[beat.Value()] {
		case 0:
			noteColor = ChartoGopher.BTN_OPEN
			break
		case 1:
			noteColor = ChartoGopher.BTN_GREEN
			break
		case 2:
			noteColor = ChartoGopher.BTN_RED
			break
		case 3:
			noteColor = ChartoGopher.BTN_YELLOW
			break
		case 4:
			noteColor = ChartoGopher.BTN_BLUE
			break
		case 5:
			noteColor = ChartoGopher.BTN_ORANGE
			break
		}
		chartTrack.AddNote(beat.Position(), noteColor, beat.Duration(), false, false)
	}
}

//The over-arching song is divided into 6 "buckets" to get a good pitch reference
func findBucketResult(toFind int8, percentiles []float64) int {
	if InRange(float64(toFind), 0, percentiles[0]) {
		return 0
	}
	for i := 1; i < len(percentiles); i++ {
		if InRange(float64(toFind), percentiles[i-1], percentiles[i]) {
			return i
		}
	}

	return len(percentiles)
}

func ProcessPattern(pattern Pattern, chartTrack *ChartoGopher.Track, percentiles []float64) {
	numUniques := pattern.numUniques
	//Only one way to do this if every spot is taken
	if numUniques == 6-DISABLE_OPENS {
		AssignColours(pattern, chartTrack, 0+DISABLE_OPENS, 1)
	} else {
		//Determine the lowest colour that the highest note can be
		lowestPossibleMax := -1 + numUniques
		maxBucketResult := findBucketResult(pattern.max, percentiles)
		for maxBucketResult < lowestPossibleMax {
			maxBucketResult++
		}

		//Determine the highest colour that the lowest note can be
		highestPossibleMin := 6 - DISABLE_OPENS - numUniques
		minBucketResult := findBucketResult(pattern.min, percentiles)
		for minBucketResult > highestPossibleMin {
			minBucketResult--
		}

		//Make sure everything fits
		mover := 0
		for maxBucketResult-minBucketResult < numUniques-1 {
			move := mover % 2
			mover++
			switch move {
			case 0:
				if maxBucketResult < 5 {
					maxBucketResult++
				} else {
					minBucketResult--
				}
				break
			case 1:
				if minBucketResult > 1+DISABLE_OPENS {
					minBucketResult--
				} else {
					maxBucketResult++
				}
				break
			}
		}
		incr := (maxBucketResult - minBucketResult + 1) / numUniques
		AssignColours(pattern, chartTrack, minBucketResult+DISABLE_OPENS, incr)
	}
}
