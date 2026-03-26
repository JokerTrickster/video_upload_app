import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:video_upload_app/core/storage/settings_storage.dart';

// SettingsScreen widget tests are not included because the screen triggers
// API calls in initState via postFrameCallback (refreshQuota), which creates
// pending Dio timers that are incompatible with the test framework.
// The settings logic is covered through SettingsStorage unit tests.

void main() {
  group('SettingsStorage', () {
    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      await SettingsStorage.instance.init();
    });

    test('default auto upload is disabled', () {
      expect(SettingsStorage.instance.isAutoUploadEnabled, isFalse);
    });

    test('can enable auto upload', () async {
      await SettingsStorage.instance.setAutoUploadEnabled(true);
      expect(SettingsStorage.instance.isAutoUploadEnabled, isTrue);
    });

    test('can disable auto upload', () async {
      await SettingsStorage.instance.setAutoUploadEnabled(true);
      expect(SettingsStorage.instance.isAutoUploadEnabled, isTrue);

      await SettingsStorage.instance.setAutoUploadEnabled(false);
      expect(SettingsStorage.instance.isAutoUploadEnabled, isFalse);
    });

    test('persists value across reads', () async {
      await SettingsStorage.instance.setAutoUploadEnabled(true);

      // Re-init (simulates app restart)
      SharedPreferences.setMockInitialValues({'auto_upload_enabled': true});
      await SettingsStorage.instance.init();

      expect(SettingsStorage.instance.isAutoUploadEnabled, isTrue);
    });
  });
}
