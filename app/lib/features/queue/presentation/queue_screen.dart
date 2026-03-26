import 'dart:io';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import '../../../core/utils/responsive.dart';
import '../../../shared/models/queue_model.dart';
import '../../../shared/widgets/error_snackbar.dart';
import 'queue_provider.dart';

class QueueScreen extends StatefulWidget {
  const QueueScreen({super.key});

  @override
  State<QueueScreen> createState() => _QueueScreenState();
}

class _QueueScreenState extends State<QueueScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<QueueProvider>().loadQueue();
    });
  }

  Future<void> _addVideos() async {
    final picker = ImagePicker();
    final videos = await picker.pickMultipleMedia();

    if (videos.isNotEmpty && mounted) {
      final provider = context.read<QueueProvider>();
      int added = 0;
      int failed = 0;
      for (final video in videos) {
        final file = File(video.path);
        final stat = await file.stat();
        try {
          await provider.addToQueue(
            filePath: video.path,
            filename: video.name,
            fileSizeBytes: stat.size,
          );
          added++;
        } catch (e) {
          failed++;
          debugPrint('Failed to add ${video.name} to queue: $e');
        }
      }
      if (mounted) {
        if (failed > 0) {
          showErrorSnackBar(context, '$added added, $failed failed to queue');
        } else if (added > 0) {
          showSuccessSnackBar(context, '$added video(s) added to queue');
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Scaffold(
      appBar: AppBar(
        title: Text('Upload Queue', style: TextStyle(fontSize: r.titleMedium)),
        leading: IconButton(
          icon: Icon(Icons.arrow_back, size: r.iconMedium),
          onPressed: () => context.go('/media'),
        ),
        actions: [
          IconButton(
            icon: Icon(Icons.refresh, size: r.iconMedium),
            onPressed: () => context.read<QueueProvider>().loadQueue(),
          ),
        ],
      ),
      body: Consumer<QueueProvider>(
        builder: (context, queue, _) {
          if (queue.isLoading && queue.items.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          return RefreshIndicator(
            onRefresh: () => queue.loadQueue(),
            child: CustomScrollView(
              slivers: [
                // Quota card
                if (queue.quota != null)
                  SliverToBoxAdapter(
                    child: _QuotaCard(quota: queue.quota!),
                  ),

                // Stats row
                SliverToBoxAdapter(
                  child: Padding(
                    padding: EdgeInsets.symmetric(
                      horizontal: r.horizontalPadding,
                      vertical: r.h(8),
                    ),
                    child: Row(
                      children: [
                        _StatChip('Pending', queue.pendingCount, Colors.orange),
                        SizedBox(width: r.w(8)),
                        _StatChip('Done', queue.completedCount, Colors.green),
                        SizedBox(width: r.w(8)),
                        _StatChip('Failed', queue.failedCount, Colors.red),
                      ],
                    ),
                  ),
                ),

                // Error
                if (queue.error != null)
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: EdgeInsets.all(r.horizontalPadding),
                      child: Text(queue.error!,
                          style: TextStyle(color: Colors.red, fontSize: r.bodySmall)),
                    ),
                  ),

                // Queue items
                if (queue.items.isEmpty)
                  SliverFillRemaining(
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.queue_outlined,
                              size: r.iconXLarge, color: Colors.grey[400]),
                          SizedBox(height: r.h(16)),
                          Text('Queue is empty',
                              style: TextStyle(fontSize: r.titleMedium, color: Colors.grey[600])),
                          SizedBox(height: r.h(8)),
                          Text('Add videos to auto-upload daily',
                              style: TextStyle(fontSize: r.bodyMedium)),
                        ],
                      ),
                    ),
                  )
                else
                  SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) {
                        final item = queue.items[index];
                        return _QueueItemCard(
                          item: item,
                          onRemove: item.isPending
                              ? () => queue.removeItem(item.queueId)
                              : null,
                        );
                      },
                      childCount: queue.items.length,
                    ),
                  ),
              ],
            ),
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _addVideos,
        icon: Icon(Icons.add, size: r.iconMedium),
        label: Text('Add to Queue', style: TextStyle(fontSize: r.bodyMedium)),
      ),
    );
  }
}

class _QuotaCard extends StatelessWidget {
  final QuotaModel quota;
  const _QuotaCard({required this.quota});

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Padding(
      padding: EdgeInsets.all(r.horizontalPadding),
      child: Card(
        color: quota.canUpload ? Colors.blue[50] : Colors.red[50],
        child: Padding(
          padding: EdgeInsets.all(r.w(16)),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text("Today's Quota",
                      style: TextStyle(fontSize: r.bodyLarge, fontWeight: FontWeight.bold)),
                  Text(quota.date,
                      style: TextStyle(fontSize: r.bodySmall, color: Colors.grey[600])),
                ],
              ),
              SizedBox(height: r.h(12)),
              ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: LinearProgressIndicator(
                  value: quota.usagePercent / 100,
                  minHeight: r.h(8).clamp(6, 12),
                  backgroundColor: Colors.grey[200],
                  color: quota.canUpload ? Colors.blue : Colors.red,
                ),
              ),
              SizedBox(height: r.h(8)),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('${quota.uploadsToday} uploaded today',
                      style: TextStyle(fontSize: r.bodySmall)),
                  Text('${quota.remainingUploads} remaining',
                      style: TextStyle(
                        fontSize: r.bodyMedium,
                        fontWeight: FontWeight.bold,
                        color: quota.canUpload ? Colors.blue : Colors.red,
                      )),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatChip extends StatelessWidget {
  final String label;
  final int count;
  final Color color;
  const _StatChip(this.label, this.count, this.color);

  @override
  Widget build(BuildContext context) {
    return Chip(
      label: Text('$label: $count',
          style: const TextStyle(color: Colors.white, fontSize: 12)),
      backgroundColor: color,
      padding: EdgeInsets.zero,
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }
}

class _QueueItemCard extends StatelessWidget {
  final QueueItemModel item;
  final VoidCallback? onRemove;
  const _QueueItemCard({required this.item, this.onRemove});

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Padding(
      padding: EdgeInsets.symmetric(horizontal: r.horizontalPadding, vertical: r.h(4)),
      child: Card(
        child: Padding(
          padding: EdgeInsets.all(r.w(12)),
          child: Row(
            children: [
              _statusIcon(r),
              SizedBox(width: r.w(12)),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(item.filename,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(fontSize: r.bodyLarge, fontWeight: FontWeight.w500)),
                    SizedBox(height: r.h(2)),
                    Text('${item.fileSizeFormatted} - ${item.status}',
                        style: TextStyle(
                          fontSize: r.bodySmall,
                          color: item.isFailed ? Colors.red : Colors.grey[600],
                        )),
                    if (item.errorMessage != null)
                      Text(item.errorMessage!,
                          style: TextStyle(color: Colors.red, fontSize: r.bodySmall),
                          maxLines: 2, overflow: TextOverflow.ellipsis),
                  ],
                ),
              ),
              if (onRemove != null)
                IconButton(
                  icon: Icon(Icons.close, size: r.iconSmall),
                  onPressed: onRemove,
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _statusIcon(Responsive r) {
    final size = r.iconLarge;
    if (item.isCompleted) {
      return Icon(Icons.check_circle, color: Colors.green, size: size);
    }
    if (item.isFailed) {
      return Icon(Icons.error, color: Colors.red, size: size);
    }
    if (item.isProcessing) {
      return SizedBox(
        width: size, height: size,
        child: const CircularProgressIndicator(strokeWidth: 3),
      );
    }
    return Icon(Icons.schedule, color: Colors.orange, size: size);
  }
}
