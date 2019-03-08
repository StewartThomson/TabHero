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

//Provide a fitness. Patterns with >6 beats are severely punished. Longer & more diverse patterns are rewarded
//Occurrences of beats that are similar on either side of a barrier are only lightly punished
func (b Beats) Evaluate() (float64, error) {
	fitness := 0.0
	numUniques := 0
	sumOfUniques := 0
	niceToHave := 0.0
	reward := 0
	numTotal := 0
	numBarriers := 0
	var pitchMap = map[int8]bool{}
	var pitch int8

	for i := 0; i < b.setLength; i++ {
		pitch = b.beatSet[i].Value()
		if pitch == -1 {
			numBarriers++
			if numUniques != 0 {
				if numUniques > MAX_UNIQUES {
					fitness += math.Pow(float64(numUniques+sumOfUniques), 3)
				} else {
					reward -= numTotal * numUniques
				}
				if i != 0 && i != b.setLength-1 {
					back := b.beatSet[i-1].Value()
					forward := b.beatSet[i+1].Value()
					j := i + 1
					for forward == -1 && j != b.setLength-1 {
						forward = b.beatSet[j].Value()
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
				pitchMap = map[int8]bool{}
				numTotal = 0
			}
		} else {
			if _, ok := pitchMap[pitch]; !ok {
				numUniques++
				sumOfUniques += int(pitch)
				pitchMap[pitch] = true
			}
			numTotal++
		}
	}
	if pitch != -1 {
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
func (b Beats) Mutate(rng *rand.Rand) {
	for i := 0; i < b.setLength; i++ {
		if b.beatSet[i].Value() == -1 {
			direction := rng.Intn(6)
			velocity := rng.Intn(1) + 1
			pos := i
			//move forward
			if direction == 0 {
				for pos != b.setLength-1 && velocity != 0 {
					//swap
					b.beatSet[pos], b.beatSet[pos+1] = b.beatSet[pos+1], b.beatSet[pos]
					pos++
					velocity--
				}
			} else if direction == 1 { //move backward
				for pos != 0 && velocity != 0 {
					//swap
					b.beatSet[pos], b.beatSet[pos-1] = b.beatSet[pos-1], b.beatSet[pos]
					pos--
					velocity--
				}
			} else if direction == 2 { //delete
				//if pos != b.setLength-1 {
				//	b.beatSet = append(b.beatSet[:pos], b.beatSet[pos+1:]...)
				//	b.setLength--
				//}
			}
		}
	}
}

//Can't really be implemented here.
func (b Beats) Crossover(Y eaopt.Genome, rng *rand.Rand) {}

func (b Beats) Clone() eaopt.Genome {
	Y := Beats{
		beatSet:      make([]Beat, len(b.beatSet)),
		globalRange:  b.globalRange,
		numRealBeats: b.numRealBeats,
		setLength:    b.setLength,
	}
	copy(Y.beatSet, b.beatSet)
	return Y
}

//Randomly distribute -1 throughout the array. This denotes a barrier between patterns
func NoteFactory(noteArr Beats, rng *rand.Rand) eaopt.Genome {
	numNotes := noteArr.numRealBeats
	mindivs := 0
	numdivs := rng.Intn(numNotes-mindivs) + mindivs
	var arr = make([]Beat, numNotes+numdivs)
	copy(arr, noteArr.beatSet)
	for i := 0; i < numdivs; i++ {
		spot := rng.Intn(len(arr))
		copy(arr[spot+1:], arr[spot:])
		arr[spot] = Barrier{}
	}
	//Some barriers may have been pushed out of the end, just replace them here
	for i := len(arr) - 1; i > 0; i-- {
		if arr[i] == nil {
			arr[i] = Barrier{}
		}
	}
	return Beats{
		beatSet:      arr,
		numRealBeats: noteArr.numRealBeats,
		globalRange:  noteArr.globalRange,
		setLength:    len(arr),
	}
}

//Failed patterns defined as a pattern with over 6 unique beats
func (b Beats) CountFailedPatterns() (int, []int) {
	ret := 0
	numUniques := 0.0
	uniqueSum := 0
	var pitch int8
	var pitchMap = map[int8]bool{}
	var uniqueSums []int
	for i := 0; i < b.setLength; i++ {
		pitch = b.beatSet[i].Value()
		if pitch == -1 {
			if numUniques != 0 {
				if numUniques > MAX_UNIQUES {
					ret += 1
					uniqueSums = append(uniqueSums, int(numUniques))
				}
				numUniques = 0
				uniqueSum = 0
				pitchMap = map[int8]bool{}
			}
		} else {
			if _, ok := pitchMap[pitch]; !ok {
				numUniques++
				uniqueSum += int(pitch)
				pitchMap[pitch] = true
			}
		}
	}
	if pitch != -1 {
		if numUniques > MAX_UNIQUES {
			ret += 1
			uniqueSums = append(uniqueSums, int(numUniques))
		}
	}
	return ret, uniqueSums
}

func PatternizeNotes(X Beats) (patternized Beats, err error) {
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		return
	}
	ga.ParallelEval = false

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

	best := ga.HallOfFame[0].Genome.(Beats)
	a, b := best.CountFailedPatterns()
	fmt.Println("Best ", a, b)

	patternized = best.Clone().(Beats)

	return
}
