// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/gombz"
	"github.com/tbogdala/groggy"
)

// Skeleton contains data for Bones and all of the matrix transforms
// for them based on animations.
type Skeleton struct {
	// Bones is a slice of Bone objects defined for the mesh. This
	// will not be modified by the Skeleton methods and can be shared
	// between many instances of Skeleton.
	Bones []gombz.Bone

	// Animations is a slice of Animation objects that are compatible
	// with the Bones of the Skeleton.
	Animations []gombz.Animation

	// PoseTransforms are the final updated matrices for each bone
	// that represents it current position, scale and rotation in the pose.
	// This is the matrix stack that can get passed to OpenGL to transform vertices.
	PoseTransforms []mgl.Mat4

	// localTransforms are the updated matrices for each bone that
	// is relative to the parent. These matrices are local to the skeleton
	// because they are specific to the last calcualted animation.
	localTransforms []mgl.Mat4

	// globalTransforms are the updated matrices for each bone that
	// represents the global position of the bone in an animation.
	// They are local to the skeleton since it depends on the last calculated
	// animation.
	globalTransforms []mgl.Mat4
}

// NewSkeleton creates a new Skeleton that shares a bones slice.
func NewSkeleton(bones []gombz.Bone, animations []gombz.Animation) *Skeleton {
	// create a new object and store a reference to the bones
	skel := new(Skeleton)
	skel.Bones = bones
	skel.Animations = animations

	// create the transform slices based on the number of bones passed in.
	boneCount := len(bones)
	skel.localTransforms = make([]mgl.Mat4, boneCount)
	skel.globalTransforms = make([]mgl.Mat4, boneCount)
	skel.PoseTransforms = make([]mgl.Mat4, boneCount)

	// setup the transforms with idenity matrixes
	for i := range skel.PoseTransforms {
		skel.PoseTransforms[i] = mgl.Ident4()
	}

	return skel
}

// Animate interpolates the animation at the given time then calculates
// the bone transformation matrixes.
func (skel *Skeleton) Animate(animation *gombz.Animation, time float32) {
	// sanity checks
	if animation == nil {
		return
	}

	skel.updateLocalTransforms(animation, time)
	skel.updateGlobalTransforms()
	skel.updatePoseTransforms(animation)
}

// getAnimationChannel returns the Channel for a given bone id or nil on error.
func getAnimationChannel(animation *gombz.Animation, boneId int32) *gombz.AnimationChannel {
	for _, c := range animation.Channels {
		if c.BoneId == boneId {
			return &c
		}
	}

	return nil
}

func interpolateKeyVec3(keys []gombz.AnimationVec3Key, time float32) mgl.Vec3 {
	// if there's only one key, just return it
	if len(keys) == 1 {
		return keys[0].Key
	}

	// Note: at this point, there should be more than one key so we should
	// be able to interpolate between two of them.

	// find the first key index that has a Time greater than the current animation time ...
	keyForTime := -1
	for aniKeyIndex := 0; aniKeyIndex < len(keys)-1; aniKeyIndex++ {
		// check to see if the current time
		if time < keys[aniKeyIndex+1].Time {
			keyForTime = aniKeyIndex
			break
		}
	}

	// if we didn't find a key with a Time greater then the animation time has
	// overflowed what is defined in the channel -- just return the last key
	if keyForTime == -1 {
		return keys[len(keys)-1].Key
	}

	// get the data to interpolate
	keyTime := keys[keyForTime].Time
	nextKeyTime := keys[keyForTime+1].Time
	key := keys[keyForTime].Key
	nextKey := keys[keyForTime+1].Key

	keyTimeDelta := nextKeyTime - keyTime
	factor := (time - keyTime) / keyTimeDelta

	// get the raw difference between vectors
	diffVec := nextKey.Sub(key)

	// scale the difference based on the current time factor
	scaledVec := diffVec.Mul(factor)

	// add the difference back to the original key for the final result
	interpVec := scaledVec.Add(key)

	return interpVec
}

