import 'dart:convert';

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
