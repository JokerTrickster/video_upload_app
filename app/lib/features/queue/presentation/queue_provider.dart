import 'package:flutter/foundation.dart';
import '../../../shared/models/queue_model.dart';
import '../data/queue_repository.dart';

class QueueProvider extends ChangeNotifier {
  final QueueRepository _queueRepository;

  List<QueueItemModel> _items = [];
  QuotaModel? _quota;
  bool _isLoading = false;
  String? _error;

  QueueProvider(this._queueRepository);

  List<QueueItemModel> get items => _items;
  QuotaModel? get quota => _quota;
  bool get isLoading => _isLoading;
  String? get error => _error;

  int get pendingCount => _items.where((i) => i.isPending).length;
  int get completedCount => _items.where((i) => i.isCompleted).length;
  int get failedCount => _items.where((i) => i.isFailed).length;

  Future<void> loadQueue() async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final results = await Future.wait([
        _queueRepository.getQueue(),
        _queueRepository.getQuota(),
      ]);
      _items = results[0] as List<QueueItemModel>;
      _quota = results[1] as QuotaModel;
    } catch (e) {
      _error = 'Failed to load queue: $e';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> addToQueue({
    required String filePath,
    required String filename,
    required int fileSizeBytes,
    String? title,
  }) async {
    try {
      final item = await _queueRepository.addToQueue(
        filePath: filePath,
        filename: filename,
        fileSizeBytes: fileSizeBytes,
        title: title,
      );
      _items.insert(0, item);
      notifyListeners();
    } catch (e) {
      _error = 'Failed to add to queue: $e';
      notifyListeners();
      rethrow;
    }
  }

  Future<void> removeItem(String queueId) async {
    try {
      await _queueRepository.removeFromQueue(queueId);
      _items.removeWhere((i) => i.queueId == queueId);
      notifyListeners();
    } catch (e) {
      _error = 'Failed to remove: $e';
      notifyListeners();
    }
  }

  Future<void> refreshQuota() async {
    try {
      _quota = await _queueRepository.getQuota();
      notifyListeners();
    } catch (_) {}
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
