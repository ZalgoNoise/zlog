package config

import "reflect"

// Config interface describes what a single configuration should be able to do
//
// - Its `Name()` method returns a config name, to identify it for debugging
// - Its `Apply(*Configs)` method will sync this Config with the Configs map
// - Its `Is(interface{}) bool` method will allow type comparisons to scope different
// types of Config
type Config interface {
	Name() string
	Apply(confs *Configs)
	Is(interface{}) bool
}

// Configs type is a placeholder map to store Config, which are referenced by their name
// for quicker / direct access. While this is also an insurance that two references for
// the same Config are not held, it's important to notice that similar keys will overwrite
// existing Config when applied.
type Configs map[string]Config

// configuration struct will represent the body of a Config object, which is composed of
// a name (string) and two empty interfaces:
//
// - a parent (to scope implementations to certain types)
// - a value (optional)
type configuration struct {
	name   string      // name to identify the config, for debugging
	parent interface{} // parent type to constrain implementations
	value  interface{} // value of the configuration
}

func NewMap(config ...Config) *Configs {
	c := &Configs{}
	for _, conf := range config {
		conf.Apply(c)
	}
	return c
}

// New function will create a new, empty Config, based on the input name string (the key),
// and a parent (type) which this config should be associated to
func New(name string, parent interface{}) Config {
	return &configuration{
		name:   name,
		parent: parent,
	}
}

// WithValue function will take in a Config and set its value to the input v. This is done
// separately in the same pattern that the context package defines different types of contexts
// for added simplicity on nil-value configs (toggles).
func WithValue(c Config, v interface{}) Config {
	c.(*configuration).value = v
	return c
}

// Map method will allow easier conversion of a Configs type into a map[string]Config type
func (c *Configs) Map() map[string]Config {
	return *c
}

// Get method will take in an input key string and return its value from the Configs map.
//
// This operation will return nil if the value is unset.
func (c *Configs) Get(key string) interface{} {
	cfg := *c
	return cfg[key].(*configuration).value

}

func (c *Configs) IsSet(key string) bool {
	cfg := *c
	_, ok := cfg[key].(*configuration)
	return ok
}

// Name method will return the Config name (string) to identify this configuration for
// debugging purposes
func (c *configuration) Name() string {
	return c.name
}

// Apply method will add this Config to the Configs map, by setting the Config name as key and the
// entire object as value. Finally, it will update the Configs pointer to the new, updated map
func (c *configuration) Apply(confs *Configs) {
	out := *confs
	out[c.name] = c
	confs = &out
}

// Is method will take in any type, and using `reflect.TypeOf(interface{})` it will check
// if the input v and the parent types match, returning a bool value accordingly
func (c *configuration) Is(v interface{}) bool {
	p := reflect.TypeOf(c.parent)
	t := reflect.TypeOf(v)

	return p == t
}
