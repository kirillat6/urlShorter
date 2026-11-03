package main

import "testing"

func TestSaveLink(t *testing.T) {
	storage := NewLinkStorage()
	res := storage.SaveLink("abc123", "https://google.com")
	if res != true {
		t.Errorf("Ожидали true, получили: %v", res)
	}

	res2 := storage.SaveLink("abc123", "https://example.com")
	if res2 {
		t.Error("Данный code уже существует!")
	}
}
