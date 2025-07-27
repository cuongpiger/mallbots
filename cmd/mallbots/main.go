package mallbots

import "github.com/cuongpiger/mallbots/internal/config"

func main() {

}

func run() (err error) {
	var (
		cfg config.AppConfig
	)

	// parse config/env/...
	cfg, err = config.InitConfig()
	if err != nil {
		return
	}

	return nil
}
