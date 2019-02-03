package hooks

import (
	"net/http"
)

// A NamedHook is a struct that contains a name and hook.
type NamedHook struct {
	Name     string
	Apply    func(w http.ResponseWriter, r *http.Request) bool
	Priority int
}

// A HookList manages zero or more hook(s) in a list.
type HookList struct {
	list []NamedHook

	// Called after each server hook in the list is invoked. If set
	// and the func returns true the HookList will continue to iterate
	// over the server hooks. If false is returned the HookList
	// will stop iterating.
	//
	// Should be used if extra logic to be performed between each hook
	// in the list. This can be used to terminate a list's iteration
	// based on a condition such as logging like NewServerDebugLogHook.
	AfterEach func(item HookItem) bool
}

// Has returns true if named hook exists, otherwise returns false
func (l *HookList) Has(name string) bool {
	for i := 0; i < len(l.list); i++ {
		if l.list[i].Name == name {
			return true
		}
	}

	return false
}

// Copy creates a copy of the hook list.
func (l *HookList) Copy() HookList {
	list := HookList{
		AfterEach: l.AfterEach,
	}

	if len(l.list) > 0 {
		list.list = make([]NamedHook, len(l.list))
		copy(list.list, l.list)
	}

	return list
}

// Len returns the number of hooks in the list.
func (l *HookList) Len() int {
	return len(l.list)
}

// Hooks returns the hooks in the list
func (l *HookList) Hooks() []NamedHook {
	return l.list
}

// Clear clears the hook list.
func (l *HookList) Clear() {
	l.list = l.list[0:0]
}

// PushBack pushes hook fn to the back of the hook list.
func (l *HookList) PushBack(fn func(w http.ResponseWriter, r *http.Request) bool) {
	l.PushBackNamed(NamedHook{"__anonymous", fn, -1})
}

// PushBackNamed pushes named hook to the back of the hook list.
func (l *HookList) PushBackNamed(hook NamedHook) {
	if cap(l.list) == 0 {
		l.list = make([]NamedHook, 0, 3)
	}

	l.list = append(l.list, hook)
}

// PushFront pushes hook fn to the front of the hook list.
func (l *HookList) PushFront(fn func(w http.ResponseWriter, r *http.Request) bool) {
	l.PushFrontNamed(NamedHook{"__anonymous", fn, -1})
}

// PushFrontNamed pushes named hook to the front of the hook list.
func (l *HookList) PushFrontNamed(hook NamedHook) {
	if cap(l.list) == len(l.list) {
		// Allocating new list required
		l.list = append([]NamedHook{hook}, l.list...)
	} else {
		// Enough room to prepend into list.
		l.list = append(l.list, NamedHook{})
		copy(l.list[1:], l.list)

		l.list[0] = hook
	}
}

// Pop removes a NamedHook of hook.Name and returns it
func (l *HookList) Pop(hook NamedHook) NamedHook {
	return l.PopNamed(hook.Name)
}

// PopNamed removes a NamedHook of name and returns it if exist.
func (l *HookList) PopNamed(name string) NamedHook {
	var hook NamedHook

	for i := 0; i < len(l.list); i++ {
		h := l.list[i]
		if h.Name == name {
			hook = h

			// Shift array preventing creating new arrays
			copy(l.list[i:], l.list[i+1:])
			l.list[len(l.list)-1] = NamedHook{}
			l.list = l.list[:len(l.list)-1]

			// decrement list so next check to length is correct
			i--
		}
	}

	return hook
}

// SwapNamed will swap out any existing hooks with the same name as the
// passed in NamedHook returning true if hooks were swapped. False is
// returned otherwise.
func (l *HookList) SwapNamed(hook NamedHook) (swapped bool) {
	for i := 0; i < len(l.list); i++ {
		if l.list[i].Name == hook.Name {
			l.list[i].Apply = hook.Apply
			swapped = true
		}
	}

	return swapped
}

// Swap will swap out all hooks matching the name passed in. The matched
// hooks will be swapped in. True is returned if the hooks were swapped.
func (l *HookList) Swap(name string, hook NamedHook) bool {
	var swapped bool

	for i := 0; i < len(l.list); i++ {
		if l.list[i].Name == name {
			l.list[i] = hook
			swapped = true
		}
	}

	return swapped
}

// SetBackNamed will replace the named hook if it exists in the hook list.
// If the hook does not exist the named hook will be added to the end of the list.
func (l *HookList) SetBackNamed(hook NamedHook) {
	if !l.SwapNamed(hook) {
		l.PushBackNamed(hook)
	}
}

// SetFrontNamed will replace the named hook if it exists in the hook list.
// If the hook does not exist the named hook will be added to the beginning of
// the list.
func (l *HookList) SetFrontNamed(n NamedHook) {
	if !l.SwapNamed(n) {
		l.PushFrontNamed(n)
	}
}

// A HookItem represents an entry in the HookList which
// is being run.
type HookItem struct {
	Index    int
	Hook     NamedHook
	Response http.ResponseWriter
	Request  *http.Request
}

// Run executes all hooks in the list with given http.ResponseWriter and *http.Request.
func (l *HookList) Run(w http.ResponseWriter, r *http.Request) bool {
	if l == nil {
		return true
	}

	for i, h := range l.list {
		if !h.Apply(w, r) {
			return false
		}

		if l.AfterEach != nil {
			item := HookItem{
				Index:    i,
				Hook:     h,
				Response: w,
				Request:  r,
			}

			if !l.AfterEach(item) {
				return false
			}
		}
	}

	return true
}
