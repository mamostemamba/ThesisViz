package sanitize

import "testing"

func TestTikZ_TextCommands(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "textbf to bfseries",
			in:   `|[fill=primaryFill]| \textbf{Hello}`,
			want: `|[fill=primaryFill]| {\bfseries Hello}`,
		},
		{
			name: "textit to itshape",
			in:   `\textit{world}`,
			want: `{\itshape world}`,
		},
		{
			name: "emph to itshape",
			in:   `\emph{important}`,
			want: `{\itshape important}`,
		},
		{
			name: "underline removed",
			in:   `\underline{text}`,
			want: `text`,
		},
		{
			name: "text wrapper removed",
			in:   `\text{plain}`,
			want: `plain`,
		},
		{
			name: "mbox wrapper removed",
			in:   `\mbox{content}`,
			want: `content`,
		},
		{
			name: "font size commands removed",
			in:   `\footnotesize Hello \large World`,
			want: `Hello World`,
		},
		{
			name: "scriptsize removed",
			in:   `\scriptsize text`,
			want: `text`,
		},
		{
			name: "tabular replaced with plain text",
			in:   `\begin{tabular}{c}A & B \\ C & D\end{tabular}`,
			want: `A B C D`,
		},
		{
			name: "multiple replacements in one string",
			in:   `|[fill=primaryFill]| \textbf{Module} \footnotesize v2`,
			want: `|[fill=primaryFill]| {\bfseries Module} v2`,
		},
		{
			name: "definecolor still removed",
			in:   `\definecolor{mycolor}{RGB}{255,0,0}` + "\n" + `\node {ok};`,
			want: `\node {ok};`,
		},
		{
			name: "no change for safe content",
			in:   `|[fill=primaryFill]| {\bfseries Hello}`,
			want: `|[fill=primaryFill]| {\bfseries Hello}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TikZ(tt.in)
			if got != tt.want {
				t.Errorf("TikZ(%q)\n  got:  %q\n  want: %q", tt.in, got, tt.want)
			}
		})
	}
}
