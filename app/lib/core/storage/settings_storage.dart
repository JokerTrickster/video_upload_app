import 'package:shared_preferences/shared_preferences.dart';

class SettingsStorage {
  static const _keyAutoUpload = 'auto_upload_enabled';

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
}
