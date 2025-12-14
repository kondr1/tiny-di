package container

func (c *callSite[T]) tryFastInit(resolved *T, deps []any) (bool, error) {
	switch len(deps) {
	case 0:
		if init, ok := any(resolved).(Initializable0); ok {
			return true, init.Init()
		}
	}
	return false, nil
}
