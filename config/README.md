# Config

Standardizing environment configuration for GO applications

Uses [godotenv](https://github.com/joho/godotenv) to load .env files from Vault into environment variables

Uses [envconfig](https://github.com/kelseyhightower/envconfig) to process environment variables into config structs

Contains helpful structs for common configurations to force consistency

## Usage

``` go
import (
    "github.com/jesse0michael/pkg/config"
)

type Config struct {
    Hostname string `envconfig:"HOSTNAME"`
    MySQL    config.MysqlConfig
}

func main() {
    var cfg Config
    if err := config.Process(os.Getenv("ENV_FILES_DIR"), &cfg); err != nil {
        panic(err)
    }
}
```
