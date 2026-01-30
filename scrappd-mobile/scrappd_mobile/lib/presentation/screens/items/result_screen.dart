import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/constants/theme_constants.dart';
import '../../../domain/entities/item.dart';
import '../../providers/items_provider.dart';

class ResultScreen extends StatefulWidget {
  const ResultScreen({super.key, required this.item});

  final Item item;

  @override
  State<ResultScreen> createState() => _ResultScreenState();
}

class _ResultScreenState extends State<ResultScreen> {
  bool _showProcessed = true;

  @override
  Widget build(BuildContext context) {
    final imageUrl = _showProcessed && widget.item.processedImageUrl != null
        ? widget.item.processedImageUrl!
        : widget.item.originalImageUrl;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Result'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () {
            context.read<ItemsProvider>().resetUpload();
            Navigator.pop(context);
          },
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(AppTheme.spacing16),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
                child: Container(
                  color: Colors.white,
                  child: CachedNetworkImage(
                    imageUrl: imageUrl,
                    fit: BoxFit.contain,
                    placeholder: (context, url) => const Center(
                      child: CircularProgressIndicator(),
                    ),
                    errorWidget: (context, url, error) => const Center(
                      child: Icon(Icons.broken_image_outlined, size: 48),
                    ),
                  ),
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppTheme.spacing24),
            child: Row(
              children: [
                Expanded(
                  child: ChoiceChip(
                    label: const Text('Processed'),
                    selected: _showProcessed,
                    onSelected: widget.item.processedImageUrl == null
                        ? null
                        : (value) {
                            setState(() {
                              _showProcessed = true;
                            });
                          },
                  ),
                ),
                const SizedBox(width: AppTheme.spacing12),
                Expanded(
                  child: ChoiceChip(
                    label: const Text('Original'),
                    selected: !_showProcessed,
                    onSelected: (value) {
                      setState(() {
                        _showProcessed = false;
                      });
                    },
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: AppTheme.spacing16),
          Padding(
            padding: const EdgeInsets.all(AppTheme.spacing24),
            child: Column(
              children: [
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () {
                      context.read<ItemsProvider>().resetUpload();
                      Navigator.pop(context);
                    },
                    child: const Text('Create another'),
                  ),
                ),
                const SizedBox(height: AppTheme.spacing12),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    onPressed: () {
                      context.read<ItemsProvider>().resetUpload();
                      Navigator.popUntil(context, (route) => route.isFirst);
                    },
                    child: const Text('Back to gallery'),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
