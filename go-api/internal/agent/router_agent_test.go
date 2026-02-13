package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlexString_PlainString(t *testing.T) {
	var f flexString
	if err := json.Unmarshal([]byte(`"hello world"`), &f); err != nil {
		t.Fatalf("unmarshal plain string: %v", err)
	}
	if f != "hello world" {
		t.Errorf("got %q, want %q", f, "hello world")
	}
}

func TestFlexString_JSONObject_PreservesOrder(t *testing.T) {
	var f flexString
	input := `{"Design Intent":"desc","Content":"items","Layout":"grid"}`
	if err := json.Unmarshal([]byte(input), &f); err != nil {
		t.Fatalf("unmarshal object: %v", err)
	}
	want := "Design Intent:\ndesc\n\nContent:\nitems\n\nLayout:\ngrid"
	if f.String() != want {
		t.Errorf("key order not preserved:\ngot:  %q\nwant: %q", f.String(), want)
	}
}

func TestFlexString_RawFallback(t *testing.T) {
	var f flexString
	if err := json.Unmarshal([]byte(`[1,2,3]`), &f); err != nil {
		t.Fatalf("unmarshal array: %v", err)
	}
	if f.String() != "[1,2,3]" {
		t.Errorf("got %q, want %q", f, "[1,2,3]")
	}
}

func TestFlexString_Empty(t *testing.T) {
	var f flexString
	if err := json.Unmarshal([]byte(`""`), &f); err != nil {
		t.Fatalf("unmarshal empty: %v", err)
	}
	if f != "" {
		t.Errorf("got %q, want empty", f)
	}
}

func TestFlexString_InRecommendation(t *testing.T) {
	input := `{"title":"Flow","description":"A flow chart","drawing_prompt":"draw a flow","priority":1}`
	var rec Recommendation
	if err := json.Unmarshal([]byte(input), &rec); err != nil {
		t.Fatalf("unmarshal recommendation: %v", err)
	}
	if rec.DrawingPrompt.String() != "draw a flow" {
		t.Errorf("drawing_prompt: got %q, want %q", rec.DrawingPrompt, "draw a flow")
	}
	if !strings.Contains(rec.Title, "Flow") {
		t.Errorf("title: got %q", rec.Title)
	}
}
