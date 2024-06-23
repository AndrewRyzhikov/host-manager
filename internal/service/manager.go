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
	const op = "NewFileSystemDNSManager"
	return &FileSystemDNSManager{log: log.With().Str("op", op).Logger()}
}

func backupHostname() (string, error) {
	hostname, err := os.ReadFile(hostnameFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s for backup: %w", hostnameFilePath, err)
	}

	backupFileName := fmt.Sprintf("%s-%d", hostnameFilePath, time.Now().Unix())
	err = os.WriteFile(backupFileName, hostname, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	log.Info().Str("backupFileName", backupFileName).Msg("Hostname backup created successfully")
	return backupFileName, nil
}

func revertHostname(backupFileName string) error {
	backup, err := os.ReadFile(backupFileName)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	err = os.WriteFile(hostnameFilePath, backup, 0644)
	if err != nil {
		return fmt.Errorf("failed to restore backup to %s: %w", hostnameFilePath, err)
	}

	err = exec.Command("hostname", strings.TrimSpace(string(backup))).Run()
	if err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	log.Info().Str("backupFileName", backupFileName).Msg("Hostname reverted successfully")
	return nil
}

func (m *FileSystemDNSManager) SetHostname(ctx context.Context, hostname string) error {
	m.log.Info().Str("hostname", hostname).Msg("Setting hostname")

	backupFileName, err := backupHostname()
	if err != nil {
		return err
	}

	err = exec.CommandContext(ctx, "hostname", hostname).Run()
	if err != nil {
		if revertErr := revertHostname(backupFileName); revertErr != nil {
			m.log.Error().Err(revertErr).Msg("Failed to revert hostname")
		}
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	err = os.WriteFile(hostnameFilePath, []byte(hostname+"\n"), 0644)
	if err != nil {
		if revertErr := revertHostname(backupFileName); revertErr != nil {
			m.log.Error().Err(revertErr).Msg("Failed to revert hostname")
		}
		return fmt.Errorf("failed to write to %s: %w", hostnameFilePath, err)
	}

	m.log.Info().Str("hostname", hostname).Msg("Hostname set successfully")
	return nil
}

func backupResolvConf() (string, error) {
	input, err := os.Open(resolvConfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open %s for backup: %w", resolvConfPath, err)
	}
	defer func() {
		if err := input.Close(); err != nil {
			log.Warn().Str("failed to close input file", fmt.Sprint(err))
		}
	}()

	backupFileName := fmt.Sprintf("%s-%d", resolvConfPath, time.Now().Unix())
	output, err := os.Create(backupFileName)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer closeFile(output)

	_, err = io.Copy(output, input)
	if err != nil {
		return "", fmt.Errorf("failed to copy contents to backup file: %w", err)
	}

	log.Info().Str("backupFileName", backupFileName).Msg("Backup created successfully")
	return backupFileName, nil
}

func revertResolvConf(backupFileName string) error {
	input, err := os.Open(backupFileName)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer closeFile(input)

	output, err := os.Create(resolvConfPath)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", resolvConfPath, err)
	}
	defer closeFile(output)

	_, err = io.Copy(output, input)
	if err != nil {
		return fmt.Errorf("failed to copy contents from backup file: %w", err)
	}

	log.Info().Str("backupFileName", backupFileName).Msg("Reverted to backup successfully")
	return nil
}

func (m *FileSystemDNSManager) findDNSServer(file *os.File, server string) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == server {
				return fmt.Errorf("DNS server %s already exists", server)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

func (m *FileSystemDNSManager) AddDNSServer(ctx context.Context, server string) error {
	m.log.Info().Str("server", server).Msg("Adding DNS server")
	backupFileName, err := backupResolvConf()
	if err != nil {
		return err
	}

	file, err := os.Open(resolvConfPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", resolvConfPath, err)
	}
	defer closeFile(file)

	if err = m.findDNSServer(file, server); err != nil {
		return err
	}

	file, err = os.OpenFile(resolvConfPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", resolvConfPath, err)
	}
	defer closeFile(file)

	if _, err := file.WriteString(fmt.Sprintf("nameserver %s\n", server)); err != nil {
		if revertErr := revertResolvConf(backupFileName); revertErr != nil {
			m.log.Error().Err(revertErr).Msg("Failed to revert")
		}
		return fmt.Errorf("failed to write to %s: %w", resolvConfPath, err)
	}

	m.log.Info().Str("server", server).Msg("DNS server added successfully")
	return nil
}

func (m *FileSystemDNSManager) RemoveDNSServer(ctx context.Context, server string) error {
	m.log.Info().Str("server", server).Msg("Removing DNS server")
	backupFileName, err := backupResolvConf()
	if err != nil {
		return err
	}

	file, err := os.Open(resolvConfPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", resolvConfPath, err)
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
		return fmt.Errorf("error reading %s: %w", resolvConfPath, err)
	}

	file, err = os.OpenFile(resolvConfPath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", resolvConfPath, err)
	}
	defer closeFile(file)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			if revertErr := revertResolvConf(backupFileName); revertErr != nil {
				m.log.Error().Err(revertErr).Msg("Failed to revert")
			}
			return fmt.Errorf("failed to write to %s: %w", resolvConfPath, err)
		}
	}

	if err := writer.Flush(); err != nil {
		if revertErr := revertResolvConf(backupFileName); revertErr != nil {
			m.log.Error().Err(revertErr).Msg("Failed to revert")
		}
		return fmt.Errorf("failed to flush writes to %s: %w", resolvConfPath, err)
	}

	m.log.Info().Str("server", server).Msg("DNS server removed successfully")
	return nil
}

func (m *FileSystemDNSManager) ListDNSServers(ctx context.Context) ([]string, error) {
	m.log.Info().Msg("Listing DNS servers")
	file, err := os.Open(resolvConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", resolvConfPath, err)
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
		return nil, fmt.Errorf("error reading %s: %w", resolvConfPath, err)
	}

	m.log.Info().Int("count", len(dnsServers)).Msg("Listed DNS servers successfully")
	return dnsServers, nil
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close file")
	}
}
