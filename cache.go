package giocache

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// FIXME: Reset the cached operations when the percent of deleted items
// reaches some level.

// ID identifies an operation within the cache.
type ID struct {
	value int
}

// OK returns true if the ID is valid.
func (id ID) OK() bool {
	return id.value > 0
}

// Cache holds cached rendering operations for gio.
type Cache struct {
	ops    op.Ops
	id     int
	lookup map[int]op.CallOp
	dead   int
}

// New prepares a new operation cache.
func New() *Cache {
	return &Cache{lookup: make(map[int]op.CallOp)}
}

// Context returns a graphical context for the cache with the given
// constraints and metric.
//
// TODO: Consider making constraints cache-wide, provide an update function,
// and invalidate the cache when the constraints (and/or metric?) change.
func (c *Cache) Context(constraints layout.Constraints, metric unit.Metric) layout.Context {
	return layout.Context{
		Constraints: constraints,
		Metric:      metric,
		Ops:         &c.ops,
	}
}

// Add adds the given call to the cache and returns an ID for future use.
func (c *Cache) Add(cb func(op.MacroOp)) (op.CallOp, ID) {
	c.id++
	macro := op.Record(&c.ops)
	cb(macro)
	call := macro.Stop()
	c.lookup[c.id] = call
	return call, ID{c.id}
}

// Get retrieves a call from the cache for id. The call is only valid until
// the cache is cleared or purged. It returns false if the id is no longer
// valid.
func (c *Cache) Get(id ID) (call op.CallOp, ok bool) {
	call, ok = c.lookup[id.value]
	return
}

// Delete marks the call in the cache with the given ID as dead.
func (c *Cache) Delete(id ID) {
	if _, found := c.lookup[id.value]; found {
		delete(c.lookup, id.value)
		c.dead++
	}
}

// Purge clears the cache if the number of dead items exceed threshold.
// It returns true if the cache was cleared.
func (c *Cache) Purge(threshold int) bool {
	if c.dead <= threshold {
		return false
	}
	for k := range c.lookup {
		delete(c.lookup, k)
	}
	c.ops.Reset()
	c.dead = 0
	return true
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	for id := range c.lookup {
		delete(c.lookup, id)
	}
	c.ops.Reset()
}
