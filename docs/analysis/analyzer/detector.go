package analyzer

import (
	"strings"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/analyzer"
	"github.com/blugelabs/bluge/analysis/lang/ar"
	"github.com/blugelabs/bluge/analysis/lang/cjk"
	"github.com/blugelabs/bluge/analysis/lang/da"
	"github.com/blugelabs/bluge/analysis/lang/de"
	"github.com/blugelabs/bluge/analysis/lang/en"
	"github.com/blugelabs/bluge/analysis/lang/es"
	"github.com/blugelabs/bluge/analysis/lang/fa"
	"github.com/blugelabs/bluge/analysis/lang/fi"
	"github.com/blugelabs/bluge/analysis/lang/fr"
	"github.com/blugelabs/bluge/analysis/lang/hi"
	"github.com/blugelabs/bluge/analysis/lang/hu"
	"github.com/blugelabs/bluge/analysis/lang/it"
	"github.com/blugelabs/bluge/analysis/lang/nl"
	"github.com/blugelabs/bluge/analysis/lang/no"
	"github.com/blugelabs/bluge/analysis/lang/pt"
	"github.com/blugelabs/bluge/analysis/lang/ro"
	"github.com/blugelabs/bluge/analysis/lang/ru"
	"github.com/blugelabs/bluge/analysis/lang/sv"
	"github.com/blugelabs/bluge/analysis/lang/tr"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
	"github.com/mosuka/phalanx/analysis/lang/bg"
	"github.com/mosuka/phalanx/analysis/lang/ca"
	"github.com/mosuka/phalanx/analysis/lang/cs"
	"github.com/mosuka/phalanx/analysis/lang/el"
	"github.com/mosuka/phalanx/analysis/lang/eu"
	"github.com/mosuka/phalanx/analysis/lang/ga"
	"github.com/mosuka/phalanx/analysis/lang/hy"
	"github.com/mosuka/phalanx/analysis/lang/id"
	"github.com/mosuka/phalanx/analysis/lang/in"
	lingua "github.com/pemistahl/lingua-go"
	"go.uber.org/zap"
)

type AnalyzerDetector struct {
	languages []lingua.Language
	analyzers map[string]*analysis.Analyzer
	detector  lingua.LanguageDetector
	logger    *zap.Logger
}

func NewAnalyzerDetector(logger *zap.Logger) (*AnalyzerDetector, error) {

	languages := []lingua.Language{
		lingua.Arabic,     // ar
		lingua.Bulgarian,  // bg
		lingua.Catalan,    // ca
		lingua.Czech,      // cs
		lingua.Danish,     // da
		lingua.German,     // de
		lingua.Greek,      // el
		lingua.English,    // en
		lingua.Spanish,    // es
		lingua.Basque,     // eu
		lingua.Persian,    // fa
		lingua.Finnish,    // fi
		lingua.French,     // fr
		lingua.Irish,      // ga
		lingua.Hindi,      // hi
		lingua.Urdu,       // hi
		lingua.Hungarian,  // hu
		lingua.Armenian,   // hy
		lingua.Indonesian, // id
		lingua.Bengali,    // in
		lingua.Punjabi,    // in
		lingua.Marathi,    // in
		lingua.Gujarati,   // in
		lingua.Italian,    // it
		lingua.Japanese,   // ja
		lingua.Korean,     // ko
		lingua.Dutch,      // nl
		lingua.Nynorsk,    // no
		lingua.Bokmal,     // no
		lingua.Portuguese, // pt
		lingua.Romanian,   // ro
		lingua.Russian,    // ru
		lingua.Swedish,    // sv
		lingua.Turkish,    // tr
		lingua.Chinese,    // zh
	}

	analyzers := map[string]*analysis.Analyzer{
		"ar": ar.Analyzer(),    // ar
		"br": bg.Analyzer(),    // bg
		"ca": ca.Analyzer(),    // ca
		"cs": cs.Analyzer(),    // cs
		"da": da.Analyzer(),    // da
		"de": de.Analyzer(),    // de
		"el": el.Analyzer(),    // el
		"en": en.NewAnalyzer(), // en
		"es": es.Analyzer(),    // es
		"eu": eu.Analyzer(),    // eu
		"fa": fa.Analyzer(),    // fa
		"fi": fi.Analyzer(),    // fi
		"fr": fr.Analyzer(),    // fr
		"ga": ga.Analyzer(),    // ga
		"hi": hi.Analyzer(),    // hi
		"ur": hi.Analyzer(),    // ur -> hi
		"hu": hu.Analyzer(),    // hu
		"hy": hy.Analyzer(),    // hy
		"id": id.Analyzer(),    // id
		"bn": in.Analyzer(),    // bn -> in
		"pa": in.Analyzer(),    // pa -> in
		"mr": in.Analyzer(),    // mr -> in
		"gu": in.Analyzer(),    // gu -> in
		"it": it.Analyzer(),    // it
		"ja": ja.Analyzer(),    // ja
		"ko": cjk.Analyzer(),   // ko
		"nl": nl.Analyzer(),    // nl
		"nn": no.Analyzer(),    // nn -> no
		"nb": no.Analyzer(),    // nb -> no
		"pt": pt.Analyzer(),    // pt
		"ro": ro.Analyzer(),    // ro
		"ru": ru.Analyzer(),    // ru
		"sv": sv.Analyzer(),    // sv
		"tr": tr.Analyzer(),    // tr
		"zh": cjk.Analyzer(),   // zh
	}

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	return &AnalyzerDetector{
		languages: languages,
		analyzers: analyzers,
		detector:  detector,
		logger:    logger,
	}, nil
}

func (d *AnalyzerDetector) detectLanguage(text string) (string, bool) {
	language, exists := d.detector.DetectLanguageOf(text)
	if !exists {
		d.logger.Warn("language not detected", zap.String("language", language.String()), zap.String("text", text))
		return "", exists
	}

	iso639_1 := strings.ToLower(language.IsoCode639_1().String())
	// d.logger.Debug("language detected", zap.String("language", language.String()), zap.String("iso639_1", iso639_1), zap.String("text", text))
	return iso639_1, exists
}

func (d *AnalyzerDetector) getAnalyzer(iso639_1 string) *analysis.Analyzer {
	langAnalyzer, ok := d.analyzers[iso639_1]
	if !ok {
		d.logger.Warn("unsupported language. return standard analyzer", zap.String("iso639_1", iso639_1))
		return analyzer.NewStandardAnalyzer()
	}

	// d.logger.Debug("analyzer found", zap.String("iso639_1", iso639_1))
	return langAnalyzer
}

func (d *AnalyzerDetector) DetectAnalyzer(text string) *analysis.Analyzer {
	iso639_1, exists := d.detectLanguage(text)
	if !exists {
		return analyzer.NewStandardAnalyzer()
	}

	return d.getAnalyzer(iso639_1)
}
