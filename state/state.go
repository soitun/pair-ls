package state

import (
	"log"
	"reflect"
	"sync"

	"github.com/asaskevich/EventBus"
	"github.com/sourcegraph/go-lsp"
)

type WorkspaceState struct {
	files  map[string]File
	view   *View
	mu     sync.Mutex
	events EventBus.Bus
	logger *log.Logger
}

type File struct {
	Filename string   `json:"filename"`
	Lines    []string `json:"lines"`
	Language string   `json:"language"`
}

type View struct {
	Filename  string     `json:"filename"`
	Line      int        `json:"line"`
	Character int        `json:"character"`
	Range     *lsp.Range `json:"range"`
}

type OpenFileEvent struct {
	Filename string `json:"filename"`
	Language string `json:"language"`
}

type CloseFileEvent struct {
	Filename string `json:"filename"`
}

type ReplaceTextEvent struct {
	Filename string   `json:"filename"`
	Text     []string `json:"text"`
}

type ChangeViewEvent struct {
	View View `json:"view"`
}

const TOPIC = "all"

func NewState(logger *log.Logger) *WorkspaceState {
	return &WorkspaceState{
		files:  make(map[string]File),
		events: EventBus.New(),
		logger: logger,
	}
}

func (s *WorkspaceState) publish(value interface{}) {
	s.events.Publish(TOPIC, value)
}

func (s *WorkspaceState) Subscribe(fn interface{}) error {
	return s.events.Subscribe(TOPIC, fn)
}

func (s *WorkspaceState) Unsubscribe(fn interface{}) error {
	return s.events.Unsubscribe(TOPIC, fn)
}

func (s *WorkspaceState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.files {
		delete(s.files, k)
	}
	s.view = nil
}

func (s *WorkspaceState) CursorMove(filename string, position lsp.Position, rng *lsp.Range) {
	s.mu.Lock()
	defer s.mu.Unlock()
	anyChanges := false
	if s.view.Filename != filename {
		s.view.Filename = filename
		anyChanges = true
	}
	if s.view.Line != position.Line {
		s.view.Line = position.Line
		anyChanges = true
	}
	if s.view.Character != position.Character {
		s.view.Character = position.Character
		anyChanges = true
	}
	if rng != nil {
		if !reflect.DeepEqual(s.view.Range, rng) {
			anyChanges = true
			s.view.Range = rng
		}
	} else if s.view.Range != nil {
		s.view.Range = nil
		anyChanges = true
	}
	if anyChanges {
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
}

func (s *WorkspaceState) OpenFile(filename string, text []string, language string, updateCursor bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[filename] = File{
		Filename: filename,
		Lines:    text,
		Language: language,
	}
	s.publish(OpenFileEvent{
		Filename: filename,
		Language: language,
	})

	if updateCursor || s.view == nil {
		s.view = &View{
			Filename:  filename,
			Line:      0,
			Character: 0,
		}
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
}

func (s *WorkspaceState) CloseFile(filename string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, filename)
	s.publish(CloseFileEvent{
		Filename: filename,
	})
}

func (s *WorkspaceState) ReplaceText(filename string, text []string, updateCursor bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev := s.files[filename]
	s.files[filename] = File{
		Filename: prev.Filename,
		Language: prev.Language,
		Lines:    text,
	}
	s.publish(ReplaceTextEvent{
		Filename: filename,
		Text:     text,
	})

	if updateCursor {
		lnum := 0
		col := 0
		lnum = -1
		for i, line := range prev.Lines {
			if i >= len(text) {
				lnum = i
				break
			} else if line != text[i] {
				lnum = i
				col = longestCommonPrefix(line, text[i])
				break
			}
		}
		if lnum == -1 {
			lnum = len(text)
		}
		s.view = &View{
			Filename:  filename,
			Line:      lnum,
			Character: col,
		}
		s.publish(ChangeViewEvent{
			View: *s.view,
		})
	}
}

func (s *WorkspaceState) GetFiles() map[string]File {
	s.mu.Lock()
	defer s.mu.Unlock()

	ret := make(map[string]File)
	for key, value := range s.files {
		ret[key] = value
	}
	return ret
}

func (s *WorkspaceState) GetText(filename string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.files[filename].Lines
}

func (s *WorkspaceState) GetView() *View {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.view
}

func longestCommonPrefix(s1 string, s2 string) int {
	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	if len(s1) < len(s2) {
		return len(s1)
	} else {
		return len(s2)
	}
}