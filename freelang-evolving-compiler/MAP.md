# 프로젝트 맵

## 파일 구조

```
internal/ast/nodes.go
internal/codegen/codegen.go
internal/codegen/codegen_test.go
internal/evolution/evolution_test.go
internal/evolution/recorder.go
internal/evolution/regression.go
internal/ir/generator.go
internal/ir/ir.go
internal/ir/ir_test.go
internal/lexer/lexer.go
internal/lexer/lexer_test.go
internal/optimizer/adaptive.go
internal/optimizer/optimizer_test.go
internal/optimizer/rule.go
internal/parser/parser.go
internal/parser/parser_test.go
internal/profiler/collector.go
internal/profiler/db.go
internal/profiler/pattern.go
internal/profiler/profiler_test.go
main.go
```

## 함수 목록

### lexer_test.go

- `TestLexer`
- `TestOperators`

### lexer.go

- `New`
- `isDigit`
- `isLetter`
- `lookupKeyword`

### parser.go

- `New`
- `precedence`

### parser_test.go

- `TestParseBinaryExpressions`
- `TestParseFnDecl`
- `TestParseForStmt`
- `TestParseIfStmt`
- `TestParseLetDecl`
- `TestParseMultipleStatements`
- `TestParseNesting`

### profiler_test.go

- `TestAverageBuildTime`
- `TestCollectorBasic`
- `TestCollectorConstantExpressions`
- `TestDatabaseCreation`
- `TestDatabasePersistence`
- `TestDatabaseTopPatterns`
- `TestDatabaseUpdate`
- `TestDeadAssignDetection`
- `TestEmptyProgram`
- `TestGetPatternStats`
- `TestMultipleBuildHistory`
- `TestPatternContextAnalysis`
- `TestPatternSignature`
- `TestPatternTypeDetection`
- `TestRegressionDetection`

### db.go

- `LoadFromFile`
- `NewDatabase`

### pattern.go

- `NewPatternContext`
- `PatternSignature`
- `PatternType`
- `findConstExpr`
- `findFunctionDef`
- `isLiteral`
- `nodeKindStr`

### collector.go

- `NewCollector`
- `formatNs`
- `isFunctionInlinable`

### adaptive.go

- `NewAdaptiveOptimizer`
- `countNodes`
- `nodeSignature`
- `parsePatternKind`

### rule.go

- `DefaultRules`
- `evaluateConstExpr`
- `formatInt`
- `init`
- `initCommonSubexpressionRule`
- `initConstantFoldingRule`
- `initDeadCodeEliminationRule`
- `initInliningRule`
- `initLoopInvariantMovementRule`
- `parseIntLiteral`

### optimizer_test.go

- `TestAdaptiveOptimizerCreation`
- `TestAdaptiveOptimizerPriorities`
- `TestConstantFolding`
- `TestConstantFoldingDivision`
- `TestConstantFoldingMultiplication`
- `TestConstantFoldingNested`
- `TestConstantFoldingSubtraction`
- `TestCountNodes`
- `TestCountNodesNil`
- `TestDivisionByZero`
- `TestFormatInt`
- `TestNodeSignature`
- `TestOptimizeAST`
- `TestOptimizeNilNode`
- `TestOptimizeWithStats`
- `TestParsePatternKind`
- `TestRuleOrder`

### recorder.go

- `NewEvolutionRecorder`

### regression.go

- `NewRegressionDetector`

### evolution_test.go

- `TestAnalyzeFull`
- `TestAverageBuildTime`
- `TestAverageOptimizationsApplied`
- `TestDetectOutlier`
- `TestDetectRegression`
- `TestDetectTrendRegression`
- `TestDetectionCreation`
- `TestGetBuildsSince`
- `TestGetLastBuild`
- `TestGetOptimizationFrequency`
- `TestHealthStatusDegraded`
- `TestHealthStatusHealthy`
- `TestLatestBuildTime`
- `TestMetricsFields`
- `TestMultipleBuilds`
- `TestNoRegressionBeforeBaseline`
- `TestRecordBuild`
- `TestRecorderCreation`
- `TestSummaryOutput`

### generator.go

- `NewGenerator`

### ir_test.go

- `TestGenerateBinaryAdd`
- `TestGenerateCallExpr`
- `TestGenerateEmpty`
- `TestGenerateFnDecl`
- `TestGenerateForStmt`
- `TestGenerateIfStmt`
- `TestGenerateIntLit`
- `TestGenerateLetDecl`
- `TestGenerateReturn`
- `TestProgramByteSize`

### codegen.go

- `New`

### codegen_test.go

- `TestCodeGenAddExpr`
- `TestCodeGenCallExpr`
- `TestCodeGenEmpty`
- `TestCodeGenForLoop`
- `TestCodeGenFunction`
- `TestCodeGenIfStmt`
- `TestCodeGenIntLit`
- `TestCodeGenReturn`
- `TestCodeGenRoundTrip`
- `TestResultByteSize`

### main.go

- `compileCode`
- `lexCode`
- `main`
- `parseCode`
- `printAST`
- `profileCode`
- `profileReport`

