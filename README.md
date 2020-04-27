# Config
Config is a simple package to load config from a JSON formatted file and to watch for changes in the file. After the initial configurate load, Config watches the file and broadcasts a message whenever there are changes.

## Example
The following example reads a config file located at `./config.json` into the struct `Config`.

```
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
		conf,
		logger.StandardLogger(&hclog.StandardLoggerOptions{}),
		func() {
			logger.Info("Config file updated", "config", conf)
		},
	)

  for{}
}
```

A full example can be found in the `example` folder.