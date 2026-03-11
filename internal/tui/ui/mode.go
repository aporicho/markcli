package ui

type AppMode string

const (
	ModeBrowsing   AppMode = "browsing"
	ModeReading    AppMode = "reading"
	ModeSelecting  AppMode = "selecting"
	ModeAnnotating AppMode = "annotating"
	ModeOverview   AppMode = "overview"
)
