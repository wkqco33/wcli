// config_demo는 wcli의 초경량 설정 바인딩 기능(BindConfig)을 보여주는 예제 앱입니다.
//
// 실행 가이드:
//
//	# 1. 설정파일 자동 탐색 테스트
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
	"io/ioutil"
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

const tempJSONConfig = `{
  "database": {
    "host": "config-db-host",
    "port": 5432,
    "user": "postgres"
  }
}`

func main() {
	// 임시 설정 파일 생성
	configFile := "demo_config.json"
	_ = ioutil.WriteFile(configFile, []byte(tempJSONConfig), 0644)
	defer os.Remove(configFile)

	// 1. wcli 설정 로드
	wcli.SetConfigFile(configFile)
	wcli.SetConfigType("json")
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
			rich.Println("[bold][green]⚡ 설정값 바인딩 결과:[/green][/bold]")
			fmt.Printf("  - 데이터베이스 호스트  : %s\n", dbHost)
			fmt.Printf("  - 데이터베이스 포트    : %d\n", dbPort)
			fmt.Printf("  - 데이터베이스 사용자  : %s\n", dbUser)
			return nil
		},
	}

	// 플래그 정의
	rootCmd.Flags().StringVar(&dbHost, "host", "H", "localhost", "데이터베이스 호스트")
	rootCmd.Flags().IntVar(&dbPort, "port", "p", 3306, "데이터베이스 포트")
	rootCmd.Flags().StringVar(&dbUser, "user", "u", "root", "데이터베이스 사용자")

	// 2. 환경변수 및 설정파일 키 연동 바인딩
	_ = rootCmd.Flags().BindEnv("host", "DB_HOST")
	_ = rootCmd.Flags().BindConfig("host", "database.host")

	_ = rootCmd.Flags().BindConfig("port", "database.port")
	_ = rootCmd.Flags().BindConfig("user", "database.user")

	if err := rootCmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
