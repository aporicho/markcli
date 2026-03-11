package ui

type AppMode string

const (
	ModeReading    AppMode = "reading"
	ModeSelecting  AppMode = "selecting"
	ModeAnnotating AppMode = "annotating"
	ModeOverview   AppMode = "overview"
)
