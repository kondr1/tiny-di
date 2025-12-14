package container

// Type switch optimization for 0-5 dependencies (fast path).
// Avoids reflection by directly calling Init methods for common cases.
// Returns error if initialization fails, or falls back to reflection if type doesn't match.
func (c *callSite[T]) tryFastInit(resolved *T, deps []any) (bool, error) {
	switch len(deps) {
	case 0:
		if init, ok := any(resolved).(Initializable0); ok {
			return true, init.Init()
		}
	case 1:
		if init, ok := any(resolved).(Initializable1); ok {
			return true, init.Init(deps[0])
		}
	case 2:
		if init, ok := any(resolved).(Initializable2); ok {
			return true, init.Init(deps[0], deps[1])
		}
	case 3:
		if init, ok := any(resolved).(Initializable3); ok {
			return true, init.Init(deps[0], deps[1], deps[2])
		}
	case 4:
		if init, ok := any(resolved).(Initializable4); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3])
		}
	case 5:
		if init, ok := any(resolved).(Initializable5); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4])
		}
	case 6:
		if init, ok := any(resolved).(Initializable6); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5])
		}
	case 7:
		if init, ok := any(resolved).(Initializable7); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6])
		}
	case 8:
		if init, ok := any(resolved).(Initializable8); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7])
		}
	case 9:
		if init, ok := any(resolved).(Initializable9); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8])
		}
	case 10:
		if init, ok := any(resolved).(Initializable10); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9])
		}
	case 11:
		if init, ok := any(resolved).(Initializable11); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10])
		}
	case 12:
		if init, ok := any(resolved).(Initializable12); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11])
		}
	case 13:
		if init, ok := any(resolved).(Initializable13); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12])
		}
	case 14:
		if init, ok := any(resolved).(Initializable14); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13])
		}
	case 15:
		if init, ok := any(resolved).(Initializable15); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14])
		}
	case 16:
		if init, ok := any(resolved).(Initializable16); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14], deps[15])
		}
	case 17:
		if init, ok := any(resolved).(Initializable17); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14], deps[15], deps[16])
		}
	case 18:
		if init, ok := any(resolved).(Initializable18); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14], deps[15], deps[16], deps[17])
		}
	case 19:
		if init, ok := any(resolved).(Initializable19); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14], deps[15], deps[16], deps[17], deps[18])
		}
	case 20:
		if init, ok := any(resolved).(Initializable20); ok {
			return true, init.Init(deps[0], deps[1], deps[2], deps[3], deps[4], deps[5], deps[6], deps[7], deps[8], deps[9], deps[10], deps[11], deps[12], deps[13], deps[14], deps[15], deps[16], deps[17], deps[18], deps[19])
		}
	}
	return false, nil
}
