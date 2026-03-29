import 'package:flutter/material.dart';
import '../services/sse_service.dart';
import '../models/sse_event.dart';

class FeedScreen extends StatefulWidget {
  const FeedScreen({Key? key}) : super(key: key);

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends State<FeedScreen> {
  late Stream<SseEvent> _feedStream;
  final List<SseEvent> _events = [];

  @override
  void initState() {
    super.initState();
    _feedStream = SseService.feedStream();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('📡 Real-time Feed'),
        backgroundColor: Colors.black,
      ),
      body: StreamBuilder<SseEvent>(
        stream: _feedStream,
        builder: (context, snapshot) {
          if (snapshot.hasData) {
            final event = snapshot.data!;

            // Add to list and limit to 20 recent events
            if (!event.isWaiting) {
              _events.insert(0, event);
              if (_events.length > 20) {
                _events.removeLast();
              }
            }

            if (_events.isEmpty && event.isWaiting) {
              return const Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    CircularProgressIndicator(color: Color(0xFF00FF41)),
                    SizedBox(height: 16),
                    Text(
                      'Waiting for events...',
                      style: TextStyle(color: Colors.white70),
                    ),
                  ],
                ),
              );
            }

            if (_events.isEmpty) {
              return const Center(
                child: Text(
                  'No events',
                  style: TextStyle(color: Colors.white70),
                ),
              );
            }

            return ListView.builder(
              itemCount: _events.length,
              itemBuilder: (context, index) => _buildEventTile(_events[index]),
            );
          } else if (snapshot.hasError) {
            return Center(
              child: Text(
                'Error: ${snapshot.error}',
                style: const TextStyle(color: Colors.red),
              ),
            );
          }

          return const Center(
            child: CircularProgressIndicator(color: Color(0xFF00FF41)),
          );
        },
      ),
    );
  }

  Widget _buildEventTile(SseEvent event) {
    Color badgeColor;
    String badgeText;

    switch (event.eventType) {
      case 'commit':
        badgeColor = Colors.green[700] ?? Colors.green;
        badgeText = '🔗 COMMIT';
        break;
      case 'test':
        badgeColor = Colors.blue[700] ?? Colors.blue;
        badgeText = '✓ TEST';
        break;
      case 'build':
        badgeColor = Colors.orange[700] ?? Colors.orange;
        badgeText = '🔨 BUILD';
        break;
      default:
        badgeColor = Colors.grey[700] ?? Colors.grey;
        badgeText = '⏳ WAIT';
    }

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
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: badgeColor,
                  borderRadius: BorderRadius.circular(3),
                ),
                child: Text(
                  badgeText,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              if (event.timestamp != null)
                Text(
                  DateTime.fromMillisecondsSinceEpoch(event.timestamp! * 1000)
                      .toString()
                      .substring(0, 19),
                  style: const TextStyle(color: Colors.white54, fontSize: 11),
                ),
            ],
          ),
          const SizedBox(height: 8),
          if (event.message != null)
            Text(
              event.message!,
              style: const TextStyle(color: Colors.white, fontSize: 13),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          if (event.data != null) ...[
            const SizedBox(height: 8),
            Text(
              'Data: ${event.data!.substring(0, event.data!.length > 50 ? 50 : event.data!.length)}...',
              style: const TextStyle(color: Colors.white54, fontSize: 11),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ],
      ),
    );
  }
}
