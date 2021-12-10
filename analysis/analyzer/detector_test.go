package analyzer

import (
	"reflect"
	"testing"

	"github.com/blugelabs/bluge/analysis"
	"github.com/mosuka/phalanx/logging"
)

func TestNewLanguageDetectorEn(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	detector, err := NewAnalyzerDetector(logger)
	if err != nil {
		t.Fatalf("failed to create language detector\n")
	}

	text := "Hello world."
	analyzer := detector.DetectAnalyzer(text)

	ts := analyzer.Analyze([]byte(text))
	freqs, _ := analysis.TokenFrequency(ts, true, 0)
	tokens := make([]string, 0)
	for token := range freqs {
		tokens = append(tokens, token)
	}
	if !reflect.DeepEqual(tokens, []string{"hello", "world"}) {
		t.Fatalf("unexpected %v\n", tokens)
	}
}

func TestNewLanguageDetectorJa(t *testing.T) {
	logger := logging.NewLogger("WARN", "", 500, 3, 30, false)

	detector, err := NewAnalyzerDetector(logger)
	if err != nil {
		t.Fatalf("failed to create language detector\n")
	}

	text := "本日は晴天なり"
	analyzer := detector.DetectAnalyzer(text)

	ts := analyzer.Analyze([]byte(text))
	freqs, _ := analysis.TokenFrequency(ts, true, 0)
	tokens := make([]string, 0)
	for token := range freqs {
		tokens = append(tokens, token)
	}
	if !reflect.DeepEqual(tokens, []string{"本日", "晴天"}) {
		t.Fatalf("unexpected %v\n", tokens)
	}
}
