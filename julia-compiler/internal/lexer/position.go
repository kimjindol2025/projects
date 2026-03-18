package lexer

import "fmt"

// Position - 소스 코드의 위치 정보
type Position struct {
	File   string // 파일명
	Line   int    // 줄 번호 (1부터 시작)
	Column int    // 열 번호 (1부터 시작)
	Offset int    // 파일 시작부터의 바이트 오프셋
}

// Range - 시작부터 종료까지의 범위
type Range struct {
	Start Position
	End   Position
}

func (p Position) String() string {
	return p.File + ":" + fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token - 토큰 구조체 (위치 정보 포함)
type Token struct {
	Type   TokenType
	Lexeme string
	Value  interface{} // 실제 값 (정수, 부동소수 등)
	Pos    Position
}

func (t Token) String() string {
	return t.Type.String() + "(" + t.Lexeme + ")"
}
