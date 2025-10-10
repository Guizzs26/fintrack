package config

type Config struct {
	PasswordPepper string `env:"PASSWORD_PEPPER,required"`
}
