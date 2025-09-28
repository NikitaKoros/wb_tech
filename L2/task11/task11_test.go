package main

import (
	"reflect"
	"testing"
)

func TestFindAnagrams(t *testing.T) {
	testCases := []struct {
		name  string
		words []string
		want  map[string][]string
	}{
		{
			name:  "empty slice",
			words: []string{},
			want:  map[string][]string{},
		},
		{
			name:  "slice with empty string",
			words: []string{""},
			want:  map[string][]string{},
		},
		{
			name:  "no anagrams",
			words: []string{"кот", "собака", "дом"},
			want:  map[string][]string{},
		},
		{
			name:  "simple anagrams",
			words: []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"},
			want: map[string][]string{
				"пятак":  {"пятак", "пятка", "тяпка"},
				"листок": {"листок", "слиток", "столик"},
			},
		},
		{
			name:  "case insensitivity",
			words: []string{"ПЯТАК", "пятка", "Тяпка"},
			want: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "duplicates in input",
			words: []string{"пятак", "пятак", "пятка", "тяпка"},
			want: map[string][]string{
				"пятак": {"пятак", "пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "latin anagrams",
			words: []string{"listen", "silent", "enlist", "stone", "tones"},
			want: map[string][]string{
				"enlist": {"enlist", "listen", "silent"},
				"stone":  {"stone", "tones"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindAnagrams(tc.words)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("FindAnagrams(%v) = %#v, want %#v", tc.words, got, tc.want)
			}
		})
	}

}
