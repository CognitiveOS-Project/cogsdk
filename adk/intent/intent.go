package intent

import (
	"fmt"
	"strings"
)

// Intent is a recognized user intent with associated parameters.
type Intent struct {
	Name   string
	Params map[string]interface{}
	Score  float64
}

// Recognizer matches raw input against a set of known intents.
type Recognizer struct {
	intents map[string]*IntentPattern
}

// IntentPattern defines a single intent pattern with keywords to match.
type IntentPattern struct {
	Name     string
	Keywords []string
	Params   map[string]string
}

// NewRecognizer creates a new intent recognizer with no patterns.
func NewRecognizer() *Recognizer {
	return &Recognizer{
		intents: make(map[string]*IntentPattern),
	}
}

// RegisterIntent registers a new intent pattern with the recognizer.
func (r *Recognizer) RegisterIntent(name string, keywords []string, params map[string]string) {
	r.intents[name] = &IntentPattern{
		Name:     name,
		Keywords: keywords,
		Params:   params,
	}
}

// Recognize classifies input text into the best matching intent.
//
// Returns the matched intent and a confidence score. Returns nil intent
// with score 0 if no pattern matches.
func (r *Recognizer) Recognize(input string) (*Intent, error) {
	inputLower := strings.ToLower(input)

	var best *IntentPattern
	var bestScore float64

	for _, p := range r.intents {
		score := 0.0
		for _, kw := range p.Keywords {
			if strings.Contains(inputLower, strings.ToLower(kw)) {
				score++
			}
		}
		if score > 0 && score > bestScore {
			best = p
			bestScore = score
		}
	}

	if best == nil || bestScore == 0 {
		return nil, nil
	}

	params := make(map[string]interface{})
	for k, v := range best.Params {
		params[k] = v
	}

	return &Intent{
		Name:   best.Name,
		Params: params,
		Score:  bestScore / float64(len(best.Keywords)),
	}, nil
}

// String returns a string representation of the intent.
func (i *Intent) String() string {
	return fmt.Sprintf("%s(score=%.2f)", i.Name, i.Score)
}