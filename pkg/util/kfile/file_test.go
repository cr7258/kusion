//go:build !arm64
// +build !arm64

package kfile

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
)

const (
	mockHomeDir = "/home/test"
	mockToken   = "testtoken"
)

func TestFileExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "file exists",
			args: args{
				filename: "./testdata/test.txt",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "file does not exists",
			args: args{
				filename: "test.txt",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileExists(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKusionDataFolder(t *testing.T) {
	// Run test
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    filepath.Join(mockHomeDir, ".kusion"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			// Mock data
			os.Setenv(EnvKusionPath, "")
			mockUserCurrent()
			mockMkdirall()
			got, err := KusionDataFolder()
			if (err != nil) != tt.wantErr {
				t.Errorf("KusionDataFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KusionDataFolder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCredentialsToken(t *testing.T) {
	// Run test
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: mockToken,
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			// Mock data
			mockUserCurrent()
			mockMkdirall()
			mockReadFile()
			if got := GetCredentialsToken(); got != tt.want {
				t.Errorf("GetCredentialsToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockUserCurrent() {
	mockey.Mock(user.Current).To(func() (*user.User, error) {
		return &user.User{
			HomeDir: mockHomeDir,
		}, nil
	}).Build()
}

func mockMkdirall() {
	mockey.Mock(os.MkdirAll).To(func(path string, perm os.FileMode) error {
		return nil
	}).Build()
}

func mockReadFile() {
	mockey.Mock(os.ReadFile).To(func(filename string) ([]byte, error) {
		return []byte(fmt.Sprintf(`{"token": "%s"}`, mockToken)), nil
	}).Build()
}
