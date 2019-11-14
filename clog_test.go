package clog

import "testing"

// see logLevels map
var testValidMap = map[Level]string{
	InvalidLevel:  "",
	DisabledLevel: "disabled",
	ErrorLevel:    "error",
	WarnLevel:     "warning",
	InfoLevel:     "info",
	DebugLevel:    "debug",
}

func TestLevelFromString(t *testing.T) {
	for l, ls := range testValidMap {
		lx, err := LevelFromString(ls)
		if err != nil {
			t.Errorf("processing: %q - unexpected error: %v", ls, err)
		}
		if lx != l {
			t.Errorf("processing: %q - expected: %d (%q), got: %d, (%q)",
				ls, l, l.String(), lx, lx.String())
		}
	}

	str := "nonsense"
	l, err := LevelFromString(str)
	if err == nil {
		t.Errorf("processing: %q - expected error, got nil", str)
	}
	if l != InvalidLevel {
		t.Errorf("processing: %q - expected %d (%q), got %d (%q)",
			str, InvalidLevel, InvalidLevel.String(), l, l.String())
	}
}

func TestLevelString(t *testing.T) {
	for l, ls := range testValidMap {
		if l.String() != ls {
			t.Errorf("(%d).String() expected: %s, got: %s",
				l, ls, l.String())
		}
	}

	l := Level(999)
	if l.String() != "" {
		t.Errorf("invalid level 999 should return empty string, got %s", l.String())
	}
}

func TestLevelValidate(t *testing.T) {
	for l := range testValidMap {
		err := l.Validate()
		if err != nil {
			if l == InvalidLevel {
				continue // expected error
			}
			t.Errorf("%d (%q) should be valid, unexpected error: %v",
				l, l.String(), err)
		}
	}

	if Level(999).Validate() == nil {
		t.Error("invalid level 999 should return error, got nil")
	}
}
