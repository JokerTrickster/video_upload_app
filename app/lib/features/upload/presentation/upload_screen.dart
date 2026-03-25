import 'dart:io';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';
import 'upload_provider.dart';

class UploadScreen extends StatelessWidget {
  const UploadScreen({super.key});

  Future<void> _pickVideos(BuildContext context) async {
    final picker = ImagePicker();
    final videos = await picker.pickMultipleMedia();

    if (videos.isNotEmpty) {
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
    return Scaffold(
      appBar: AppBar(
        title: const Text('Upload Videos'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/media'),
        ),
      ),
      body: Consumer<UploadProvider>(
        builder: (context, upload, _) {
          return Column(
            children: [
              // File list
              Expanded(
                child: upload.files.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.video_call_outlined,
                                size: 64, color: Colors.grey[400]),
                            const SizedBox(height: 16),
                            Text('No files selected',
                                style: TextStyle(
                                    fontSize: 18, color: Colors.grey[600])),
                            const SizedBox(height: 8),
                            const Text('Tap + to select videos'),
                          ],
                        ),
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
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

              // Error message
              if (upload.error != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Text(
                    upload.error!,
                    style: const TextStyle(color: Colors.red),
                  ),
                ),

              // Progress bar (during upload)
              if (upload.isUploading)
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      LinearProgressIndicator(
                        value: upload.overallProgress / 100,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '${upload.completedCount}/${upload.files.length} completed '
                        '(${upload.overallProgress.toStringAsFixed(0)}%)',
                      ),
                    ],
                  ),
                ),

              // Action buttons
              SafeArea(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      if (!upload.isUploading)
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: () => _pickVideos(context),
                            icon: const Icon(Icons.add),
                            label: const Text('Add Files'),
                          ),
                        ),
                      if (!upload.isUploading && upload.files.isNotEmpty)
                        const SizedBox(width: 12),
                      if (upload.files.isNotEmpty)
                        Expanded(
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
                            icon: Icon(upload.isUploading
                                ? Icons.cancel
                                : Icons.cloud_upload),
                            label: Text(
                                upload.isUploading ? 'Cancel' : 'Start Upload'),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: upload.isUploading
                                  ? Colors.red
                                  : Theme.of(context).colorScheme.primary,
                              foregroundColor: Colors.white,
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
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: _buildStatusIcon(),
        title: Text(
          file.filename,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(_formatSize(file.size)),
            if (file.status == 'uploading')
              Padding(
                padding: const EdgeInsets.only(top: 4),
                child: LinearProgressIndicator(
                  value: file.progress / 100,
                ),
              ),
            if (file.error != null)
              Text(file.error!,
                  style: const TextStyle(color: Colors.red, fontSize: 12)),
          ],
        ),
        trailing: onRemove != null
            ? IconButton(
                icon: const Icon(Icons.close),
                onPressed: onRemove,
              )
            : null,
      ),
    );
  }

  Widget _buildStatusIcon() {
    switch (file.status) {
      case 'completed':
        return const Icon(Icons.check_circle, color: Colors.green, size: 32);
      case 'failed':
        return const Icon(Icons.error, color: Colors.red, size: 32);
      case 'uploading':
        return const SizedBox(
          width: 32,
          height: 32,
          child: CircularProgressIndicator(strokeWidth: 3),
        );
      default:
        return const Icon(Icons.video_file, color: Colors.grey, size: 32);
    }
  }

  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) {
      return '${(bytes / 1024).toStringAsFixed(1)} KB';
    }
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }
}
