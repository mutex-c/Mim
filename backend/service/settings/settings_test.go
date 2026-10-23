package settings_test

import (
	"testing"

	"mim/service/settings"
)

func TestDFAFilterAndSettings(t *testing.T) {
	initial := settings.SystemConfig{
		SiteTitle:       "Mim Enterprise",
		MaxUploadSizeMB: 100,
		RecallTimeLimit: 120,
		SensitiveWords:  []string{"spam", "scam", "badword"},
	}

	svc := settings.NewService(initial)

	// 1. Test Sensitive Word Filtering
	input := "Hello, this is not a scam or spam message, please enjoy badword free chat."
	filtered := svc.SanitizeText(input)
	expected := "Hello, this is not a **** or **** message, please enjoy ******* free chat."

	if filtered != expected {
		t.Errorf("filtered output mismatch:\ngot:      %s\nexpected: %s", filtered, expected)
	}

	// 2. Test Dynamic Update
	newCfg := svc.GetConfig()
	newCfg.SensitiveWords = append(newCfg.SensitiveWords, "fraud")
	svc.UpdateConfig(newCfg)

	fraudText := "Detect fraud immediately"
	newFiltered := svc.SanitizeText(fraudText)
	if newFiltered != "Detect ***** immediately" {
		t.Errorf("expected fraud to be masked, got: %s", newFiltered)
	}
}
