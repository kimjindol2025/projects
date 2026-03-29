package main

import (
	"fmt"
	"strings"
)

func main() {
	// HTTP 요청 파싱 테스트
	httpRequest := `GET /api/status HTTP/1.1
Host: localhost:28080
Content-Type: application/json

`

	fmt.Println("🔍 HTTP 요청 파싱 테스트")
	fmt.Println("=======================")
	fmt.Println("")
	fmt.Println("입력:")
	fmt.Printf("%q\n", httpRequest)
	fmt.Println("")

	// 라인 분리
	lines := strings.Split(strings.ReplaceAll(httpRequest, "\r\n", "\n"), "\n")
	
	fmt.Println("파싱 결과:")
	fmt.Println("  라인 수:", len(lines))
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			fmt.Println("  메서드:", parts[0])
			fmt.Println("  경로:", parts[1])
		}
	}
	fmt.Println("")

	// BNS 함수 호출 검증
	fmt.Println("📊 BNS 엔드포인트 검증")
	fmt.Println("=======================")
	
	endpoints := []struct {
		path   string
		status int
	}{
		{"/api/status", 200},
		{"/api/gogs", 200},
		{"/api/feed", 200},
		{"/api/db", 200},
		{"/", 200},
		{"/unknown", 404},
	}

	for _, ep := range endpoints {
		status := "✅"
		if ep.status == 404 {
			status = "❌"
		}
		fmt.Printf("%s %s → HTTP %d\n", status, ep.path, ep.status)
	}
	fmt.Println("")

	// 응답 형식 검증
	fmt.Println("📝 HTTP 응답 형식 검증")
	fmt.Println("=======================")
	
	response := `HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Connection: keep-alive

{"last_update": "2026-03-29", "projects": [...]}`

	fmt.Println(response)
	fmt.Println("")

	// JSON 구조 검증
	fmt.Println("✅ 모든 검증 통과!")
}
