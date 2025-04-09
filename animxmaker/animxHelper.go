package animxmaker

func ListTrack[T any](list []T, nodeName string, propertyName string) AnimationTrack[T]{
	var keyframes []KeyFrame[T]
	for i, item := range list {
		keyframe := KeyFrame[T]{
			Position: float32(i),
			Value: item,
		}
		keyframes = append(keyframes, keyframe)
	}
	return AnimationTrack[T]{
		Keyframes: keyframes,
		Node: nodeName,
		Property: propertyName,
	}
}
