package gofs

type Flag uint32

const (
	FLAG__RECURSIVE Flag = 1 << iota
	FLAG__WITH_INFOS
)
