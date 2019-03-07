package main

import (
	"github.com/go-audio/midi"
	"golang.org/x/tools/container/intsets"
	"os"
)

type MidiInfo struct {
	notes       Notes
	tickRate    uint16
	highestNote int
	lowestNote  int
	tempos      []Tempo
	timeSigs    []TimeSignature
	noteRange   int
}

type Tempo struct {
	bpm      int
	position int
}

type TimeSignature struct {
	numerator   int
	denominator int
	position    int
}

func GetMidiNotes(filename string, trackToParse int) (info MidiInfo, err error) {
	info = MidiInfo{}

	data, err := os.Open(filename)
	if err != nil {
		return
	}
	defer data.Close()

	decoder := midi.NewDecoder(data)
	decoder.Debug = true
	if err = decoder.Decode(); err != nil {
		return
	}

	//chart tick rate is the same as midi tick rate
	info.tickRate = decoder.TicksPerQuarterNote

	//Tempo is always stored in first position??
	for _, ev := range decoder.Tracks[0].Events {
		if ev.Cmd == midi.MetaByteMap["Tempo"] {
			info.tempos = append(info.tempos, Tempo{
				bpm:      int(ev.Bpm),
				position: int(ev.AbsTicks),
			})
		}
	}

	info.highestNote = intsets.MinInt
	info.lowestNote = intsets.MaxInt

	track := decoder.Tracks[trackToParse]
	tickOffset := decoder.Tracks[trackToParse-1].Events[len(decoder.Tracks[trackToParse-1].Events)-1].AbsTicks

	for _, ev := range decoder.Tracks[1].Events {
		if ev.Cmd == midi.MetaByteMap["Time Signature"] {
			info.timeSigs = append(info.timeSigs, TimeSignature{
				numerator:   int(ev.TimeSignature.Numerator),
				denominator: int(ev.TimeSignature.Denum()),
				position:    int(ev.AbsTicks),
			})
		}
	}

	info.notes = Notes{
		noteSet:      make([]Note, 0),
		numRealNotes: 0,
	}
	noteStatus := map[uint8]uint64{}
	for _, ev := range track.Events {
		if ev.MsgType == midi.EventByteMap["NoteOn"] {
			noteStatus[ev.Note] = ev.AbsTicks
		}
		if ev.MsgType == midi.EventByteMap["NoteOff"] {
			distance := float64(ev.AbsTicks - noteStatus[ev.Note])
			if distance <= float64(info.tickRate/2) {
				distance = 0
			}
			note := int(ev.Note)
			if note > info.highestNote {
				info.highestNote = note
			}
			if note < info.lowestNote {
				info.lowestNote = note
			}
			info.notes.noteSet = append(info.notes.noteSet, Note{
				value:    note,
				duration: int(distance * 0.85),
				position: int(noteStatus[ev.Note] - tickOffset),
			})
			info.notes.numRealNotes++
		}
	}
	info.noteRange = info.highestNote - info.lowestNote
	info.notes.globalRange = info.noteRange
	return
}
