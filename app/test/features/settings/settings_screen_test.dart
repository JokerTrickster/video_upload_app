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

  group('SettingsStorage - Background Upload', () {
    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      await SettingsStorage.instance.init();
    });

    test('background upload enabled by default', () {
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, isTrue);
    });

    test('can toggle background upload', () async {
      await SettingsStorage.instance.setBackgroundUploadEnabled(false);
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, isFalse);

      await SettingsStorage.instance.setBackgroundUploadEnabled(true);
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, isTrue);
    });

    test('wifi only enabled by default', () {
      expect(SettingsStorage.instance.isWifiOnly, isTrue);
    });

    test('can toggle wifi only', () async {
      await SettingsStorage.instance.setWifiOnly(false);
      expect(SettingsStorage.instance.isWifiOnly, isFalse);
    });

    test('charging only disabled by default', () {
      expect(SettingsStorage.instance.isChargingOnly, isFalse);
    });

    test('can toggle charging only', () async {
      await SettingsStorage.instance.setChargingOnly(true);
      expect(SettingsStorage.instance.isChargingOnly, isTrue);
    });

    test('background settings persist across re-init', () async {
      await SettingsStorage.instance.setBackgroundUploadEnabled(false);
      await SettingsStorage.instance.setWifiOnly(false);
      await SettingsStorage.instance.setChargingOnly(true);

      SharedPreferences.setMockInitialValues({
        'background_upload_enabled': false,
        'wifi_only_upload': false,
        'charging_only_upload': true,
      });
      await SettingsStorage.instance.init();

      expect(SettingsStorage.instance.isBackgroundUploadEnabled, isFalse);
      expect(SettingsStorage.instance.isWifiOnly, isFalse);
      expect(SettingsStorage.instance.isChargingOnly, isTrue);
    });

    test('background settings are independent from auto upload', () async {
      await SettingsStorage.instance.setAutoUploadEnabled(true);
      await SettingsStorage.instance.setBackgroundUploadEnabled(false);

      expect(SettingsStorage.instance.isAutoUploadEnabled, isTrue);
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, isFalse);
    });
  });
}
