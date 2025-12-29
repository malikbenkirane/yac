package snake

import "strings"

type Option func(Config) Config

func Case(s string, opts ...Option) string {
	if len(s) == 0 {
		return ""
	}
	cfg := Config{
		placeHolder: '-',
		toLower:     true,
		exception:   make(map[string]struct{}),
	}
	for _, opt := range opts {
		cfg = opt(cfg)
	}
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteRune(cfg.placeHolder)
			}
			if cfg.toLower {
				b.WriteRune(r - 'A' + 'a')
			} else {
				b.WriteRune(r)
			}
		} else {
			b.WriteRune(r)
		}
	}
	if cfg.plural {
		{
			parts := strings.Split(b.String(), string(cfg.placeHolder))
			if _, ok := cfg.exception[strings.ToLower(parts[len(parts)-1])]; ok {
				return b.String()
			}
		}
		if strings.HasSuffix(b.String(), "ed") ||
			strings.HasSuffix(b.String(), "ing") {
			return b.String()
		}
		for _, suffix := range []string{"s", "ss", "sh", "ch", "x", "z"} {
			found := strings.HasSuffix(b.String(), suffix)
			if !found {
				continue
			}
			if suffix == "z" {
				b.WriteRune('z')
			}
			return b.String() + "es"
		}
		for i := 'b'; i <= 'z'; i++ {
			if i == 'e' || i == 'i' || i == 'o' || i == 'u' {
				continue
			} else {
				root, found := strings.
					CutSuffix(b.String(), string([]byte{byte(i), 'y'}))
				if !found {
					continue
				}
				return root + string(i) + "ies"
			}
		}
		return b.String() + "s"
	}
	return b.String()
}

func WithSpace() Option {
	return func(c Config) Config {
		c.placeHolder = ' '
		return c
	}
}

func WithKeepCase() Option {
	return func(c Config) Config {
		c.toLower = false
		return c
	}
}

// WithPlural option configure snake.Case to make plural for last words in the
// provide camel case.  The words listed in exceptions will be skipped.
//
// It is safe to set WithPlural option multiple times.
func WithPlural(exceptions ...string) Option {
	return func(c Config) Config {
		c.plural = true
		for _, e := range exceptions {
			c.exception[strings.ToLower(e)] = struct{}{}
		}
		return c
	}
}

type Config struct {
	placeHolder rune
	toLower     bool
	plural      bool
	exception   map[string]struct{}
}
