package main

import (
	"fmt"
	"github.com/MaxHalford/eaopt"
	"gopkg.in/cheggaaa/pb.v1"
	"math"
	"math/rand"
	"time"
)

const (
	MAX_UNIQUES = 5
)

//Provide a fitness. Patterns with >6 notes are severely punished. Longer & more diverse patterns are rewarded
//Occurrences of notes that are similar on either side of a barrier are only lightly punished
func (X Notes) Evaluate() (float64, error) {
	fitness := 0.0
	numUniques := 0
	sumOfUniques := 0
	note := 0
	niceToHave := 0.0
	reward := 0
	numTotal := 0
	numBarriers := 0
	var noteMap = map[int]bool{}

	for i := 0; i < X.setLength; i++ {
		note = X.noteSet[i].value
		if note == -1 {
			numBarriers++
			if numUniques != 0 {
				if numUniques > MAX_UNIQUES {
					fitness += math.Pow(float64(numUniques+sumOfUniques), 3)
				} else {
					reward -= numTotal * numUniques
				}
				if i != 0 && i != X.setLength-1 {
					back := X.noteSet[i-1].value
					forward := X.noteSet[i+1].value
					j := i + 1
					for forward == -1 && j != X.setLength-1 {
						forward = X.noteSet[j].value
						j++
					}
					if forward != -1 {
						diff := math.Abs(float64(forward - back))
						if diff <= 10 {
							niceToHave += 11 - diff
						}
					}
				}
				numUniques = 0
				sumOfUniques = 0
				noteMap = map[int]bool{}
				numTotal = 0
			}
		} else {
			if _, ok := noteMap[note]; !ok {
				numUniques++
				sumOfUniques += note
				noteMap[note] = true
			}
			numTotal++
		}
	}
	if note != -1 {
		if numUniques > MAX_UNIQUES {
			fitness += math.Pow(float64(numUniques+sumOfUniques), 3)
		} else {
			reward -= numTotal * numUniques
		}
	}
	if fitness <= 0 {
		fitness += niceToHave
		fitness += float64(reward / numBarriers)
	}
	return fitness, nil
}

//1 in 2 chance of mutating each barrier. The barrier is then moved forward/backward/ or deleted
func (X Notes) Mutate(rng *rand.Rand) {
	for i := 0; i < X.setLength; i++ {
		if X.noteSet[i].value == -1 {
			direction := rng.Intn(6)
			velocity := rng.Intn(1) + 1
			pos := i
			//move forward
			if direction == 0 {
				for pos != X.setLength-1 && velocity != 0 {
					//swap
					X.noteSet[pos], X.noteSet[pos+1] = X.noteSet[pos+1], X.noteSet[pos]
					pos++
					velocity--
				}
			} else if direction == 1 { //move backward
				for pos != 0 && velocity != 0 {
					//swap
					X.noteSet[pos], X.noteSet[pos-1] = X.noteSet[pos-1], X.noteSet[pos]
					pos--
					velocity--
				}
			} else if direction == 2 { //delete
				if pos != X.setLength-1 {
					X.noteSet = append(X.noteSet[:pos], X.noteSet[pos+1:]...)
					X.setLength--
				}
			}
		}
	}
}

//Can't really be implemented here.
func (X Notes) Crossover(Y eaopt.Genome, rng *rand.Rand) {}

func (X Notes) Clone() eaopt.Genome {
	Y := Notes{
		noteSet:      make([]Note, len(X.noteSet)),
		globalRange:  X.globalRange,
		numRealNotes: X.numRealNotes,
		setLength:    X.setLength,
	}
	copy(Y.noteSet, X.noteSet)
	return Y
}

//Randomly distribute -1 throughout the array. This denotes a barrier between patterns
func NoteFactory(noteArr Notes, rng *rand.Rand) eaopt.Genome {
	numNotes := noteArr.numRealNotes
	mindivs := 0
	numdivs := rng.Intn(numNotes-mindivs) + mindivs
	var arr = make([]Note, numNotes+numdivs)
	copy(arr, noteArr.noteSet)
	for i := 0; i < numdivs; i++ {
		spot := rng.Intn(numNotes + numdivs - 2)
		copy(arr[spot+1:], arr[spot:])
		arr[spot].value = -1
	}
	return Notes{
		noteSet:      arr,
		numRealNotes: noteArr.numRealNotes,
		globalRange:  noteArr.globalRange,
		setLength:    len(arr),
	}
}

//Failed patterns defined as a pattern with over 6 unique notes
func (X Notes) CountFailedPatterns() (int, []int) {
	ret := 0
	numUniques := 0.0
	uniqueSum := 0
	note := 0
	var noteMap = map[int]bool{}
	var uniqueSums []int
	for i := 0; i < X.setLength; i++ {
		note = X.noteSet[i].value
		if note == -1 {
			if numUniques != 0 {
				if numUniques > MAX_UNIQUES {
					ret += 1
					uniqueSums = append(uniqueSums, int(numUniques))
				}
				numUniques = 0
				uniqueSum = 0
				noteMap = map[int]bool{}
			}
		} else {
			if _, ok := noteMap[note]; !ok {
				numUniques++
				uniqueSum += note
				noteMap[note] = true
			}
		}
	}
	if note != -1 {
		if numUniques > MAX_UNIQUES {
			ret += 1
			uniqueSums = append(uniqueSums, int(numUniques))
		}
	}
	return ret, uniqueSums
}

func PatternizeNotes(X Notes) (patternized Notes, err error) {
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		return
	}
	ga.ParallelEval = true

	//Modify these!
	ga.NGenerations = 750
	ga.NPops = 4
	ga.PopSize = 2500

	genMap := map[float64]bool{}
	//Progress bar
	bar := pb.StartNew(int(ga.NGenerations))
	ga.Callback = func(ga *eaopt.GA) {
		bar.Increment()
		if _, ok := genMap[ga.HallOfFame[0].Fitness]; !ok {
			bar.Prefix(fmt.Sprintf("Best fitness %f", ga.HallOfFame[0].Fitness))
			genMap[ga.HallOfFame[0].Fitness] = true
		}
	}

	start := time.Now()
	err = ga.Minimize(func(rng *rand.Rand) eaopt.Genome {
		return NoteFactory(X, rng)
	})
	if err != nil {
		return
	}
	fmt.Printf("Process took %s\n", time.Since(start))

	best := ga.HallOfFame[0].Genome.(Notes)
	a, b := best.CountFailedPatterns()
	fmt.Println("Best ", a, b)

	patternized = best.Clone().(Notes)

	return
}
