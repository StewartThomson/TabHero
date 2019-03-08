package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

type Pattern struct {
	beats      []Beat
	numUniques int
	min        int8
	max        int8
	uniqueSum  int
}

//For Pattern to implement Sort interface
func (pattern Pattern) Len() int {
	return len(pattern.beats)
}
func (pattern Pattern) Swap(i, j int) {
	pattern.beats[i], pattern.beats[j] = pattern.beats[j], pattern.beats[i]
}
func (pattern Pattern) Less(i, j int) bool {
	return pattern.beats[i].Value() < pattern.beats[j].Value()
}

//Divide the noteset into patterns. -1 denotes a pattern barrier
func GetPatternsFromNoteSet(beats Beats) (patterns []Pattern, err error) {
	beatSet := beats.beatSet
	numUniques := 0
	uniqueSum := 0
	var noteMap = map[int8]bool{}
	patterns = make([]Pattern, 0)
	var note Note
	var pattern Pattern
	pattern = Pattern{
		beats: make([]Beat, 0),
		min:   math.MaxInt8,
		max:   math.MinInt8,
	}

	for i := 0; i < beats.setLength; i++ {
		beat := beatSet[i]
		if beat.Value() == -1 {
			if len(pattern.beats) != 0 {
				if numUniques > 6 {
					err = errors.New(fmt.Sprintf("pattern found with %d unique beats", numUniques))
					return
				}
				pattern.numUniques = numUniques
				pattern.uniqueSum = uniqueSum
				patterns = append(patterns, pattern)
				pattern = Pattern{
					beats: make([]Beat, 0),
					min:   math.MaxInt8,
					max:   math.MinInt8,
				}
				numUniques = 0
				uniqueSum = 0
				noteMap = map[int8]bool{}
			}
		} else {
			if _, ok := noteMap[beat.Value()]; !ok {
				if beat.Value() < pattern.min {
					pattern.min = beat.Value()
				}
				if beat.Value() > pattern.max {
					pattern.max = beat.Value()
				}
				numUniques++
				uniqueSum += int(beat.Value())
				noteMap[beat.Value()] = true
			}
			pattern.beats = append(pattern.beats, beat)
		}
	}

	if note.value != -1 {
		if numUniques > 6 {
			err = errors.New(fmt.Sprintf("pattern found with %d unique beats", numUniques))
			return
		}
		if len(pattern.beats) != 0 {
			pattern.numUniques = numUniques
			pattern.uniqueSum = uniqueSum
			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

func (pattern Pattern) GetSortedCopy() Pattern {
	copyToSort := Pattern{
		beats: make([]Beat, len(pattern.beats)),
	}
	copy(copyToSort.beats, pattern.beats)
	sort.Sort(copyToSort)

	return copyToSort
}
