import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/constants/theme_constants.dart';
import '../providers/image_provider.dart';

class ProcessingScreen extends StatelessWidget {
  const ProcessingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Consumer<ImageProcessingProvider>(
          builder: (context, provider, child) {
            return Container(
              decoration: const BoxDecoration(
                gradient: AppTheme.primaryGradient,
              ),
              child: Center(
                child: Padding(
                  padding: const EdgeInsets.all(AppTheme.spacing32),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // Processing Animation
                      Stack(
                        alignment: Alignment.center,
                        children: [
                          // Outer ring
                          SizedBox(
                            width: 120,
                            height: 120,
                            child: CircularProgressIndicator(
                              value: provider.progress,
                              strokeWidth: 8,
                              backgroundColor: Colors.white.withValues(alpha: 0.2),
                              valueColor: const AlwaysStoppedAnimation<Color>(
                                Colors.white,
                              ),
                            ),
                          ),
                          // Inner icon
                          Container(
                            width: 80,
                            height: 80,
                            decoration: BoxDecoration(
                              color: Colors.white.withValues(alpha: 0.2),
                              shape: BoxShape.circle,
                            ),
                            child: const Icon(
                              Icons.auto_fix_high,
                              size: 40,
                              color: Colors.white,
                            ),
                          ),
                        ],
                      ),

                      const SizedBox(height: AppTheme.spacing32),

                      // Status Text
                      Text(
                        _getStatusText(provider.state),
                        style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                        textAlign: TextAlign.center,
                      ),

                      const SizedBox(height: AppTheme.spacing12),

                      // Subtitle
                      Text(
                        'This may take 10-15 seconds',
                        style: TextStyle(
                          fontSize: 16,
                          color: Colors.white.withValues(alpha: 0.8),
                        ),
                        textAlign: TextAlign.center,
                      ),

                      const SizedBox(height: AppTheme.spacing12),

                      // Diagnostics
                      Column(
                        children: [
                          Text(
                            'Elapsed: ${_formatDuration(provider.elapsed)}',
                            style: TextStyle(
                              fontSize: 14,
                              color: Colors.white.withValues(alpha: 0.8),
                            ),
                            textAlign: TextAlign.center,
                          ),
                          if (provider.uploadBytes != null)
                            Text(
                              'Upload size: ${_formatBytes(provider.uploadBytes!)}',
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.white.withValues(alpha: 0.8),
                              ),
                              textAlign: TextAlign.center,
                            ),
                        ],
                      ),

                      const SizedBox(height: AppTheme.spacing32),

                      // Progress percentage
                      Text(
                        '${(provider.progress * 100).toInt()}%',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: Colors.white.withValues(alpha: 0.9),
                        ),
                      ),

                      const SizedBox(height: AppTheme.spacing48),

                      // Fun fact or tip
                      Container(
                        padding: const EdgeInsets.all(AppTheme.spacing16),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(
                            AppTheme.radiusMedium,
                          ),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              Icons.lightbulb_outline,
                              color: Colors.white.withValues(alpha: 0.9),
                              size: 24,
                            ),
                            const SizedBox(width: AppTheme.spacing12),
                            Expanded(
                              child: Text(
                                'Pro tip: Well-lit photos give better results',
                                style: TextStyle(
                                  fontSize: 14,
                                  color: Colors.white.withValues(alpha: 0.9),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  String _getStatusText(ProcessingState state) {
    switch (state) {
      case ProcessingState.uploading:
        return 'Uploading image...';
      case ProcessingState.processing:
        return 'Removing background...';
      default:
        return 'Processing...';
    }
  }

  String _formatDuration(Duration duration) {
    final seconds = duration.inSeconds;
    if (seconds < 60) return '${seconds}s';
    final minutes = seconds ~/ 60;
    final remainder = seconds % 60;
    return '${minutes}m ${remainder}s';
  }

  String _formatBytes(int bytes) {
    final mb = bytes / (1024 * 1024);
    if (mb >= 1) return '${mb.toStringAsFixed(1)}MB';
    final kb = bytes / 1024;
    return '${kb.toStringAsFixed(0)}KB';
  }
}
