package main

import (
	"github.com/go-audio/midi"
	"math"
	"os"
)

type MidiInfo struct {
	beats       Beats
	tickRate    uint16
	highestNote int8
	lowestNote  int8
	tempos      []Tempo
	timeSigs    []TimeSignature
	noteRange   int8
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

	info.highestNote = math.MinInt8
	info.lowestNote = math.MaxInt8

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

	info.beats = Beats{
		numRealBeats: 0,
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
			pitch := int8(ev.Note)
			if pitch > info.highestNote {
				info.highestNote = pitch
			}
			if pitch < info.lowestNote {
				info.lowestNote = pitch
			}
			note := Note{
				value:    pitch,
				duration: int(distance * 0.85),
				position: int(noteStatus[ev.Note] - tickOffset),
			}
			if len(info.beats.beatSet) > 0 {
				if info.beats.beatSet[len(info.beats.beatSet)-1].Position() == int(noteStatus[ev.Note]-tickOffset) {
					info.beats.AppendToChord(note)
				} else {
					info.beats.AddNote(note)
				}
			} else {
				info.beats.AddNote(note)
			}
			info.beats.unstructuredNotes = append(info.beats.unstructuredNotes, note)
		}
	}
	info.noteRange = info.highestNote - info.lowestNote
	info.beats.globalRange = info.noteRange
	return
}
