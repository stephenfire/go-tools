package tools

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SplitNetAddr like host:port to host and port
// if there's no valid host exists, return ""
// if there's no valid port exists, return -1
func SplitNetAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	switch len(parts) {
	case 0:
		return "", -1
	case 1:
		host = strings.TrimSpace(parts[0])
	default:
		host = strings.TrimSpace(parts[0])
		if p, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			port = p
		}
	}
	if port <= 0 || port > 65535 {
		port = -1
	}
	return host, port
}

// ValidAndCreateDir 将传入的绝对路径pstr当作文件目录，确认其存在，如果不存在则创建。并返回包含最后的路径分隔符的绝对路径字符串
func ValidAndCreateDir(pstr string, checkAndCreateIfAbsent bool) (string, error) {
	formatted := filepath.Clean(pstr)
	if formatted == "" || !filepath.IsAbs(pstr) {
		return "", errors.New("tools: absolute path required")
	}
	if checkAndCreateIfAbsent {
		fi, err := os.Stat(formatted)
		if err != nil {
			if os.IsNotExist(err) {
				if err = os.MkdirAll(formatted, 0755); err != nil {
					return "", fmt.Errorf("tools: failed to create dir %s: %w", formatted, err)
				} else {
					slog.Debug("dir created", "dir", formatted)
					fi, err = os.Stat(formatted)
					if err != nil {
						return "", fmt.Errorf("tools: failed to get stat after creating dir %s: %w", formatted, err)
					}
				}
			} else {
				return "", fmt.Errorf("tools: failed to stat dir %s: %w", formatted, err)
			}
		}
		if !fi.IsDir() {
			return "", fmt.Errorf("tools: %s is not a dir", formatted)
		}
	}
	if formatted[len(formatted)-1] != os.PathSeparator {
		formatted = formatted + string([]byte{os.PathSeparator})
	}
	return formatted, nil
}

func OpenWriteFile(fullFilePathName string) (*os.File, error) {
	dir := filepath.Dir(fullFilePathName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	out, err := os.Create(fullFilePathName)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func RemoveFile(fullFilePathName string) error {
	if err := os.Remove(fullFilePathName); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return errors.New("tools: file still exists")
	}
	return nil
}
