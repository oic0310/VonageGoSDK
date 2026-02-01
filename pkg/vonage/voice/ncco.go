package voice

import "encoding/json"

// ========================================
// NCCO (Nexmo Call Control Objects)
// ========================================

// NCCO represents a list of NCCO actions
type NCCO []Action

// JSON returns the NCCO as a JSON byte slice
func (n NCCO) JSON() ([]byte, error) {
	return json.Marshal(n)
}

// Action represents a single NCCO action
type Action struct {
	// Common fields
	ActionType string `json:"action"`

	// Talk action
	Text      string `json:"text,omitempty"`
	VoiceName string `json:"voiceName,omitempty"`
	Language  string `json:"language,omitempty"`
	Style     int    `json:"style,omitempty"`
	Premium   bool   `json:"premium,omitempty"`
	Level     int    `json:"level,omitempty"`
	BargeIn   *bool  `json:"bargeIn,omitempty"`
	Loop      int    `json:"loop,omitempty"`

	// Stream action
	StreamURL []string `json:"streamUrl,omitempty"`

	// Input action
	Type         []string `json:"type,omitempty"`
	EventURL     []string `json:"eventUrl,omitempty"`
	EventMethod  string   `json:"eventMethod,omitempty"`
	EndOnSilence float64  `json:"endOnSilence,omitempty"`
	StartTimeout int      `json:"startTimeout,omitempty"`
	MaxDuration  int      `json:"maxDuration,omitempty"`

	// DTMF specific (within Input)
	MaxDigits  int    `json:"maxDigits,omitempty"`
	SubmitOnHash bool `json:"submitOnHash,omitempty"`
	TimeOut    int    `json:"timeOut,omitempty"`

	// Notify action
	Payload map[string]interface{} `json:"payload,omitempty"`

	// Record action
	Format       string   `json:"format,omitempty"`
	BeepStart    *bool    `json:"beepStart,omitempty"`
	EndOnKey     string   `json:"endOnKey,omitempty"`
	Channels     int      `json:"channels,omitempty"`
	Split        string   `json:"split,omitempty"`
}

// ========================================
// NCCO Builder
// ========================================

// NCCOBuilder provides a fluent API for building NCCO
type NCCOBuilder struct {
	actions []Action
}

// NewNCCO creates a new NCCO builder
func NewNCCO() *NCCOBuilder {
	return &NCCOBuilder{
		actions: make([]Action, 0),
	}
}

// Build returns the final NCCO
func (b *NCCOBuilder) Build() NCCO {
	return NCCO(b.actions)
}

// ========================================
// Talk Action
// ========================================

// TalkBuilder builds a talk action
type TalkBuilder struct {
	parent *NCCOBuilder
	action Action
}

// Talk adds a talk action to the NCCO
func (b *NCCOBuilder) Talk(text string) *TalkBuilder {
	return &TalkBuilder{
		parent: b,
		action: Action{
			ActionType: "talk",
			Text:       text,
		},
	}
}

// VoiceName sets the voice name
func (t *TalkBuilder) VoiceName(name string) *TalkBuilder {
	t.action.VoiceName = name
	return t
}

// Language sets the language
func (t *TalkBuilder) Language(lang string) *TalkBuilder {
	t.action.Language = lang
	return t
}

// Style sets the voice style
func (t *TalkBuilder) Style(style int) *TalkBuilder {
	t.action.Style = style
	return t
}

// Premium enables premium voice
func (t *TalkBuilder) Premium() *TalkBuilder {
	t.action.Premium = true
	return t
}

// Level sets the volume level (-1 to 1)
func (t *TalkBuilder) Level(level int) *TalkBuilder {
	t.action.Level = level
	return t
}

// BargeIn allows the caller to interrupt the talk
func (t *TalkBuilder) BargeIn() *TalkBuilder {
	bargeIn := true
	t.action.BargeIn = &bargeIn
	return t
}

// Loop sets the number of times to loop (0 = infinite)
func (t *TalkBuilder) Loop(count int) *TalkBuilder {
	t.action.Loop = count
	return t
}

