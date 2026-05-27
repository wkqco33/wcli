// config_demo는 wcli의 설정 기능을 보여주는 예제 앱입니다.
//
// 실행 가이드:
//
//	# 1. 기본 실행 (설정파일 자동 탐색)
//	go run ./examples/config_demo
//
//	# 2. 환경변수 덮어쓰기 테스트
//	DB_HOST="env-db-host" go run ./examples/config_demo
//
//	# 3. CLI 플래그 최우선 적용 테스트
//	DB_HOST="env-db-host" go run ./examples/config_demo --host cli-db-host
package main

import (
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

// AppConfig struct binding 방식의 설정 구조체
type AppConfig struct {
	AppName string `wcli:"NAME" default:"MyApp"`
	Server  struct {
		Host string `wcli:"HOST" default:"0.0.0.0"`
		Port int    `wcli:"PORT" default:"8080"`
	} `wcli:"SERVER"`
	Database struct {
		Host string `wcli:"HOST" default:"localhost"`
		Port int    `wcli:"PORT" default:"5432"`
		User string `wcli:"USER" default:"postgres"`
	} `wcli:"DATABASE"`
}

const tempJSONConfig = `{
  "database": {
    "host": "config-db-host",
    "port": 5432,
    "user": "postgres"
  }
}`

func main() {
	rich.Println("[bold][cyan]== wcli Config Demo ==[/cyan][/bold]")
	fmt.Println()

	// --- 방식 1: 글로벌 config store (SetConfigFile / Get) ---
	rich.Println("[bold]1. 글로벌 config store (JSON 파일)[/bold]")

	configFile := "demo_config.json"
	_ = os.WriteFile(configFile, []byte(tempJSONConfig), 0644)
	defer os.Remove(configFile)

	wcli.SetConfigFile(configFile)
	if err := wcli.ReadInConfig(); err != nil {
		fmt.Printf("설정 로드 실패: %v\n", err)
		os.Exit(1)
	}

	var (
		dbHost string
		dbPort int
		dbUser string
	)

	rootCmd := &wcli.Command{
		Use:   "config_demo",
		Short: "설정 파일 바인딩 데모",
		Run: func(ctx *wcli.Context) error {
			rich.Println("[green]  [글로벌 config store][/green]")
			fmt.Printf("  DB Host: %s\n", dbHost)
			fmt.Printf("  DB Port: %d\n", dbPort)
			fmt.Printf("  DB User: %s\n", dbUser)
			fmt.Println()

			// --- 방식 2: struct binding (Load) ---
			rich.Println("[green]  [struct binding (Load)][/green]")
			demoStructBinding()
			return nil
		},
	}

	rootCmd.Flags().StringVar(&dbHost, "host", "H", "localhost", "데이터베이스 호스트")
	rootCmd.Flags().IntVar(&dbPort, "port", "p", 3306, "데이터베이스 포트")
	rootCmd.Flags().StringVar(&dbUser, "user", "u", "root", "데이터베이스 사용자")

	_ = rootCmd.Flags().BindEnv("host", "DB_HOST")
	_ = rootCmd.Flags().BindConfig("host", "database.host")
	_ = rootCmd.Flags().BindConfig("port", "database.port")
	_ = rootCmd.Flags().BindConfig("user", "database.user")

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func demoStructBinding() {
	// 임시 YAML 파일 생성
	yamlContent := "SERVER:\n  HOST: yaml-server\n  PORT: 9090\n"
	_ = os.WriteFile("demo.yaml", []byte(yamlContent), 0644)
	defer os.Remove("demo.yaml")

	// 임시 .env 파일 생성
	envContent := "NAME=DemoApp\n"
	_ = os.WriteFile("demo.env", []byte(envContent), 0644)
	defer os.Remove("demo.env")

	var cfg AppConfig
	if err := wcli.Load(&cfg,
		wcli.WithDotEnv("demo.env"),
		wcli.WithFiles("demo.yaml"),
		wcli.WithEnv(),
	); err != nil {
		fmt.Printf("  Load 실패: %v\n", err)
		return
	}

	fmt.Printf("  AppName:     %s\n", cfg.AppName)
	fmt.Printf("  Server:      %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  DB Host:     %s (default)\n", cfg.Database.Host)
	fmt.Printf("  DB Port:     %d (default)\n", cfg.Database.Port)
}
