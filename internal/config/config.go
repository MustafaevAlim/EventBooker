package config

import "github.com/wb-go/wbf/config"

type Config struct {
	Postgre PostgreConfig
	Server  ServerConfig
	TgBot   TgBotConfig
}

type ServerConfig struct {
	Port   string
	JwtKey string
}

type PostgreConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
}

type TgBotConfig struct {
	Token string
}

func NewConfig() (*Config, error) {
	c := config.New()
	err := c.Load(".env", "", "")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Postgre: PostgreConfig{
			User:     c.GetString("POSTGRES_USER"),
			Password: c.GetString("POSTGRES_PASSWORD"),
			DBName:   c.GetString("POSTGRES_DB"),
			Host:     c.GetString("POSTGRES_HOST"),
		},
		Server: ServerConfig{
			Port:   c.GetString("PORT"),
			JwtKey: c.GetString("JWT_SECREY_KEY"),
		},
		TgBot: TgBotConfig{
			Token: c.GetString("TG_TOKEN"),
		},
	}
	return cfg, nil
}
