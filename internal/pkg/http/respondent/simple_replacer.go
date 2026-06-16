package respondent

import "errors"

type SimpleReplacer struct {
	rules []replacementRule
}

type replacementRule struct {
	original    error
	replacement error
}

func NewSimpleReplacer() *SimpleReplacer {
	return &SimpleReplacer{
		rules: make([]replacementRule, 0),
	}
}

func (sr *SimpleReplacer) ReplaceBy(original, replacement error) *SimpleReplacer {
	if original == nil || replacement == nil || errors.Is(original, replacement) {
		return sr
	}
	sr.rules = append(sr.rules, replacementRule{original: original, replacement: replacement})
	return sr
}

func (sr *SimpleReplacer) Replace(err error) error {
	if err == nil {
		return nil
	}
	for _, rule := range sr.rules {
		if errors.Is(err, rule.original) {
			return rule.replacement
		}
	}
	return err
}
