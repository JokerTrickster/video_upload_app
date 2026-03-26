import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../shared/models/media_asset_model.dart';
import '../../../shared/widgets/error_snackbar.dart';
import 'media_provider.dart';

class MediaDetailScreen extends StatelessWidget {
  final String assetId;

  const MediaDetailScreen({super.key, required this.assetId});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Media Detail'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/media'),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete, color: Colors.red),
            onPressed: () => _deleteAsset(context),
          ),
        ],
      ),
      body: Consumer<MediaProvider>(
        builder: (context, media, _) {
          final asset = media.assets
              .where((a) => a.assetId == assetId)
              .firstOrNull;

          if (asset == null) {
            return const Center(child: Text('Asset not found'));
          }

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Thumbnail hero
                if (asset.effectiveThumbnailUrl != null) ...[
                  ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: AspectRatio(
                      aspectRatio: 16 / 9,
                      child: Image.network(
                        asset.effectiveThumbnailUrl!,
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => Container(
                          color: Colors.grey[200],
                          child: const Icon(Icons.video_file, size: 64, color: Colors.grey),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                ],

                // Status card
                _StatusCard(asset: asset),
                const SizedBox(height: 16),

                // File info
                _InfoSection(
                  title: 'File Information',
                  items: {
                    'Filename': asset.originalFilename,
                    'Size': asset.fileSizeFormatted,
                    'Type': asset.mediaType,
                    'Created': _formatDate(asset.createdAt),
                  },
                ),
                const SizedBox(height: 16),

                // Upload info
                _InfoSection(
                  title: 'Upload Details',
                  items: {
                    'Status': asset.syncStatus,
                    'Retry Count': '${asset.retryCount}',
                    if (asset.uploadStartedAt != null)
                      'Started': _formatDate(asset.uploadStartedAt!),
                    if (asset.uploadCompletedAt != null)
                      'Completed': _formatDate(asset.uploadCompletedAt!),
                    if (asset.errorMessage != null)
                      'Error': asset.errorMessage!,
                  },
                ),

                // YouTube link
                if (asset.youtubeVideoId != null) ...[
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: () => _openYouTube(asset.youtubeVideoId!),
                      icon: const Icon(Icons.play_circle_outline),
                      label: const Text('Open in YouTube'),
                    ),
                  ),
                ],
              ],
            ),
          );
        },
      ),
    );
  }

  Future<void> _openYouTube(String videoId) async {
    final uri = Uri.parse('https://www.youtube.com/watch?v=$videoId');
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    }
  }

  Future<void> _deleteAsset(BuildContext context) async {
    final confirmed = await showConfirmDialog(
      context,
      title: 'Delete Media',
      message: 'This will permanently delete this media record.',
      confirmText: 'Delete',
      isDestructive: true,
    );

    if (confirmed && context.mounted) {
      await context.read<MediaProvider>().deleteAsset(assetId);
      if (context.mounted) {
        showSuccessSnackBar(context, 'Media deleted');
        context.go('/media');
      }
    }
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} '
        '${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
  }
}

class _StatusCard extends StatelessWidget {
  final MediaAssetModel asset;

  const _StatusCard({required this.asset});

  @override
  Widget build(BuildContext context) {
    Color color;
    IconData icon;
    String label;

    if (asset.isCompleted) {
      color = Colors.green;
      icon = Icons.check_circle;
      label = 'Upload Complete';
    } else if (asset.isFailed) {
      color = Colors.red;
      icon = Icons.error;
      label = 'Upload Failed';
    } else if (asset.isUploading) {
      color = Colors.blue;
      icon = Icons.cloud_upload;
      label = 'Uploading...';
    } else {
      color = Colors.orange;
      icon = Icons.pending;
      label = 'Pending';
    }

    return Card(
      color: color.withAlpha(25),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Icon(icon, color: color, size: 40),
            const SizedBox(width: 16),
            Text(label, style: TextStyle(fontSize: 18, color: color, fontWeight: FontWeight.bold)),
          ],
        ),
      ),
    );
  }
}

class _InfoSection extends StatelessWidget {
  final String title;
  final Map<String, String> items;

  const _InfoSection({required this.title, required this.items});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const Divider(),
            ...items.entries.map((e) => Padding(
              padding: const EdgeInsets.symmetric(vertical: 4),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: 100,
                    child: Text(e.key, style: TextStyle(color: Colors.grey[600])),
                  ),
                  Expanded(child: Text(e.value)),
                ],
              ),
            )),
          ],
        ),
      ),
    );
  }
}
