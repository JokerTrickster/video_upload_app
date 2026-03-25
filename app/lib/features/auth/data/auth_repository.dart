import '../../../core/api/api_client.dart';
import '../../../core/constants/api_constants.dart';
import '../../../shared/models/user_model.dart';

class AuthRepository {
  final ApiClient _apiClient;

  AuthRepository(this._apiClient);

  Future<String> getGoogleAuthUrl() async {
    final response = await _apiClient.get(ApiConstants.authGoogleUrl);
    return response.data['data']['auth_url'] as String;
  }

  Future<AuthResponse> handleCallback(String code, String state) async {
    final response = await _apiClient.post(
      ApiConstants.authGoogleCallback,
      data: {'code': code, 'state': state},
    );
    return AuthResponse.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<UserModel> getCurrentUser() async {
    final response = await _apiClient.get(ApiConstants.authMe);
    return UserModel.fromJson(response.data['data'] as Map<String, dynamic>);
  }

  Future<void> logout() async {
    await _apiClient.post(ApiConstants.authLogout);
    await _apiClient.clearTokens();
  }
}
