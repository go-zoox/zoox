package logger

import (
	"fmt"
	"strings"

	"github.com/go-zoox/logger"
	cst "github.com/go-zoox/logger/components/transport"
	"github.com/go-zoox/logger/transport/console"
	"github.com/go-zoox/logger/transport/file"
	"github.com/go-zoox/zoox/config"
)

// Build builds the application logger from config.Logger.
// No transports configured → console only; otherwise console + each configured transport.
func Build(cfg *config.Logger) (*logger.Logger, error) {
	if cfg == nil {
		cfg = &config.Logger{}
	}

	transports := map[string]cst.Transport{
		"console": console.New(),
	}

	for i, spec := range cfg.Transports {
		key, t, err := transportFromSpec(i, &spec)
		if err != nil {
			return nil, err
		}
		if t == nil {
			continue
		}
		transports[key] = t
	}

	level := strings.TrimSpace(cfg.Level)
	if level == "" {
		level = "INFO"
	}

	return logger.New(func(opt *logger.Option) {
		opt.Transports = transports
		opt.Level = level
	}), nil
}

func transportFromSpec(i int, spec *config.Transport) (string, cst.Transport, error) {
	typ := strings.ToLower(strings.TrimSpace(spec.Type))
	switch typ {
	case "", "console":
		return "", nil, nil
	case "file":
		fc := fileConfigFromSpec(spec)
		if fc == nil {
			return "", nil, nil
		}
		key := "file"
		if i > 0 {
			key = fmt.Sprintf("file-%d", i)
		}
		return key, file.New(fc), nil
	default:
		return "", nil, fmt.Errorf("unknown logger transport type: %q", spec.Type)
	}
}

func fileConfigFromSpec(spec *config.Transport) *file.Config {
	path := strings.TrimSpace(spec.Path)
	levels := make(map[string]string, len(spec.Levels))
	for k, v := range spec.Levels {
		levels[normalizeLogLevelKey(k)] = strings.TrimSpace(v)
	}
	if path == "" && len(levels) == 0 {
		return nil
	}
	return &file.Config{
		FilePath: path,
		Levels:   levels,
	}
}

func normalizeLogLevelKey(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	switch k {
	case "warning":
		return "warn"
	case "err":
		return "error"
	default:
		return k
	}
}
