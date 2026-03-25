import 'package:flutter/foundation.dart';
import '../../../shared/models/media_asset_model.dart';
import '../data/media_repository.dart';

class MediaProvider extends ChangeNotifier {
  final MediaRepository _mediaRepository;

  List<MediaAssetModel> _assets = [];
  bool _isLoading = false;
  String? _error;
  int _currentPage = 1;
  int _totalPages = 1;
  int _total = 0;
  String? _filterMediaType;
  String? _filterSyncStatus;
  String _sort = 'created_at_desc';

  MediaProvider(this._mediaRepository);

  List<MediaAssetModel> get assets => _assets;
  bool get isLoading => _isLoading;
  String? get error => _error;
  int get currentPage => _currentPage;
  int get totalPages => _totalPages;
  int get total => _total;
  bool get hasMore => _currentPage < _totalPages;

  Future<void> loadAssets({bool refresh = false}) async {
    if (refresh) _currentPage = 1;
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final result = await _mediaRepository.listAssets(
        page: _currentPage,
        mediaType: _filterMediaType,
        syncStatus: _filterSyncStatus,
        sort: _sort,
      );
      if (refresh || _currentPage == 1) {
        _assets = result.assets;
      } else {
        _assets.addAll(result.assets);
      }
      _totalPages = result.totalPages;
      _total = result.total;
    } catch (e) {
      _error = 'Failed to load media: $e';
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> loadMore() async {
    if (!hasMore || _isLoading) return;
    _currentPage++;
    await loadAssets();
  }

  Future<void> deleteAsset(String assetId) async {
    try {
      await _mediaRepository.deleteAsset(assetId);
      _assets.removeWhere((a) => a.assetId == assetId);
      _total--;
      notifyListeners();
    } catch (e) {
      _error = 'Failed to delete: $e';
      notifyListeners();
    }
  }

  void setFilter({String? mediaType, String? syncStatus}) {
    _filterMediaType = mediaType;
    _filterSyncStatus = syncStatus;
    loadAssets(refresh: true);
  }

  void setSort(String sort) {
    _sort = sort;
    loadAssets(refresh: true);
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }
}
