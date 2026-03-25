import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../../core/storage/settings_storage.dart';
import '../../../core/utils/responsive.dart';
import '../../auth/presentation/auth_provider.dart';
import '../../queue/presentation/queue_provider.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  bool _autoUpload = false;

  @override
  void initState() {
    super.initState();
    _autoUpload = SettingsStorage.instance.isAutoUploadEnabled;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<QueueProvider>().refreshQuota();
    });
  }

  @override
  Widget build(BuildContext context) {
    final r = context.responsive;

    return Scaffold(
      appBar: AppBar(
        title: Text('Settings', style: TextStyle(fontSize: r.titleMedium)),
        leading: IconButton(
          icon: Icon(Icons.arrow_back, size: r.iconMedium),
          onPressed: () => context.go('/media'),
        ),
      ),
      body: ListView(
        padding: EdgeInsets.all(r.horizontalPadding),
        children: [
          Card(
            child: SwitchListTile(
              title: Text('Auto Upload',
                  style: TextStyle(fontSize: r.bodyLarge)),
              subtitle: Text(
                'Automatically add selected videos to the upload queue',
                style: TextStyle(fontSize: r.bodySmall),
              ),
              secondary: Icon(Icons.cloud_upload_outlined, size: r.iconLarge),
              value: _autoUpload,
              onChanged: (value) async {
                await SettingsStorage.instance.setAutoUploadEnabled(value);
                setState(() => _autoUpload = value);
              },
            ),
          ),
          SizedBox(height: r.h(16)),
          Consumer<QueueProvider>(
            builder: (context, queue, _) {
              final quota = queue.quota;
              if (quota == null) return const SizedBox.shrink();

              final canUpload = quota.canUpload;

              return Card(
                child: Padding(
                  padding: EdgeInsets.all(r.horizontalPadding),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(Icons.today, size: r.iconMedium),
                          SizedBox(width: r.w(8)),
                          Text("Today's Upload Quota",
                              style: TextStyle(
                                  fontSize: r.bodyLarge,
                                  fontWeight: FontWeight.w600)),
                        ],
                      ),
                      SizedBox(height: r.h(12)),
                      Center(
                        child: Text(
                          '${quota.remainingUploads}',
                          style: TextStyle(
                            fontSize: r.titleMedium * 2,
                            fontWeight: FontWeight.bold,
                            color: canUpload ? Colors.green : Colors.red,
                          ),
                        ),
                      ),
                      Center(
                        child: Text(
                          canUpload
                              ? 'uploads available today'
                              : 'no uploads remaining today',
                          style: TextStyle(
                              fontSize: r.bodyMedium,
                              color: Colors.grey[600]),
                        ),
                      ),
                      SizedBox(height: r.h(12)),
                      LinearProgressIndicator(
                        value: quota.usagePercent / 100,
                        backgroundColor: Colors.grey[200],
                        color: canUpload ? null : Colors.red,
                      ),
                      SizedBox(height: r.h(8)),
                      Text(
                        '${quota.unitsUsed} / ${quota.unitsMax} units used  ·  ${quota.uploadsToday} uploaded today',
                        style: TextStyle(fontSize: r.bodySmall, color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
          SizedBox(height: r.h(24)),
          SizedBox(
            width: double.infinity,
            height: r.h(48).clamp(40.0, 56.0),
            child: OutlinedButton.icon(
              onPressed: () async {
                await context.read<AuthProvider>().logout();
                if (context.mounted) context.go('/login');
              },
              icon: Icon(Icons.logout, size: r.iconSmall, color: Colors.red),
              label: Text('Logout',
                  style: TextStyle(fontSize: r.bodyMedium, color: Colors.red)),
              style: OutlinedButton.styleFrom(
                side: const BorderSide(color: Colors.red),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
