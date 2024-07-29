package main

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

func LoadDotEnv[T interface{}](config *T) {
	godotenv.Load()
	if err := env.Parse(config); err != nil {
		log.Fatalln(err)
	}
}

type _Env struct {
	Port string `env:"PORT"`
	DRTS_API  string `env:"DRTS_API,required"`
}

var Env _Env

func init() {
	LoadDotEnv(&Env)

	if len(Env.Port) == 0 {
		Env.Port = "5432"
	}
}
