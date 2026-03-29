import 'dart:convert';

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
