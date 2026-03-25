import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
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
    return Scaffold(
      appBar: AppBar(
        title: const Text('My Media'),
        actions: [
          PopupMenuButton<String>(
            icon: const Icon(Icons.filter_list),
            onSelected: (value) {
              final provider = context.read<MediaProvider>();
              switch (value) {
                case 'all':
                  provider.setFilter();
                  break;
                case 'completed':
                  provider.setFilter(syncStatus: 'COMPLETED');
                  break;
                case 'failed':
                  provider.setFilter(syncStatus: 'FAILED');
                  break;
                case 'uploading':
                  provider.setFilter(syncStatus: 'UPLOADING');
                  break;
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
            icon: const Icon(Icons.logout),
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
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(media.error!, style: const TextStyle(color: Colors.red)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => media.loadAssets(refresh: true),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (media.assets.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.video_library_outlined,
                      size: 64, color: Colors.grey[400]),
                  const SizedBox(height: 16),
                  Text('No media yet',
                      style: TextStyle(
                          fontSize: 18, color: Colors.grey[600])),
                  const SizedBox(height: 8),
                  const Text('Upload your first video'),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () => media.loadAssets(refresh: true),
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: media.assets.length + (media.hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index == media.assets.length) {
                  return const Center(
                    child: Padding(
                      padding: EdgeInsets.all(16),
                      child: CircularProgressIndicator(),
                    ),
                  );
                }
                return _MediaCard(asset: media.assets[index]);
              },
            ),
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.go('/upload'),
        icon: const Icon(Icons.upload),
        label: const Text('Upload'),
      ),
    );
  }
}

class _MediaCard extends StatelessWidget {
  final MediaAssetModel asset;

  const _MediaCard({required this.asset});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: _buildStatusIcon(),
        title: Text(
          asset.originalFilename,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          '${asset.fileSizeFormatted} - ${asset.syncStatus}',
          style: TextStyle(
            color: asset.isFailed ? Colors.red : Colors.grey[600],
          ),
        ),
        trailing: PopupMenuButton<String>(
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
      ),
    );
  }

  Widget _buildStatusIcon() {
    if (asset.isCompleted) {
      return const CircleAvatar(
        backgroundColor: Colors.green,
        child: Icon(Icons.check, color: Colors.white, size: 20),
      );
    }
    if (asset.isFailed) {
      return const CircleAvatar(
        backgroundColor: Colors.red,
        child: Icon(Icons.error, color: Colors.white, size: 20),
      );
    }
    if (asset.isUploading) {
      return const CircleAvatar(
        backgroundColor: Colors.blue,
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(
            strokeWidth: 2,
            color: Colors.white,
          ),
        ),
      );
    }
    return CircleAvatar(
      backgroundColor: Colors.grey[300],
      child: const Icon(Icons.video_file, color: Colors.grey, size: 20),
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
