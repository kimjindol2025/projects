import 'dart:convert';
import 'dart:async';
import 'package:http/http.dart' as http;
import '../models/sse_event.dart';

const String _baseUrl = 'http://localhost:28080';

class SseService {
  static Stream<SseEvent> feedStream() async* {
    while (true) {
      try {
        final response = await http
            .get(Uri.parse('$_baseUrl/api/feed'))
            .timeout(const Duration(seconds: 10));

        if (response.statusCode == 200) {
          // Parse "data: {...}\n\n" format
          final body = response.body.trim();
          String jsonStr = body;

          if (body.startsWith('data: ')) {
            jsonStr = body.substring(6).trim();
          }

          try {
            final json = jsonDecode(jsonStr) as Map<String, dynamic>;
            yield SseEvent.fromJson(json);
          } catch (e) {
            // If parsing fails, return waiting state
            yield SseEvent.waiting();
          }
        } else {
          yield SseEvent.waiting();
        }
      } catch (e) {
        // On network error, yield waiting state
        yield SseEvent.waiting();
      }

      // Poll every 3 seconds
      await Future.delayed(const Duration(seconds: 3));
    }
  }
}
