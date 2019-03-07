package main

import (
	"errors"
	"fmt"
	"golang.org/x/tools/container/intsets"
	"sort"
)

type Pattern struct {
	notes      []Note
	numUniques int
	min        int
	max        int
	uniqueSum  int
}

//For Pattern to implement Sort interface
func (pattern Pattern) Len() int {
	return len(pattern.notes)
}
func (pattern Pattern) Swap(i, j int) {
	pattern.notes[i], pattern.notes[j] = pattern.notes[j], pattern.notes[i]
}
func (pattern Pattern) Less(i, j int) bool {
	return pattern.notes[i].value < pattern.notes[j].value
}

//Divide the noteset into patterns. -1 denotes a pattern barrier
func GetPatternsFromNoteSet(notes Notes) (patterns []Pattern, err error) {
	noteSet := notes.noteSet
	numUniques := 0
	uniqueSum := 0
	var noteMap = map[int]bool{}
	patterns = make([]Pattern, 0)
	var note Note
	var pattern Pattern
	pattern = Pattern{
		notes: make([]Note, 0),
		min:   intsets.MaxInt,
		max:   intsets.MinInt,
	}

	for i := 0; i < notes.setLength; i++ {
		note := noteSet[i]
		if note.value == -1 {
			if len(pattern.notes) != 0 {
				if numUniques > 6 {
					err = errors.New(fmt.Sprintf("pattern found with %d unique notes", numUniques))
					return
				}
				pattern.numUniques = numUniques
				pattern.uniqueSum = uniqueSum
				patterns = append(patterns, pattern)
				pattern = Pattern{
					notes: make([]Note, 0),
					min:   intsets.MaxInt,
					max:   intsets.MinInt,
				}
				numUniques = 0
				uniqueSum = 0
				noteMap = map[int]bool{}
			}
		} else {
			if _, ok := noteMap[note.value]; !ok {
				if note.value < pattern.min {
					pattern.min = note.value
				}
				if note.value > pattern.max {
					pattern.max = note.value
				}
				numUniques++
				uniqueSum += note.value
				noteMap[note.value] = true
			}
			pattern.notes = append(pattern.notes, note)
		}
	}

	if note.value != -1 {
		if numUniques > 6 {
			err = errors.New(fmt.Sprintf("pattern found with %d unique notes", numUniques))
			return
		}
		if len(pattern.notes) != 0 {
			pattern.numUniques = numUniques
			pattern.uniqueSum = uniqueSum
			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

func (pattern Pattern) GetSortedCopy() Pattern {
	copyToSort := Pattern{
		notes: make([]Note, len(pattern.notes)),
	}
	copy(copyToSort.notes, pattern.notes)
	sort.Sort(copyToSort)

	return copyToSort
}
