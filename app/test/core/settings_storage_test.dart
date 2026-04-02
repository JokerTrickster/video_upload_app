import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:video_upload_app/core/storage/settings_storage.dart';

void main() {
  group('SettingsStorage - Background Upload Settings', () {
    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      await SettingsStorage.instance.init();
    });

    test('isBackgroundUploadEnabled defaults to true', () {
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, true);
    });

    test('setBackgroundUploadEnabled persists value', () async {
      await SettingsStorage.instance.setBackgroundUploadEnabled(false);
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, false);

      await SettingsStorage.instance.setBackgroundUploadEnabled(true);
      expect(SettingsStorage.instance.isBackgroundUploadEnabled, true);
    });

    test('isWifiOnly defaults to true', () {
      expect(SettingsStorage.instance.isWifiOnly, true);
    });

    test('setWifiOnly persists value', () async {
      await SettingsStorage.instance.setWifiOnly(false);
      expect(SettingsStorage.instance.isWifiOnly, false);

      await SettingsStorage.instance.setWifiOnly(true);
      expect(SettingsStorage.instance.isWifiOnly, true);
    });

    test('isChargingOnly defaults to false', () {
      expect(SettingsStorage.instance.isChargingOnly, false);
    });

    test('setChargingOnly persists value', () async {
      await SettingsStorage.instance.setChargingOnly(true);
      expect(SettingsStorage.instance.isChargingOnly, true);

      await SettingsStorage.instance.setChargingOnly(false);
      expect(SettingsStorage.instance.isChargingOnly, false);
    });

    test('isAutoUploadEnabled defaults to false (existing behavior)', () {
      expect(SettingsStorage.instance.isAutoUploadEnabled, false);
    });

    test('all settings are independent', () async {
      await SettingsStorage.instance.setBackgroundUploadEnabled(false);
      await SettingsStorage.instance.setWifiOnly(false);
      await SettingsStorage.instance.setChargingOnly(true);
      await SettingsStorage.instance.setAutoUploadEnabled(true);

      expect(SettingsStorage.instance.isBackgroundUploadEnabled, false);
      expect(SettingsStorage.instance.isWifiOnly, false);
      expect(SettingsStorage.instance.isChargingOnly, true);
      expect(SettingsStorage.instance.isAutoUploadEnabled, true);
    });

    test('reads from pre-set SharedPreferences values', () async {
      SharedPreferences.setMockInitialValues({
        'background_upload_enabled': false,
        'wifi_only_upload': false,
        'charging_only_upload': true,
      });
      await SettingsStorage.instance.init();

      expect(SettingsStorage.instance.isBackgroundUploadEnabled, false);
      expect(SettingsStorage.instance.isWifiOnly, false);
      expect(SettingsStorage.instance.isChargingOnly, true);
    });
  });
}
