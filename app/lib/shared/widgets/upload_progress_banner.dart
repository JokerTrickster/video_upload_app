import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/utils/responsive.dart';
import '../../features/upload/presentation/upload_provider.dart';

class UploadProgressBanner extends StatelessWidget {
  const UploadProgressBanner({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<UploadProvider>(
      builder: (context, upload, _) {
        if (!upload.isUploading) return const SizedBox.shrink();

        final r = context.responsive;
        final currentFile = upload.files
            .where((f) => f.status == 'uploading')
            .firstOrNull;
        final filename = currentFile?.filename ?? 'Uploading...';

        return GestureDetector(
          onTap: () => context.go('/upload'),
          child: Container(
            width: double.infinity,
            color: Theme.of(context).colorScheme.primaryContainer,
            child: SafeArea(
              top: false,
              child: Padding(
                padding: EdgeInsets.symmetric(
                  horizontal: r.horizontalPadding,
                  vertical: r.h(8),
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        SizedBox(
                          width: r.iconSmall,
                          height: r.iconSmall,
                          child: const CircularProgressIndicator(strokeWidth: 2),
                        ),
                        SizedBox(width: r.w(8)),
                        Expanded(
                          child: Text(
                            filename,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(fontSize: r.bodySmall),
                          ),
                        ),
                        Text(
                          '${upload.overallProgress.toStringAsFixed(0)}%',
                          style: TextStyle(
                            fontSize: r.bodySmall,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                    SizedBox(height: r.h(4)),
                    LinearProgressIndicator(
                      value: upload.overallProgress / 100,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}
