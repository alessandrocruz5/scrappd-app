import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../../../core/constants/theme_constants.dart';
import '../../providers/items_provider.dart';
import '../shell/main_shell.dart';

class CreateItemScreen extends StatefulWidget {
  const CreateItemScreen({super.key});

  @override
  State<CreateItemScreen> createState() => _CreateItemScreenState();
}

class _CreateItemScreenState extends State<CreateItemScreen> {
  final _formKey = GlobalKey<FormState>();
  final _itemNameController = TextEditingController();
  final _categoryController = TextEditingController();
  final _tagsController = TextEditingController();
  final ImagePicker _picker = ImagePicker();
  File? _selectedImage;
  String _format = 'png';

  @override
  void dispose() {
    _itemNameController.dispose();
    _categoryController.dispose();
    _tagsController.dispose();
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
    final XFile? image = await _picker.pickImage(
      source: source,
      imageQuality: 92,
      maxWidth: 4096,
      maxHeight: 4096,
    );

    if (image == null) return;
    setState(() {
      _selectedImage = File(image.path);
    });
  }

  List<String> _parseTags() {
    final raw = _tagsController.text.trim();
    if (raw.isEmpty) return [];
    return raw
        .split(',')
        .map((tag) => tag.trim())
        .where((tag) => tag.isNotEmpty)
        .toList();
  }

  Future<void> _submit() async {
    if (_selectedImage == null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Please select an image.')));
      return;
    }

    if (!_formKey.currentState!.validate()) return;

    final provider = context.read<ItemsProvider>();
    provider.resetUpload();

    await provider.createItem(
      imageFile: _selectedImage!,
      itemName: _itemNameController.text.trim().isEmpty
          ? null
          : _itemNameController.text.trim(),
      itemCategory: _categoryController.text.trim().isEmpty
          ? null
          : _categoryController.text.trim(),
      tags: _parseTags(),
      format: _format,
    );

    if (!mounted) return;

    if (provider.uploadState == UploadState.success) {
      final shell = MainShell.of(context);
      if (shell != null) {
        shell.setIndex(2);
      } else {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (_) => const MainShell(initialIndex: 2)),
        );
      }
    } else if (provider.uploadState == UploadState.error &&
        provider.errorMessage != null) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(provider.errorMessage!)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppTheme.spacing24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            height: 200,
            decoration: BoxDecoration(
              color: AppTheme.surfaceColor,
              borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
              border: Border.all(color: const Color(0xFFE5E7EB)),
            ),
            child: _selectedImage == null
                ? Center(
                    child: Text(
                      'Select an image to start',
                      style: TextStyle(
                        color: AppTheme.textSecondary.withValues(alpha: 0.8),
                      ),
                    ),
                  )
                : ClipRRect(
                    borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
                    child: Image.file(
                      _selectedImage!,
                      fit: BoxFit.cover,
                      width: double.infinity,
                    ),
                  ),
          ),
          const SizedBox(height: AppTheme.spacing16),
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _pickImage(ImageSource.camera),
                  icon: const Icon(Icons.camera_alt_outlined),
                  label: const Text('Camera'),
                ),
              ),
              const SizedBox(width: AppTheme.spacing12),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _pickImage(ImageSource.gallery),
                  icon: const Icon(Icons.photo_library_outlined),
                  label: const Text('Gallery'),
                ),
              ),
            ],
          ),
          const SizedBox(height: AppTheme.spacing24),
          Form(
            key: _formKey,
            child: Column(
              children: [
                TextFormField(
                  controller: _itemNameController,
                  decoration: const InputDecoration(
                    labelText: 'Item name (optional)',
                  ),
                ),
                const SizedBox(height: AppTheme.spacing16),
                TextFormField(
                  controller: _categoryController,
                  decoration: const InputDecoration(
                    labelText: 'Category (optional)',
                  ),
                ),
                const SizedBox(height: AppTheme.spacing16),
                TextFormField(
                  controller: _tagsController,
                  decoration: const InputDecoration(
                    labelText: 'Tags (comma separated)',
                  ),
                ),
                const SizedBox(height: AppTheme.spacing16),
                DropdownButtonFormField<String>(
                  initialValue: _format,
                  decoration: const InputDecoration(labelText: 'Output format'),
                  items: const [
                    DropdownMenuItem(value: 'png', child: Text('PNG')),
                    DropdownMenuItem(value: 'webp', child: Text('WEBP')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _format = value ?? 'png';
                    });
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: AppTheme.spacing24),
          Consumer<ItemsProvider>(
            builder: (context, provider, _) {
              final isUploading = provider.uploadState == UploadState.uploading;
              return SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: isUploading ? null : _submit,
                  child: isUploading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Upload & Process'),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}
