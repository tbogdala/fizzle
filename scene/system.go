// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package scene

// System describes the basic interface for a 'controller' type object
// that gets executed from the scene manager.
type System interface {
	// Update should get called to run updates for the system every frame
	// by the owning Manager object.
	Update(frameDelta float32)

	// OnAddEntity should get called by the scene Manager each time a new entity
	// has been added to the scene.
	OnAddEntity(newEntity Entity)

	// OnRemoveEntity should get called by the scene Manager each time an entity
	// has been removed from the scene.
	OnRemoveEntity(oldEntity Entity)

	// GetRequestedPriority returns the requested priority level for the System
	// which may be of significance to a Manager if they want to order Update() calls.
	GetRequestedPriority() float32

	// GetName returns the name of the system that can be used to identify
	// the System within Manager.
	GetName() string
}

// SystemsByPriority is a type alias that will implement sort.Interface to sort
// the slice of Systems by priority.
type SystemsByPriority []System

// Len is the length of the slice.
func (s SystemsByPriority) Len() int {
	return len(s)
}

// Swap changes the values at the two indices.
func (s SystemsByPriority) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns true if the priority of system i is less than system j.
func (s SystemsByPriority) Less(i, j int) bool {
	return s[i].GetRequestedPriority() < s[j].GetRequestedPriority()
}
