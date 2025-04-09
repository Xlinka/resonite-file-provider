package tests

import (
	"os"
	"resonite-file-provider/animxmaker"
)

func bllrpypackage() []byte {
	keyframes := []animxmaker.KeyFrame[string]{
		animxmaker.KeyFrame[string]{
			Position: 0,
			Value: "FirstFrame",
		},
		animxmaker.KeyFrame[string]{
			Position: 1,
			Value: "SecondFrame",
		},
	}

	track := animxmaker.AnimationTrack[string]{
		Node: "FirstTrack",
		Property: "FirstTrackProperty",
		Keyframes: keyframes,
	}
	track2 := animxmaker.AnimationTrack[string]{
		Node: "SecondTrack",
		Property: "SecondTrackProperty",
		Keyframes: keyframes,
	}
	animation := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			animxmaker.AnimationTrackWrapper(&track),
			animxmaker.AnimationTrackWrapper(&track2),
		},
	}
	return animation.EncodeAnimation("test")
	

}
func helperTest() []byte {
	listtrack := animxmaker.ListTrack[int]([]int{1,76,2,4,6,1}, "listtrack", "int")
	stringtrack := animxmaker.ListTrack[string]([]string{"a", "b", "c"}, "listtrack", "string")
	animation := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			animxmaker.AnimationTrackWrapper(&listtrack),
			animxmaker.AnimationTrackWrapper(&stringtrack),
		},
	}
	return animation.EncodeAnimation("test")
}
func Main(){
	animationBytes := helperTest()
	os.WriteFile("listTrack.animx", animationBytes, 0644)
}
