package settings

import (
	"strings"
	"sync"
)

type SystemConfig struct {
	SiteTitle         string   `json:"site_title"`
	MaxUploadSizeMB   int      `json:"max_upload_size_mb"`
	RecallTimeLimit   int      `json:"recall_time_limit"`
	Notice            string   `json:"notice"`
	AllowRegistration bool     `json:"allow_registration"`
	SensitiveWords    []string `json:"sensitive_words"`
}

type DFAFilter struct {
	root  map[rune]interface{}
	isEnd map[rune]bool
	mu    sync.RWMutex
}

func NewDFAFilter(words []string) *DFAFilter {
	f := &DFAFilter{
		root:  make(map[rune]interface{}),
		isEnd: make(map[rune]bool),
	}
	f.Build(words)
	return f
}

func (f *DFAFilter) Build(words []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.root = make(map[rune]interface{})
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		current := f.root
		runes := []rune(word)
		for i, r := range runes {
			if _, exists := current[r]; !exists {
				current[r] = make(map[rune]interface{})
			}
			if i == len(runes)-1 {
				// Mark word end
				current[r].(map[rune]interface{})[0] = true
			}
			current = current[r].(map[rune]interface{})
		}
	}
}

// Filter replaces sensitive words with replacement string (e.g. "***")
func (f *DFAFilter) Filter(text string, replaceRune rune) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	runes := []rune(text)
	length := len(runes)
	result := make([]rune, length)
	copy(result, runes)

	for i := 0; i < length; i++ {
		current := f.root
		matchLen := 0
		foundEnd := false

		for j := i; j < length; j++ {
			r := runes[j]
			next, ok := current[r]
			if !ok {
				break
			}
			matchLen++
			nextMap := next.(map[rune]interface{})
			if _, ok := nextMap[0]; ok {
				foundEnd = true
				break
			}
			current = nextMap
		}

		if foundEnd {
			for k := i; k < i+matchLen; k++ {
				result[k] = replaceRune
			}
			i += matchLen - 1
		}
	}

	return string(result)
}

type Service struct {
	mu     sync.RWMutex
	cfg    SystemConfig
	filter *DFAFilter
}

func NewService(initial SystemConfig) *Service {
	return &Service{
		cfg:    initial,
		filter: NewDFAFilter(initial.SensitiveWords),
	}
}

func (s *Service) GetConfig() SystemConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Service) UpdateConfig(newCfg SystemConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = newCfg
	s.filter.Build(newCfg.SensitiveWords)
}

func (s *Service) SanitizeText(text string) string {
	return s.filter.Filter(text, '*')
}
