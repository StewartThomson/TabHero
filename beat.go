package main

type Beat interface {
	Position() int
	Value() int8
	Duration() int
}

type Beats struct {
	beatSet           []Beat
	unstructuredNotes []Note
	globalRange       int8
	numRealBeats      int
	setLength         int
}

func (b *Beats) AppendToChord(note Note) {
	chord, ok := b.beatSet[len(b.beatSet)-1].(Chord)
	if ok {
		chord.notes = append(chord.notes, note)
	} else {
		//Take previous note and make a chord from it
		prevNote, ok := b.beatSet[len(b.beatSet)-1].(Note)
		if !ok {
			panic("Something other than a chord or note appeared!")
		}
		b.beatSet[len(b.beatSet)-1] = Chord{
			notes: []Note{
				prevNote,
				note,
			},
			position: note.position,
		}
	}
}

func (b *Beats) AddNote(note Note) {
	b.beatSet = append(b.beatSet, note)
	b.numRealBeats++
}

type Note struct {
	value    int8
	duration int
	position int
}

func (n Note) Value() int8 {
	return n.value
}

func (n Note) Position() int {
	return n.position
}

func (n Note) Duration() int {
	return n.duration
}

type Chord struct {
	notes    []Note
	position int
}

func (c Chord) Value() int8 {
	return c.notes[0].value
}

func (c Chord) Position() int {
	return c.position
}

func (c Chord) Duration() int {
	return c.notes[0].duration
}

type Barrier struct{}

func (Barrier) Value() int8 {
	return -1
}

func (Barrier) Position() int {
	return -1
}

func (Barrier) Duration() int {
	return -1
}
