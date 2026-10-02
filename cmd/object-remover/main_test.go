package main

import (
	"testing"
)

func TestParseFlags_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "both --bucket and --prefix are missing",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "--bucket specified but --prefix is missing",
			args:    []string{"--bucket", "my-bucket"},
			wantErr: true,
		},
		{
			name:    "--prefix specified but --bucket is missing",
			args:    []string{"--prefix", "my-prefix/"},
			wantErr: true,
		},
		{
			name:    "both --bucket and --prefix are specified",
			args:    []string{"--bucket", "my-bucket", "--prefix", "my-prefix/"},
			wantErr: false,
		},
		{
			name:    "all flags specified",
			args:    []string{"--bucket", "my-bucket", "--prefix", "my-prefix/", "--region", "ap-northeast-1", "--yes"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFlags(tt.args)
			if tt.wantErr && err == nil {
				t.Errorf("expected an error, but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseFlags_Values(t *testing.T) {
	args := []string{
		"--bucket", "my-bucket",
		"--prefix", "my-prefix/",
		"--region", "ap-northeast-1",
		"--yes",
	}

	f, err := parseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.bucket != "my-bucket" {
		t.Errorf("bucket: got %q, want %q", f.bucket, "my-bucket")
	}
	if f.prefix != "my-prefix/" {
		t.Errorf("prefix: got %q, want %q", f.prefix, "my-prefix/")
	}
	if f.region != "ap-northeast-1" {
		t.Errorf("region: got %q, want %q", f.region, "ap-northeast-1")
	}
	if !f.yes {
		t.Errorf("yes: got false, want true")
	}
}

func TestParseFlags_Help(t *testing.T) {
	f, err := parseFlags([]string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.help {
		t.Errorf("help: got false, want true")
	}
}
