import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../models/project_status.dart';

class StatusScreen extends StatefulWidget {
  const StatusScreen({Key? key}) : super(key: key);

  @override
  State<StatusScreen> createState() => _StatusScreenState();
}

class _StatusScreenState extends State<StatusScreen> {
  late Future<ApiStatusResponse> _futureStatus;

  @override
  void initState() {
    super.initState();
    _futureStatus = ApiService.fetchStatus();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('📊 Project Status'),
        backgroundColor: Colors.black,
      ),
      body: FutureBuilder<ApiStatusResponse>(
        future: _futureStatus,
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
            final status = snapshot.data!;
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(
                  'Last Update: ${status.lastUpdate}',
                  style: const TextStyle(color: Color(0xFF00FF41), fontSize: 12),
                ),
                const SizedBox(height: 16),
                ...status.projects.map((project) => _buildProjectCard(project)),
                const SizedBox(height: 16),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    border: Border.all(color: const Color(0xFF00FF41)),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Totals',
                        style: TextStyle(
                          color: Color(0xFF00FF41),
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Lines: ${status.totalLines}',
                        style: const TextStyle(color: Colors.white70),
                      ),
                      Text(
                        'Tests: ${status.totalTests}',
                        style: const TextStyle(color: Colors.white70),
                      ),
                    ],
                  ),
                ),
              ],
            );
          }
          return const Center(child: Text('No data'));
        },
      ),
    );
  }

  Widget _buildProjectCard(ProjectStatus project) {
    final progress = project.phase / 11.0;
    return Card(
      color: Colors.grey[900],
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  project.name,
                  style: const TextStyle(
                    color: Color(0xFF00FF41),
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  project.status,
                  style: const TextStyle(color: Colors.white70, fontSize: 12),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Phase ${project.phase}/11',
              style: const TextStyle(color: Colors.white70),
            ),
            const SizedBox(height: 8),
            LinearProgressIndicator(
              value: progress,
              backgroundColor: Colors.grey[800],
              valueColor: const AlwaysStoppedAnimation<Color>(Color(0xFF00FF41)),
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildStat('Lines', '${project.lines}'),
                _buildStat('Files', '${project.files}'),
                _buildStat('Tests', '${project.tests}'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStat(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Color(0xFF00FF41),
            fontSize: 14,
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
}
