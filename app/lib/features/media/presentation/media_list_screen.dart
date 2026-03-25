import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../../core/utils/responsive.dart';
import '../../../features/auth/presentation/auth_provider.dart';
import '../../../shared/models/media_asset_model.dart';
import 'media_provider.dart';

class MediaListScreen extends StatefulWidget {
  const MediaListScreen({super.key});

  @override
  State<MediaListScreen> createState() => _MediaListScreenState();
}

class _MediaListScreenState extends State<MediaListScreen> {
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<MediaProvider>().loadAssets(refresh: true);
    });
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 200) {
      context.read<MediaProvider>().loadMore();
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Scaffold(
      appBar: AppBar(
        title: Text('My Media', style: TextStyle(fontSize: r.titleMedium)),
        actions: [
          PopupMenuButton<String>(
            icon: Icon(Icons.filter_list, size: r.iconMedium),
            onSelected: (value) {
              final provider = context.read<MediaProvider>();
              switch (value) {
                case 'all':
                  provider.setFilter();
                case 'completed':
                  provider.setFilter(syncStatus: 'COMPLETED');
                case 'failed':
                  provider.setFilter(syncStatus: 'FAILED');
                case 'uploading':
                  provider.setFilter(syncStatus: 'UPLOADING');
              }
            },
            itemBuilder: (_) => [
              const PopupMenuItem(value: 'all', child: Text('All')),
              const PopupMenuItem(value: 'completed', child: Text('Completed')),
              const PopupMenuItem(value: 'uploading', child: Text('Uploading')),
              const PopupMenuItem(value: 'failed', child: Text('Failed')),
            ],
          ),
          IconButton(
            icon: Icon(Icons.logout, size: r.iconMedium),
            onPressed: () async {
              await context.read<AuthProvider>().logout();
              if (context.mounted) context.go('/login');
            },
          ),
        ],
      ),
      body: Consumer<MediaProvider>(
        builder: (context, media, _) {
          if (media.isLoading && media.assets.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (media.error != null && media.assets.isEmpty) {
            return Center(
              child: Padding(
                padding: EdgeInsets.symmetric(horizontal: r.horizontalPadding),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(media.error!, style: TextStyle(color: Colors.red, fontSize: r.bodyMedium)),
                    SizedBox(height: r.h(16)),
                    ElevatedButton(
                      onPressed: () => media.loadAssets(refresh: true),
                      child: const Text('Retry'),
                    ),
                  ],
                ),
              ),
            );
          }

          if (media.assets.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.video_library_outlined,
                      size: r.iconXLarge, color: Colors.grey[400]),
                  SizedBox(height: r.h(16)),
                  Text('No media yet',
                      style: TextStyle(fontSize: r.titleMedium, color: Colors.grey[600])),
                  SizedBox(height: r.h(8)),
                  Text('Upload your first video',
                      style: TextStyle(fontSize: r.bodyMedium)),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () => media.loadAssets(refresh: true),
            child: ListView.builder(
              controller: _scrollController,
              padding: EdgeInsets.all(r.horizontalPadding),
              itemCount: media.assets.length + (media.hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index == media.assets.length) {
                  return Center(
                    child: Padding(
                      padding: EdgeInsets.all(r.h(16)),
                      child: const CircularProgressIndicator(),
                    ),
                  );
                }
                final asset = media.assets[index];
                return GestureDetector(
                  onTap: () => context.go('/media/${asset.assetId}'),
                  child: _MediaCard(asset: asset),
                );
              },
            ),
          );
        },
      ),
      floatingActionButton: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          FloatingActionButton.small(
            heroTag: 'queue',
            onPressed: () => context.go('/queue'),
            child: Icon(Icons.queue, size: r.iconSmall),
          ),
          SizedBox(height: r.h(8)),
          FloatingActionButton.extended(
            heroTag: 'upload',
            onPressed: () => context.go('/upload'),
            icon: Icon(Icons.upload, size: r.iconMedium),
            label: Text('Upload', style: TextStyle(fontSize: r.bodyMedium)),
          ),
        ],
      ),
    );
  }
}

class _MediaCard extends StatelessWidget {
  final MediaAssetModel asset;

  const _MediaCard({required this.asset});

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Card(
      margin: EdgeInsets.only(bottom: r.h(12)),
      child: Padding(
        padding: EdgeInsets.symmetric(
          horizontal: r.w(12),
          vertical: r.h(8),
        ),
        child: Row(
          children: [
            _buildStatusIcon(r),
            SizedBox(width: r.w(12)),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    asset.originalFilename,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(fontSize: r.bodyLarge, fontWeight: FontWeight.w500),
                  ),
                  SizedBox(height: r.h(4)),
                  Text(
                    '${asset.fileSizeFormatted} - ${asset.syncStatus}',
                    style: TextStyle(
                      fontSize: r.bodySmall,
                      color: asset.isFailed ? Colors.red : Colors.grey[600],
                    ),
                  ),
                ],
              ),
            ),
            PopupMenuButton<String>(
              iconSize: r.iconMedium,
              onSelected: (value) {
                if (value == 'delete') {
                  _confirmDelete(context);
                }
              },
              itemBuilder: (_) => [
                const PopupMenuItem(
                  value: 'delete',
                  child: Text('Delete', style: TextStyle(color: Colors.red)),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusIcon(Responsive r) {
    final size = r.iconLarge;
    final iconSize = r.iconSmall;

    if (asset.isCompleted) {
      return CircleAvatar(
        radius: size / 2,
        backgroundColor: Colors.green,
        child: Icon(Icons.check, color: Colors.white, size: iconSize),
      );
    }
    if (asset.isFailed) {
      return CircleAvatar(
        radius: size / 2,
        backgroundColor: Colors.red,
        child: Icon(Icons.error, color: Colors.white, size: iconSize),
      );
    }
    if (asset.isUploading) {
      return CircleAvatar(
        radius: size / 2,
        backgroundColor: Colors.blue,
        child: SizedBox(
          width: iconSize,
          height: iconSize,
          child: const CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
        ),
      );
    }
    return CircleAvatar(
      radius: size / 2,
      backgroundColor: Colors.grey[300],
      child: Icon(Icons.video_file, color: Colors.grey, size: iconSize),
    );
  }

  void _confirmDelete(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete Media'),
        content: Text('Delete "${asset.originalFilename}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              context.read<MediaProvider>().deleteAsset(asset.assetId);
              Navigator.pop(ctx);
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}
