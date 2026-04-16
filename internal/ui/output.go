package ui

// Mode represents the output mode for the CLI.
type Mode string

const (
	ModeInteractive Mode = "interactive"
	ModeJSON        Mode = "json"
	ModeQuiet       Mode = "quiet"
)

// Event represents a structured output event emitted by the CLI.
type Event struct {
	Type       string `json:"event,omitempty"`
	TotalLinks int    `json:"total_links,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Bytes      int64  `json:"bytes,omitempty"`
	Speed      string `json:"speed,omitempty"`
	Path       string `json:"path,omitempty"`
	ElapsedMs  int64  `json:"elapsed_ms,omitempty"`
	Link       string `json:"link,omitempty"`
	Code       int    `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	Total      int    `json:"total,omitempty"`
	Succeeded  int    `json:"succeeded,omitempty"`
	Failed     int    `json:"failed,omitempty"`
	Attempt    int    `json:"attempt,omitempty"`
	WaitMs     int64  `json:"wait_ms,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// Emitter is the interface for outputting structured events.
type Emitter interface {
	Emit(Event)
}

// DetectMode determines the output mode based on flags, environment, config,
// and TTY detection. Priority order: json flag > quiet flag > env > config > TTY.
func DetectMode(jsonFlag, quietFlag bool, envMode, cfgMode string, isTTY bool) Mode {
	if jsonFlag {
		return ModeJSON
	}
	if quietFlag {
		return ModeQuiet
	}
	if envMode != "" {
		switch Mode(envMode) {
		case ModeJSON:
			return ModeJSON
		case ModeQuiet:
			return ModeQuiet
		case ModeInteractive:
			return ModeInteractive
		}
	}
	if cfgMode != "" {
		switch Mode(cfgMode) {
		case ModeJSON:
			return ModeJSON
		case ModeQuiet:
			return ModeQuiet
		case ModeInteractive:
			return ModeInteractive
		}
	}
	if isTTY {
		return ModeInteractive
	}
	return ModeJSON
}
