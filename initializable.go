package container

// Initializable interfaces for type switch optimization.
// These interfaces allow avoiding reflection for constructors with 0-5 dependencies.

type Initializable0 interface {
	Init() error
}

type Initializable1 interface {
	Init(dep1 any) error
}

type Initializable2 interface {
	Init(dep1, dep2 any) error
}

type Initializable3 interface {
	Init(dep1, dep2, dep3 any) error
}

type Initializable4 interface {
	Init(dep1, dep2, dep3, dep4 any) error
}

type Initializable5 interface {
	Init(dep1, dep2, dep3, dep4, dep5 any) error
}

type Initializable6 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6 any) error
}

type Initializable7 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7 any) error
}

type Initializable8 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8 any) error
}

type Initializable9 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9 any) error
}

type Initializable10 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10 any) error
}

type Initializable11 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11 any) error
}

type Initializable12 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12 any) error
}

type Initializable13 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13 any) error
}

type Initializable14 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14 any) error
}

type Initializable15 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15 any) error
}

type Initializable16 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15, dep16 any) error
}

type Initializable17 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15, dep16, dep17 any) error
}

type Initializable18 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15, dep16, dep17, dep18 any) error
}

type Initializable19 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15, dep16, dep17, dep18, dep19 any) error
}

type Initializable20 interface {
	Init(dep1, dep2, dep3, dep4, dep5, dep6, dep7, dep8, dep9, dep10, dep11, dep12, dep13, dep14, dep15, dep16, dep17, dep18, dep19, dep20 any) error
}
