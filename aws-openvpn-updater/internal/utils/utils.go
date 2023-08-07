package utils

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func GetFireSignalsChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		syscall.SIGHUP,  // "terminal is disconnected"
	)
	return c
}

func CreateFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func ReadContentFromFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	// Iterate from the end of the file content and remove empty lines.
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lines = lines[:i+1]
			break
		}
	}

	trimmedContent := strings.Join(lines, "\n")
	return trimmedContent, nil
}

func ReadCertificateFromFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	var certificate []string
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "-----") {
			found = true
		}
		if found && strings.TrimSpace(lines[i]) != "" {
			certificate = append(certificate, lines[i])
		}
	}

	certificateContent := strings.Join(certificate, "\n")
	return certificateContent, nil
}

func ExtractEmail(user string, domain string) (string, error) {
	pattern := `^[a-zA-Z-]+\.[a-zA-Z-]+`
	re := regexp.MustCompile(pattern)
	matches := re.FindString(user)
	mail := matches + "@" + domain
	if len(mail) == (len(domain) + 1) {
		return "", errors.New("Couldn't extract user email")
	}
	return mail, nil
}
