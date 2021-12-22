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

	actual := analyzer.Analyze([]byte(text))

	expected := analysis.TokenStream{
		&analysis.Token{
			Start:        0,
			End:          5,
			PositionIncr: 1,
			Type:         0,
			Term:         []byte("hello"),
		},
		&analysis.Token{
			Start:        6,
			End:          11,
			PositionIncr: 1,
			Type:         0,
			Term:         []byte("world"),
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected %v\n", actual)
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

	actual := analyzer.Analyze([]byte(text))

	expected := analysis.TokenStream{
		&analysis.Token{
			Start:        0,
			End:          6,
			PositionIncr: 1,
			Type:         1,
			Term:         []byte("本日"),
		},
		&analysis.Token{
			Start:        9,
			End:          15,
			PositionIncr: 2,
			Type:         1,
			Term:         []byte("晴天"),
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected %v\n", actual)
	}
}
