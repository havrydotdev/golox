package env

type Env struct {
	outer *Env

	values map[string]any
}

func New() *Env {
	return &Env{values: make(map[string]any), outer: nil}
}

func NewChild(outer *Env) *Env {
	return &Env{values: make(map[string]any), outer: outer}
}

func (e *Env) Define(name string, value any) {
	e.values[name] = value
}

func (e *Env) Assign(name string, value any) bool {
	_, ok := e.values[name]
	if !ok {
		if e.outer != nil {
			return e.outer.Assign(name, value)
		}

		return false
	}

	e.values[name] = value
	return true
}

func (e *Env) Get(name string) (any, bool) {
	val, ok := e.values[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}

	return val, ok
}
