package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	resolvConfPath   = "/etc/resolv.conf"
	hostnameFilePath = "/etc/hostname"
)

type FileSystemDNSManager struct {
	log zerolog.Logger
}

func NewFileSystemDNSManager(log zerolog.Logger) *FileSystemDNSManager {
	return &FileSystemDNSManager{log: log}
}

func backupHostname() (string, error) {
	const op = "backupHostname"
	l := zerolog.New(os.Stdout).With().Str("op", op).Logger()

	hostname, err := os.ReadFile(hostnameFilePath)
	if err != nil {
		return "", fmt.Errorf("op: %s, failed to read %s for backup: %w", op, hostnameFilePath, err)
	}

	backupFileName := fmt.Sprintf("%s-%d", hostnameFilePath, time.Now().Unix())
	err = os.WriteFile(backupFileName, hostname, 0644)
	if err != nil {
		return "", fmt.Errorf("op: %s, failed to create backup file: %w", op, err)
	}

	l.Info().Str("backupFileName", backupFileName).Msg("Hostname backup created successfully")
	return backupFileName, nil
}

func revertHostname(backupFileName string) error {
	const op = "revertHostname"
	l := log.With().Str("op", op).Logger()

	backup, err := os.ReadFile(backupFileName)
	if err != nil {
		return fmt.Errorf("op: %s, failed to read backup file: %w", op, err)
	}

	err = os.WriteFile(hostnameFilePath, backup, 0644)
	if err != nil {
		return fmt.Errorf("op: %s, failed to restore backup to %s: %w", op, hostnameFilePath, err)
	}

	err = exec.Command("hostname", strings.TrimSpace(string(backup))).Run()
	if err != nil {
		return fmt.Errorf("op: %s, failed to set hostname: %w", op, err)
	}

	l.Info().Str("backupFileName", backupFileName).Msg("Hostname reverted successfully")
	return nil
}

func (m *FileSystemDNSManager) SetHostname(ctx context.Context, hostname string) error {
	const op = "SetHostname"
	l := m.log.With().Str("op", op).Logger()

	l.Info().Str("hostname", hostname).Msg("Setting hostname")

	backupFileName, err := backupHostname()
	if err != nil {
		return err
	}

	err = exec.CommandContext(ctx, "hostname", hostname).Run()
	if err != nil {
		if revertErr := revertHostname(backupFileName); revertErr != nil {
			l.Error().Err(revertErr).Msg("Failed to revert hostname")
		}
		return fmt.Errorf("op: %s, failed to set hostname: %w", op, err)
	}

	err = os.WriteFile(hostnameFilePath, []byte(hostname+"\n"), 0644)
	if err != nil {
		if revertErr := revertHostname(backupFileName); revertErr != nil {
			l.Error().Err(revertErr).Msg("Failed to revert hostname")
		}
		return fmt.Errorf("op: %s, failed to write to %s: %w", op, hostnameFilePath, err)
	}

	l.Info().Str("hostname", hostname).Msg("Hostname set successfully")
	return nil
}

func backupResolvConf() (string, error) {
	const op = "backupResolvConf"
	l := log.With().Str("op", op).Logger()

	input, err := os.Open(resolvConfPath)
	if err != nil {
		return "", fmt.Errorf("op: %s, failed to open %s for backup: %w", op, resolvConfPath, err)
	}
	defer func() {
		if err := input.Close(); err != nil {
			l.Warn().Str("failed to close input file", fmt.Sprint(err))
		}
	}()

	backupFileName := fmt.Sprintf("%s-%d", resolvConfPath, time.Now().Unix())
	output, err := os.Create(backupFileName)
	if err != nil {
		return "", fmt.Errorf("op: %s, failed to create backup file: %w", op, err)
	}
	defer closeFile(output)

	_, err = io.Copy(output, input)
	if err != nil {
		return "", fmt.Errorf("failed to copy contents to backup file: %w", err)
	}

	l.Info().Str("backupFileName", backupFileName).Msg("Backup created successfully")
	return backupFileName, nil
}

