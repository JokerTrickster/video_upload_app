import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../../shared/models/upload_session_model.dart';
import '../data/upload_repository.dart';

class SessionStatusScreen extends StatefulWidget {
  final String sessionId;

  const SessionStatusScreen({super.key, required this.sessionId});

  @override
  State<SessionStatusScreen> createState() => _SessionStatusScreenState();
}

class _SessionStatusScreenState extends State<SessionStatusScreen> {
  UploadSessionModel? _session;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadStatus();
  }

  Future<void> _loadStatus() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = context.read<UploadRepository>();
      final session = await repo.getSessionStatus(widget.sessionId);
      if (mounted) {
        setState(() {
          _session = session;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Session Status'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/media'),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadStatus,
          ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_error!, style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 16),
            ElevatedButton(onPressed: _loadStatus, child: const Text('Retry')),
          ],
        ),
      );
    }

    final session = _session!;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          // Progress circle
          Card(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  SizedBox(
                    width: 120,
                    height: 120,
                    child: Stack(
                      fit: StackFit.expand,
                      children: [
                        CircularProgressIndicator(
                          value: session.progress / 100,
                          strokeWidth: 10,
                          backgroundColor: Colors.grey[200],
                        ),
                        Center(
                          child: Text(
                            '${session.progress.toStringAsFixed(0)}%',
                            style: const TextStyle(
                              fontSize: 28,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  _StatusChip(status: session.sessionStatus),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Stats
          Row(
            children: [
              _StatCard(
                label: 'Completed',
                value: '${session.completedFiles}',
                color: Colors.green,
              ),
              const SizedBox(width: 8),
              _StatCard(
                label: 'Failed',
                value: '${session.failedFiles}',
                color: Colors.red,
              ),
              const SizedBox(width: 8),
              _StatCard(
                label: 'Pending',
                value: '${session.pendingFiles}',
                color: Colors.orange,
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Details
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Details',
                      style:
                          TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  const Divider(),
                  _DetailRow('Session ID', session.sessionId.substring(0, 8)),
                  _DetailRow('Total Files', '${session.totalFiles}'),
                  _DetailRow('Total Size', _formatBytes(session.totalBytes)),
                  _DetailRow(
                      'Uploaded', _formatBytes(session.uploadedBytes)),
                  _DetailRow('Started', _formatDate(session.startedAt)),
                  if (session.completedAt != null)
                    _DetailRow(
                        'Completed', _formatDate(session.completedAt!)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024 * 1024) {
      return '${(bytes / 1024).toStringAsFixed(1)} KB';
    }
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} '
        '${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
  }
}

class _StatusChip extends StatelessWidget {
  final String status;
  const _StatusChip({required this.status});

  @override
  Widget build(BuildContext context) {
    Color color;
    switch (status) {
      case 'ACTIVE':
        color = Colors.blue;
        break;
      case 'COMPLETED':
        color = Colors.green;
        break;
      case 'CANCELLED':
        color = Colors.red;
        break;
      default:
        color = Colors.grey;
    }
    return Chip(
      label: Text(status, style: const TextStyle(color: Colors.white)),
      backgroundColor: color,
    );
  }
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  const _StatCard(
      {required this.label, required this.value, required this.color});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Text(value,
                  style: TextStyle(
                      fontSize: 24, fontWeight: FontWeight.bold, color: color)),
              const SizedBox(height: 4),
              Text(label, style: TextStyle(color: Colors.grey[600])),
            ],
          ),
        ),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  final String label;
  final String value;
  const _DetailRow(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          SizedBox(
            width: 110,
            child: Text(label, style: TextStyle(color: Colors.grey[600])),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}
