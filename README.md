# Config
Config is a simple package to load config from a JSON formatted file and to watch for changes in the file. After the initial configurate load, Config watches the file and broadcasts a message whenever there are changes.

## Example
The following example reads a config file located at `./config.json` into the struct `Config`.

```go
type Config struct {
	DBConnection   string `json:"db_connection"`
	BindAddress    string `json:"bind_address"`
	MetricsAddress string `json:"metrics_address"`
}

func main() {
	conf = &Config{}

	// Create a new config watcher
	c, err := config.New(
		"./config.json",
		logger.StandardLogger(&hclog.StandardLoggerOptions{}),
		func(c *Config) {
      // callback when config file is updated
      // returns a copy of the read config

			logger.Info("Config file updated", "config", c)
		},
	)

  // read the config
  conf = c.Get()

  for{}
}
```

A full example can be found in the `example` folder.