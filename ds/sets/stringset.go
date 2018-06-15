package sets

import (
	"sort"
	"sync"
)

// StringSet operations for string keys
type StringSet interface {
	Insert(items ...string)
	Delete(items ...string)
	Clear() []string
	Replace(items ...string)
	Has(item string) bool
	HasAll(items ...string) bool
	HasAny(items ...string) bool
	IsSuperset(right StringSet) bool
	IsSubset(right StringSet) bool
	Equal(right StringSet) bool
	Diff(right StringSet) StringSet
	Union(right StringSet) StringSet
	Intersection(right StringSet) StringSet
	List() []string
	SortedList() []string
	PopAny() (string, bool)
	Len() int
}

// NewStringSet create a StringSet from a string slice
func NewStringSet(items ...string) StringSet {
	ss := &ssetImpl{
		set:   make(map[string]struct{}),
		mutex: new(sync.RWMutex),
	}
	if len(items) > 0 {
		ss.Insert(items...)
	}
	return ss
}

type ssetImpl struct {
	set   map[string]struct{}
	mutex *sync.RWMutex
}

func (ss *ssetImpl) Insert(items ...string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for _, item := range items {
		ss.set[item] = struct{}{}
	}
}
func (ss *ssetImpl) Delete(items ...string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for _, item := range items {
		delete(ss.set, item)
	}
}

func (ss *ssetImpl) Clear() []string {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	result := make([]string, 0, len(ss.set))
	for item := range ss.set {
		result = append(result, item)
	}
	// use a new map to replace with the old one, and the old one will be gced.
	ss.set = make(map[string]struct{})

	return result
}
func (ss *ssetImpl) Replace(items ...string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// use a new map to replace with the old one, and the old one will be gced.
	ss.set = make(map[string]struct{})
	for _, item := range items {
		ss.set[item] = struct{}{}
	}
}
func (ss *ssetImpl) Has(item string) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	return ss.hasImpl(item)
}
func (ss *ssetImpl) hasImpl(item string) bool {
	_, exists := ss.set[item]
	return exists
}

func (ss *ssetImpl) HasAll(items ...string) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	for _, item := range items {
		if !ss.hasImpl(item) {
			return false
		}
	}
	return true
}
func (ss *ssetImpl) HasAny(items ...string) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	for _, item := range items {
		if ss.hasImpl(item) {
			return true
		}
	}
	return false
}
func (ss *ssetImpl) IsSuperset(right StringSet) bool {
	rightImpl := right.(*ssetImpl)

	ss.mutex.RLock()
	rightImpl.mutex.RLock()
	defer ss.mutex.RUnlock()
	defer rightImpl.mutex.RUnlock()

	for item := range rightImpl.set {
		if !ss.hasImpl(item) {
			return false
		}
	}
	return true
}
func (ss *ssetImpl) IsSubset(right StringSet) bool {
	return right.IsSuperset(ss)
}
func (ss *ssetImpl) Equal(right StringSet) bool {
	return ss.Len() == right.Len() && ss.IsSuperset(right)
}
func (ss *ssetImpl) Diff(right StringSet) StringSet {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	result := NewStringSet()
	for item := range ss.set {
		if !right.Has(item) {
			result.Insert(item)
		}
	}
	return result
}
func (ss *ssetImpl) Union(right StringSet) StringSet {
	rightImpl := right.(*ssetImpl)

	ss.mutex.RLock()
	rightImpl.mutex.RLock()
	defer ss.mutex.RUnlock()
	defer rightImpl.mutex.RUnlock()

	result := NewStringSet()
	for item := range ss.set {
		result.Insert(item)
	}
	for item := range rightImpl.set {
		result.Insert(item)
	}
	return result
}
func (ss *ssetImpl) Intersection(right StringSet) StringSet {
	s1 := ss
	s2 := right.(*ssetImpl)
	if s1.Len() > s2.Len() {
		s1, s2 = s2, s1
	}

	s1.mutex.RLock()
	s2.mutex.RLock()
	defer s1.mutex.RUnlock()
	defer s2.mutex.RUnlock()

	result := NewStringSet()
	for item := range s1.set {
		if s2.hasImpl(item) {
			result.Insert(item)
		}
	}
	return result
}
func (ss *ssetImpl) List() []string {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	result := make([]string, 0, len(ss.set))
	for item := range ss.set {
		result = append(result, item)
	}
	return result
}
func (ss *ssetImpl) SortedList() []string {
	list := ss.List()
	sort.Strings(list)
	return list
}
func (ss *ssetImpl) PopAny() (string, bool) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for item := range ss.set {
		delete(ss.set, item)
		return item, true
	}
	return "", false
}
func (ss *ssetImpl) Len() int {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	return len(ss.set)
}
