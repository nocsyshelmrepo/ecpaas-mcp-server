package constants

import "os"

type Environment struct {
	env string
	def string
}

var (
	Kubesphere Environment = Environment{
		env: "KUBESPHERE_CONTEXT",
		def: "kubesphere",
	}
)

func Getenv(env Environment) string {
	val, ok := os.LookupEnv(env.env)
	if !ok {
		return env.def
	}
	return val
}
