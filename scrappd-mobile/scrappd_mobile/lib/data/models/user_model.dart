import '../../domain/entities/user.dart';

class UserModel extends User {
  UserModel({
    required super.id,
    required super.email,
    required super.username,
    super.displayName,
    super.profileImageUrl,
    required super.subscriptionTier,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] ?? '',
      email: json['email'] ?? '',
      username: json['username'] ?? '',
      displayName: json['display_name'],
      profileImageUrl: json['profile_image_url'],
      subscriptionTier: json['subscription_tier'] ?? 'free',
    );
  }
}
