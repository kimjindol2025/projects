import 'dart:convert';

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
