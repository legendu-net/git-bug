package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleParser(t *testing.T) {
	type args struct {
		diff string
	}
	type want struct {
		title string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple addition (html)",
			args: args{
				diff: `<p>changed title from <code class="idiff">simple title</code> to <code class="idiff">simple title<span class="idiff left addition"> addition</span></code></p>`,
			},
			want: want{
				title: "simple title addition",
			},
		},
		{
			name: "simple addition (markdown)",
			args: args{
				diff: `changed title from **simple title** to **simple title{+ addition+}**`,
			},
			want: want{
				title: "simple title addition",
			},
		},
		{
			name: "simple deletion (html)",
			args: args{
				diff: `<p>changed title from <code class="idiff">simple<span class="idiff left right deletion"> deleted</span> title</code> to <code class="idiff">simple title</code></p>`,
			},
			want: want{
				title: "simple title",
			},
		},
		{
			name: "simple deletion (markdown)",
			args: args{
				diff: `changed title from **simple{- deleted-} title** to **simple title**`,
			},
			want: want{
				title: "simple title",
			},
		},
		{
			name: "tail replacement (html)",
			args: args{
				diff: `<p>changed title from <code class="idiff">tail <span class="idiff left right deletion">title</span></code> to <code class="idiff">tail <span class="idiff left addition">replacement</span></code></p>`,
			},
			want: want{
				title: "tail replacement",
			},
		},
		{
			name: "tail replacement (markdown)",
			args: args{
				diff: `changed title from **tail {-title-}** to **tail {+replacement+}**`,
			},
			want: want{
				title: "tail replacement",
			},
		},
		{
			name: "head replacement (html)",
			args: args{
				diff: `<p>changed title from <code class="idiff"><span class="idiff left right deletion">title</span> replacement</code> to <code class="idiff"><span class="idiff left addition">head</span> replacement</code></p>`,
			},
			want: want{
				title: "head replacement",
			},
		},
		{
			name: "head replacement (markdown)",
			args: args{
				diff: `changed title from **{-title-} replacement** to **{+head+} replacement**`,
			},
			want: want{
				title: "head replacement",
			},
		},
		{
			name: "complex multi-section diff (html)",
			args: args{
				diff: `<p>changed title from <code class="idiff">this <span class="idiff left right deletion">is</span> an <span class="idiff left right deletion">issue</span></code> to <code class="idiff">this <span class="idiff left addition">may be</span> an <span class="idiff left right addition">amazing bug</span></code></p>`,
			},
			want: want{
				title: "this may be an amazing bug",
			},
		},
		{
			name: "complex multi-section diff (markdown)",
			args: args{
				diff: `changed title from **this {-is-} an {-issue-}** to **this {+may be+} an {+amazing bug+}**`,
			},
			want: want{
				title: "this may be an amazing bug",
			},
		},
		{
			name: "title containing bold markers (markdown)",
			args: args{
				diff: `changed title from **old** to **new **bold** title**`,
			},
			want: want{
				title: "new **bold** title",
			},
		},
		{
			name: "old and new titles containing bold markers (markdown)",
			args: args{
				diff: `changed title from **foo **bar**** to **foo **baz****`,
			},
			want: want{
				title: "foo **baz**",
			},
		},
		{
			name: "title containing word to (markdown)",
			args: args{
				diff: `changed title from **a to b** to **a to c**`,
			},
			want: want{
				title: "a to c",
			},
		},
		{
			name: "html without paragraph wrapper",
			args: args{
				diff: `changed title from <code class="idiff">old title</code> to <code class="idiff">new title</code>`,
			},
			want: want{
				title: "new title",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := NewWithInput(TitleParser, tt.args.diff).Parse()
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.want.title, title)
		})
	}
}

func TestTitleParserRejectsNonTitleNotes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "multi-span bold markdown after to",
			input: "We should sample down to **10% – 20% of the total number of examples**. Large datasets are defined as **10M+ examples** in our benchmark.",
		},
		{
			name:  "single-span bold markdown after to",
			input: "Query ran much faster when switched to **Distributed Streaming MPP** under `ANALYTICS_REGULAR`!",
		},
		{
			name:  "inline code inside bold after to",
			input: "This causes the launcher to **ignore the user's `PATH`**",
		},
		{
			name:  "inline code only inside bold after to",
			input: "On Ubuntu, `fd` is installed as a symlink to **`fdfind`**",
		},
		{
			name:  "numeric range bold after to",
			input: "Bump memory limit to **512MB or 1GB**",
		},
		{
			name:  "non-title system note with bold after to",
			input: "changed due date to **2026-12-01**",
		},
		{
			name:  "title change prefix with trailing non-title text",
			input: "changed title from **old** to **new** and more text",
		},
		{
			name:  "multiline note starting with title change text",
			input: "changed title from **old** to **new**\nextra line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := NewWithInput(TitleParser, tt.input).Parse()
			assert.Error(t, err)
			assert.Empty(t, title)
		})
	}
}
