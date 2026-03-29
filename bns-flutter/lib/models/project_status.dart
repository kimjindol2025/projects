import 'dart:convert';

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
