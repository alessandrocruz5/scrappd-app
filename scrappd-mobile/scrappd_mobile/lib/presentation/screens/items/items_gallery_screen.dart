import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/constants/theme_constants.dart';
import '../../providers/items_provider.dart';
import 'item_detail_screen.dart';

class ItemsGalleryScreen extends StatefulWidget {
  const ItemsGalleryScreen({super.key});

  @override
  State<ItemsGalleryScreen> createState() => _ItemsGalleryScreenState();
}

class _ItemsGalleryScreenState extends State<ItemsGalleryScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => context.read<ItemsProvider>().loadItems());
  }

  Future<bool> _confirmAction({
    required String title,
    required String message,
    required String confirmLabel,
  }) async {
    final result = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(confirmLabel),
          ),
        ],
      ),
    );
    return result ?? false;
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<ItemsProvider>(
      builder: (context, provider, child) {
        if (provider.isLoading && provider.items.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }

        if (provider.items.isEmpty) {
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(AppTheme.spacing24),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.photo_album_outlined, size: 64),
                  const SizedBox(height: AppTheme.spacing16),
                  const Text('No items yet'),
                  const SizedBox(height: AppTheme.spacing8),
                  Text(
                    'Upload your first photo to start your gallery.',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: AppTheme.textSecondary.withValues(alpha: 0.9),
                    ),
                  ),
                ],
              ),
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () => provider.loadItems(refresh: true),
          child: Column(
            children: [
              if (provider.errorMessage != null)
                Padding(
                  padding: const EdgeInsets.all(AppTheme.spacing16),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.error_outline,
                        color: AppTheme.errorColor,
                      ),
                      const SizedBox(width: AppTheme.spacing8),
                      Expanded(
                        child: Text(
                          provider.errorMessage!,
                          style: const TextStyle(color: AppTheme.errorColor),
                        ),
                      ),
                    ],
                  ),
                ),
              Expanded(
                child: GridView.builder(
                  padding: const EdgeInsets.all(AppTheme.spacing16),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    mainAxisSpacing: AppTheme.spacing16,
                    crossAxisSpacing: AppTheme.spacing16,
                    childAspectRatio: 0.9,
                  ),
                  itemCount: provider.items.length,
                  itemBuilder: (context, index) {
                    final item = provider.items[index];
                    final imageUrl =
                        item.processedImageUrl ?? item.originalImageUrl;
                    final isProcessing = item.processingStatus != 'completed';
                    final isFailed = item.processingStatus == 'failed';
                    return GestureDetector(
                      onTap: isProcessing
                          ? null
                          : () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => ItemDetailScreen(item: item),
                                ),
                              );
                            },
                      child: Container(
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(
                            AppTheme.radiusLarge,
                          ),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.06),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: Column(
                          children: [
                            Expanded(
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(
                                  AppTheme.radiusLarge,
                                ),
                                child: Stack(
                                  fit: StackFit.expand,
                                  children: [
                                    ColorFiltered(
                                      colorFilter: isProcessing
                                          ? const ColorFilter.mode(
                                              Colors.grey,
                                              BlendMode.saturation,
                                            )
                                          : const ColorFilter.mode(
                                              Colors.transparent,
                                              BlendMode.multiply,
                                            ),
                                      child: CachedNetworkImage(
                                        imageUrl: imageUrl,
                                        fit: BoxFit.cover,
                                        width: double.infinity,
                                        placeholder: (context, url) =>
                                            const Center(
                                              child: CircularProgressIndicator(
                                                strokeWidth: 2,
                                              ),
                                            ),
                                        errorWidget: (context, url, error) =>
                                            const Center(
                                              child: Icon(
                                                Icons.broken_image_outlined,
                                              ),
                                            ),
                                      ),
                                    ),
                                    if (isProcessing)
                                      Positioned.fill(
                                        child: Container(
                                          color: Colors.black.withValues(
                                            alpha: 0.25,
                                          ),
                                          child: Align(
                                            alignment: Alignment.bottomCenter,
                                            child: Container(
                                              height: 36,
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 12,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: isFailed
                                                    ? AppTheme.errorColor
                                                          .withValues(
                                                            alpha: 0.9,
                                                          )
                                                    : Colors.black.withValues(
                                                        alpha: 0.75,
                                                      ),
                                              ),
                                              child: Row(
                                                children: [
                                                  Expanded(
                                                    child: Text(
                                                      isFailed
                                                          ? 'Processing failed'
                                                          : 'Processing...',
                                                      style: const TextStyle(
                                                        color: Colors.white,
                                                        fontSize: 12,
                                                        fontWeight:
                                                            FontWeight.w600,
                                                      ),
                                                      maxLines: 1,
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                    ),
                                                  ),
                                                  Row(
                                                    children: [
                                                      if (isFailed)
                                                        TextButton(
                                                          onPressed: () =>
                                                              provider
                                                                  .retryItem(
                                                                    item,
                                                                  ),
                                                          style: TextButton.styleFrom(
                                                            foregroundColor:
                                                                Colors.white,
                                                            padding:
                                                                const EdgeInsets.symmetric(
                                                                  horizontal: 8,
                                                                ),
                                                            minimumSize:
                                                                const Size(
                                                                  0,
                                                                  28,
                                                                ),
                                                          ),
                                                          child: const Text(
                                                            'Retry',
                                                            style: TextStyle(
                                                              fontSize: 12,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .w600,
                                                            ),
                                                          ),
                                                        )
                                                      else
                                                        TextButton(
                                                          onPressed: () async {
                                                            final ok = await _confirmAction(
                                                              title:
                                                                  'Cancel processing?',
                                                              message:
                                                                  'This stops background removal and keeps the original.',
                                                              confirmLabel:
                                                                  'Cancel',
                                                            );
                                                            if (ok) {
                                                              await provider
                                                                  .cancelItem(
                                                                    item.id,
                                                                  );
                                                            }
                                                          },
                                                          style: TextButton.styleFrom(
                                                            foregroundColor:
                                                                Colors.white,
                                                            padding:
                                                                const EdgeInsets.symmetric(
                                                                  horizontal: 6,
                                                                ),
                                                            minimumSize:
                                                                const Size(
                                                                  0,
                                                                  28,
                                                                ),
                                                          ),
                                                          child: const Text(
                                                            'Cancel',
                                                            style: TextStyle(
                                                              fontSize: 12,
                                                              fontWeight:
                                                                  FontWeight
                                                                      .w600,
                                                            ),
                                                          ),
                                                        ),
                                                      const SizedBox(width: 6),
                                                      TextButton(
                                                        onPressed: () async {
                                                          final ok = await _confirmAction(
                                                            title:
                                                                'Delete item?',
                                                            message:
                                                                'This will remove the item from your gallery.',
                                                            confirmLabel:
                                                                'Delete',
                                                          );
                                                          if (ok) {
                                                            await provider
                                                                .deleteItem(
                                                                  item.id,
                                                                );
                                                          }
                                                        },
                                                        style: TextButton.styleFrom(
                                                          foregroundColor:
                                                              Colors.white,
                                                          padding:
                                                              const EdgeInsets.symmetric(
                                                                horizontal: 6,
                                                              ),
                                                          minimumSize:
                                                              const Size(0, 28),
                                                        ),
                                                        child: const Text(
                                                          'Delete',
                                                          style: TextStyle(
                                                            fontSize: 12,
                                                            fontWeight:
                                                                FontWeight.w600,
                                                          ),
                                                        ),
                                                      ),
                                                      if (!isFailed) ...[
                                                        const SizedBox(
                                                          width: 6,
                                                        ),
                                                        SizedBox(
                                                          width: 50,
                                                          child: LinearProgressIndicator(
                                                            color: Colors.white,
                                                            backgroundColor:
                                                                Colors.white
                                                                    .withValues(
                                                                      alpha:
                                                                          0.3,
                                                                    ),
                                                            minHeight: 4,
                                                          ),
                                                        ),
                                                      ],
                                                    ],
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ),
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                            ),
                            Padding(
                              padding: const EdgeInsets.all(AppTheme.spacing12),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    item.itemName ?? 'Untitled',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                  const SizedBox(height: AppTheme.spacing4),
                                  Text(
                                    item.itemCategory ?? 'Uncategorized',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: TextStyle(
                                      fontSize: 12,
                                      color: AppTheme.textSecondary.withValues(
                                        alpha: 0.9,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
