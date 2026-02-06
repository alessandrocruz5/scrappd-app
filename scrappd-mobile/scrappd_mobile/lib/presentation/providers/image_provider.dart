import 'dart:async';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:image_picker/image_picker.dart';
import 'package:path_provider/path_provider.dart';
import '../../core/utils/image_preprocessor.dart';
import '../../data/services/api_service.dart';

enum ProcessingState {
  idle,
  picking,
  uploading,
  processing,
  success,
  error,
}

class ImageProcessingProvider extends ChangeNotifier {
  final ApiService _apiService;
  final ImagePicker _imagePicker = ImagePicker();

  ProcessingState _state = ProcessingState.idle;
  File? _originalImage;
  File? _processedImage;
  Uint8List? _processedImageBytes; // Raw bytes for flexibility
  String? _errorMessage;
  double _progress = 0.0;
  DateTime? _startTime;
  Duration _elapsed = Duration.zero;
  int? _uploadBytes;
  Timer? _timer;

  ImageProcessingProvider(this._apiService);

  // Getters
  ProcessingState get state => _state;
  File? get originalImage => _originalImage;
  File? get processedImage => _processedImage;
  Uint8List? get processedImageBytes => _processedImageBytes;
  String? get errorMessage => _errorMessage;
  double get progress => _progress;
  Duration get elapsed => _elapsed;
  int? get uploadBytes => _uploadBytes;
  bool get isProcessing => _state == ProcessingState.processing || 
                           _state == ProcessingState.uploading;

  // Max image dimension for capture/pick (preprocessor enforces size limit)
  static final double _maxImageDimension = ImagePreprocessor.maxDimension.toDouble();

  /// Pick image from camera
  Future<void> pickFromCamera() async {
    try {
      _setState(ProcessingState.picking);

      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        imageQuality: 92,
        maxWidth: _maxImageDimension,
        maxHeight: _maxImageDimension,
      );

      if (image != null) {
        _originalImage = File(image.path);
        await _processImage();
      } else {
        _setState(ProcessingState.idle);
      }
    } catch (e) {
      _setError('Failed to capture image: $e');
    }
  }

  /// Pick image from gallery
  Future<void> pickFromGallery() async {
    try {
      _setState(ProcessingState.picking);

      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.gallery,
        imageQuality: 92,
        maxWidth: _maxImageDimension,
        maxHeight: _maxImageDimension,
      );

      if (image != null) {
        _originalImage = File(image.path);
        await _processImage();
      } else {
        _setState(ProcessingState.idle);
      }
    } catch (e) {
      _setError('Failed to pick image: $e');
    }
  }

  /// Process the selected image
  Future<void> _processImage() async {
    if (_originalImage == null) return;

    try {
      _setState(ProcessingState.uploading);
      _updateProgress(0.1);

      _startTimer();
      final uploadFile = await ImagePreprocessor.prepareForUpload(_originalImage!);
      _uploadBytes = await uploadFile.length();

      _setState(ProcessingState.processing);
      
      // Call API to remove background - returns Uint8List
      final Uint8List processedBytes = await _apiService.removeBackground(
        uploadFile,
        onProgress: (sent, total) {
          // Update progress during upload (0.1 to 0.4)
          if (total > 0) {
            _updateProgress(0.1 + (sent / total) * 0.3);
          }
        },
      );
      
      _updateProgress(0.8);
      
      // Store raw bytes
      _processedImageBytes = processedBytes;
      
      // Save bytes to a temporary file for compatibility
      _processedImage = await _saveBytesToTempFile(processedBytes);
      
      _updateProgress(1.0);
      _setState(ProcessingState.success);
      _stopTimer();

    } on ImageTooLargeException catch (e) {
      _setError(e.toString());
      _stopTimer();
    } on ApiException catch (e) {
      _setError(e.message);
      _stopTimer();
    } catch (e) {
      _setError('Processing failed: $e');
      _stopTimer();
    }
  }

  /// Save bytes to a temporary file
  Future<File> _saveBytesToTempFile(Uint8List bytes) async {
    final tempDir = await getTemporaryDirectory();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final file = File('${tempDir.path}/processed_$timestamp.png');
    await file.writeAsBytes(bytes);
    return file;
  }

  /// Retry processing
  Future<void> retry() async {
    if (_originalImage != null) {
      _clearError();
      await _processImage();
    }
  }

  /// Reset state
  void reset() {
    _originalImage = null;
    _processedImage = null;
    _processedImageBytes = null;
    _errorMessage = null;
    _progress = 0.0;
    _startTime = null;
    _elapsed = Duration.zero;
    _uploadBytes = null;
    _stopTimer();
    _setState(ProcessingState.idle);
  }

  /// Clear error
  void _clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  /// Set error state
  void _setError(String message) {
    _errorMessage = message;
    _setState(ProcessingState.error);
  }

  /// Update processing state
  void _setState(ProcessingState newState) {
    _state = newState;
    notifyListeners();
  }

  /// Update progress
  void _updateProgress(double value) {
    _progress = value;
    notifyListeners();
  }

  void _startTimer() {
    _startTime = DateTime.now();
    _elapsed = Duration.zero;
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (_startTime == null) return;
      _elapsed = DateTime.now().difference(_startTime!);
      notifyListeners();
    });
  }

  void _stopTimer() {
    _timer?.cancel();
    _timer = null;
  }

  @override
  void dispose() {
    // Clean up temp files if needed
    _stopTimer();
    super.dispose();
  }
}