// Japanese is a convenience method for Japanese TTS with Mizuki voice
func (t *TalkBuilder) Japanese() *TalkBuilder {
	t.action.VoiceName = "Mizuki"
	t.action.Language = "ja-JP"
	return t
}

// Done finalizes the talk action and returns the NCCO builder
func (t *TalkBuilder) Done() *NCCOBuilder {
	t.parent.actions = append(t.parent.actions, t.action)
	return t.parent
}

// ========================================
// Stream Action
// ========================================

// StreamBuilder builds a stream action
type StreamBuilder struct {
	parent *NCCOBuilder
	action Action
}

// Stream adds a stream action to the NCCO
func (b *NCCOBuilder) Stream(urls ...string) *StreamBuilder {
	return &StreamBuilder{
		parent: b,
		action: Action{
			ActionType: "stream",
			StreamURL:  urls,
		},
	}
}

// Level sets the volume level (-1 to 1)
func (s *StreamBuilder) Level(level int) *StreamBuilder {
	s.action.Level = level
	return s
}

// BargeIn allows the caller to interrupt the stream
func (s *StreamBuilder) BargeIn() *StreamBuilder {
	bargeIn := true
	s.action.BargeIn = &bargeIn
	return s
}

// Loop sets the number of times to loop (0 = infinite)
func (s *StreamBuilder) Loop(count int) *StreamBuilder {
	s.action.Loop = count
	return s
}

// Done finalizes the stream action and returns the NCCO builder
func (s *StreamBuilder) Done() *NCCOBuilder {
	s.parent.actions = append(s.parent.actions, s.action)
	return s.parent
}

// ========================================
// Input Action
// ========================================

// InputBuilder builds an input action
type InputBuilder struct {
	parent *NCCOBuilder
	action Action
}

// Input adds an input action to the NCCO
func (b *NCCOBuilder) Input() *InputBuilder {
	return &InputBuilder{
		parent: b,
		action: Action{
			ActionType: "input",
		},
	}
}

// Speech enables speech recognition input
func (i *InputBuilder) Speech() *InputBuilder {
	i.action.Type = appendUnique(i.action.Type, "speech")
	return i
}

// DTMF enables DTMF input
func (i *InputBuilder) DTMF() *InputBuilder {
	i.action.Type = appendUnique(i.action.Type, "dtmf")
	return i
}

// SpeechAndDTMF enables both speech and DTMF input
func (i *InputBuilder) SpeechAndDTMF() *InputBuilder {
	i.action.Type = []string{"speech", "dtmf"}
	return i
}

// EventURL sets the event URL for input results
func (i *InputBuilder) EventURL(url string) *InputBuilder {
	i.action.EventURL = []string{url}
	return i
}

// EventMethod sets the HTTP method for the event URL
func (i *InputBuilder) EventMethod(method string) *InputBuilder {
	i.action.EventMethod = method
	return i
}

// EndOnSilence sets the silence detection threshold (seconds)
func (i *InputBuilder) EndOnSilence(seconds float64) *InputBuilder {
	i.action.EndOnSilence = seconds
	return i
}

// StartTimeout sets the start timeout (seconds)
func (i *InputBuilder) StartTimeout(seconds int) *InputBuilder {
	i.action.StartTimeout = seconds
	return i
}

// MaxDuration sets the maximum input duration (seconds)
func (i *InputBuilder) MaxDuration(seconds int) *InputBuilder {
	i.action.MaxDuration = seconds
	return i
}

// MaxDigits sets the maximum number of DTMF digits
func (i *InputBuilder) MaxDigits(digits int) *InputBuilder {
	i.action.MaxDigits = digits
	return i
}

// SubmitOnHash ends DTMF input when # is pressed
func (i *InputBuilder) SubmitOnHash() *InputBuilder {
	i.action.SubmitOnHash = true
	return i
}

// TimeOut sets the DTMF timeout (seconds)
func (i *InputBuilder) TimeOut(seconds int) *InputBuilder {
	i.action.TimeOut = seconds
	return i
}

