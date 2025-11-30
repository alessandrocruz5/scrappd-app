import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/constants/theme_constants.dart';
import '../providers/image_provider.dart';
import '../widgets/custom_button.dart';
import 'processing_screen.dart';
import 'result_screen.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Consumer<ImageProcessingProvider>(
          builder: (context, provider, child) {
            // Navigate based on state
            if (provider.state == ProcessingState.processing ||
                provider.state == ProcessingState.uploading) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const ProcessingScreen(),
                  ),
                );
              });
            } else if (provider.state == ProcessingState.success) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const ResultScreen(),
                  ),
                );
              });
            }

            return _buildHomeContent(context, provider);
          },
        ),
      ),
    );
  }

  Widget _buildHomeContent(
    BuildContext context,
    ImageProcessingProvider provider,
  ) {
    return Container(
      decoration: const BoxDecoration(
        gradient: AppTheme.primaryGradient,
      ),
      child: Column(
        children: [
          // Header
          Padding(
            padding: const EdgeInsets.all(AppTheme.spacing24),
            child: Column(
              children: [
                const SizedBox(height: AppTheme.spacing32),
                const Text(
                  "Scrapp'd",
                  style: TextStyle(
                    fontSize: 48,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: -1,
                  ),
                ),
                const SizedBox(height: AppTheme.spacing8),
                Text(
                  'Remove backgrounds with AI',
                  style: TextStyle(
                    fontSize: 18,
                    color: Colors.white.withValues(alpha: 0.9),
                  ),
                ),
              ],
            ),
          ),

          const Spacer(),

          // Main Content
          Padding(
            padding: const EdgeInsets.all(AppTheme.spacing24),
            child: Column(
              children: [
                // Camera Icon
                Container(
                  width: 120,
                  height: 120,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.2),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.camera_alt_rounded,
                    size: 60,
                    color: Colors.white,
                  ),
                ),

                const SizedBox(height: AppTheme.spacing32),

                // Take Photo Button
                CustomButton(
                  text: 'Take Photo',
                  icon: Icons.camera_alt,
                  onPressed: provider.isProcessing
                      ? null
                      : () => provider.pickFromCamera(),
                  isPrimary: false,
                ),

                const SizedBox(height: AppTheme.spacing16),

                // Choose from Gallery Button
                CustomButton(
                  text: 'Choose from Gallery',
                  icon: Icons.photo_library,
                  onPressed: provider.isProcessing
                      ? null
                      : () => provider.pickFromGallery(),
                  isPrimary: false,
                ),

                const SizedBox(height: AppTheme.spacing32),

                // Error Message
                if (provider.errorMessage != null)
                  Container(
                    padding: const EdgeInsets.all(AppTheme.spacing16),
                    decoration: BoxDecoration(
                      color: AppTheme.errorColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
                    ),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.error_outline,
                          color: AppTheme.errorColor,
                        ),
                        const SizedBox(width: AppTheme.spacing12),
                        Expanded(
                          child: Text(
                            provider.errorMessage!,
                            style: const TextStyle(
                              color: AppTheme.errorColor,
                              fontSize: 14,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),

          const Spacer(),

          // Footer
          Padding(
            padding: const EdgeInsets.all(AppTheme.spacing24),
            child: Text(
              'Powered by BiRefNet AI',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.7),
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }
}