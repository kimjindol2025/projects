import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../models/gogs_commit.dart';

class GogsScreen extends StatefulWidget {
  const GogsScreen({Key? key}) : super(key: key);

  @override
  State<GogsScreen> createState() => _GogsScreenState();
}

class _GogsScreenState extends State<GogsScreen> {
  late Future<ApiGogsResponse> _futureGogs;

  @override
  void initState() {
    super.initState();
    _futureGogs = ApiService.fetchGogs();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('🔧 Recent Commits'),
        backgroundColor: Colors.black,
      ),
      body: FutureBuilder<ApiGogsResponse>(
        future: _futureGogs,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(
              child: CircularProgressIndicator(color: Color(0xFF00FF41)),
            );
          } else if (snapshot.hasError) {
            return Center(
              child: Text(
                'Error: ${snapshot.error}',
                style: const TextStyle(color: Colors.red),
              ),
            );
          } else if (snapshot.hasData) {
            final gogs = snapshot.data!;
            return ListView(
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  color: Colors.grey[900],
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _buildHeader('Repos', '${gogs.repoCount}'),
                      _buildHeader('Commits', '${gogs.totalCommits}'),
                    ],
                  ),
                ),
                const Divider(color: Color(0xFF00FF41)),
                ...gogs.recentCommits.map((commit) => _buildCommitTile(commit)),
              ],
            );
          }
          return const Center(child: Text('No data'));
        },
      ),
    );
  }

  Widget _buildHeader(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Color(0xFF00FF41),
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          label,
          style: const TextStyle(color: Colors.white54, fontSize: 12),
        ),
      ],
    );
  }

  Widget _buildCommitTile(GogsCommit commit) {
    return Container(
      padding: const EdgeInsets.all(12),
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        border: Border.all(color: const Color(0xFF00FF41), width: 0.5),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                commit.hash,
                style: const TextStyle(
                  color: Color(0xFF00FF41),
                  fontFamily: 'monospace',
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                commit.date,
                style: const TextStyle(color: Colors.white54, fontSize: 11),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            commit.message,
            style: const TextStyle(color: Colors.white, fontSize: 13),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 6),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Files: ${commit.filesChanged}',
                style: const TextStyle(color: Colors.white54, fontSize: 11),
              ),
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: Colors.green[900],
                      borderRadius: BorderRadius.circular(2),
                    ),
                    child: Text(
                      '+${commit.insertions}',
                      style: const TextStyle(
                        color: Colors.green,
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(width: 4),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: Colors.red[900],
                      borderRadius: BorderRadius.circular(2),
                    ),
                    child: Text(
                      '-${commit.deletions}',
                      style: const TextStyle(
                        color: Colors.red,
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }
}
