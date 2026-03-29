import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/project_status.dart';
import '../models/gogs_commit.dart';
import '../models/db_status.dart';

const String _baseUrl = 'http://localhost:28080';

class ApiService {
  static Future<ApiStatusResponse> fetchStatus() async {
    try {
      final response = await http
          .get(Uri.parse('$_baseUrl/api/status'))
          .timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        return ApiStatusResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
      } else {
        throw Exception('Failed to load status: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching status: $e');
    }
  }

  static Future<ApiGogsResponse> fetchGogs() async {
    try {
      final response = await http
          .get(Uri.parse('$_baseUrl/api/gogs'))
          .timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        return ApiGogsResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
      } else {
        throw Exception('Failed to load gogs: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching gogs: $e');
    }
  }

  static Future<DbStatus> fetchDb() async {
    try {
      final response = await http
          .get(Uri.parse('$_baseUrl/api/db'))
          .timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        return DbStatus.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
      } else {
        throw Exception('Failed to load db: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error fetching db: $e');
    }
  }
}
