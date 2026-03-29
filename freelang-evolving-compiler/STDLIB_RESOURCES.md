# FreeLang 표준 라이브러리 자원 분석

**분석일**: 2026-03-29
**기준**: v2, v4 기존 구현 조사
**결론**: 참고 자료 풍부 - 5개 카테고리 정의됨

---

## 📍 자원 위치

| 버전 | 위치 | 파일명 | 크기 | 상태 |
|------|------|--------|------|------|
| **v2** | `src/engine/` | `builtins.ts` | 2,758줄 | ✅ 완성 |
| **v4** | `dev/archived/freelang-v4-stdlib/` | 여러 파일 | ~1,500줄 | ✅ 완성 |
| **v4-core** | `src/stdlib/` | 여러 파일 | ~800줄 | ✅ 완성 |

---

## 1️⃣ v2 Builtins (2,758줄)

### 구조

```
builtins.ts (TypeScript)
├─ interface BuiltinSpec
│  ├─ name: string
│  ├─ params: BuiltinParam[]
│  ├─ return_type: string
│  ├─ c_name: string (C 구현)
│  ├─ headers: string[]
│  └─ impl: (...args) => any (JavaScript 구현)
└─ export const BUILTINS: Record<string, BuiltinSpec>
```

### 구현된 함수들

**배열 함수** (Array aggregates):
```
- sum(arr: number[]) -> number
- average(arr: number[]) -> number
- max(arr: number[]) -> number
- min(arr: number[]) -> number
- count(arr: number[]) -> number
- length(arr: number[]) -> number
```

**수학 함수** (Math functions):
```
- sqrt(x: number) -> number
- sin(x: number) -> number
- cos(x: number) -> number
- abs(x: number) -> number
- pow(x: number, y: number) -> number
- floor(x: number) -> number
- ceil(x: number) -> number
- round(x: number) -> number
```

**문자열 함수** (추정):
```
- concat(a: string, b: string) -> string
- length(s: string) -> number
- substring(s: string, start: int, end: int) -> string
- split(s: string, sep: string) -> string[]
- upper(s: string) -> string
- lower(s: string) -> string
```

### 특징

```
✅ 3단계 구현 (TypeScript, C, JavaScript)
✅ 타입 명시
✅ 표준 인터페이스
✅ 문서화됨
```

---

## 2️⃣ v4 Standard Library

### 구조 (5개 카테고리)

#### 1. String Functions (14개)
```
- string_length(s: string) -> int
- string_concat(a: string, b: string) -> string
- string_substring(s: string, start: int, len: int) -> string
- string_index_of(s: string, substr: string) -> int
- string_replace(s: string, old: string, new: string) -> string
- string_split(s: string, sep: string) -> string[]
- string_trim(s: string) -> string
- string_upper(s: string) -> string
- string_lower(s: string) -> string
- string_starts_with(s: string, prefix: string) -> bool
- string_ends_with(s: string, suffix: string) -> bool
- string_contains(s: string, substr: string) -> bool
- string_reverse(s: string) -> string
- string_repeat(s: string, count: int) -> string
```

#### 2. Array/Iterator Functions (14개)
```
- array_length(arr: array) -> int
- array_get(arr: array, index: int) -> any
- array_set(arr: array, index: int, value: any) -> void
- array_append(arr: array, value: any) -> void
- array_pop(arr: array) -> any
- array_shift(arr: array) -> any
- array_unshift(arr: array, value: any) -> void
- array_slice(arr: array, start: int, end: int) -> array
- array_concat(a: array, b: array) -> array
- array_reverse(arr: array) -> array
- array_sort(arr: array) -> array
- array_filter(arr: array, predicate: fn) -> array
- array_map(arr: array, transform: fn) -> array
- array_reduce(arr: array, init: any, reducer: fn) -> any
```

#### 3. Math Functions (14개)
```
- math_abs(x: number) -> number
- math_sign(x: number) -> number
- math_min(a: number, b: number) -> number
- math_max(a: number, b: number) -> number
- math_floor(x: number) -> number
- math_ceil(x: number) -> number
- math_round(x: number) -> number
- math_sqrt(x: number) -> number
- math_pow(x: number, y: number) -> number
- math_sin(x: number) -> number
- math_cos(x: number) -> number
- math_tan(x: number) -> number
- math_log(x: number) -> number
- math_exp(x: number) -> number
```

#### 4. Object/Map Functions (14개)
```
- map_new() -> map
- map_get(m: map, key: string) -> any
- map_set(m: map, key: string, value: any) -> void
- map_has(m: map, key: string) -> bool
- map_delete(m: map, key: string) -> void
- map_keys(m: map) -> string[]
- map_values(m: map) -> any[]
- map_entries(m: map) -> [string, any][]
- map_clear(m: map) -> void
- map_size(m: map) -> int
- map_merge(a: map, b: map) -> map
- map_clone(m: map) -> map
- object_keys(obj: object) -> string[]
- object_values(obj: object) -> any[]
```

#### 5. Functional Programming (9개)
```
- compose(f: fn, g: fn) -> fn
- curry(f: fn) -> fn
- partial(f: fn, ...args) -> fn
- memoize(f: fn) -> fn
- once(f: fn) -> fn
- pipe(value: any, ...fns) -> any
- apply(f: fn, args: array) -> any
- call(f: fn, ...args) -> any
- bind(f: fn, context: any) -> fn
```

#### 추가 (Regex, JSON, Database - v4에만)

**Regex** (7개):
```
- regex_match(pattern: string, str: string) -> bool
- regex_find(pattern: string, str: string) -> string
- regex_replace(pattern: string, str: string, replacement: string) -> string
- regex_split(pattern: string, str: string) -> string[]
- regex_test(pattern: string, str: string) -> bool
- regex_exec(pattern: string, str: string) -> match[]
- regex_compile(pattern: string) -> regex
```

