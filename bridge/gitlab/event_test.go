package gitlab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/api/client-go"
)

var _ Event = mockEvent(0)

type mockEvent int64

func (m mockEvent) ID() string           { panic("implement me") }
func (m mockEvent) UserID() int64        { panic("implement me") }
func (m mockEvent) Kind() EventKind      { panic("implement me") }
func (m mockEvent) CreatedAt() time.Time { return time.Unix(int64(m), 0) }

func TestSortedEvents(t *testing.T) {
	makeInput := func(times ...int64) chan Event {
		out := make(chan Event)
		go func() {
			for _, t := range times {
				out <- mockEvent(t)
			}
			close(out)
		}()
		return out
	}

	sorted := SortedEvents(
		makeInput(),
		makeInput(1, 7, 9, 19),
		makeInput(2, 8, 23),
		makeInput(35, 48, 59, 64, 721),
	)

	var previous Event
	for event := range sorted {
		if previous != nil {
			require.True(t, previous.CreatedAt().Before(event.CreatedAt()))
		}
		previous = event
	}
}

func TestNoteEventKind(t *testing.T) {
	tests := []struct {
		name      string
		system    bool
		body      string
		wantKind  EventKind
		wantTitle string
	}{
		{
			name:     "non-system multi-span bold after to",
			system:   false,
			body:     "We should sample down to **10% – 20% of the total number of examples**. Large datasets are defined as **10M+ examples** in our benchmark.",
			wantKind: EventComment,
		},
		{
			name:     "non-system single-span bold after to",
			system:   false,
			body:     "Query ran much faster when switched to **Distributed Streaming MPP** under `ANALYTICS_REGULAR`!",
			wantKind: EventComment,
		},
		{
			name:     "non-system inline code inside bold",
			system:   false,
			body:     "This causes the launcher to **ignore the user's `PATH`**",
			wantKind: EventComment,
		},
		{
			name:     "non-system inline code only inside bold",
			system:   false,
			body:     "On Ubuntu, `fd` is installed as a symlink to **`fdfind`**",
			wantKind: EventComment,
		},
		{
			name:     "non-system numeric range bold after to",
			system:   false,
			body:     "Bump memory limit to **512MB or 1GB**",
			wantKind: EventComment,
		},
		{
			name:     "non-system user comment literally mimicking markdown title change",
			system:   false,
			body:     "changed title from **old** to **new**",
			wantKind: EventComment,
		},
		{
			name:     "non-system user comment literally mimicking html title change",
			system:   false,
			body:     `<p>changed title from <code class="idiff">old</code> to <code class="idiff">new</code></p>`,
			wantKind: EventComment,
		},
		{
			name:     "non-title system note changing due date with bold",
			system:   true,
			body:     "changed due date to **2026-12-01**",
			wantKind: EventChangedDuedate,
		},
		{
			name:     "non-title unknown system note with bold after to",
			system:   true,
			body:     "changed weight to **5**",
			wantKind: EventUnknown,
		},
		{
			name:      "valid markdown title change system note",
			system:    true,
			body:      "changed title from **old title** to **old title{+ updated+}**",
			wantKind:  EventTitleChanged,
			wantTitle: "old title updated",
		},
		{
			name:      "valid markdown title change with bold markers in title",
			system:    true,
			body:      "changed title from **old** to **new **bold** title**",
			wantKind:  EventTitleChanged,
			wantTitle: "new **bold** title",
		},
		{
			name:      "valid html title change system note",
			system:    true,
			body:      `<p>changed title from <code class="idiff">old title</code> to <code class="idiff">old title<span class="idiff left addition"> updated</span></code></p>`,
			wantKind:  EventTitleChanged,
			wantTitle: "old title updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var note gitlab.Note
			note.System = tt.system
			note.Body = tt.body
			ev := NoteEvent{note}

			require.Equal(t, tt.wantKind, ev.Kind())
			if tt.wantKind == EventTitleChanged {
				gotTitle, err := ev.Title()
				require.NoError(t, err)
				require.Equal(t, tt.wantTitle, gotTitle)
			}
		})
	}
}
