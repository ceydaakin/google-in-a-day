package crawler

import (
	"testing"
)

func TestWordFrequencyEmpty(t *testing.T) {
	freq := wordFrequency("")
	if len(freq) != 0 {
		t.Errorf("expected empty map, got %v", freq)
	}
}

func TestWordFrequencySingleWord(t *testing.T) {
	freq := wordFrequency("hello")
	if freq["hello"] != 1 {
		t.Errorf("expected hello=1, got %d", freq["hello"])
	}
}

func TestWordFrequencyRepeated(t *testing.T) {
	freq := wordFrequency("go go go")
	if freq["go"] != 3 {
		t.Errorf("expected go=3, got %d", freq["go"])
	}
}

func TestWordFrequencySingleCharExcluded(t *testing.T) {
	freq := wordFrequency("I am a developer")
	if _, ok := freq["i"]; ok {
		t.Error("single char 'i' should be excluded")
	}
	if _, ok := freq["a"]; ok {
		t.Error("single char 'a' should be excluded")
	}
	if freq["am"] != 1 {
		t.Errorf("expected am=1, got %d", freq["am"])
	}
	if freq["developer"] != 1 {
		t.Errorf("expected developer=1, got %d", freq["developer"])
	}
}

func TestWordFrequencyCaseNormalized(t *testing.T) {
	freq := wordFrequency("Go GO gO go")
	if freq["go"] != 4 {
		t.Errorf("expected go=4, got %d", freq["go"])
	}
}

func TestWordFrequencyPunctuation(t *testing.T) {
	freq := wordFrequency("hello, world! hello.")
	if freq["hello"] != 2 {
		t.Errorf("expected hello=2, got %d", freq["hello"])
	}
	if freq["world"] != 1 {
		t.Errorf("expected world=1, got %d", freq["world"])
	}
}
