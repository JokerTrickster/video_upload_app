import 'dart:io';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import '../../../core/storage/settings_storage.dart';
import '../../../core/utils/responsive.dart';
import '../../queue/presentation/queue_provider.dart';
import 'upload_provider.dart';

class UploadScreen extends StatelessWidget {
  const UploadScreen({super.key});

  Future<void> _pickVideos(BuildContext context) async {
    final picker = ImagePicker();
    final videos = await picker.pickMultipleMedia();

    if (videos.isEmpty || !context.mounted) return;

    final isAutoUpload = SettingsStorage.instance.isAutoUploadEnabled;

    if (isAutoUpload) {
      final queueProvider = context.read<QueueProvider>();
      var addedCount = 0;
      var failedCount = 0;
      for (final video in videos) {
        final file = File(video.path);
        final stat = await file.stat();
        try {
          await queueProvider.addToQueue(
            filePath: video.path,
            filename: video.name,
            fileSizeBytes: stat.size,
          );
          addedCount++;
        } catch (e) {
          failedCount++;
          debugPrint('Failed to add ${video.name} to queue: $e');
        }
      }
      if (context.mounted) {
        final message = failedCount > 0
            ? '$addedCount added, $failedCount failed to queue'
            : '$addedCount file(s) added to queue';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(message),
            backgroundColor: failedCount > 0 ? Colors.orange : null,
          ),
        );
      }
    } else {
      final uploadFiles = <UploadFile>[];
      for (final video in videos) {
        final file = File(video.path);
        final stat = await file.stat();
        uploadFiles.add(UploadFile(
          path: video.path,
          filename: video.name,
          size: stat.size,
        ));
      }
      if (context.mounted) {
        context.read<UploadProvider>().addFiles(uploadFiles);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Scaffold(
      appBar: AppBar(
        title: Text('Upload Videos', style: TextStyle(fontSize: r.titleMedium)),
        leading: IconButton(
          icon: Icon(Icons.arrow_back, size: r.iconMedium),
          onPressed: () => context.go('/media'),
        ),
      ),
      body: Consumer<UploadProvider>(
        builder: (context, upload, _) {
          return Column(
            children: [
              Expanded(
                child: upload.files.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.video_call_outlined,
                                size: r.iconXLarge, color: Colors.grey[400]),
                            SizedBox(height: r.h(16)),
                            Text('No files selected',
                                style: TextStyle(fontSize: r.titleMedium, color: Colors.grey[600])),
                            SizedBox(height: r.h(8)),
                            Text('Tap + to select videos',
                                style: TextStyle(fontSize: r.bodyMedium)),
                          ],
                        ),
                      )
                    : ListView.builder(
                        padding: EdgeInsets.all(r.horizontalPadding),
                        itemCount: upload.files.length,
                        itemBuilder: (context, index) {
                          final file = upload.files[index];
                          return _UploadFileCard(
                            file: file,
                            onRemove: upload.isUploading
                                ? null
                                : () => upload.removeFile(index),
                          );
                        },
                      ),
              ),

              if (upload.error != null)
                Padding(
                  padding: EdgeInsets.symmetric(horizontal: r.horizontalPadding),
                  child: Text(upload.error!,
                      style: TextStyle(color: Colors.red, fontSize: r.bodySmall)),
                ),

              if (upload.isUploading)
                Padding(
                  padding: EdgeInsets.all(r.horizontalPadding),
                  child: Column(
                    children: [
                      LinearProgressIndicator(value: upload.overallProgress / 100),
                      SizedBox(height: r.h(8)),
                      Text(
                        '${upload.completedCount}/${upload.files.length} completed '
                        '(${upload.overallProgress.toStringAsFixed(0)}%)',
                        style: TextStyle(fontSize: r.bodyMedium),
                      ),
                    ],
                  ),
                ),

              SafeArea(
                child: Padding(
                  padding: EdgeInsets.all(r.horizontalPadding),
                  child: Row(
                    children: [
                      if (!upload.isUploading)
                        Expanded(
                          child: SizedBox(
                            height: r.h(48).clamp(40.0, 56.0),
                            child: OutlinedButton.icon(
                              onPressed: () => _pickVideos(context),
                              icon: Icon(Icons.add, size: r.iconSmall),
                              label: Text('Add Files', style: TextStyle(fontSize: r.bodyMedium)),
                            ),
                          ),
                        ),
                      if (!upload.isUploading && upload.files.isNotEmpty)
                        SizedBox(width: r.w(12)),
                      if (upload.files.isNotEmpty)
                        Expanded(
                          child: SizedBox(
                            height: r.h(48).clamp(40.0, 56.0),
                            child: ElevatedButton.icon(
                              onPressed: upload.isUploading
                                  ? () => upload.cancelUpload()
                                  : () async {
                                      await upload.startUpload();
                                      if (context.mounted &&
                                          !upload.isUploading &&
                                          upload.failedCount == 0) {
                                        upload.clearFiles();
                                        context.go('/media');
                                      }
                                    },
                              icon: Icon(
                                upload.isUploading ? Icons.cancel : Icons.cloud_upload,
                                size: r.iconSmall,
                              ),
                              label: Text(
                                upload.isUploading ? 'Cancel' : 'Start Upload',
                                style: TextStyle(fontSize: r.bodyMedium),
                              ),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: upload.isUploading
                                    ? Colors.red
                                    : Theme.of(context).colorScheme.primary,
                                foregroundColor: Colors.white,
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _UploadFileCard extends StatelessWidget {
  final UploadFile file;
  final VoidCallback? onRemove;

  const _UploadFileCard({required this.file, this.onRemove});

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Card(
      margin: EdgeInsets.only(bottom: r.h(8)),
      child: Padding(
        padding: EdgeInsets.symmetric(horizontal: r.w(12), vertical: r.h(8)),
        child: Row(
          children: [
            _buildStatusIcon(r),
            SizedBox(width: r.w(12)),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(file.filename,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(fontSize: r.bodyLarge)),
                  SizedBox(height: r.h(2)),
                  Text(_formatSize(file.size),
                      style: TextStyle(fontSize: r.bodySmall, color: Colors.grey[600])),
                  if (file.status == 'uploading')
                    Padding(
                      padding: EdgeInsets.only(top: r.h(4)),
                      child: LinearProgressIndicator(value: file.progress / 100),
                    ),
                  if (file.error != null)
                    Text(file.error!,
                        style: TextStyle(color: Colors.red, fontSize: r.bodySmall)),
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
    );
  }

  Widget _buildStatusIcon(Responsive r) {
    final size = r.iconLarge;
    switch (file.status) {
      case 'completed':
        return Icon(Icons.check_circle, color: Colors.green, size: size);
      case 'failed':
        return Icon(Icons.error, color: Colors.red, size: size);
      case 'uploading':
        return SizedBox(
          width: size, height: size,
          child: const CircularProgressIndicator(strokeWidth: 3),
        );
      default:
        return Icon(Icons.video_file, color: Colors.grey, size: size);
    }
  }

  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }
}
