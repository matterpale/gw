package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// worktreeRepo makes a real git repository with a branch "foo" and a tmux stub on
// PATH, so the test never touches a running tmux server.
func worktreeRepo(t *testing.T, ignoreCase string) string {
	t.Helper()
	root := t.TempDir()
	stubs := filepath.Join(root, "bin")
	if err := os.MkdirAll(stubs, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stubs, "tmux"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", stubs+":"+os.Getenv("PATH"))

	repo := filepath.Join(root, "repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", repo},
		{"-C", repo, "config", "core.ignorecase", ignoreCase},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "i"},
		{"-C", repo, "branch", "foo"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return repo
}

func TestAddWorktreeRejectsBranchThatDiffersOnlyByCase(t *testing.T) {
	repo := worktreeRepo(t, "true")

	_, _, err := addWorktree(repo, "repo", "Foo", "")

	if err == nil {
		t.Fatal("expected a case collision error")
	}
	if !strings.Contains(err.Error(), `"foo"`) {
		t.Fatalf("error does not name the existing branch: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(filepath.Dir(repo), "repo-Foo")); statErr == nil {
		t.Fatal("a worktree was created despite the collision")
	}
}

func TestAddWorktreeAcceptsExactBranchName(t *testing.T) {
	repo := worktreeRepo(t, "true")

	wt, _, err := addWorktree(repo, "repo", "foo", "")

	if err != nil {
		t.Fatalf("addWorktree: %v", err)
	}
	if wt.Branch != "foo" {
		t.Fatalf("branch = %q", wt.Branch)
	}
}
