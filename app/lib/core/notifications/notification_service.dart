import 'package:flutter_local_notifications/flutter_local_notifications.dart';

class NotificationService {
  static final NotificationService _instance = NotificationService._();
  factory NotificationService() => _instance;
  NotificationService._();

  final FlutterLocalNotificationsPlugin _plugin =
      FlutterLocalNotificationsPlugin();

  Future<void> init() async {
    const androidSettings =
        AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosSettings = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );
    const settings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );
    await _plugin.initialize(settings);
  }

  Future<void> showUploadComplete(String filename) async {
    await _plugin.show(
      DateTime.now().millisecondsSinceEpoch ~/ 1000,
      'Upload Complete',
      '$filename has been uploaded to YouTube',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'upload_channel',
          'Upload Notifications',
          channelDescription: 'Notifications for video upload status',
          importance: Importance.defaultImportance,
          priority: Priority.defaultPriority,
        ),
        iOS: DarwinNotificationDetails(),
      ),
    );
  }

  Future<void> showUploadFailed(String filename, String error) async {
    await _plugin.show(
      DateTime.now().millisecondsSinceEpoch ~/ 1000,
      'Upload Failed',
      '$filename: $error',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'upload_channel',
          'Upload Notifications',
          channelDescription: 'Notifications for video upload status',
          importance: Importance.high,
          priority: Priority.high,
        ),
        iOS: DarwinNotificationDetails(),
      ),
    );
  }

  Future<void> showQueueProgress(int completed, int total) async {
    await _plugin.show(
      0, // fixed ID for progress notification
      'Auto Upload Progress',
      '$completed/$total videos uploaded today',
      const NotificationDetails(
        android: AndroidNotificationDetails(
          'queue_channel',
          'Queue Notifications',
          channelDescription: 'Notifications for auto-upload queue',
          importance: Importance.low,
          priority: Priority.low,
        ),
        iOS: DarwinNotificationDetails(),
      ),
    );
  }
}
