import 'package:shared_preferences/shared_preferences.dart';

class SettingsStorage {
  static const _keyAutoUpload = 'auto_upload_enabled';
  static const _keyBackgroundUpload = 'background_upload_enabled';
  static const _keyWifiOnly = 'wifi_only_upload';
  static const _keyChargingOnly = 'charging_only_upload';

  static SettingsStorage? _instance;
  SharedPreferences? _prefs;

  SettingsStorage._();

  static SettingsStorage get instance {
    _instance ??= SettingsStorage._();
    return _instance!;
  }

  Future<void> init() async {
    _prefs ??= await SharedPreferences.getInstance();
  }

  bool get isAutoUploadEnabled {
    return _prefs?.getBool(_keyAutoUpload) ?? false;
  }

  Future<void> setAutoUploadEnabled(bool value) async {
    await init();
    await _prefs!.setBool(_keyAutoUpload, value);
  }

  bool get isBackgroundUploadEnabled {
    return _prefs?.getBool(_keyBackgroundUpload) ?? true;
  }

  Future<void> setBackgroundUploadEnabled(bool value) async {
    await init();
    await _prefs!.setBool(_keyBackgroundUpload, value);
  }

  bool get isWifiOnly {
    return _prefs?.getBool(_keyWifiOnly) ?? true;
  }

  Future<void> setWifiOnly(bool value) async {
    await init();
    await _prefs!.setBool(_keyWifiOnly, value);
  }

  bool get isChargingOnly {
    return _prefs?.getBool(_keyChargingOnly) ?? false;
  }

  Future<void> setChargingOnly(bool value) async {
    await init();
    await _prefs!.setBool(_keyChargingOnly, value);
  }
}
