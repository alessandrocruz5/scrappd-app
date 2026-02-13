import 'dart:io';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:gal/gal.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import '../../core/constants/theme_constants.dart';
import '../providers/image_provider.dart';
import '../widgets/custom_button.dart';

class ResultScreen extends StatefulWidget {
  const ResultScreen({super.key});

  @override
  State<ResultScreen> createState() => _ResultScreenState();
}

class _ResultScreenState extends State<ResultScreen> {
  bool _showComparison = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text('Result'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () {
            final provider = context.read<ImageProcessingProvider>();
            provider.reset();
            Navigator.of(context).popUntil((route) => route.isFirst);
          },
        ),
      ),
      body: Consumer<ImageProcessingProvider>(
        builder: (context, provider, child) {
          if (provider.processedImage == null) {
            return const Center(
              child: Text('No processed image available'),
            );
          }

          return Column(
            children: [
              // Image Display
              Expanded(
                child: Container(
                  margin: const EdgeInsets.all(AppTheme.spacing16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.1),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
                    child: Stack(
                      children: [
                        // Checkerboard background
                        _buildCheckerboard(),
                        
                        // Image
                        Center(
                          child: _showComparison
                              ? _buildComparisonView(provider)
                              : Image.file(
                                  provider.processedImage!,
                                  fit: BoxFit.contain,
                                ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),

              // Controls
              Container(
                padding: const EdgeInsets.all(AppTheme.spacing24),
                decoration: BoxDecoration(
                  color: Colors.white,
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.05),
                      blurRadius: 10,
                      offset: const Offset(0, -4),
                    ),
                  ],
                ),
                child: Column(
                  children: [
                    // Compare Toggle
                    if (provider.originalImage != null)
                      SwitchListTile(
                        title: const Text('Compare with original'),
                        value: _showComparison,
                        onChanged: (value) {
                          setState(() {
                            _showComparison = value;
                          });
                        },
                        activeThumbColor: AppTheme.primaryColor,
                      ),

                    const SizedBox(height: AppTheme.spacing16),

                    // Action Buttons
                    Row(
                      children: [
                        Expanded(
                          child: CustomButton(
                            text: 'Save',
                            icon: Icons.download,
                            onPressed: () => _saveImage(
                              context,
                              provider.processedImage!,
                            ),
                            isPrimary: true,
                          ),
                        ),
                        const SizedBox(width: AppTheme.spacing12),
                        Expanded(
                          child: CustomButton(
                            text: 'Share',
                            icon: Icons.share,
                            onPressed: () => _shareImage(
                              provider.processedImage!,
                            ),
                            isPrimary: false,
                          ),
                        ),
                      ],
                    ),

                    const SizedBox(height: AppTheme.spacing12),

                    // New Photo Button
                    CustomButton(
                      text: 'Process Another Photo',
                      icon: Icons.add_photo_alternate,
                      onPressed: () {
                        provider.reset();
                        Navigator.of(context).popUntil((route) => route.isFirst);
                      },
                      isPrimary: false,
                    ),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildCheckerboard() {
    return Container(
      decoration: BoxDecoration(
        image: DecorationImage(
          image: const AssetImage('assets/images/checkerboard.png'),
          repeat: ImageRepeat.repeat,
          colorFilter: ColorFilter.mode(
            Colors.grey.withValues(alpha: 0.1),
            BlendMode.srcOver,
          ),
        ),
      ),
    );
  }

  Widget _buildComparisonView(ImageProcessingProvider provider) {
    return Row(
      children: [
        Expanded(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text(
                'Original',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: AppTheme.textSecondary,
                ),
              ),
              const SizedBox(height: AppTheme.spacing8),
              Expanded(
                child: Image.file(
                  provider.originalImage!,
                  fit: BoxFit.contain,
                ),
              ),
            ],
          ),
        ),
        Container(
          width: 2,
          color: AppTheme.primaryColor,
        ),
        Expanded(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text(
                'Processed',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: AppTheme.primaryColor,
                ),
              ),
              const SizedBox(height: AppTheme.spacing8),
              Expanded(
                child: Image.file(
                  provider.processedImage!,
                  fit: BoxFit.contain,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _saveImage(BuildContext context, File imageFile) async {
    try {
      if (kIsWeb || !(Platform.isAndroid || Platform.isIOS)) {
        throw Exception(
          'Saving to gallery is only supported on Android and iOS.',
        );
      }
      await Gal.putImage(imageFile.path);
  
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Image saved to gallery!'),
            backgroundColor: AppTheme.successColor,
          ),
        );
      }
    } catch (e) {
      if (context.mounted) {
        final message = e is MissingPluginException
            ? 'Save failed. Please fully restart the app after adding plugins.'
            : 'Failed to save: $e';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(message),
            backgroundColor: AppTheme.errorColor,
          ),
        );
      }
    }
  }

  Future<void> _shareImage(File imageFile) async {
    // TODO: Implement share functionality
    // You'll need to add share_plus package
    debugPrint('Share image: ${imageFile.path}');
  }
}
