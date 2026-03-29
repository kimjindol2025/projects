import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../models/db_status.dart';

class DbScreen extends StatefulWidget {
  const DbScreen({Key? key}) : super(key: key);

  @override
  State<DbScreen> createState() => _DbScreenState();
}

class _DbScreenState extends State<DbScreen> {
  late Future<DbStatus> _futureDb;

  @override
  void initState() {
    super.initState();
    _futureDb = ApiService.fetchDb();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('💾 Database Status'),
        backgroundColor: Colors.black,
      ),
      body: FutureBuilder<DbStatus>(
        future: _futureDb,
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
            final db = snapshot.data!;
            final hitRatePercent = (db.performance.indexHitRate * 100).toInt();

            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(
                  db.name,
                  style: const TextStyle(
                    color: Color(0xFF00FF41),
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Phase ${db.phase} | ${db.modules} Modules | ${db.totalLines} Lines',
                  style: const TextStyle(color: Colors.white70, fontSize: 12),
                ),
                const SizedBox(height: 24),
                // Index Hit Rate Gauge
                Center(
                  child: Column(
                    children: [
                      SizedBox(
                        width: 150,
                        height: 150,
                        child: Stack(
                          alignment: Alignment.center,
                          children: [
                            CircularProgressIndicator(
                              value: db.performance.indexHitRate,
                              strokeWidth: 8,
                              backgroundColor: Colors.grey[800],
                              valueColor:
                                  const AlwaysStoppedAnimation<Color>(
                                Color(0xFF00FF41),
                              ),
                            ),
                            Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  '$hitRatePercent%',
                                  style: const TextStyle(
                                    color: Color(0xFF00FF41),
                                    fontSize: 28,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                const Text(
                                  'Index Hit Rate',
                                  style: TextStyle(
                                    color: Colors.white54,
                                    fontSize: 11,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 24),
                      Text(
                        'Performance Metrics',
                        style: const TextStyle(
                          color: Color(0xFF00FF41),
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                _buildMetricCard(
                  '⏱️ Query Latency',
                  '${db.performance.queryLatencyMs.toStringAsFixed(1)}ms',
                  'Average time per query',
                ),
                _buildMetricCard(
                  '⚡ Insert Throughput',
                  '${db.performance.insertThroughputPerSec} ops/sec',
                  'Records inserted per second',
                ),
                _buildMetricCard(
                  '💾 Memory Usage',
                  '${db.memoryUsageMb.toStringAsFixed(1)} MB',
                  'Current memory footprint',
                ),
                _buildMetricCard(
                  '🔄 Active Transactions',
                  '${db.activeTransactions}',
                  'Running transactions',
                ),
                _buildMetricCard(
                  '📋 Cached Queries',
                  '${db.cachedQueries}',
                  'Cached query results',
                ),
              ],
            );
          }
          return const Center(child: Text('No data'));
        },
      ),
    );
  }

  Widget _buildMetricCard(String label, String value, String description) {
    return Container(
      padding: const EdgeInsets.all(12),
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        border: Border.all(color: const Color(0xFF00FF41), width: 0.5),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(
              color: Color(0xFF00FF41),
              fontSize: 12,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            description,
            style: const TextStyle(color: Colors.white54, fontSize: 11),
          ),
        ],
      ),
    );
  }
}
