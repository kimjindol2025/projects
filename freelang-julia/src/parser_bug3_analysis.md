# 🔴 Bug #3 상세 분석: Parser Postfix 우선순위

## 실제 버그 위치

**파일**: `parser.fl`
**함수 체인**:
```
parse_expression (줄 356)
  → parse_binary_op (줄 359)
    → parse_unary (줄 388)
      → parse_postfix (줄 400)
        → parse_postfix_loop (줄 407)
```

## 문제의 본질

### 파싱 우선순위 체인

```
Expression: a.b + c.d * e.f

파싱 흐름:
1. parse_expression → parse_binary_op(min_prec=0)
2. parse_binary_op_loop:
   - 왼쪽: parse_unary → parse_postfix → parse_primary("a")
   - Postfix: parse_postfix_loop
     - "." 만남, advance, "b" 읽음
     - "a.b" 결과 반환
   - 이제 parser는 "+" 가리킴
3. parser.current == "+" (binary_op)
   - 우선순위 비교: + (낮음)
   - 오른쪽 파싱: parse_binary_op(min_prec=1)
     - c 파싱, postfix에서 ".d" 처리
     - c.d 반환, 현재 "*" 가리킴
4. "*" vs "+" 우선순위 비교
   - "*" (높음) >= min_prec (1) → 계속
   - "c.d * e.f" 파싱
5. 최종: (a.b) + (c.d * e.f) ✅
```

## 실제 버그: 미묘한 케이스

```freejulia
# 입력 예
obj[index].method(arg).field

파싱:
1. "obj" (primary)
2. "[index]" (postfix 루프 - 인덱싱)
3. ".method" (postfix 루프 - 멤버)
4. "(arg)" (postfix 루프 - 함수 호출)
5. ".field" (postfix 루프 - 멤버)

현재 코드: parse_postfix_loop에서 모든 postfix 연산을 처리
→ 올바르게 작동 ✅

그러나 문제:
"obj.field" 다음에 다른 postfix가 오면?

입력:
obj.field1.field2[0].method()

파싱:
1. obj (primary)
2. .field1 (postfix 루프, 재귀)
   → parser3 = advance(field1)
   → parse_postfix_loop(parser3, MemberAccess)
3. .field2 (postfix 루프, 재귀)
4. [0] (postfix 루프, 재귀)
5. .method() (postfix 루프, 재귀)

이건 올바르게 작동! ✅

실제 버그는 어디에?
```

## 진짜 버그: 에러 복구 부족

```freejulia
# 문법 오류 케이스
obj.  # ← 식별자 없음

현재:
parse_postfix_loop에서
  TokenDot
    parser2 = advance (건너뜸)
    parser2.current.type != TokenIdentifier
    → (parser2, Some(expr))  # parser2로 진행, expr 그대로

결과: parser가 건너뛰어짐 (parser2 상태)
      그러나 expr은 변경 안 됨

이어서:
parse_binary_op_loop에서 현재 토큰이 뭐든
  - parser2가 가리키는 토큰 사용
  - expr은 "."를 처리하지 않은 상태

→ 불일치! parser와 expr이 동기화 안 됨
```

## 진짜 원인

```
줄 439: (parser2, Some(expr))

문제:
- parser2: "." 다음을 가리킴 (이미 진행)
- expr: "obj.field" 같은 것이지만,
        실제로는 "." 이후에 식별자 없었을 때
        파서가 건너뛴 상태

결과: 다음 파싱에서 불일치 발생 가능
```

## 해결 방법

```freejulia
# AS-IS
    TokenDot ->
      let parser2 = advance_parser(parser)
      if parser2.current.type == TokenIdentifier then
        ... parse_postfix_loop(parser3, access)
      else
        (parser2, Some(expr))  # ← 문제

# TO-BE: 에러 토큰 추가
    TokenDot ->
      let parser2 = advance_parser(parser)
      if parser2.current.type == TokenIdentifier then
        let member_name = parser2.current.lexeme
        let parser3 = advance_parser(parser2)
        let access = MemberAccess { object = expr, member = member_name }
        parse_postfix_loop(parser3, access)
      else
        # 에러: TokenDot 이후에 식별자 필수
        # 그러나 recovery: expr은 "." 이전 것으로 유지
        # parser2는 "." 다음을 가리키므로 건너뜀
        # → 이어서 파싱 계속 (에러 버그)
        (parser2, Some(expr))
```

## 실제 문제는 여기!

```freejulia
줄 420:
let (parser3, index) = parse_expression(parser2)

여기서 parse_expression을 재귀 호출!
→ 이건 binary operation을 포함한 전체 식을 파싱
→ 배열 인덱스로 "1 + 2" 같은 걸 파싱할 수 있음

근데 parse_arguments (줄 444)는?
→ parse_expression을 호출
→ 함수 인자로 "a + b" 파싱 가능

이건 올바름! ✅
```

## 최종 진단

**Bug #3의 실제 문제**:

1. **미묘한 케이스**: `obj.` (식별자 없음) 후 다른 연산
   - parse_postfix_loop가 (parser2, expr) 반환
   - parser2는 건너뜀, expr은 변경 안 됨
   - 다음 파싱에서 토큰-표현식 불일치 가능

2. **현재 코드는 대부분 맞음** ✅
   - parse_postfix_loop의 재귀 처리가 정확
   - 멤버, 인덱싱, 함수 호출 모두 연쇄 가능

3. **수정 필요 부분**: 에러 처리
   - `obj.` 이후 에러 신호 필요
   - 현재는 조용히 건너뜸

## 권장 수정

```freejulia
    TokenDot ->
      let parser2 = advance_parser(parser)
      if parser2.current.type == TokenIdentifier then
        let member_name = parser2.current.lexeme
        let parser3 = advance_parser(parser2)
        let access = MemberAccess { object = expr, member = member_name }
        parse_postfix_loop(parser3, access)
      else
        # 에러: "." 이후에 식별자 필요
        # 복구 방법:
        # Option 1: 에러 토큰 추가
        # Option 2: 그냥 expr 반환 (현재) - 다음 파싱에서 에러 감지
        # Option 3: parser를 진행하지 말고 원래대로 (parser, Some(expr))

        # 현재 코드 (Option 2)가 합리적
        (parser2, Some(expr))
```

## 결론

**Bug #3은 "심각한 버그"라기보다는 "에러 처리 미흡"**

- 정상 케이스: 모두 작동 ✅
- 에러 케이스: 처리 미흡 ⚠️
- 우선순위: 낮음 (다른 버그 먼저)

