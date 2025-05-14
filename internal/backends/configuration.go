package backends

import "time"

type Configuration struct {
	Backends []struct {
		Name string `yaml:"name"`
		Url  string `yaml:"url"`
	} `yaml:"backends"`

	HealthCheck struct {
		Timeout  time.Duration `yaml:"timeout"`
		Interval time.Duration `yaml:"interval"`
		Url      string        `yaml:"url"`
		Method   string        `yaml:"method"`
	} `yaml:"health_check"`
}
