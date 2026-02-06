import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/constants/theme_constants.dart';
import '../../../data/datasources/auth_remote_datasource.dart';

class EmailVerificationPendingScreen extends StatefulWidget {
  const EmailVerificationPendingScreen({
    this.email,
    super.key,
  });

  final String? email;

  @override
  State<EmailVerificationPendingScreen> createState() =>
      _EmailVerificationPendingScreenState();
}

class _EmailVerificationPendingScreenState
    extends State<EmailVerificationPendingScreen> {
  bool _isLoading = false;
  String? _message;
  String? _error;

  Future<void> _resendVerification() async {
    final email = widget.email;
    if (email == null || email.trim().isEmpty) {
      setState(() {
        _error = 'No email available to resend verification.';
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _message = null;
      _error = null;
    });

    try {
      await context.read<AuthRemoteDataSource>().resendVerification(email: email);
      if (!mounted) return;
      setState(() {
        _message = 'Verification email sent again.';
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = _friendlyError(e);
      });
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  String _friendlyError(Object error) {
    if (error is DioException && error.response?.statusCode == 404) {
      return 'Email verification is not enabled on the backend yet.';
    }
    return error.toString();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Verify your email')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppTheme.spacing24),
          child: Column(
            children: [
              const Icon(
                Icons.mark_email_unread_outlined,
                size: 64,
                color: AppTheme.primaryColor,
              ),
              const SizedBox(height: AppTheme.spacing16),
              const Text(
                'Check your inbox',
                style: TextStyle(
                  fontSize: 26,
                  fontWeight: FontWeight.bold,
                  color: AppTheme.textPrimary,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppTheme.spacing8),
              Text(
                widget.email == null
                    ? 'We sent a verification link to your email.'
                    : 'We sent a verification link to ${widget.email}.',
                style: const TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 15,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppTheme.spacing24),
              if (_error != null)
                Text(
                  _error!,
                  style: const TextStyle(color: AppTheme.errorColor),
                  textAlign: TextAlign.center,
                ),
              if (_message != null)
                Text(
                  _message!,
                  style: const TextStyle(color: AppTheme.successColor),
                  textAlign: TextAlign.center,
                ),
              const Spacer(),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _resendVerification,
                  child: _isLoading
                      ? const SizedBox(
                          height: 18,
                          width: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : const Text('Resend verification email'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
