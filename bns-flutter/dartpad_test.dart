import 'dart:convert';

/// Phase 4 모델 클래스 (Flutter 의존성 없음)

class ProjectStatus {
  final String name;
  final int phase;
  final int lines;
  final int files;
  final int tests;
  final String status;

  ProjectStatus({
    required this.name,
    required this.phase,
    required this.lines,
    required this.files,
    required this.tests,
    required this.status,
  });

  factory ProjectStatus.fromJson(Map<String, dynamic> json) {
    return ProjectStatus(
      name: json['name'] as String,
      phase: json['phase'] as int,
      lines: json['lines'] as int,
      files: json['files'] as int,
      tests: json['tests'] as int,
      status: json['status'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'phase': phase,
        'lines': lines,
        'files': files,
        'tests': tests,
        'status': status,
      };
}

class ApiStatusResponse {
  final String lastUpdate;
  final List<ProjectStatus> projects;
  final int totalLines;
  final int totalTests;

  ApiStatusResponse({
    required this.lastUpdate,
    required this.projects,
    required this.totalLines,
    required this.totalTests,
  });

  factory ApiStatusResponse.fromJson(Map<String, dynamic> json) {
    return ApiStatusResponse(
      lastUpdate: json['last_update'] as String,
      projects: (json['projects'] as List<dynamic>)
          .map((p) => ProjectStatus.fromJson(p as Map<String, dynamic>))
          .toList(),
      totalLines: json['total_lines'] as int,
      totalTests: json['total_tests'] as int,
    );
  }

  Map<String, dynamic> toJson() => {
        'last_update': lastUpdate,
        'projects': projects.map((p) => p.toJson()).toList(),
        'total_lines': totalLines,
        'total_tests': totalTests,
      };
}

class GogsCommit {
  final String hash;
  final String message;
  final String repo;
  final String date;
  final int filesChanged;
  final int insertions;
  final int deletions;

  GogsCommit({
    required this.hash,
    required this.message,
    required this.repo,
    required this.date,
    required this.filesChanged,
    required this.insertions,
    required this.deletions,
  });

  factory GogsCommit.fromJson(Map<String, dynamic> json) {
    return GogsCommit(
      hash: json['hash'] as String,
      message: json['message'] as String,
      repo: json['repo'] as String,
      date: json['date'] as String,
      filesChanged: json['files_changed'] as int,
      insertions: json['insertions'] as int,
      deletions: json['deletions'] as int,
    );
  }

  Map<String, dynamic> toJson() => {
        'hash': hash,
        'message': message,
        'repo': repo,
        'date': date,
        'files_changed': filesChanged,
        'insertions': insertions,
        'deletions': deletions,
      };
}

class ApiGogsResponse {
  final List<GogsCommit> recentCommits;
  final int repoCount;
  final int totalCommits;

  ApiGogsResponse({
    required this.recentCommits,
    required this.repoCount,
    required this.totalCommits,
  });

  factory ApiGogsResponse.fromJson(Map<String, dynamic> json) {
    return ApiGogsResponse(
      recentCommits: (json['recent_commits'] as List<dynamic>)
          .map((c) => GogsCommit.fromJson(c as Map<String, dynamic>))
          .toList(),
      repoCount: json['repo_count'] as int,
      totalCommits: json['total_commits'] as int,
    );
  }

  Map<String, dynamic> toJson() => {
        'recent_commits': recentCommits.map((c) => c.toJson()).toList(),
        'repo_count': repoCount,
        'total_commits': totalCommits,
      };
}

class SseEvent {
  final String? eventType;
  final String? message;
  final String? status;
  final String? data;
  final int? timestamp;

  SseEvent({
    this.eventType,
    this.message,
    this.status,
    this.data,
    this.timestamp,
  });

  bool get isWaiting => status == 'waiting';

  factory SseEvent.fromJson(Map<String, dynamic> json) {
    return SseEvent(
      eventType: json['event_type'] as String?,
      message: json['message'] as String?,
      status: json['status'] as String?,
      data: json['data'] as String?,
      timestamp: json['timestamp'] as int?,
    );
  }

  factory SseEvent.waiting() {
    return SseEvent(
      status: 'waiting',
      message: 'No events yet',
    );
  }

  Map<String, dynamic> toJson() => {
        if (eventType != null) 'event_type': eventType,
        if (message != null) 'message': message,
        if (status != null) 'status': status,
        if (data != null) 'data': data,
        if (timestamp != null) 'timestamp': timestamp,
      };
}

class DbPerformance {
  final double queryLatencyMs;
  final int insertThroughputPerSec;
  final double indexHitRate;

  DbPerformance({
    required this.queryLatencyMs,
    required this.insertThroughputPerSec,
    required this.indexHitRate,
  });

  factory DbPerformance.fromJson(Map<String, dynamic> json) {
    return DbPerformance(
      queryLatencyMs: (json['query_latency_ms'] as num).toDouble(),
      insertThroughputPerSec: json['insert_throughput_per_sec'] as int,
      indexHitRate: (json['index_hit_rate'] as num).toDouble(),
    );
  }

  Map<String, dynamic> toJson() => {
        'query_latency_ms': queryLatencyMs,
        'insert_throughput_per_sec': insertThroughputPerSec,
        'index_hit_rate': indexHitRate,
      };
}

class DbStatus {
  final String name;
  final int phase;
  final int modules;
  final int totalLines;
  final double memoryUsageMb;
  final int activeTransactions;
  final int cachedQueries;
  final DbPerformance performance;

  DbStatus({
    required this.name,
    required this.phase,
    required this.modules,
    required this.totalLines,
    required this.memoryUsageMb,
    required this.activeTransactions,
    required this.cachedQueries,
    required this.performance,
  });

  factory DbStatus.fromJson(Map<String, dynamic> json) {
    return DbStatus(
      name: json['name'] as String,
      phase: json['phase'] as int,
      modules: json['modules'] as int,
      totalLines: json['total_lines'] as int,
      memoryUsageMb: (json['memory_usage_mb'] as num).toDouble(),
      activeTransactions: json['active_transactions'] as int,
      cachedQueries: json['cached_queries'] as int,
      performance: DbPerformance.fromJson(
        json['performance'] as Map<String, dynamic>,
      ),
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'phase': phase,
        'modules': modules,
        'total_lines': totalLines,
        'memory_usage_mb': memoryUsageMb,
        'active_transactions': activeTransactions,
        'cached_queries': cachedQueries,
        'performance': performance.toJson(),
      };
}

/// DartPad 테스트 실행
void main() {
  print('🧪 BNS Phase 4 - Dart 모델 JSON 검증\n');

  testProjectStatus();
  testApiStatusResponse();
  testGogsCommit();
  testApiGogsResponse();
  testSseEvent();
  testDbStatus();

  print('\n✅ 모든 JSON 직렬화 검증 완료!\n');
}

void testProjectStatus() {
  print('테스트 1️⃣ : ProjectStatus fromJson/toJson');

  final json = {
    'name': 'Zero-Copy-DB',
    'phase': 11,
    'lines': 22439,
    'files': 56,
    'tests': 182,
    'status': '✅ COMPLETE',
  };

  final p = ProjectStatus.fromJson(json);
  assert(p.name == 'Zero-Copy-DB', 'name 불일치');
  assert(p.phase == 11, 'phase 불일치');
  assert(p.lines == 22439, 'lines 불일치');
  assert(p.files == 56, 'files 불일치');
  assert(p.tests == 182, 'tests 불일치');
  assert(p.status == '✅ COMPLETE', 'status 불일치');

  final json2 = p.toJson();
  assert(json2['name'] == json['name'], 'toJson name 불일치');
  assert(json2['phase'] == json['phase'], 'toJson phase 불일치');

  print('  ✅ ProjectStatus: ${p.name} (phase ${p.phase})\n');
}

void testApiStatusResponse() {
  print('테스트 2️⃣ : ApiStatusResponse fromJson (중첩 배열)');

  final json = {
    'last_update': '2026-03-29',
    'projects': [
      {
        'name': 'Zero-Copy-DB',
        'phase': 11,
        'lines': 22439,
        'files': 56,
        'tests': 182,
        'status': '✅ COMPLETE',
      },
      {
        'name': 'Compiler',
        'phase': 8,
        'lines': 4435,
        'files': 20,
        'tests': 80,
        'status': '✅ COMPLETE',
      },
    ],
    'total_lines': 26874,
    'total_tests': 262,
  };

  final resp = ApiStatusResponse.fromJson(json);
  assert(resp.projects.length == 2, 'projects 배열 길이 불일치');
  assert(resp.projects[0].name == 'Zero-Copy-DB', '첫 번째 프로젝트 이름 불일치');
  assert(resp.totalLines == 26874, 'totalLines 불일치');

  print('  ✅ ApiStatusResponse: ${resp.projects.length} 프로젝트\n');
}

void testGogsCommit() {
  print('테스트 3️⃣ : GogsCommit fromJson');

  final json = {
    'hash': '2fcb909',
    'message': '🎯 Claude Code 심화 시스템',
    'repo': 'freelang-ecosystem',
    'date': '2026-03-29',
    'files_changed': 5,
    'insertions': 89,
    'deletions': 45,
  };

  final c = GogsCommit.fromJson(json);
  assert(c.hash == '2fcb909', 'hash 불일치');
  assert(c.insertions == 89, 'insertions 불일치');
  assert(c.deletions == 45, 'deletions 불일치');

  print('  ✅ GogsCommit: +${c.insertions} -${c.deletions}\n');
}

void testApiGogsResponse() {
  print('테스트 4️⃣ : ApiGogsResponse fromJson (중첩 배열)');

  final json = {
    'repo_count': 6,
    'total_commits': 100,
    'recent_commits': [
      {
        'hash': '2fcb909',
        'message': '🎯 Claude Code',
        'repo': 'freelang',
        'date': '2026-03-29',
        'files_changed': 5,
        'insertions': 89,
        'deletions': 45,
      },
      {
        'hash': 'd52e189',
        'message': '📚 메모리 업데이트',
        'repo': 'freelang',
        'date': '2026-03-28',
        'files_changed': 3,
        'insertions': 42,
        'deletions': 12,
      },
    ],
  };

  final resp = ApiGogsResponse.fromJson(json);
  assert(resp.recentCommits.length == 2, 'recent_commits 길이 불일치');
  assert(resp.recentCommits[0].hash == '2fcb909', '첫 커밋 hash 불일치');

  print('  ✅ ApiGogsResponse: ${resp.recentCommits.length} 커밋\n');
}

void testSseEvent() {
  print('테스트 5️⃣ : SseEvent isWaiting + factory');

  final waiting = SseEvent.waiting();
  assert(waiting.isWaiting == true, 'waiting.isWaiting false');
  assert(waiting.status == 'waiting', 'waiting.status 불일치');

  final json = {
    'event_type': 'commit',
    'message': 'Push to master',
    'data': 'hash=2fcb909',
    'timestamp': 1743200000,
  };

  final event = SseEvent.fromJson(json);
  assert(event.isWaiting == false, 'event.isWaiting true');
  assert(event.eventType == 'commit', 'eventType 불일치');

  print('  ✅ SseEvent: isWaiting=${waiting.isWaiting}, eventType=${event.eventType}\n');
}

void testDbStatus() {
  print('테스트 6️⃣ : DbStatus (중첩 DbPerformance)');

  final json = {
    'name': 'Zero-Copy-DB',
    'phase': 11,
    'modules': 11,
    'total_lines': 22439,
    'memory_usage_mb': 8.5,
    'active_transactions': 3,
    'cached_queries': 12,
    'performance': {
      'query_latency_ms': 2.5,
      'insert_throughput_per_sec': 5000,
      'index_hit_rate': 0.94,
    },
  };

  final db = DbStatus.fromJson(json);
  assert(db.name == 'Zero-Copy-DB', 'name 불일치');
  assert(db.phase == 11, 'phase 불일치');
  assert(db.performance.indexHitRate == 0.94, 'indexHitRate 불일치');

  final json2 = db.toJson();
  assert(json2['performance']['index_hit_rate'] == 0.94, 'toJson indexHitRate 불일치');

  print('  ✅ DbStatus: ${db.name} (index_hit_rate=${db.performance.indexHitRate})\n');
}
