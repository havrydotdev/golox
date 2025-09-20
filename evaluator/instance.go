package eval

type Instance struct {
	Class
	fields map[string]any
}

func (i Instance) String() string {
	return i.Name + " instance"
}

func (i Instance) Get(key string) (any, bool) {
	val, ok := i.fields[key]
	return val, ok
}

func (i Instance) Set(key string, value any) {
	i.fields[key] = value
}
