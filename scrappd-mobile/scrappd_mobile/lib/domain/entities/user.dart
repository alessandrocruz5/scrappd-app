class User {
  final String id;
  final String email;
  final String username;
  final String? displayName;
  final String? profileImageUrl;
  final String subscriptionTier;

  User({
    required this.id,
    required this.email,
    required this.username,
    this.displayName,
    this.profileImageUrl,
    required this.subscriptionTier,
  });
}
