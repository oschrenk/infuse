package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestRemoveExclude(t *testing.T) {
	cases := []struct {
		name, before, remove, after string
	}{
		{"middle line", "# c\na\ntasks/X.md\nb\n", "tasks/X.md", "# c\na\nb\n"},
		{"absent line", "a\nb\n", "tasks/X.md", "a\nb\n"},
		{"prefix collision", "tasks/done/X.md\ntasks/X.md\n", "tasks/X.md", "tasks/done/X.md\n"},
		{"substring collision", "tasks/X.md\nX.md\n", "X.md", "tasks/X.md\n"},
		{"last line without newline", "a\ntasks/X.md", "tasks/X.md", "a"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if _, err := git.PlainInit(dir, false); err != nil {
				t.Fatal(err)
			}
			excludePath := filepath.Join(dir, ".git", "info", "exclude")
			if err := os.MkdirAll(filepath.Dir(excludePath), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(excludePath, []byte(c.before), 0644); err != nil {
				t.Fatal(err)
			}
			repo, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if err := repo.RemoveExclude(c.remove); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(excludePath)
			if string(got) != c.after {
				t.Errorf("got %q, want %q", got, c.after)
			}
		})
	}
}

func TestRemoveExcludeMissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := git.PlainInit(dir, false); err != nil {
		t.Fatal(err)
	}
	repo, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.RemoveExclude("x"); err != nil {
		t.Fatal(err)
	}
}
