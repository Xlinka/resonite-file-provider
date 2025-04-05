package animxmaker

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MAGIC_STRING  = "AnimX"
	ANIMX_VERSION = int32(1)
)

type KeyFrame struct {
	Time  float32
	Value string
}

type AnimationTrack struct {
	Node      string
	Property  string
	KeyFrames []KeyFrame
}

type Animx struct {
	TrackCount     uint32
	GlobalDuration float32
	Name           string
	Tracks         []AnimationTrack
}

func write7BitEncodedLength(w *bytes.Buffer, length int) error {
	for {
		b := byte(length & 0x7F)
		length >>= 7
		if length > 0 {
			b |= 0x80
		}
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
		if length == 0 {
			break
		}
	}
	return nil
}

func write7BitEncodedString(w *bytes.Buffer, str string) error {
	if err := write7BitEncodedLength(w, len(str)); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}

func (a *Animx) EncodeBinary() ([]byte, error) {
	var buf bytes.Buffer

	// Little endian for binary encoding
	order := binary.LittleEndian

	// Magic String
	if err := write7BitEncodedString(&buf, MAGIC_STRING); err != nil {
		return nil, fmt.Errorf("writing magic string: %w", err)
	}

	// Version
	if err := binary.Write(&buf, order, ANIMX_VERSION); err != nil {
		return nil, fmt.Errorf("writing version: %w", err)
	}

	// Track count
	if err := write7BitEncodedLength(&buf, int(a.TrackCount)); err != nil {
		return nil, fmt.Errorf("writing track count: %w", err)
	}

	// Global duration
	if err := binary.Write(&buf, order, a.GlobalDuration); err != nil {
		return nil, fmt.Errorf("writing duration: %w", err)
	}

	// Animation name
	if err := write7BitEncodedString(&buf, a.Name); err != nil {
		return nil, fmt.Errorf("writing name: %w", err)
	}

	// Encoding type (0)
	if err := binary.Write(&buf, order, byte(0)); err != nil {
		return nil, fmt.Errorf("writing encoding type: %w", err)
	}

	for _, track := range a.Tracks {
		// Track type
		if err := binary.Write(&buf, order, byte(1)); err != nil {
			return nil, fmt.Errorf("writing track type: %w", err)
		}

		// Value type
		if err := binary.Write(&buf, order, byte(39)); err != nil {
			return nil, fmt.Errorf("writing value type: %w", err)
		}

		if err := write7BitEncodedString(&buf, track.Node); err != nil {
			return nil, fmt.Errorf("writing track node: %w", err)
		}
		if err := write7BitEncodedString(&buf, track.Property); err != nil {
			return nil, fmt.Errorf("writing track property: %w", err)
		}

		if err := write7BitEncodedLength(&buf, len(track.KeyFrames)); err != nil {
			return nil, fmt.Errorf("writing keyframe count: %w", err)
		}

		for _, key := range track.KeyFrames {
			if err := binary.Write(&buf, order, key.Time); err != nil {
				return nil, fmt.Errorf("writing keyframe time: %w", err)
			}
			if err := binary.Write(&buf, order, byte(1)); err != nil {
				return nil, fmt.Errorf("writing useless number: %w", err)
			}
			if err := write7BitEncodedString(&buf, key.Value); err != nil {
				return nil, fmt.Errorf("writing keyframe value: %w", err)
			}
		}
	}

	return buf.Bytes(), nil
}


func ListTrack[T any](list []T, name string) AnimationTrack {
	keyframes := make([]KeyFrame, len(list))
	for i, val := range list {
		keyframes[i] = KeyFrame{
			Time:  float32(i),
			Value: fmt.Sprintf("%v", val),
		}
	}
	return AnimationTrack{
		Node:      "default_node",
		Property:  "values",
		KeyFrames: keyframes,
	}
}