func revertResolvConf(backupFileName string) error {
	const op = "revertResolvConf"
	l := log.With().Str("op", op).Logger()

	input, err := os.Open(backupFileName)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open backup file: %w", op, err)
	}
	defer closeFile(input)

	output, err := os.Create(resolvConfPath)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open %s for writing: %w", op, resolvConfPath, err)
	}
	defer closeFile(output)

	_, err = io.Copy(output, input)
	if err != nil {
		return fmt.Errorf("op: %s, failed to copy contents from backup file: %w", op, err)
	}

	l.Info().Str("backupFileName", backupFileName).Msg("Reverted to backup successfully")
	return nil
}

func (m *FileSystemDNSManager) findDNSServer(file *os.File, server string) error {
	const op = "findDNSServer"

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == server {
				return fmt.Errorf("op: %s, DNS server %s already exists", op, server)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("op: %s, error reading file: %w", op, err)
	}

	return nil
}

func (m *FileSystemDNSManager) AddDNSServer(ctx context.Context, server string) error {
	const op = "AddDNSServer"
	l := log.With().Str("op", op).Logger()

	l.Info().Str("server", server).Msg("Adding DNS server")

	backupFileName, err := backupResolvConf()
	if err != nil {
		return err
	}

	file, err := os.Open(resolvConfPath)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open %s: %w", op, resolvConfPath, err)
	}
	defer closeFile(file)

	if err = m.findDNSServer(file, server); err != nil {
		return err
	}

	file, err = os.OpenFile(resolvConfPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open %s: %w", op, resolvConfPath, err)
	}
	defer closeFile(file)

	if _, err := file.WriteString(fmt.Sprintf("nameserver %s\n", server)); err != nil {
		if revertErr := revertResolvConf(backupFileName); revertErr != nil {
			l.Error().Err(revertErr).Msg("Failed to revert")
		}
		return fmt.Errorf("op: %s, failed to write to %s: %w", op, resolvConfPath, err)
	}

	l.Info().Str("server", server).Msg("DNS server added successfully")
	return nil
}

func (m *FileSystemDNSManager) RemoveDNSServer(ctx context.Context, server string) error {
	const op = "RemoveDNSServer"
	l := log.With().Str("op", op).Logger()

	l.Info().Str("server", server).Msg("Removing DNS server")
	backupFileName, err := backupResolvConf()
	if err != nil {
		return err
	}

	file, err := os.Open(resolvConfPath)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open %s: %w", op, resolvConfPath, err)
	}
	defer closeFile(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != fmt.Sprintf("nameserver %s", server) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("op: %s, error reading %s: %w", op, resolvConfPath, err)
	}

	file, err = os.OpenFile(resolvConfPath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("op: %s, failed to open %s for writing: %w", op, resolvConfPath, err)
	}
	defer closeFile(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			if revertErr := revertResolvConf(backupFileName); revertErr != nil {
				l.Error().Err(revertErr).Msg("Failed to revert")
			}
			return fmt.Errorf("op: %s, failed to write to %s: %w", op, resolvConfPath, err)
		}
	}

	if err := writer.Flush(); err != nil {
		if revertErr := revertResolvConf(backupFileName); revertErr != nil {
			l.Error().Err(revertErr).Msg("Failed to revert")
		}
		return fmt.Errorf("op: %s, failed to flush writes to %s: %w", op, resolvConfPath, err)
	}

	l.Info().Str("server", server).Msg("DNS server removed successfully")
	return nil
}

func (m *FileSystemDNSManager) ListDNSServers(ctx context.Context) ([]string, error) {
	const op = "ListDNSServers"
	l := log.With().Str("op", op).Logger()

	l.Info().Msg("Listing DNS servers")

	file, err := os.Open(resolvConfPath)
	if err != nil {
		return nil, fmt.Errorf("op: %s, failed to open %s: %w", op, resolvConfPath, err)
	}
	defer closeFile(file)

	var dnsServers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				dnsServers = append(dnsServers, fields[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("op: %s, error reading %s: %w", op, resolvConfPath, err)
	}

	l.Info().Int("count", len(dnsServers)).Msg("Listed DNS servers successfully")
	return dnsServers, nil
}

func closeFile(f *os.File) {
	const op = "closeFile"
	l := log.With().Str("op", op).Logger()

	if err := f.Close(); err != nil {
		l.Warn().Err(err).Msg("Failed to close file")
	}
}
