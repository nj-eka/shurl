package config

import "time"

type AppConfig struct {
	ShutdownTimeout time.Duration    `mapstructure:"shutdown-timeout"`
	Logging         *LoggingConfig   `mapstructure:"logging"`
	Server          *ServerConfig    `mapstructure:"server"`
	Router          *RouterConfig    `mapstructure:"router"`
	Store           *StoreConfig     `mapstructure:"store"`
	Tokenizer       *TokenizerConfig `mapstructure:"tokenizer"`
}

// logging:
//  path: "./log/shurl.log"
//  level: debug
//  format: text
type LoggingConfig struct {
	// Path to log output file; empty = os.Stdout
	FilePath string `mapstructure:"path"`
	// logging levels: panic, fatal, error, warn / warning, info, debug, trace
	Level string `mapstructure:"level"`
	// supported logging formats: text, json
	Format string `mapstructure:"format"`
}

// server:
//  addr: "0.0.0.0:8443"
//  timeout: 3s
type ServerConfig struct {
	// Address string        `mapstructure:"addr"`
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// router:
//  web-path: "web"
type RouterConfig struct {
	WebPath string `mapstructure:"web-path"`
}

// store:
//  bolt:
//    path: "links.db"
//    timeout: 1s
type StoreConfig struct {
	Bolt *BoltStoreConfig `mapstructure:"bolt"`
	Mem  *MemStoreConfig  `mapstructure:"mem"`
}

type BoltStoreConfig struct {
	FilePath string        `mapstructure:"path"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type MemStoreConfig struct {
	FilePath string `mapstructure:"path"`
}

// tokenizer:
//  hashid:
//    salt: "ecafbaf0-1bcc-11ec-9621-0242ac130002"
//    min-length: 5
//    alphabet: "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
type TokenizerConfig struct {
	Hashid *HashidTokenizerConfig `mapstructure:"hashid"`
}

type HashidTokenizerConfig struct {
	Salt      string `mapstructure:"salt"`
	MinLength int    `mapstructure:"min-length"`
	Alphabet  string `mapstructure:"alphabet"`
}
