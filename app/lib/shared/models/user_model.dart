class UserModel {
  final String id;
  final String email;
  final String googleId;
  final String? youtubeChannelId;
  final String? youtubeChannelName;
  final String? profileImageUrl;
  final DateTime createdAt;

  UserModel({
    required this.id,
    required this.email,
    required this.googleId,
    this.youtubeChannelId,
    this.youtubeChannelName,
    this.profileImageUrl,
    required this.createdAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] as String,
      email: json['email'] as String,
      googleId: json['google_id'] as String,
      youtubeChannelId: json['youtube_channel_id'] as String?,
      youtubeChannelName: json['youtube_channel_name'] as String?,
      profileImageUrl: json['profile_image_url'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class AuthTokens {
  final String accessToken;
  final String refreshToken;
  final int expiresIn;
  final String tokenType;

  AuthTokens({
    required this.accessToken,
    required this.refreshToken,
    required this.expiresIn,
    required this.tokenType,
  });

  factory AuthTokens.fromJson(Map<String, dynamic> json) {
    return AuthTokens(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
      expiresIn: json['expires_in'] as int,
      tokenType: json['token_type'] as String,
    );
  }
}

class AuthResponse {
  final AuthTokens tokens;
  final UserModel user;

  AuthResponse({required this.tokens, required this.user});

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      tokens: AuthTokens(
        accessToken: json['access_token'] as String,
        refreshToken: json['refresh_token'] as String,
        expiresIn: json['expires_in'] as int,
        tokenType: json['token_type'] as String,
      ),
      user: UserModel.fromJson(json['user'] as Map<String, dynamic>),
    );
  }
}
