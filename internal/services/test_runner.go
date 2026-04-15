package services

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type TestRunner struct {
	dockerClient *client.Client
}

type TestResult struct {
	Grade    int
	Comments string
}

func NewTestRunner() (*TestRunner, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &TestRunner{
		dockerClient: cli,
	}, nil
}

func (tr *TestRunner) Pytest(filename, studentCode, testCode string) (*TestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create container
	// pip install cmd
	pytestRunCmd := "pip install pytest > /dev/null 2>&1 && python -m pytest --no-header --color=no"
	resp, err := tr.dockerClient.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config: &container.Config{
				Image: "python:3.11-slim",
				Cmd: []string{
					"sh", "-c",
					pytestRunCmd,
				},
				WorkingDir: "/app",
				Tty:        true,
			},
			HostConfig: &container.HostConfig{
				Resources: container.Resources{
					Memory:   128 * 1024 * 1024, // 128 MB
					NanoCPUs: 1_000_000_000,     // 1 CPU
				},
			},
			NetworkingConfig: nil,
			Platform:         nil,
			Name:             "",
		},
	)

	if err != nil {
		return nil, err
	}

	containerId := resp.ID
	defer tr.cleanup(ctx, containerId)

	if err := tr.copyFilesToContainer(ctx, containerId, filename, studentCode, testCode); err != nil {
		return nil, err
	}

	if _, err := tr.dockerClient.ContainerStart(ctx, containerId, client.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	res := tr.dockerClient.ContainerWait(ctx, containerId, client.ContainerWaitOptions{})

	select {
	case err := <-res.Error:
		if err != nil {
			return nil, err
		}

	case <-res.Result:
	// Container finished successfully

	case <-ctx.Done():
		return nil, fmt.Errorf("test execution timeout")
	}

	// get pytest ouptut
	logs, err := tr.dockerClient.ContainerLogs(ctx, containerId, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		return nil, err
	}
	defer logs.Close()

	stdouterr, err := io.ReadAll(logs) // blank
	if err != nil {
		return nil, err
	}
	return tr.parseTestResults(string(stdouterr)), nil
}

func (tr *TestRunner) writeToTar(tw *tar.Writer, name string, content []byte) error {
	tarHeader := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(content)),
	}

	err := tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	_, err = tw.Write(content)
	return err
}

func (tr *TestRunner) copyFilesToContainer(ctx context.Context, containerID, filename, studentCode, testCode string) error {
	// Read conftest.py
	conftestContent, err := os.ReadFile("conftest/conftest.py")
	if err != nil {
		return err
	}

	// Create tar archive with both files
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if err := tr.writeToTar(tw, filename, []byte(studentCode)); err != nil {
		return err
	}

	if err := tr.writeToTar(tw, "test_"+filename, []byte(testCode)); err != nil {
		return err
	}

	if err := tr.writeToTar(tw, "conftest.py", []byte(conftestContent)); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	// Copy tar archive to container
	_, err = tr.dockerClient.CopyToContainer(
		ctx,
		containerID,
		client.CopyToContainerOptions{
			DestinationPath:           "/app",
			Content:                   &buf,
			AllowOverwriteDirWithFile: false,
			CopyUIDGID:                true,
		},
	)
	return err
}

func (tr *TestRunner) parseTestResults(output string) *TestResult {
	// Parse pytest output for html

	result := &TestResult{
		Grade:    0,
		Comments: output,
	}

	var passed, total int
	// Parse pytest output
	// Look for pattern like "1 passed, 2 failed"
	re := regexp.MustCompile(`(\d+) passed`)
	if matches := re.FindStringSubmatch(output); len(matches) > 1 {
		passed, _ = strconv.Atoi(matches[1])
	}

	// Count total tests
	reTotal := regexp.MustCompile(`(\d+) (passed|failed)`)
	matches := reTotal.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		count, _ := strconv.Atoi(match[1])
		total += count
	}

	if total == 0 {
		result.Grade = 0
	} else {
		result.Grade = (passed*4 + total/2) / total
	}

	return result
}

func (tr *TestRunner) cleanup(ctx context.Context, containerID string) {
	tr.dockerClient.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
		Force: true,
	})
}
