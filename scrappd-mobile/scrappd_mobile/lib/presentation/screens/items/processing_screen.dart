import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/constants/theme_constants.dart';
import '../../providers/items_provider.dart';
import 'result_screen.dart';

class ProcessingScreen extends StatelessWidget {
  const ProcessingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Consumer<ItemsProvider>(
          builder: (context, provider, child) {
            if (provider.uploadState == UploadState.success &&
                provider.latestItem != null) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                Navigator.pushReplacement(
                  context,
                  MaterialPageRoute(
                    builder: (_) => ResultScreen(item: provider.latestItem!),
                  ),
                );
              });
            }

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
                      const CircularProgressIndicator(
                        valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                      const SizedBox(height: AppTheme.spacing24),
                      Text(
                        provider.uploadState == UploadState.uploading
                            ? 'Processing your item...'
                            : 'Preparing your upload...',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: AppTheme.spacing16),
                      Text(
                        'Hang tight, this usually takes a few seconds.',
                        style: TextStyle(
                          color: Colors.white.withValues(alpha: 0.85),
                        ),
                        textAlign: TextAlign.center,
                      ),
                      if (provider.uploadState == UploadState.error)
                        Padding(
                          padding: const EdgeInsets.only(
                            top: AppTheme.spacing24,
                          ),
                          child: Column(
                            children: [
                              Text(
                                provider.errorMessage ??
                                    'Upload failed. Please try again.',
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontWeight: FontWeight.w600,
                                ),
                                textAlign: TextAlign.center,
                              ),
                              const SizedBox(height: AppTheme.spacing16),
                              ElevatedButton(
                                onPressed: () {
                                  provider.resetUpload();
                                  Navigator.pop(context);
                                },
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: Colors.white,
                                  foregroundColor: AppTheme.primaryColor,
                                ),
                                child: const Text('Back'),
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
}
