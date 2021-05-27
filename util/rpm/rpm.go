package rpm

import (
	"regexp"
	"strings"
)

// Spec spec information
type Spec struct {
	values map[string]string
	lines  []string
}

// NewSpec new a spec instance include information about a spec file
func NewSpec(content string) *Spec {
	s := &Spec{
		values: make(map[string]string),
	}
	s.lines = strings.Split(content, "\n")
	err := s.parse()
	if err != nil {
		return nil
	}
	return s
}

// preproccess read macros define
func (s *Spec) preproccess() map[string]string {
	re := regexp.MustCompile(`^\s*(?:%define|%global)\s+(?P<key>\w+)\s+(?P<value>[^\s]+)`)
	macros := make(map[string]string)
	for _, line := range s.lines {
		match := re.FindStringSubmatch(line)
		if match != nil {
			macros[match[1]] = match[2]
		}
	}
	return macros
}

// substitute replace macros by value
func (s *Spec) substitute(macros map[string]string) {
	for k, v := range macros {
		re := regexp.MustCompile(`%{\??` + k + `}`)
		for i, line := range s.lines {
			s.lines[i] = re.ReplaceAllString(line, v)
		}
	}
}

// extract
func (s *Spec) extract() {
	keys := []string{"Version", "Release"}
	for _, key := range keys {
		re := regexp.MustCompile(`^\s*` + key + `\s*:\s+([^\s]+)`)
		for _, line := range s.lines {
			match := re.FindStringSubmatch(line)
			if match != nil {
				s.values[key] = match[1]
			}
		}
	}
}

func (s *Spec) parse() error {
	macros := s.preproccess()
	s.substitute(macros)
	s.extract()
	return nil
}

// Version get Version from spec
func (s *Spec) Version() string {
	return s.values["Version"]
}

// Release get Release from spec
func (s *Spec) Release() string {
	return s.values["Release"]
}
