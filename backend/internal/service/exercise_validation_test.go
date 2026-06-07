package service

import "testing"

func TestMuscleGroupSlug(t *testing.T) {
	cases := map[string]string{
		"Pecho":     "chest",     // Spanish label
		"pecho":     "chest",     // lower case
		"Tríceps":   "triceps",   // accented label
		"Hombro":    "shoulders", // singular of a plural label ("Hombros")
		"Hombros":   "shoulders", // plural label
		"chest":     "chest",     // already a slug
		"full body": "full_body", // spaced slug
		"Glúteo":    "glutes",    // singular accented label
		"unknown":   "",          // unresolved
		"":          "",          // blank
	}

	for input, want := range cases {
		if got := muscleGroupSlug(input); got != want {
			t.Errorf("muscleGroupSlug(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestMuscleGroupSlugsDedupesAndDropsBlanks(t *testing.T) {
	got := muscleGroupSlugs([]string{"Pecho", "Hombro", "Tríceps", "chest", "", "   "})

	want := []string{"chest", "shoulders", "triceps"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
