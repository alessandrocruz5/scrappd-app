class User {
  final String id;
  final String email;
  final String username;
  final String? displayName;
  final String? profileImageUrl;
  final String? bio;
  final String subscriptionTier;
  final String subscriptionStatus;
  final DateTime? subscriptionExpiresAt;
  final int monthlyBgRemovalsUsed;
  final int monthlyBgRemovalsLimit;
  final double monthlyStorageUsedMb;
  final int monthlyStorageLimitMb;
  final int followerCount;
  final int followingCount;
  final bool isVerified;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? lastLoginAt;

  User({
    required this.id,
    required this.email,
    required this.username,
    this.displayName,
    this.profileImageUrl,
    this.bio,
    required this.subscriptionTier,
    required this.subscriptionStatus,
    this.subscriptionExpiresAt,
    required this.monthlyBgRemovalsUsed,
    required this.monthlyBgRemovalsLimit,
    required this.monthlyStorageUsedMb,
    required this.monthlyStorageLimitMb,
    required this.followerCount,
    required this.followingCount,
    required this.isVerified,
    required this.createdAt,
    required this.updatedAt,
    this.lastLoginAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      email: json['email'] ?? '',
      username: json['username'] ?? '',
      displayName: json['display_name'],
      profileImageUrl: json['profile_image_url'],
      bio: json['bio'],
      subscriptionTier: json['subscription_tier'] ?? 'free',
      subscriptionStatus: json['subscription_status'] ?? 'inactive',
      subscriptionExpiresAt: json['subscription_expires_at'] != null
          ? DateTime.parse(json['subscription_expires_at'])
          : null,
      monthlyBgRemovalsUsed: json['monthly_bg_removals_used'] ?? 0,
      monthlyBgRemovalsLimit: json['monthly_bg_removals_limit'] ?? 5,
      monthlyStorageUsedMb:
          (json['monthly_storage_used_mb'] ?? 0.0).toDouble(),
      monthlyStorageLimitMb: json['monthly_storage_limit_mb'] ?? 100,
      followerCount: json['follower_count'] ?? 0,
      followingCount: json['following_count'] ?? 0,
      isVerified: json['is_verified'] ?? false,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
      lastLoginAt: json['last_login_at'] != null
          ? DateTime.parse(json['last_login_at'])
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'email': email,
      'username': username,
      'display_name': displayName,
      'profile_image_url': profileImageUrl,
      'bio': bio,
      'subscription_tier': subscriptionTier,
      'subscription_status': subscriptionStatus,
      'subscription_expires_at': subscriptionExpiresAt?.toIso8601String(),
      'monthly_bg_removals_used': monthlyBgRemovalsUsed,
      'monthly_bg_removals_limit': monthlyBgRemovalsLimit,
      'monthly_storage_used_mb': monthlyStorageUsedMb,
      'monthly_storage_limit_mb': monthlyStorageLimitMb,
      'follower_count': followerCount,
      'following_count': followingCount,
      'is_verified': isVerified,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'last_login_at': lastLoginAt?.toIso8601String(),
    };
  }
}

class LoginRequest {
  final String email;
  final String password;

  LoginRequest({
    required this.email,
    required this.password,
  });

  Map<String, dynamic> toJson() {
    return {
      'email': email,
      'password': password,
    };
  }
}

class RegisterRequest {
  final String email;
  final String username;
  final String password;
  final String? displayName;

  RegisterRequest({
    required this.email,
    required this.username,
    required this.password,
    this.displayName,
  });

  Map<String, dynamic> toJson() {
    return {
      'email': email,
      'username': username,
      'password': password,
      if (displayName != null) 'display_name': displayName,
    };
  }
}

class AuthResponse {
  final User user;
  final String accessToken;
  final String refreshToken;

  AuthResponse({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      user: User.fromJson(json['user']),
      accessToken: json['access_token'] ?? '',
      refreshToken: json['refresh_token'] ?? '',
    );
  }
}