// Done finalizes the input action and returns the NCCO builder
func (i *InputBuilder) Done() *NCCOBuilder {
	// Default to POST method
	if i.action.EventMethod == "" {
		i.action.EventMethod = "POST"
	}
	i.parent.actions = append(i.parent.actions, i.action)
	return i.parent
}

// ========================================
// Record Action
// ========================================

// RecordBuilder builds a record action
type RecordBuilder struct {
	parent *NCCOBuilder
	action Action
}

// Record adds a record action to the NCCO
func (b *NCCOBuilder) Record() *RecordBuilder {
	return &RecordBuilder{
		parent: b,
		action: Action{
			ActionType: "record",
		},
	}
}

// Format sets the recording format (mp3 or wav)
func (r *RecordBuilder) Format(format string) *RecordBuilder {
	r.action.Format = format
	return r
}

// EndOnSilence sets the silence detection threshold
func (r *RecordBuilder) EndOnSilence(seconds float64) *RecordBuilder {
	r.action.EndOnSilence = seconds
	return r
}

// EndOnKey sets the key to end recording
func (r *RecordBuilder) EndOnKey(key string) *RecordBuilder {
	r.action.EndOnKey = key
	return r
}

// BeepStart plays a beep at the start of recording
func (r *RecordBuilder) BeepStart() *RecordBuilder {
	beepStart := true
	r.action.BeepStart = &beepStart
	return r
}

// EventURL sets the event URL for recording results
func (r *RecordBuilder) EventURL(url string) *RecordBuilder {
	r.action.EventURL = []string{url}
	return r
}

// Split enables split recording for conversations
func (r *RecordBuilder) Split() *RecordBuilder {
	r.action.Split = "conversation"
	return r
}

// Channels sets the number of channels
func (r *RecordBuilder) Channels(channels int) *RecordBuilder {
	r.action.Channels = channels
	return r
}

// Done finalizes the record action and returns the NCCO builder
func (r *RecordBuilder) Done() *NCCOBuilder {
	r.parent.actions = append(r.parent.actions, r.action)
	return r.parent
}

// ========================================
// Notify Action
// ========================================

// Notify adds a notify action to the NCCO
func (b *NCCOBuilder) Notify(eventURL string, payload map[string]interface{}) *NCCOBuilder {
	b.actions = append(b.actions, Action{
		ActionType:  "notify",
		EventURL:    []string{eventURL},
		EventMethod: "POST",
		Payload:     payload,
	})
	return b
}

// ========================================
// Convenience: Quick NCCO Patterns
// ========================================

// TalkAndInput creates a common talk-then-listen NCCO pattern
// This matches the VonaTrigger pattern: speak → wait for speech input
func TalkAndInput(text, language, voiceName, inputEventURL string, endOnSilence float64) NCCO {
	return NewNCCO().
		Talk(text).VoiceName(voiceName).Language(language).Done().
		Input().Speech().EventURL(inputEventURL).
			EndOnSilence(endOnSilence).StartTimeout(5).MaxDuration(30).Done().
		Build()
}

// TalkJapanese creates a Japanese TTS NCCO (Mizuki voice)
func TalkJapanese(text string) NCCO {
	return NewNCCO().
		Talk(text).Japanese().Done().
		Build()
}

// TalkAndInputJapanese creates a Japanese talk-then-listen NCCO
func TalkAndInputJapanese(text, inputEventURL string) NCCO {
	return NewNCCO().
		Talk(text).Japanese().Done().
		Input().Speech().EventURL(inputEventURL).
			EndOnSilence(1.5).StartTimeout(5).MaxDuration(30).Done().
		Build()
}

// StreamAndInput creates a stream-then-listen NCCO pattern
func StreamAndInput(audioURL, inputEventURL string, endOnSilence float64) NCCO {
	return NewNCCO().
		Stream(audioURL).Done().
		Input().Speech().EventURL(inputEventURL).
			EndOnSilence(endOnSilence).StartTimeout(5).MaxDuration(30).Done().
		Build()
}

// ========================================
// Helpers
// ========================================

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
