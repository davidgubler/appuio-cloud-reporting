package testsuite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/hashicorp/go-getter"
)

const (
	Version = "2.32.1"

	baseFileNameTmplStr = "prometheus-{{.Version}}.{{.GOOS}}-{{.GOARCH}}"
)

var (
	baseFileNameTmpl = template.Must(template.New("baseFileName").Parse(baseFileNameTmplStr))
	downloadURLTmpl  = template.Must(template.New("downloadURL").Parse(
		"https://github.com/prometheus/prometheus/releases/download/v{{.Version}}/" + baseFileNameTmplStr + ".tar.gz",
	))
)

// StartPrometheus starts a new prometheus instance on the given port. Not safe to call concurrently.
// Cancel the context to stop.
// The returned cleanup function block until prometheus is stopped. Cleanup has to be called.
func StartPrometheus(ctx context.Context, port int) (cleanup func() error, err error) {
	promPath, err := EnsurePrometheus(ctx, Version)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "prom-data-*")
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, promPath,
		fmt.Sprintf("--web.listen-address=127.0.0.1:%d", port),
		fmt.Sprintf("--storage.tsdb.path=%s", tmpDir),
	)
	cmd.Dir = filepath.Dir(promPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return func() error {
		defer os.RemoveAll(tmpDir)
		return cmd.Wait()
	}, cmd.Start()
}

// EnsurePrometheus ensures the prometheus binary is downloaded. Not safe to call concurrently.
func EnsurePrometheus(ctx context.Context, version string) (string, error) {
	promPath, err := prometheusPath(Version)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(promPath); err != nil && errors.Is(err, os.ErrNotExist) {
		if err := downloadPrometheus(ctx, Version); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", fmt.Errorf("unknown error while checking for prometheus executable: %w", err)
	}
	return promPath, err
}

func downloadPrometheus(ctx context.Context, version string) error {
	fmt.Fprintln(os.Stderr, "Fetching prometheus 📥")
	url, err := downloadURL(version)
	if err != nil {
		return err
	}

	acd, err := appCacheDir()
	if err != nil {
		return err
	}

	err = getter.Get(acd, url, getter.WithContext(ctx))
	if err != nil {
		return err
	}

	return err
}

func appCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("could not get UserCacheDir: %w", err)
	}
	appdir := filepath.Join(dir, "appuio-cloud-reporting-test")
	if err := os.MkdirAll(appdir, 0755); err != nil {
		return "", fmt.Errorf("could not create new directory in user cache dir: %w", err)
	}
	return appdir, nil
}

func downloadURL(version string) (string, error) {
	return applyTemplate(downloadURLTmpl, version)
}

func prometheusPath(version string) (string, error) {
	acd, err := appCacheDir()
	if err != nil {
		return "", err
	}

	installDir, err := applyTemplate(baseFileNameTmpl, version)
	if err != nil {
		return "", err
	}

	return filepath.Join(acd, installDir, "prometheus"), nil
}

func applyTemplate(tmpl *template.Template, version string) (string, error) {
	b := new(strings.Builder)
	err := tmpl.Execute(b, struct {
		Version, GOOS, GOARCH string
	}{
		version, runtime.GOOS, runtime.GOARCH,
	})
	return b.String(), err
}