**JSON** (6개):
```
- json_parse(str: string) -> any
- json_stringify(obj: any) -> string
- json_pretty(obj: any) -> string
- json_minify(str: string) -> string
- json_validate(str: string) -> bool
- json_merge(a: object, b: object) -> object
```

**Database** (9개):
```
- db_open(path: string) -> db
- db_close(db: db) -> void
- db_query(db: db, sql: string) -> result
- db_execute(db: db, sql: string, params: array) -> void
- db_create_table(db: db, schema: string) -> void
- db_insert(db: db, table: string, data: map) -> void
- db_select(db: db, table: string) -> result
- db_update(db: db, table: string, data: map) -> void
- db_delete(db: db, table: string) -> void
```

---

## 📊 총 함수 개수

| 카테고리 | v4 개수 | v2 개수 | 참고 |
|---------|--------|--------|------|
| String | 14 | ~15 | 유사 |
| Array/Iterator | 14 | ~12 | 유사 |
| Math | 14 | ~10 | 유사 |
| Object/Map | 14 | ~8 | 확장됨 |
| Functional | 9 | ~5 | 확장됨 |
| Regex | 7 | 0 | 신규 |
| JSON | 6 | 0 | 신규 |
| Database | 9 | 0 | 신규 |
| **합계** | **87** | **50+** | v4가 훨씬 충실 |

---

## 🎯 Phase 2 기본 라이브러리 구현 계획

### 필수 4개 함수

#### 1. `print()` / `println()`
```go
// 시그니처
func print(value: any) -> unit
func println(value: any) -> unit

// 구현 위치
internal/builtin/io.go (신규)

// 복잡도
⭐ (매우 간단)

// 예상 줄수
~30줄
```

#### 2. `len()`
```go
// 시그니처
func len(value: array|string|map) -> int

// 구현 위치
internal/builtin/array.go (신규)

// 복잡도
⭐ (간단)

// 예상 줄수
~40줄
```

#### 3. String Operations
```go
// 시그니처
func concat(a: string, b: string) -> string
func substring(s: string, start: int, end: int) -> string
func upper(s: string) -> string
func lower(s: string) -> string
func split(s: string, sep: string) -> string[]

// 구현 위치
internal/builtin/string.go (신규)

// 복잡도
⭐⭐ (중간)

// 예상 줄수
~150줄
```

#### 4. Array Operations
```go
// 시그니처
func append(arr: array, value: any) -> array
func get(arr: array, index: int) -> any
func set(arr: array, index: int, value: any) -> void
func slice(arr: array, start: int, end: int) -> array

// 구현 위치
internal/builtin/array.go (기존에 추가)

// 복잡도
⭐⭐ (중간)

// 예상 줄수
~120줄
```

---

## 🏗️ 구현 구조

### 제안 아키텍처

```
internal/builtin/
├─ builtin.go          (레지스트리 - 모든 builtin 등록)
├─ io.go               (print, println)
├─ array.go            (len, append, get, set, slice)
├─ string.go           (concat, substring, upper, lower, split)
├─ math.go             (미래용)
└─ builtin_test.go     (통합 테스트)
```

### 레지스트리 패턴 (v2에서 배움)

```go
type BuiltinFunc struct {
	Name       string
	ParamTypes []TypeInfo
	ReturnType TypeInfo
	Impl       func(...interface{}) interface{}
}

var Builtins = map[string]BuiltinFunc{
	"print": {
		Name:       "print",
		ParamTypes: []TypeInfo{UnknownType},
		ReturnType: UnitType,
		Impl: func(args ...interface{}) interface{} {
			fmt.Print(args[0])
			return nil
		},
	},
	// ... 더 많은 함수
}
```

---

## 📈 구현 순서 (우선순위)

| 순서 | 함수 | 난이도 | 예상시간 |
|------|------|--------|---------|
| 1️⃣ | `print()` | ⭐ | 30분 |
| 2️⃣ | `len()` | ⭐ | 1시간 |
| 3️⃣ | String: `concat`, `substring` | ⭐⭐ | 2시간 |
| 4️⃣ | String: `upper`, `lower`, `split` | ⭐⭐ | 2시간 |
| 5️⃣ | Array: `append`, `get`, `set` | ⭐⭐ | 2시간 |
| 6️⃣ | Array: `slice` | ⭐⭐ | 1시간 |
| **합계** | | | **8시간** |

---

## ✅ 확인된 참고 자료

| 자료 | 파일 | 줄수 | 유용도 |
|------|------|------|--------|
| v2 Builtins | `builtins.ts` | 2,758 | ⭐⭐⭐⭐⭐ (최우수) |
| v4 String | stdlib/ | ~200 | ⭐⭐⭐⭐ (매우좋음) |
| v4 Array | stdlib/ | ~250 | ⭐⭐⭐⭐ (매우좋음) |
| v4 Math | stdlib/ | ~180 | ⭐⭐⭐ (좋음) |

---

## 🚀 Phase 2 준비 완료

**자원 검색 결과**:
- ✅ v2에서 2,758줄 builtins 구현
- ✅ v4에서 87개 함수 정의
- ✅ 패턴 확인: 레지스트리 기반
- ✅ 참고할 수 있는 모든 정보 수집

**다음 단계**: 이 자원들을 참고하여 Go로 구현 시작

---

**자원 수집 완료**: 2026-03-29
**준비 상태**: 🟢 구현 준비 완료
**예상 기간**: 2-3일 (8시간 × 여유)
