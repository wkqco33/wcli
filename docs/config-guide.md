# 설정 및 고급 플래그 가이드

[← README로 돌아가기](../README.md)

환경 변수, 설정 파일, 플래그 제약 조건과 같은 고급 설정 기능을 정리한 문서입니다.

## 플래그 확장 조건 및 검증

플래그 관계 제약 조건 및 환경 변수 자동 주입 기능을 내장하고 있습니다.

### 환경 변수 자동 매핑 (BindEnv)

특정 플래그가 지정되지 않은 경우, 대체해서 값을 읽어올 환경 변수를 지정합니다.

```go
cmd.Flags().StringVar(&token, "token", "t", "", "인증 토큰")
_ = cmd.Flags().BindEnv("token", "API_TOKEN")
```

### 상호 배제 조건 지정 (Mutually Exclusive)

지정된 옵션 그룹 중 하나만 전달되어야 하는 조건을 검증합니다.

```go
cmd.Flags().BoolVar(&jsonOut, "json", "j", false, "JSON 출력")
cmd.Flags().BoolVar(&yamlOut, "yaml", "y", false, "YAML 출력")
cmd.Flags().MarkFlagsMutuallyExclusive("json", "yaml")
```

도움말에 제약 조건이 자동으로 표시됩니다:
```
Flag Constraints:
  mutually exclusive: --json, --yaml
```

### 필수 동반 조건 지정 (Required Together)

하나라도 설정되면 그룹 내 모든 플래그가 함께 설정되어야 하는 조건입니다.

```go
cmd.Flags().StringVar(&user, "user", "u", "", "계정")
cmd.Flags().StringVar(&pass, "password", "p", "", "비밀번호")
cmd.Flags().MarkFlagsRequiredTogether("user", "password")
```

### 설정 파일 매핑 (BindConfig)

외부 구성 파일의 값을 플래그에 매핑합니다. 우선순위:
`입력된 플래그 > 바인딩된 환경변수 > 설정 파일 매핑값 > 기본값`

외부 의존성 없이 표준 라이브러리만으로 JSON, INI, YAML, TOML, .env를 파싱합니다.

**지원 포맷:** `json`, `ini`, `yaml`/`yml`, `toml`, `env`

> **파서 한계 (의도된 단순화):** YAML/TOML/INI 파서는 표준 라이브러리만으로 구현된 경량 파서입니다.
> 복잡한 문법은 지원하지 않으니 단순한 키-값 + 중첩 구조 위주로 사용하세요. JSON은 표준 `encoding/json`을 사용하므로 제약이 없습니다.
> - **미지원:** 배열/리스트(`- item`, `[a, b]`), 멀티라인 값, 앵커/별칭, 인라인 테이블, 값 안의 구분자(예: 따옴표로 감싼 `:`나 `=`)
> - YAML 들여쓰기는 **공백만** 지원하며 탭은 인식하지 않습니다.
> - 모든 스칼라 값은 **문자열**로 로드됩니다(타입 변환은 플래그 바인딩 시점에 수행).

```go
import "github.com/seoyc/wcli/config"

config.SetConfigFile("config.yaml")
if err := config.ReadInConfig(); err != nil {
    os.Exit(1)
}

var dbHost string
cmd.Flags().StringVar(&dbHost, "host", "H", "localhost", "데이터베이스 호스트")
_ = cmd.Flags().BindEnv("host", "DB_HOST")
_ = cmd.Flags().BindConfig("host", "database.host")  // 점 표기법으로 중첩 키 접근
```

또는 직접 값 조회 (타입 안전한 Get 헬퍼 포함):

```go
config.SetConfigFile("config.json")
_ = config.ReadInConfig()

// 1. 일반 점 표기법 조회
host := config.Get("database.host")

// 2. 타입 안전한 헬퍼 (자동 타입 캐스팅 지원)
port := config.GetInt("database.port")
debug := config.GetBool("app.debug")
ips := config.GetStringSlice("database.allowed_ips")
timeout := config.GetDuration("app.timeout") // time.Duration 반환

config.Set("app.debug", true)        // 런타임에 값 설정
config.SetDefault("app.port", 8080)  // 키가 없을 때만 설정
```

### 설정 파일 자동 탐색 (AutoDiscoverConfig)

앱 이름 기반으로 표준 경로를 순서대로 탐색해 첫 번째로 발견된 설정 파일을 로드합니다.

탐색 순서:
1. `extraPaths` (직접 지정 경로)
2. `./config.{yaml,yml,toml,ini,json,env}`
3. `~/.{appName}.{yaml,...}`
4. `/etc/{appName}/config.{yaml,...}`

```go
// 설정 파일을 자동 탐색하여 로드
if err := config.AutoDiscoverConfig("myapp"); err != nil {
    // 파일이 없으면 에러
}

// 추가 탐색 경로 지정
if err := config.AutoDiscoverConfig("myapp", "/opt/myapp/config.yaml"); err != nil {
    os.Exit(1)
}
```

### 환경변수 글로벌 연동 및 리로드 (Env & Reload)

* **환경변수 글로벌 연동 (`AutomaticEnv`)**: 활성화 시 `config.Get`을 통한 설정 조회 시 환경변수(예: `DATABASE_PORT`)가 존재하면 파일 내 설정 값보다 환경변수 값을 최우선으로 연동하여 반환합니다. 대소문자는 구분하지 않습니다.
* **설정 리로드 (`ReloadConfig`)**: 서버나 장기 실행 CLI 프로세스 등에서 설정 파일이 동적으로 변경되었을 때 메모리에 이미 로드된 설정을 디스크에서 다시 로드합니다.

```go
// 글로벌 환경변수 최우선 연동 활성화
config.AutomaticEnv()
config.SetEnvPrefix("APP") // 환경변수 접두사 설정 (APP_DATABASE_PORT 형태 연동)

// 디스크에서 설정 실시간 리로드
if err := config.ReloadConfig(); err != nil {
    log.Printf("설정 리로드 실패: %v", err)
}
```

### 구조체 바인딩 (Load)

설정 파일, `.env`, 환경변수 값을 구조체에 직접 바인딩합니다.

```go
import "github.com/seoyc/wcli/config"

type AppConfig struct {
    Host    string   `wcli:"HOST" default:"localhost"`
    Port    int      `wcli:"PORT" default:"8080"`
    Debug   bool     `wcli:"DEBUG"`
    IPs     []string `wcli:"IPS"` // 슬라이스 바인딩 지원 ([]string, []int 등)
    DB struct {
        User string `wcli:"USER"`
        Pass string `wcli:"PASS"`
    } `wcli:"DATABASE"`
}

var cfg AppConfig
err := config.Load(&cfg,
    config.WithDotEnv(".env"),          // .env 파일
    config.WithFiles("config.yaml"),    // YAML 또는 TOML 파일 (대소문자 구분 없이 자동 매칭 지원)
    config.WithEnv(),                   // 시스템 환경변수 (최우선)
    config.WithPrefix("APP"),           // 환경변수 접두사
)
```

소스 순서대로 병합되며, 뒤에 오는 소스가 앞의 값을 덮어씁니다.

기본값을 파일로 내보낼 수 있습니다:

```go
config.WriteDefault(&cfg, "config.yaml")   // YAML로 저장
config.WriteDefault(&cfg, "config.toml")   // TOML로 저장
config.WriteDefault(&cfg, ".env")          // .env로 저장
```
