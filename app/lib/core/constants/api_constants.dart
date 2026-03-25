class ApiConstants {
  static const String baseUrl = 'http://localhost:8080/api/v1';

  // Auth endpoints
  static const String authGoogleUrl = '/auth/google/url';
  static const String authGoogleCallback = '/auth/google/callback';
  static const String authRefresh = '/auth/refresh';
  static const String authMe = '/auth/me';
  static const String authLogout = '/auth/logout';

  // Media endpoints
  static const String mediaUploadInitiate = '/media/upload/initiate';
  static const String mediaUploadVideo = '/media/upload/video';
  static const String mediaUploadStatus = '/media/upload/status';
  static const String mediaUploadComplete = '/media/upload/complete';
  static const String mediaUploadCancel = '/media/upload/cancel';
  static const String mediaList = '/media/list';
  static const String mediaDetail = '/media'; // + /:asset_id
}
