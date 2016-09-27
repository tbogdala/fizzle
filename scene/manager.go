// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package scene

import (
	"sort"
)

// Manager is an interface for scene Manager objects that contain systems and
// entities and helps coordinate how those interact.
type Manager interface {
	// AddEntity should be called to register an entity with the scene Manager.
	// The Manager should also inform all Systems it knows of with OnAddEntity().
	AddEntity(newEntity Entity)

	// RemoveEntity should be called to unregister an entity with the scene Manager.
	// The Manager should also inform all Systems it knows of with OnRemoveEntity().
	RemoveEntity(oldEntity Entity)

	// AddSystem adds a new system to the scene manager.
	AddSystem(newSystem System)

	// RemoveSystem removes a system from the scene manager.
	RemoveSystem(oldSystem System)

	// GetSystemByName returns a System, if found, that matches the name supplied.
	GetSystemByName(name string) System

	// Update should be called each frame to update the scene manager.
	Update(frameDelta float32)

	// GetNextID should return the next ID integer available. This is not necessarilly
	// be guaranteed to be unique (a naive implemention simply increments it as the
	// overflow case will take a significant amount of time to get to under likely loads).
	GetNextID() uint64
}

// BasicSceneManager is a basic scene manager to serve as a default impelmentation
// of the Manager interface.
type BasicSceneManager struct {
	// entities are all of the tracked Entities keyed by entity ID.
	entities map[uint64]Entity

	// systems are all of the tracked Systems keyed by system Name.
	systems map[string]System

	// sortedSystems is a ordered slice of Systems ordered by increasing priority.
	// This slice is derived from the systems map.
	sortedSystems []System

	// nextID is the next ID number to return on request.
	nextID uint64
}

// NewBasicSceneManager creates a new BasicSceneManager manager object
// and initializes internal data structures.
func NewBasicSceneManager() *BasicSceneManager {
	sm := new(BasicSceneManager)
	sm.entities = make(map[uint64]Entity)
	sm.systems = make(map[string]System)
	sm.sortedSystems = []System{}
	return sm
}

// mapSystems calls the supplied function with each system in sorted order.
func (sm *BasicSceneManager) mapSystems(cb func(s System)) {
	if len(sm.sortedSystems) < 1 || cb == nil {
		return
	}
	for _, system := range sm.sortedSystems {
		cb(system)
	}
}

// GetNextID should return the next ID integer available. This is not necessarilly
// be guaranteed to be unique -- but it will be until the uint64 overflows after
// roughly 1.8e+19 increments.
func (sm *BasicSceneManager) GetNextID() uint64 {
	id := sm.nextID
	sm.nextID++
	return id
}

// AddEntity should be called to register an entity with the scene Manager.
// The Manager should also inform all Systems it knows of with OnAddEntity().
func (sm *BasicSceneManager) AddEntity(newEntity Entity) {
	if newEntity != nil {
		sm.entities[newEntity.GetID()] = newEntity
	}
	sm.mapSystems(func(s System) {
		s.OnAddEntity(newEntity)
	})
}

// RemoveEntity should be called to unregister an entity with the scene Manager.
// The Manager should also inform all Systems it knows of with OnRemoveEntity().
func (sm *BasicSceneManager) RemoveEntity(oldEntity Entity) {
	if oldEntity != nil {
		delete(sm.entities, oldEntity.GetID())
	}
	sm.mapSystems(func(s System) {
		s.OnRemoveEntity(oldEntity)
	})
}

// AddSystem adds a new system to the scene manager.
func (sm *BasicSceneManager) AddSystem(newSystem System) {
	if newSystem != nil {
		sm.systems[newSystem.GetName()] = newSystem
		sm.sortedSystems = append(sm.sortedSystems, newSystem)
		sort.Sort(SystemsByPriority(sm.sortedSystems))
	}
}

// RemoveSystem removes a system from the scene manager.
func (sm *BasicSceneManager) RemoveSystem(oldSystem System) {
	if oldSystem != nil {
		delete(sm.systems, oldSystem.GetName())

		// instead of filtering out the system being removed, we'll create
		// a new slice instead. The amount of systems in are typically few in number.
		sm.sortedSystems = make([]System, len(sm.systems))
		for _, s := range sm.systems {
			sm.sortedSystems = append(sm.sortedSystems, s)
		}
		sort.Sort(SystemsByPriority(sm.sortedSystems))
	}
}

// GetSystemByName returns a System, if found, that matches the name supplied.
func (sm *BasicSceneManager) GetSystemByName(name string) System {
	s, found := sm.systems[name]
	if !found {
		return nil
	}
	return s
}

// Update should be called each frame to update the scene manager.
func (sm *BasicSceneManager) Update(frameDelta float32) {
	// call Update on all systems
	sm.mapSystems(func(s System) {
		s.Update(frameDelta)
	})
}
