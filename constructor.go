package container

import "context"

func (c *callSite[T]) tryFastInit(resolved *T, deps []any) (bool, error) {
	switch len(deps) {
	case 0:
		if init, ok := any(resolved).(Initializable0); ok {
			return true, init.Init()
		}
	case 1:
		if init, ok := any(resolved).(Initializable1[context.Context]); ok {
			return true, init.Init(deps[0].(context.Context))
		}
	}
	return false, nil
}