func interpolateKeyQuat(keys []gombz.AnimationQuatKey, time float32) mgl.Quat {
	// if there's only one key, just return it
	if len(keys) == 1 {
		return keys[0].Key
	}

	// Note: at this point, there should be more than one key so we should
	// be able to interpolate between two of them.

	// find the first key index that has a Time greater than the current animation time ...
	keyForTime := -1
	for aniKeyIndex := 0; aniKeyIndex < len(keys)-1; aniKeyIndex++ {
		// check to see if the current time
		if time < keys[aniKeyIndex+1].Time {
			keyForTime = aniKeyIndex
			break
		}
	}

	// if we didn't find a key with a Time greater then the animation time has
	// overflowed what is defined in the channel -- just return the last key
	if keyForTime == -1 {
		return keys[len(keys)-1].Key
	}

	// get the data to interpolate
	keyTime := keys[keyForTime].Time
	nextKeyTime := keys[keyForTime+1].Time
	key := keys[keyForTime].Key
	nextKey := keys[keyForTime+1].Key

	keyTimeDelta := nextKeyTime - keyTime
	factor := (time - keyTime) / keyTimeDelta

	// for quaternions we can just SLERP them
	return mgl.QuatSlerp(key, nextKey, factor)
}

// updateLocalTransforms updates the localTransforms slice for each bone.
func (skel *Skeleton) updateLocalTransforms(animation *gombz.Animation, time float32) {
	for bi, bone := range skel.Bones {
		// get the correct channel
		channel := getAnimationChannel(animation, bone.Id)
		if channel == nil {
			groggy.Logsf("DEBUG", "updateLocalTransforms couldn't find a channel for bone %s", bone.Name)
			continue
		}

		// if there's no channel for the bone, then it doesn't get animated
		if channel == nil {
			skel.localTransforms[bi] = bone.Transform
		} else {
			// we have the channel so interpolate the scale, position and rotation keys
			scale := interpolateKeyVec3(channel.ScaleKeys, time)
			position := interpolateKeyVec3(channel.PositionKeys, time)
			rotation := interpolateKeyQuat(channel.RotationKeys, time)

			// now build up the local transform matrix for the bone
			rotMat := rotation.Mat4()
			posMat := mgl.Translate3D(position[0], position[1], position[2])
			scaleMat := mgl.Scale3D(scale[0], scale[1], scale[2])
			//skel.localTransforms[bi] = rotMat.Mul4(posMat).Mul4(scaleMat)
			skel.localTransforms[bi] = posMat.Mul4(rotMat).Mul4(scaleMat)
		}

	}
}

func (skel *Skeleton) updateGlobalTransforms() {
	for bi, bone := range skel.Bones {
		iter := &bone

		// initialize it with the local transforms
		skel.globalTransforms[bi] = skel.localTransforms[bi]

		// loop while there's a parent id
		for iter.Parent >= 0 {
			//groggy.Logsf("DEBUG", "\t\titer == %s ; iter.Parent == %s", iter.Name, skel.Bones[iter.Parent].Name)
			skel.globalTransforms[bi] = skel.localTransforms[iter.Parent].Mul4(skel.globalTransforms[bi])
			iter = &skel.Bones[iter.Parent]
		}
	}
}

func (skel *Skeleton) buildPoseTransforms(animation *gombz.Animation, b *gombz.Bone) {
	skel.PoseTransforms[b.Id] = animation.Transform.Mul4(skel.globalTransforms[b.Id].Mul4(b.Offset))

	// loop through all of the bones to find child bones of this one
	for _, possibleChild := range skel.Bones {
		// does the child's parent match this bone's id (and the parent isn't the child itself)
		if possibleChild.Parent == b.Id && possibleChild.Parent != possibleChild.Id {
			// recurse down the bone textureIndex
			//groggy.Logsf("DEBUG", "\tbuildPoseTransforms recursing. Bone=%s, Child=%s", b.Name, possibleChild.Name)
			skel.buildPoseTransforms(animation, &possibleChild)
		}
	}
}

func (skel *Skeleton) updatePoseTransforms(animation *gombz.Animation) {
	// loop through all of the root level bones
	for _, bone := range skel.Bones {
		if bone.Parent == -1 {
			skel.buildPoseTransforms(animation, &bone)
		}
	}
}
