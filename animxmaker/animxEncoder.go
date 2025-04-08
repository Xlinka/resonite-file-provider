package animxmaker

import (
	"bytes"
	"encoding/binary"
)

//const (
//	typeIds = {
//		int: 10,
//		float32: 21,
//		string: 27,
//	}
//)
//// 7-bit encoded length prefix


func write7BitEncodedInt(n int) []byte {
	var result []byte
	for {
		// Get the least significant 7 bits
		byteVal := byte(n & 0x7F)
		n >>= 7 // Shift right 7 bits to process the next chunk

		if n != 0 {
			// More bytes to come: set MSB to 1
			byteVal |= 0x80
		}

		// Append this byte to the result
		result = append(result, byteVal)

		if n == 0 {
			break
		}
	}
	return result
}

func encodeAnimString(s string, isValue bool) ([]byte) {
	var buf bytes.Buffer
	if isValue {
		buf.WriteByte(1)
	}
	// Encode length
	binary.Write(&buf, binary.LittleEndian, write7BitEncodedInt(len(s)))
	buf.Write([]byte(s))
	return buf.Bytes()
}
type KeyFrame[T any] struct {
	Position float32
	Value T 
}

func (a *KeyFrame[T]) EncodeKeyframe() []byte{
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, a.Position)
	switch v := any(a.Value).(type) {
			case string:
				buf.Write(encodeAnimString(v, true))
			case int:
				binary.Write(&buf, binary.LittleEndian, v)
			case float32:
				binary.Write(&buf, binary.LittleEndian, v)
			// pain
		}		
	return buf.Bytes()
}

type AnimationTrackWrapper interface {
	EncodeTrack() []byte
	GetTrackDuration() float32
}

type AnimationTrack[T any] struct {
	valueType int 
	testValue T
	TrackName string
	trackType byte // 0 = Raw, 1 = Discrete, 2 = Curve, 3 = Bezier
	trackDuration float32
	Node string
	Property string
	Keyframes []KeyFrame[T]
}

func (a *AnimationTrack[T]) GetTrackDuration() float32 {
	return a.trackDuration
}

func (a *AnimationTrack[T]) EncodeTrack() []byte {
	var buf bytes.Buffer
	buf.WriteByte(a.trackType)
	switch any(a.testValue).(type) {
		case string:
			buf.WriteByte(27) // 27 = String
		case int:
			buf.WriteByte(10) // 10 = Int
		case float32:
			buf.WriteByte(21) // 21 = Float
		// pain again
	}
	buf.Write(encodeAnimString(a.Node, false))
	buf.Write(encodeAnimString(a.Property, false))
	binary.Write(&buf, binary.LittleEndian, make([]byte, 0, len(a.Keyframes)))
	for _, keyframe := range a.Keyframes{
		buf.Write(keyframe.EncodeKeyframe())
	}
	return buf.Bytes()
}

type Animation struct {
	Tracks []AnimationTrackWrapper
}

func (a *Animation) EncodeAnimation(animationName string) []byte {
	var buf bytes.Buffer

	buf.Write(encodeAnimString("AnimX", false))           // "Magic Word"
	buf.WriteByte(1)                     // Version
	buf.WriteByte(byte(len(a.Tracks)))   // Number of tracks
	
	var maxDuration float32 = 0
	for _, track := range a.Tracks {
		if track.GetTrackDuration() > maxDuration {
			maxDuration = track.GetTrackDuration()
		}
	}

	binary.Write(&buf, binary.LittleEndian, maxDuration) // Max duration
	buf.Write(encodeAnimString(animationName, false))     // Animation name
	buf.WriteByte(0)                     // Encoding type (0 is default binary reader)
	for _, track := range a.Tracks {
		buf.Write(track.EncodeTrack())
	}
	
	return buf.Bytes()
}
